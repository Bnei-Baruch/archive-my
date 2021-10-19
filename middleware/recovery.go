package middleware

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rval := recover(); rval != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				// Edo: this is taken from gin's own recovery middleware
				var brokenPipe bool
				if ne, ok := rval.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				err, ok := rval.(error)
				if !ok {
					err = fmt.Errorf("panic: %+v", rval)
				}

				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
				} else {
					c.AbortWithError(http.StatusInternalServerError, err).SetType(gin.ErrorTypePrivate)
				}
			}
		}()

		c.Next()
	}
}
