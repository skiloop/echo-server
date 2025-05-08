package ja3

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net"
	"net/http"
	"os"
)

var (
	certOrKeyError = errors.New("invalid cert or key error")
	store          echo.Map
)

func init() {
	store = make(echo.Map)
}

func filepathOrContent(fileOrContent interface{}) (content []byte, err error) {
	switch v := fileOrContent.(type) {
	case string:
		return os.ReadFile(v)
	case []byte:
		return v, nil
	default:
		return nil, certOrKeyError
	}
}

func NewHttpServer(addr, cert, key string) (*http.Server, error) {
	crt, err := filepathOrContent(cert)
	if err != nil {
		return nil, err
	}
	keyData, err := filepathOrContent(key)
	if err != nil {
		return nil, err
	}
	srv := new(http.Server)
	srv.Addr = addr
	srv.TLSConfig = new(tls.Config)
	connState := srv.ConnState
	if connState == nil {
		srv.ConnState = func(c net.Conn, cs http.ConnState) {
			if cs == http.StateClosed {
				k := c.RemoteAddr().String()
				delete(store, k)
			}
		}
	} else {
		srv.ConnState = func(c net.Conn, cs http.ConnState) {
			if cs == http.StateClosed {
				k := c.RemoteAddr().String()
				fmt.Printf("remove ja3 for %s\n", k)
				delete(store, k)
			}
			connState(c, cs)
		}
	}
	srv.TLSConfig.Certificates = make([]tls.Certificate, 1)
	if srv.TLSConfig.Certificates[0], err = tls.X509KeyPair(crt, keyData); err != nil {
		return nil, err
	}
	srv.TLSConfig.NextProtos = append(srv.TLSConfig.NextProtos, "h2")

	return srv, nil
}

func getConfigForClient(info *tls.ClientHelloInfo) (*tls.Config, error) {
	ja3_, err := GenJA3Raw(*info)
	if err == nil {
		k := info.Conn.RemoteAddr().String()
		fmt.Printf("%s ja3 %s\n", k, ja3_.Md5Hash())
		store[k] = ja3_
	} else {
		fmt.Printf("get ja3 error %s\n", err)
	}
	return nil, nil
}

func NewTLSConfig(c *tls.Config) *tls.Config {
	cfg := new(tls.Config)
	var cc interface{}
	cc = c.Certificates
	cert := (cc).([]tls.Certificate)
	cfg.Certificates = cert
	cfg.NextProtos = c.NextProtos

	cfg.GetConfigForClient = getConfigForClient
	return cfg
}
