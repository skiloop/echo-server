package server

import (
	"github.com/labstack/echo/v4"
	"net"
	"time"
)

type tlsListener struct {
	*net.TCPListener
}

func (ln tlsListener) Accept() (c net.Conn, err error) {
	if c, err = ln.AcceptTCP(); err != nil {
		return
	} else if err = c.(*net.TCPConn).SetKeepAlive(true); err != nil {
		return
	}
	// Ignore error from setting the KeepAlivePeriod as some systems, such as
	// OpenBSD, do not support setting TCP_USER_TIMEOUT on IPPROTO_TCP
	_ = c.(*net.TCPConn).SetKeepAlivePeriod(3 * time.Minute)
	return
}

func newListener(address, network string) (*tlsListener, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, echo.ErrInvalidListenerNetwork
	}
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return &tlsListener{l.(*net.TCPListener)}, nil
}
