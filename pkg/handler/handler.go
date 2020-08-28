package handler

import (
	"context"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
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
	fromCtx := ctx.Value("logger")

	if fromCtx == nil {
		return log.WithFields(log.Fields{})
	}

	return fromCtx.(*log.Entry)
}

func logWithFields(ctx context.Context, fields log.Fields) context.Context {
	newLog := logFrom(ctx).WithFields(fields)
	return context.WithValue(ctx, "logger", newLog)
}
