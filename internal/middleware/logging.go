package middleware

import (
	"log/slog"
	"math/rand/v2"
	"net/http"
	"runtime/debug"

	"effective-mobile-go/internal/logger"
)

func Logging(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log := slog.Default().With("httpReqID", rand.Uint64())

		url := r.URL.String()
		log.Debug("http request begin", "fromAddr", r.RemoteAddr, "method", r.Method, "url", url)

		w = newWriteHeaderHook(w, func(statusCode int) {
			log.Debug("http request end", "statusCode", statusCode, "url", url)
		})

		ctx := logger.ContextWithLogger(r.Context(), log)
		r = r.WithContext(ctx)

		defer func() {
			if p := recover(); p != nil {
				log.Error("*** panic recovered ***", "panic", p, "stack", debug.Stack())
			}
		}()

		h.ServeHTTP(w, r)
	}
}

type writeHeaderHook struct {
	http.ResponseWriter
	hook func(statusCode int)
	flag bool // need to use atomic.Bool for thread safety
}

func newWriteHeaderHook(w http.ResponseWriter, hook func(statusCode int)) *writeHeaderHook {
	return &writeHeaderHook{
		ResponseWriter: w,
		hook:           hook,
	}
}

func (rw *writeHeaderHook) WriteHeader(statusCode int) {
	if !rw.flag {
		rw.flag = true
		rw.hook(statusCode)
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *writeHeaderHook) Write(b []byte) (int, error) {
	rw.WriteHeader(http.StatusOK)
	return rw.ResponseWriter.Write(b)
}
