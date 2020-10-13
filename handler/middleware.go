package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Location")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
			return
		}

		next.ServeHTTP(w, r)
	})
}

type contextKey string

const contextLogger = contextKey("log")

type logWrapper struct {
	logger *log.Entry
}

func newLogWrapper() *logWrapper {
	return &logWrapper{
		logger: log.NewEntry(log.StandardLogger()),
	}
}

func (w *logWrapper) getLog() *log.Entry {
	if w.logger == nil {
		w.logger = log.NewEntry(log.StandardLogger())
	}

	return w.logger
}

func (w *logWrapper) addFields(fields log.Fields) *log.Entry {
	w.logger = w.getLog().WithFields(fields)

	return w.logger
}

func ContextLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid, _ := uuid.NewRandom()

		clog := newLogWrapper()
		clog.addFields(log.Fields{
			"path":   r.URL.Path,
			"method": r.Method,
			"rid":    rid,
		})
		clog.getLog().Debug("incoming request")

		logCtx := context.WithValue(r.Context(), contextLogger, clog)
		next.ServeHTTP(w, r.WithContext(logCtx))
	})
}

func LogFrom(r *http.Request) *log.Entry {
	return r.Context().Value(contextLogger).(*logWrapper).getLog()
}

func AddLogField(r *http.Request, k string, v interface{}) *log.Entry {
	return r.Context().Value(contextLogger).(*logWrapper).addFields(log.Fields{k: v})
}

func AddLogFields(r *http.Request, fields log.Fields) *log.Entry {
	return r.Context().Value(contextLogger).(*logWrapper).addFields(fields)
}
