package middleware

import (
	"bufio"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type compressWriter struct {
	w  gin.ResponseWriter
	zw *gzip.Writer
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.Header().Del("Content-Length")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) WriteString(s string) (int, error) {
	return c.zw.Write([]byte(s))
}

func (c *compressWriter) Status() int {
	return c.w.Status()
}

func (c *compressWriter) Size() int {
	return c.w.Size()
}

func (c *compressWriter) Written() bool {
	return c.w.Written()
}

func (c *compressWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return c.w.Hijack()
}

func (c *compressWriter) CloseNotify() <-chan bool {
	return c.w.CloseNotify()
}

func (c *compressWriter) Flush() {
	c.zw.Flush()
	if f, ok := c.w.(http.Flusher); ok {
		f.Flush()
	}
}

func (c *compressWriter) Pusher() http.Pusher {
	if p, ok := c.w.(http.Pusher); ok {
		return p
	}
	return nil
}

func (c *compressWriter) WriteHeaderNow() {
	c.w.WriteHeaderNow()
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

func newCompressWriter(w gin.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipMiddleware(c *gin.Context) {
	acceptEncoding := c.Request.Header.Get("Accept-Encoding")
	supportsGzip := strings.Contains(acceptEncoding, "gzip")
	if supportsGzip {
		cw := newCompressWriter(c.Writer)
		c.Writer = cw
		defer cw.Close()
	}

	contentEncoding := c.Request.Header.Get("Content-Encoding")
	sendsGzip := strings.Contains(contentEncoding, "gzip")
	if sendsGzip {
		cr, err := newCompressReader(c.Request.Body)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		c.Request.Body = cr
		defer cr.Close()
	}

	c.Next()
}
