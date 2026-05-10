package scim

import (
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	provider UserProvider
	token    string
}

func NewHandler(provider UserProvider, token string) *Handler {
	return &Handler{provider: provider, token: token}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/scim/v2", h.authMiddleware())
	g.GET("/Users", h.listUsers)
	g.GET("/Users/:id", h.getUser)
	g.POST("/Users", h.createUser)
}

func (h *Handler) listUsers(c *gin.Context) {
	users, err := h.provider.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, scimError("500", err.Error()))
		return
	}

	if users == nil {
		users = []User{}
	}

	c.JSON(http.StatusOK, ListResponse{
		Schemas:      []Schema{SchemaListResponse},
		TotalResults: len(users),
		StartIndex:   1,
		ItemsPerPage: len(users),
		Resources:    users,
	})
}

func (h *Handler) getUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.provider.GetUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, scimError("404", "User not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, scimError("500", err.Error()))
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, scimError("400", err.Error()))
		return
	}

	user.Schemas = []Schema{SchemaCoreUser}

	created, err := h.provider.CreateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, scimError("500", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if subtle.ConstantTimeCompare([]byte(auth), []byte("Bearer "+h.token)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, scimError("401", "Unauthorized"))
			return
		}
		c.Next()
	}
}

// https://www.rfc-editor.org/rfc/rfc7644#section-3.12
func scimError(status, detail string) gin.H {
	return gin.H{
		"schemas": []Schema{SchemaError},
		"status":  status,
		"detail":  detail,
	}
}
