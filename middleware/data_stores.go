package middleware

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func DataStoresMiddleware(myDB *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("MY_DB", myDB)
		c.Next()
	}
}
