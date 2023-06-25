package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
)

const Ja3HeaderName = "X-JA3"

var (
	certOrKeyError = errors.New("invalid cert or key error")
	store          echo.Map
)

func init() {
	store = make(echo.Map)
}

// New creates an instance of Echo.
func NewEchoServer(addr, cert, key string) (e *echo.Echo) {
	e = echo.New()
	s, err := newJA3HttpServer(addr, cert, key)
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
		e.TLSListener = tls.NewListener(l, newConfig(s.TLSConfig))
		e.Use(setJa3())
	}

	return e
}

func setJa3() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			k := context.Request().RemoteAddr
			if val, ok := store[k]; ok {
				delete(store, k)
				context.Set("ja3", val)
				val := val.(Ja3)
				context.Echo().Logger.Debugf("get %s ja3 for %s", k, val.Md5Hash())
			} else {
				context.Echo().Logger.Warnf("no ja3 for %s", k)
			}
			return next(context)
		}
	}
}

func newJA3HttpServer(addr, cert, key string) (*http.Server, error) {
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
	srv.TLSConfig.Certificates = make([]tls.Certificate, 1)
	if srv.TLSConfig.Certificates[0], err = tls.X509KeyPair(crt, keyData); err != nil {
		return nil, err
	}
	srv.TLSConfig.NextProtos = append(srv.TLSConfig.NextProtos, "h2")

	return srv, nil
}

func newConfig(c *tls.Config) *tls.Config {
	cfg := new(tls.Config)
	var cc interface{}
	cc = c.Certificates
	cert := (cc).([]tls.Certificate)
	cfg.Certificates = cert
	cfg.NextProtos = c.NextProtos
	cfg.GetConfigForClient = func(info *tls.ClientHelloInfo) (*tls.Config, error) {
		ja3, err := GenJA3Raw(*info)
		if err == nil {
			k := info.Conn.RemoteAddr().String()
			fmt.Printf("%s ja3 %s\n", k, ja3.Md5Hash())
			store[k] = ja3
		}
		return nil, nil
	}
	return cfg
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

func JA3() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			ja3 := context.Get("ja3")
			if ja3 == nil {
				context.Echo().Logger.Warnf("no ja3 for request %s", context.Request().RequestURI)
			} else {
				u := ja3.(Ja3)
				context.Echo().Logger.Debugf("add ja3 header %s for %s", u.Md5Hash(), context.Request().RequestURI)
				context.Response().Header().Add(Ja3HeaderName, u.Md5Hash())
			}
			return next(context)
		}
	}
}
