package routers

import (
	"github.com/labstack/echo/v4"
	"github.com/skiloop/echo-server/utils"
	"go.uber.org/zap/buffer"
	"net/http"
	"strings"
)

func readBody(c echo.Context) *buffer.Buffer {
	buf := utils.BufferPool.Get()
	bytesBuf := make([]byte, 20)
	for {
		n, err := c.Request().Body.Read(bytesBuf)
		if n > 0 {
			_, _ = buf.Write(bytesBuf[0:n])
		}
		if err != nil {
			c.Logger().Debug(err)
			break
		}
	}
	return buf
}

func Echo(c echo.Context) error {
	buf := readBody(c)
	c.Logger().Debugf("body size: %d", len(buf.String()))
	return c.Blob(http.StatusOK, c.Request().Header.Get("Content-Type"), buf.Bytes())
}

func EchoAsJson(c echo.Context) error {
	resp := make(map[string]interface{})
	headers := make(map[string]string)
	for key, values := range c.Request().Header {
		if _, has := headers[key]; has {
			headers[key] = headers[key] + strings.Join(values, ";")
		} else {
			headers[key] = strings.Join(values, ";")
		}
	}
	resp["headers"] = headers
	// set parameters
	resp["query"] = c.QueryParams()
	// set url
	resp["url"] = c.Request().URL.RequestURI()
	resp["body"] = readBody(c).String()
	return c.JSON(http.StatusOK, resp)
}

func SetEchoRouters(e *echo.Echo, path string) {
	// echo routes
	e.GET("/echo/*", Echo)
	e.POST("/echo/*", Echo)
	e.PATCH("/echo/*", Echo)
	e.PUT("/echo/*", Echo)
	e.GET("/json/*", EchoAsJson)
	e.POST("/json/*", EchoAsJson)
	e.PATCH("/json/*", EchoAsJson)
	e.PUT("/json/*", EchoAsJson)
	e.GET("/echo", Echo)
	e.POST("/echo", Echo)
	e.PATCH("/echo", Echo)
	e.PUT("/echo", Echo)
	e.GET("/json", EchoAsJson)
	e.POST("/json", EchoAsJson)
	e.PATCH("/json", EchoAsJson)
	e.PUT("/json", EchoAsJson)
}
