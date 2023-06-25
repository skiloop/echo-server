package server

import (
	"crypto/tls"
	"fmt"
	"github.com/skiloop/echo-server/server/ctls"
	"github.com/skiloop/echo-server/utils"
	"strings"
)

type Ja3 struct {
	Hash       string   `json:"ja3"`
	MaxVersion uint16   `json:"max_version"`
	Ciphers    []uint16 `json:"ciphers"`
	Extensions []uint16 `json:"extensions"`
	Curves     []uint16 `json:"curves"`
	Points     []uint16 `json:"points"`
}

func (j *Ja3) Md5Hash() string {
	if j.Hash == "" {
		var ja3 []string
		ja3 = append(ja3, fmt.Sprintf("%d", j.MaxVersion))
		ja3 = append(ja3, getList(j.Ciphers))
		ja3 = append(ja3, getList(j.Extensions))
		ja3 = append(ja3, getList(j.Curves))
		ja3 = append(ja3, getList(j.Points))
		s := strings.Join(ja3, ",")
		j.Hash = utils.GetMD5Hash(s)
		fmt.Printf("ja3: %s, raw ja3: %s\n", j.Hash, s)
	}
	return j.Hash
}

func GenJA3(info tls.ClientHelloInfo) (string, error) {
	ja3, err := GenJA3Raw(info)
	if err != nil {
		return "", err
	}
	return ja3.Md5Hash(), nil
}

func GenJA3Raw(info tls.ClientHelloInfo) (Ja3, error) {
	var ja3 Ja3
	ja3.MaxVersion = getMaxVersion(info)
	ja3.Ciphers = info.CipherSuites
	ja3.Extensions = getExtensions(info)
	ja3.setCurves(info)
	ja3.setPoints(info)
	return ja3, nil
}

func (j *Ja3) setCurves(info tls.ClientHelloInfo) {
	for _, curve := range info.SupportedCurves {
		j.Curves = append(j.Curves, uint16(curve))
	}
}

func (j *Ja3) setPoints(info tls.ClientHelloInfo) {
	for _, curve := range info.SupportedPoints {
		j.Points = append(j.Points, uint16(curve))
	}
}

func getMaxVersion(info tls.ClientHelloInfo) uint16 {
	var maxVer uint16
	maxVer = 0
	for _, ver := range info.SupportedVersions {
		if ver > maxVer {
			maxVer = ver
		}
	}
	return maxVer
}

func getExtensions(info tls.ClientHelloInfo) []uint16 {
	var data []uint16
	if info.ServerName != "" {
		data = append(data, 0)
	}
	if len(info.SupportedPoints) > 0 {
		data = append(data, 11)
	}
	// TODO: supported groups
	if len(info.SupportedCurves) > 0 {
		data = append(data, 10)
	}
	return data
}
func getList(x interface{}) string {
	var data []string
	switch x.(type) {
	case []uint8:
		for _, item := range x.([]uint8) {
			data = append(data, fmt.Sprintf("%d", item))
		}
	case []uint16:
		for _, item := range x.([]uint16) {
			data = append(data, fmt.Sprintf("%d", item))
		}
	case []tls.CurveID:
		for _, item := range x.([]tls.CurveID) {
			data = append(data, fmt.Sprintf("%d", item))
		}
	case []tls.SignatureScheme:
		for _, item := range x.([]tls.SignatureScheme) {
			data = append(data, fmt.Sprintf("%d", item))
		}
	default:
		data = append(data, fmt.Sprintf("%d", x))
	}

	return strings.Join(data, "-")
}
