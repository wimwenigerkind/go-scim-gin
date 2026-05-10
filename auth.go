package scim

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if subtle.ConstantTimeCompare([]byte(auth), []byte("Bearer "+h.token)) != 1 {
			// https://www.rfc-editor.org/rfc/rfc7235#section-4.1
			c.Header("WWW-Authenticate", `Bearer realm="scim"`)
			scimJSON(c, http.StatusUnauthorized, scimError("401", "Unauthorized"))
			c.Abort()
			return
		}
		c.Next()
	}
}
