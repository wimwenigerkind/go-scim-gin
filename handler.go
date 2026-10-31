package scim

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	providers Providers
	token     string
	baseURL   string
	basePath  string
}

type Option func(*Handler)

// WithBaseURL sets the externally visible SCIM base URL, for example
// "https://portal.example.com/api/scim/v2". Location headers and
// meta.location are built from it instead of from the request, which is
// needed when a reverse proxy rewrites the path or the host.
func WithBaseURL(baseURL string) Option {
	return func(h *Handler) {
		h.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// NewHandler serves the given providers behind a static bearer token. An
// empty token rejects every request.
func NewHandler(providers Providers, token string, opts ...Option) *Handler {
	h := &Handler{providers: providers, token: token}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/scim/v2", h.authMiddleware())
	// Includes the prefixes of the router the caller passed in, so locations
	// stay correct when the routes are mounted below a group like /api.
	h.basePath = g.BasePath()
	h.registerUserRoutes(g)
}

func (h *Handler) resourceURL(c *gin.Context, path string) string {
	if h.baseURL != "" {
		return h.baseURL + path
	}
	return absoluteURL(c, h.basePath+path)
}
