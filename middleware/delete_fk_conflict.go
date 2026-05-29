package middleware

import (
	"bytes"
	"net/http"
	"strings"

	"auto_park/internal/apierrors"

	"github.com/gin-gonic/gin"
)

type bufferedResponseWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
	wrote  bool
}

func (w *bufferedResponseWriter) WriteHeader(code int) {
	if w.wrote {
		return
	}
	w.status = code
}

func (w *bufferedResponseWriter) WriteHeaderNow() {
	if w.wrote {
		return
	}
	w.wrote = true
	if w.status == 0 {
		w.status = http.StatusOK
	}
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	return w.body.Write(data)
}

func (w *bufferedResponseWriter) WriteString(data string) (int, error) {
	w.WriteHeaderNow()
	return w.body.WriteString(data)
}

func (w *bufferedResponseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *bufferedResponseWriter) Written() bool {
	return w.wrote
}

func DeleteForeignKeyConflictMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodDelete {
			c.Next()
			return
		}

		bw := &bufferedResponseWriter{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = bw
		c.Next()

		body := bw.body.Bytes()
		status := bw.status
		if status == 0 {
			status = http.StatusOK
		}

		raw := strings.ToLower(string(body))
		if status >= http.StatusInternalServerError &&
			(strings.Contains(raw, "violates foreign key constraint") || strings.Contains(raw, "sqlstate 23503")) {
			bw.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
			bw.ResponseWriter.WriteHeader(http.StatusConflict)
			_, _ = bw.ResponseWriter.Write([]byte(`{"success":false,"code":"` + apierrors.CodeEntityHasReferences + `","error":"Нельзя удалить: есть связанные записи"}`))
			return
		}

		bw.ResponseWriter.WriteHeader(status)
		_, _ = bw.ResponseWriter.Write(body)
	}
}
