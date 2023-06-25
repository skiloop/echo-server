package raw

import (
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"go.uber.org/zap/buffer"
	"io"
	"net"
	"time"
)

type RawHTTP2Echo struct {
	Addr       string
	CertFile   string
	KeyFile    string
	listener   net.Listener
	running    bool
	bufferPool buffer.Pool
}

func (r RawHTTP2Echo) ServeConn(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		fmt.Println("not tls connection")
		return
	}
	if err := tlsConn.Handshake(); err != nil {
		_, _ = io.WriteString(conn, "HTTP/1.0 400 Bad Request\r\n\r\nClient sent an HTTP request to an HTTPS server.\n")
		fmt.Println(err)
		return
	}
	data := make([]byte, 16)
	for {
		_ = tlsConn.SetDeadline(time.Now().Add(300 * time.Millisecond))
		n, err := tlsConn.Read(data)
		if err != nil && err != io.EOF {
			fmt.Println(err)
			break
		}
		fmt.Printf("%s\t%s\n", hex.EncodeToString(data[:n]), string(data[:n]))
		if err == io.EOF {
			break
		}
	}
}

func (r *RawHTTP2Echo) Start() (err error) {
	if r.running {
		return nil
	}
	r.running = true
	cert, err := tls.LoadX509KeyPair(r.CertFile, r.KeyFile)
	if err != nil {
		return err
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	r.listener, err = tls.Listen("tcp", r.Addr, &config)
	defer func() {
		_ = r.listener.Close()
		r.listener = nil
	}()
	if err != nil {
		return err
	}

	for r.running {
		conn, err := r.listener.Accept()
		if err != nil {
			r.running = false
			return err
		}
		go r.ServeConn(conn)
	}
	return nil
}

func StartRawHTTP2Echo(addr, certFile, keyFile string) {
	srv := RawHTTP2Echo{Addr: addr, CertFile: certFile, KeyFile: keyFile}
	srv.bufferPool = buffer.NewPool()
	if err := srv.Start(); err != nil {
		fmt.Println(err)
	}
}
