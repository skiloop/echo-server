package ctls

import (
	"crypto/elliptic"
	"errors"
	"golang.org/x/crypto/curve25519"
	"io"
	"math/big"
)

// ecdheParameters implements Diffie-Hellman with either NIST curves or X25519,
// according to RFC 8446, Section 4.2.8.2.
type ecdheParameters interface {
	CurveID() CurveID
	PublicKey() []byte
	SharedKey(peerPublicKey []byte) []byte
}

func curveForCurveID(id CurveID) (elliptic.Curve, bool) {
	switch id {
	case CurveP256:
		return elliptic.P256(), true
	case CurveP384:
		return elliptic.P384(), true
	case CurveP521:
		return elliptic.P521(), true
	default:
		return nil, false
	}
}

type x25519Parameters struct {
	privateKey []byte
	publicKey  []byte
}

func (p *x25519Parameters) CurveID() CurveID {
	return X25519
}

func (p *x25519Parameters) PublicKey() []byte {
	return p.publicKey[:]
}

func (p *x25519Parameters) SharedKey(peerPublicKey []byte) []byte {
	sharedKey, err := curve25519.X25519(p.privateKey, peerPublicKey)
	if err != nil {
		return nil
	}
	return sharedKey
}

func generateECDHEParameters(rand io.Reader, curveID CurveID) (ecdheParameters, error) {
	if curveID == X25519 {
		privateKey := make([]byte, curve25519.ScalarSize)
		if _, err := io.ReadFull(rand, privateKey); err != nil {
			return nil, err
		}
		publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
		if err != nil {
			return nil, err
		}
		return &x25519Parameters{privateKey: privateKey, publicKey: publicKey}, nil
	}

	curve, ok := curveForCurveID(curveID)
	if !ok {
		return nil, errors.New("tls: internal error: unsupported curve")
	}

	p := &nistParameters{curveID: curveID}
	var err error
	p.privateKey, p.x, p.y, err = elliptic.GenerateKey(curve, rand)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type nistParameters struct {
	privateKey []byte
	x, y       *big.Int // public key
	curveID    CurveID
}

func (p *nistParameters) CurveID() CurveID {
	return p.curveID
}

func (p *nistParameters) PublicKey() []byte {
	curve, _ := curveForCurveID(p.curveID)
	return elliptic.Marshal(curve, p.x, p.y)
}

func (p *nistParameters) SharedKey(peerPublicKey []byte) []byte {
	curve, _ := curveForCurveID(p.curveID)
	// Unmarshal also checks whether the given point is on the curve.
	x, y := elliptic.Unmarshal(curve, peerPublicKey)
	if x == nil {
		return nil
	}

	xShared, _ := curve.ScalarMult(x, y, p.privateKey)
	sharedKey := make([]byte, (curve.Params().BitSize+7)/8)
	return xShared.FillBytes(sharedKey)
}
