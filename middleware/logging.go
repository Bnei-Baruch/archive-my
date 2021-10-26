package middleware

import (
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"

	"github.com/Bnei-Baruch/archive-my/instrumentation"
)

var requestLog = zerolog.New(os.Stdout).With().Timestamp().Caller().Stack().Logger()

// zerolog helpers adapted for gin (github.com/rs/zerolog/hlog)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// request id (see hlog.RequestIDHandler)
		requestID := ksuid.New()
		c.Set("REQUEST_ID", requestID)
		c.Header("X-Request-ID", requestID.String())

		// Create a copy of the logger (see hlog.NewHandler)
		l := requestLog.With().Str("request_id", requestID.String()).Logger()
		c.Set("LOGGER", l)

		// log line (see hlog.AccessHandler)
		r := c.Request
		path := r.URL.RequestURI() // some evil middleware modify this values

		c.Next()

		latency := time.Now().Sub(start)

		l.Info().
			Str("method", r.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Int("size", c.Writer.Size()).
			Dur("duration", latency).
			Str("ip", c.ClientIP()).
			Msg("")

		instrumentation.Stats.RequestDurationHistogram.
			WithLabelValues(c.Request.Method, path, strconv.Itoa(c.Writer.Status())).
			Observe(latency.Seconds())
	}
}
