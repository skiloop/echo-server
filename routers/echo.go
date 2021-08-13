package routers

import (
	"github.com/labstack/echo/v4"
	"github.com/skiloop/echo-server/utils"
	"net/http"
	"strings"
)

func Echo(c echo.Context) error {
	buf := utils.BufferPool.Get()
	bytesBuf := make([]byte, 20)
	for {
		n, err := c.Request().Body.Read(bytesBuf)
		if err != nil {
			break
		}
		_, _ = buf.Write(bytesBuf[1:n])
	}
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
	return c.JSON(http.StatusOK, resp)
}
