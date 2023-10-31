package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger — middleware-logger for incoming HTTP-requests.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		ts := time.Now()

		c.Next()

		logger.Log.Info("got HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("code", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.Duration("duration", time.Since(ts)),
		)
	}
}

// Gzip - middleware to (de)compress request/response body using gzip.
func Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Header("Content-Encoding", "gzip")
			c.Header("Vary", "Accept-Encoding")

			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(c.Writer)

			// меняем оригинальный gin.ResponseWriter на новый
			c.Writer = cw
			defer func() {
				cw.Close()
				c.Header("Content-Length", strconv.Itoa(c.Writer.Size()))
			}()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := c.Request.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(c.Request.Body)
			if err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			// меняем тело запроса на новое
			c.Request.Body = cr
			defer cr.Close()
		}

		c.Next()
	}
}

// compressWriter implements http.ResponseWriter interface
type compressWriter struct {
	gin.ResponseWriter
	cw *gzip.Writer
}

func newCompressWriter(w gin.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		cw:             gzip.NewWriter(w),
	}
}

func (c *compressWriter) Write(b []byte) (int, error) {
	return c.cw.Write(b)
}

func (c *compressWriter) WriteString(s string) (n int, err error) {
	return c.cw.Write([]byte(s))
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.ResponseWriter.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.cw.Close()
}

// compressReader implements io.ReadCloser
type compressReader struct {
	r  io.ReadCloser
	cr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		cr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.cr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.cr.Close()
}
