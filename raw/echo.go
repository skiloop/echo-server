package raw

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"go.uber.org/zap/buffer"
	"io"
	"net"
	"net/http"
	"os"
)

type RawEcho struct {
	Addr         string
	CertFile     string
	KeyFile      string
	listener     net.Listener
	running      bool
	bufferPool   buffer.Pool
	keyLogWriter io.WriteCloser
}

func (r RawEcho) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	buf := r.bufferPool.Get()
	err := req.Write(buf)
	if err != nil {
		fmt.Println(err)
	}
	resp := make(map[string]string)
	resp["content"] = buf.String()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	err = enc.Encode(resp)
	if err != nil {
		fmt.Println(err)
	}
}
func (r *RawEcho) close() {
	if r.keyLogWriter != nil {
		_ = r.keyLogWriter.Close()
	}
}
func (r *RawEcho) Start() error {
	tlsConfig := new(tls.Config)
	tlsConfig.MaxVersion = tls.VersionTLS13

	tlsConfig.KeyLogWriter = r.keyLogWriter
	srv := http.Server{Addr: r.Addr, Handler: r, TLSConfig: tlsConfig}

	err := srv.ListenAndServeTLS(r.CertFile, r.KeyFile)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func StartRawEcho(addr, certFile, keyFile string) {
	writer, err := os.OpenFile("/Users/skiloop/.mitmproxy/ssl_key_echo.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o640)
	if err != nil {
		fmt.Println(err)
		return
	}
	srv := RawEcho{Addr: addr, CertFile: certFile, KeyFile: keyFile}
	srv.keyLogWriter = writer
	srv.bufferPool = buffer.NewPool()
	if err := srv.Start(); err != nil {
		fmt.Println(err)
	}
}
