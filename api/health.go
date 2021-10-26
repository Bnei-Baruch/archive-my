package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (a *App) HealthCheckHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	err := a.DB.PingContext(ctx)

	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("MDB ping: %s", err.Error()),
		})
		return
	}

	if ctx.Err() == context.DeadlineExceeded {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  "timeout",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (a *App) MakePrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
