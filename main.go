package main

import (
	"flag"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/skiloop/echo-server/ja3"
	"github.com/skiloop/echo-server/routers"
	"github.com/skiloop/echo-server/server"
	"os"
)

var httpsAddrDefault = "0.0.0.0:9013"
var keyFile = flag.String("key", "", "tls key file path, empty then env TLS_KEY_FILE will apply")
var certFile = flag.String("cert", "", "tls cert file path, empty then TLS_CERT_FILE will apply")
var httpAddr = flag.String("http", "0.0.0.0:9012", "http bind addr")
var httpsAddr = flag.String("https", "0.0.0.0:9013", fmt.Sprintf("https bind addr, if not set but cert and key are set, then use %s", httpsAddrDefault))

func main() {
	flag.Parse()

	addr := *httpAddr
	if "" == *keyFile {
		*keyFile = os.Getenv("TLS_KEY_FILE")
	}
	if "" == *certFile {
		*certFile = os.Getenv("TLS_CERT_FILE")
	}
	e := server.NewEchoServer(*httpsAddr, *certFile, *keyFile)
	// middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(ja3.HashMiddleware())
	e.Logger.SetLevel(log.DEBUG)

	// routers
	setUpRouters(e)

	// start
	serve(e, addr, *httpsAddr, *certFile, *keyFile)
}

func OK(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store, max-age=0")
	return c.String(200, "ok")
}

func serve(e *echo.Echo, addr, httpsAddr, cert, key string) {
	e.Logger.Debugf("cert file: %s, key file: %s", cert, key)
	tlsAddr := ""
	if "" != cert && "" != key {
		tlsAddr = httpsAddr
		if tlsAddr == "" {
			tlsAddr = os.Getenv("BIND_ADDR_TLS")
		}
		if tlsAddr == "" {
			tlsAddr = "127.0.0.1:9013"
		}
	}
	if tlsAddr == "" && addr == "" {
		e.Logger.Fatal("no server addr set")
	} else if tlsAddr != "" && addr != "" {
		go func() {
			e.Logger.Fatal(e.StartServer(e.TLSServer))
		}()
		e.Logger.Fatal(e.Start(addr))
	} else if addr != "" {
		e.Logger.Fatal(e.Start(addr))
	} else {
		e.Logger.Fatal(e.StartServer(e.TLSServer))
	}
}

func setUpRouters(e *echo.Echo) {
	routers.SetEchoRouters(e, "/")
	ja3.SetJA3Routers(e, "/")
	// ip location routes
	routers.SetLocationRouters(e, "/")
	e.GET("/mongo/:id", routers.MongoParseID)

	// short but useful
	e.GET("/ok", OK)
	e.PUT("/ok", OK)
	e.PATCH("/ok", OK)
	e.POST("/ok", OK)
	e.GET("/health", OK)
	e.GET("/_health", OK)

	// websocket
	e.GET("/ws/echo", routers.WsEcho)
}
