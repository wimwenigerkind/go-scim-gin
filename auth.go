package scim

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.authorized(c.GetHeader("Authorization")) {
			// https://www.rfc-editor.org/rfc/rfc7235#section-4.1
			c.Header("WWW-Authenticate", `Bearer realm="scim"`)
			scimJSON(c, http.StatusUnauthorized, scimError("401", "Unauthorized"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *Handler) authorized(header string) bool {
	// An unset token must lock the endpoint, not open it to "Bearer ".
	if h.token == "" {
		return false
	}
	scheme, token, ok := strings.Cut(header, " ")
	// https://www.rfc-editor.org/rfc/rfc7235#section-2.1: the auth-scheme is
	// case-insensitive.
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.TrimLeft(token, " ")), []byte(h.token)) == 1
}
