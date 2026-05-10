package scim

import "github.com/gin-gonic/gin"

type Handler struct {
	providers Providers
	token     string
}

func NewHandler(providers Providers, token string) *Handler {
	return &Handler{providers: providers, token: token}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/scim/v2", h.authMiddleware())
	h.registerUserRoutes(g)
}
