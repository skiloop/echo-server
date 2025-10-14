package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/skiloop/echo-server/ja3"
	esmw "github.com/skiloop/echo-server/middleware"
	"github.com/skiloop/echo-server/routers"
	"github.com/skiloop/echo-server/server"
)

// CLI 命令行参数定义
type CLI struct {
	HTTP               string   `short:"b" help:"HTTP bind address" default:"0.0.0.0:9012" env:"HTTP_ADDR"`
	HTTPS              string   `short:"s" help:"HTTPS bind address" default:"0.0.0.0:9013" env:"HTTPS_ADDR"`
	Cert               string   `short:"c" help:"TLS certificate file path" env:"TLS_CERT_FILE"`
	Key                string   `short:"k" help:"TLS key file path" env:"TLS_KEY_FILE"`
	Debug              bool     `short:"v" help:"Enable debug logging" default:"false"`
	AuthApiKey         string   `short:"a" help:"API key for HMAC authentication" default:"your-secret-api-key-here" env:"AUTH_API_KEY"`
	AuthTimestampValid int64    `short:"t" help:"Timestamp valid period for HMAC authentication" default:"300" env:"AUTH_TIMESTAMP_VALID"`
	AuthPaths          []string `short:"p" help:"Paths requiring HMAC authentication (supports wildcards)" default:"/upload,/upload/*" env:"AUTH_PATHS" sep:","`
}

func main() {
	// 解析命令行参数
	cli := CLI{}
	_ = kong.Parse(&cli,
		kong.Name("echo-server"),
		kong.Description("A versatile echo server with JA3 fingerprinting and file upload support"),
		kong.UsageOnError(),
	)

	// 配置日志级别
	logLevel := log.INFO
	if cli.Debug {
		logLevel = log.DEBUG
	}

	// 创建服务器
	e := server.NewEchoServer(cli.HTTPS, cli.Cert, cli.Key)
	e.Logger.SetLevel(logLevel)

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// ja3 middleware
	e.Use(ja3.Middleware())

	// use WafMiddleware
	wafCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.Use(esmw.WafMiddleware(1800, nil, wafCtx))

	// HMAC认证中间件 - 仅对指定路径生效
	if len(cli.AuthPaths) > 0 {
		e.Logger.Infof("Enabling HMAC auth for paths: %v", cli.AuthPaths)
		e.Logger.Debugf("HMAC auth key: %s", cli.AuthApiKey)
		e.Logger.Debugf("HMAC auth timestamp valid: %d", cli.AuthTimestampValid)
		e.Logger.Debugf("HMAC auth paths: %v", cli.AuthPaths)
		e.Use(esmw.HMACAuthWithConfig(cli.AuthApiKey, cli.AuthTimestampValid, cli.AuthPaths))
	}

	// routers
	setUpRouters(e)

	// start
	serve(e, cli.HTTP, cli.HTTPS, cli.Cert, cli.Key)
}

func OK(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store, max-age=0")
	return c.String(200, "ok")
}

func serve(e *echo.Echo, addr, httpsAddr, cert, key string) {
	e.Logger.Debugf("cert file: %s, key file: %s", cert, key)
	tlsAddr := ""
	if cert != "" && key != "" {
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

	// file upload routes (需要HMAC认证)
	routers.SetUploadRouters(e)

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
