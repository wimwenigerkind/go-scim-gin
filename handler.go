package scim

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// https://www.rfc-editor.org/rfc/rfc7644#section-3.1
const contentTypeSCIM = "application/scim+json"

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
		scimJSON(c, http.StatusInternalServerError, scimError("500", err.Error()))
		return
	}

	if users == nil {
		users = []User{}
	}

	for i := range users {
		h.enrichMeta(c, &users[i])
	}

	scimJSON(c, http.StatusOK, ListResponse{
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
			scimJSON(c, http.StatusNotFound, scimError("404", "User not found"))
			return
		}
		scimJSON(c, http.StatusInternalServerError, scimError("500", err.Error()))
		return
	}

	h.enrichMeta(c, user)
	scimJSON(c, http.StatusOK, user)
}

func (h *Handler) createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		scimJSON(c, http.StatusBadRequest, scimError("400", err.Error()))
		return
	}

	if user.CommonAttributes == nil {
		user.CommonAttributes = &CommonAttributes{}
	}
	user.ID = ""
	user.Meta = nil

	user.Schemas = []Schema{SchemaCoreUser}

	created, err := h.provider.CreateUser(c.Request.Context(), user)
	if err != nil {
		scimJSON(c, http.StatusInternalServerError, scimError("500", err.Error()))
		return
	}

	h.enrichMeta(c, created)
	c.Header("Location", created.Meta.Location)
	scimJSON(c, http.StatusCreated, created)
}

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

// https://www.rfc-editor.org/rfc/rfc7644#section-3.12
func scimError(status, detail string) gin.H {
	return gin.H{
		"schemas": []Schema{SchemaError},
		"status":  status,
		"detail":  detail,
	}
}

// https://www.rfc-editor.org/rfc/rfc7644#section-3.1
func scimJSON(c *gin.Context, status int, body any) {
	c.Header("Content-Type", contentTypeSCIM)
	c.JSON(status, body)
}

// https://www.rfc-editor.org/rfc/rfc7643#section-3.1
func (h *Handler) enrichMeta(c *gin.Context, user *User) {
	if user.CommonAttributes == nil {
		user.CommonAttributes = &CommonAttributes{}
	}
	if user.Meta == nil {
		user.Meta = &Meta{}
	}
	user.Meta.ResourceType = "User"
	user.Meta.Location = absoluteURL(c, "/scim/v2/Users/"+user.ID)
}

func absoluteURL(c *gin.Context, path string) string {
	scheme := "http"
	if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + path
}
