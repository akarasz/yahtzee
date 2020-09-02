package handler

import (
	"context"
	"net/http"
	"path"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type contextKey string

const (
	logger contextKey = "logger"
)

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func logFrom(ctx context.Context) *log.Entry {
	fromCtx := ctx.Value(logger)

	if fromCtx == nil {
		return log.WithFields(log.Fields{})
	}

	return fromCtx.(*log.Entry)
}

func logWithFields(ctx context.Context, fields log.Fields) context.Context {
	newLog := logFrom(ctx).WithFields(fields)
	return context.WithValue(ctx, logger, newLog)
}

func contextLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newLog := log.WithFields(log.Fields{
			"method":    r.Method,
			"path":      r.URL.Path,
			"requestID": uuid.Must(uuid.NewRandom()),
		})

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), logger, newLog)))
	})
}

func allowCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")

		next.ServeHTTP(w, r)
	})
}
