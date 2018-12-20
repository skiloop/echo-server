package main

import (
	"github.com/labstack/echo/middleware"
	"net/http"
	"github.com/labstack/echo"
	"io/ioutil"
	"strings"
	"github.com/iancoleman/orderedmap"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", echo_handler)
	e.PUT("/", echo_handler)
	e.POST("/", echo_handler)
	e.PATCH("/", echo_handler)
	e.DELETE("/", echo_handler)

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}

// Handler
func echo_handler(c echo.Context) error {
	data := orderedmap.New()
	data.Set("method", c.Request().Method)
	headers := orderedmap.New()
	for key, value := range c.Request().Header {
		headers.Set(key, strings.Join(value, ","))
	}
	data.Set("headers", headers)
	b, err := ioutil.ReadAll(c.Request().Body)
	defer c.Request().Body.Close()
	if err != nil {
		return err
	}

	data.Set("body", string(b))
	return c.JSON(http.StatusOK, data)
}
