package server

import (
	"crypto/tls"
	"github.com/labstack/echo/v4"
	"github.com/skiloop/echo-server/ja3"
)

// NewEchoServer New creates an instance of Echo.
func NewEchoServer(addr, cert, key string) (e *echo.Echo) {
	e = echo.New()
	s, err := ja3.NewHttpServer(addr, cert, key)
	if err != nil {
		e.Logger.Warn(err)
	} else {
		e.TLSServer = s
		s.Handler = e
		l, err := newListener(addr, e.ListenerNetwork)
		if err != nil {
			e.Logger.Error(err)
			return nil
		}
		e.TLSListener = tls.NewListener(l, ja3.NewTLSConfig(s.TLSConfig))
	}

	return e
}

func newConfig(c *tls.Config) *tls.Config {
	cfg := new(tls.Config)
	var cc interface{}
	cc = c.Certificates
	cert := (cc).([]tls.Certificate)
	cfg.Certificates = cert
	cfg.NextProtos = c.NextProtos
	return cfg
}
