module github.com/skiloop/echo-server

go 1.21

toolchain go1.22.5

require (
	github.com/gorilla/websocket v1.5.0
	github.com/labstack/echo/v4 v4.13.3
	github.com/labstack/gommon v0.4.2
	github.com/maxmind/mmdbinspect v0.1.1
	go.uber.org/zap v1.19.0
)

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/oschwald/maxminddb-golang v1.6.1-0.20200115172950-664f3fe38c62 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.8.0 // indirect
)

replace github.com/skiloop/echo-server/utils => ../utils

replace github.com/skiloop/echo-server/routers => ../routers
