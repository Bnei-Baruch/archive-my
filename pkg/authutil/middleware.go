package authutil

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

func AuthenticationMiddleware(verifier OIDCTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.Split(strings.TrimSpace(c.Request.Header.Get("Authorization")), " ")
		if len(authHeader) == 2 || strings.ToLower(authHeader[0]) == "bearer" {
			// Authorization header provided, let's verify.
			token, err := verifier.Verify(context.TODO(), authHeader[1])
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
				return
			}

			// ID Token is verified.
			c.Set("KC_ID", token.Subject)
		}
		c.Next()
	}
}
