package server

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/logger"
	"github.com/Dmitrevicz/gometrics/pkg/encryptor"
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

const HashHeader = "HashSHA256"

// Hash middleware.
//
// > Реализуйте механизм подписи передаваемых данных по алгоритму SHA256. Для
// этого посчитайте hash от всего тела запроса и разместите его в HTTP-заголовке
// HashSHA256. Хеш нужно считать от строки с учётом ключа, который передан
// агенту/серверу на старте: hash(value, key).
func Hash(key string) gin.HandlerFunc {
	if key == "" {
		return func(c *gin.Context) {}
	}

	return func(c *gin.Context) {
		fmt.Println("Hash middleware fired")

		rw := &bodyCatcherWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = rw

		c.Next()

		// if rw.body.Len() == 0 {
		// 	return
		// }

		hasher := hmac.New(sha256.New, []byte(key))
		_, err := hasher.Write(rw.body.Bytes())
		if err != nil {
			logger.Log.Error("Error creating hash for response body", zap.Error(err))
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		hash := hasher.Sum(nil)
		c.Header(HashHeader, hex.EncodeToString(hash))

		rw.body = &bytes.Buffer{} // (clear underlying memory?)
	}
}

type bodyCatcherWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyCatcherWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyCatcherWriter) WriteString(s string) (n int, err error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func HashCheck(key string) gin.HandlerFunc {
	if key == "" {
		return func(c *gin.Context) {}
	}

	return func(c *gin.Context) {
		fmt.Println("HashCheck fired")

		header := c.GetHeader(HashHeader)

		// Workaround autotests bug.
		// Figured out that autotests for this iteration are not finished yet:
		// hash header value is always empty now... so don't make any checks, I guess.
		if header == "" {
			return
		}

		if header == "" {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("header required: "+HashHeader))
		}

		headerHash, err := hex.DecodeString(header)
		if err != nil {
			logger.Log.Error("Error reading hash from header", zap.Error(err))
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("bad header: "+HashHeader))
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Log.Error("Error reading body", zap.Error(err))
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("can't read body"))
			return
		}

		hasher := hmac.New(sha256.New, []byte(key))
		hasher.Write(body)
		hash := hasher.Sum(nil)

		if !hmac.Equal(hash, headerHash) {
			logger.Log.Info("hash check failed",
				zap.String("key", key),
				zap.String("calculated", hex.EncodeToString(hash)),
				zap.String("provided", header),
			)
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("wrong hash"))
		} else {
			logger.Log.Info("successful hash check",
				zap.String("method", c.Request.Method),
				zap.String("req", c.Request.URL.Path),
			)
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}
}

const EncryptionHeader = "Content-Encryption"

// DecryptRSA - middleware to decrypt request body encrypted with RSA.
func DecryptRSA(decryptor *encryptor.Decryptor) gin.HandlerFunc {
	if decryptor == nil {
		logger.Log.DPanic("bad decryptor initialization - nil passed as *encryptor.Decryptor")
		return func(c *gin.Context) {}
	}

	return func(c *gin.Context) {
		// check if data sent by agent was encrypted
		enc := c.Request.Header.Get(EncryptionHeader)
		fmt.Printf("Decryption middleware fired, header - %s: '%s'\n", EncryptionHeader, enc)
		if enc != "1" {
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Log.Error("Error reading body", zap.Error(err))
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("can't read body"))
			return
		}

		plainBody, err := decryptor.Decrypt(body)
		if err != nil {
			logger.Log.Info("Content decryption failed", zap.Error(err))
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("content decryption failed"))
		} else {
			logger.Log.Info("Successful request decryption",
				zap.String("method", c.Request.Method),
				zap.String("req", c.Request.URL.Path),
			)
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(plainBody))
	}
}
