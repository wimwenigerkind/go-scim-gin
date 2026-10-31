package scim

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	userResourceType = "User"
	userEndpoint     = "/Users"
)

func (h *Handler) registerUserRoutes(g *gin.RouterGroup) {
	if h.providers.Users == nil {
		return
	}
	g.GET("/Users", h.listUsers)
	g.GET("/Users/:id", h.getUser)
	g.POST("/Users", h.createUser)
}

func (h *Handler) listUsers(c *gin.Context) {
	var filter *Filter
	if raw := c.Query("filter"); raw != "" {
		f, err := parseFilter(raw, SchemaCoreUser)
		if err != nil {
			scimJSON(c, http.StatusBadRequest,
				scimError("400", fmt.Sprintf("invalid filter: %v", err), ScimTypeInvalidFilter))
			return
		}
		filter = &f
	}

	users, err := h.providers.Users.ListUsers(c.Request.Context(), filter)
	if err != nil {
		providerError(c, err, userResourceType)
		return
	}

	if users == nil {
		users = []User{}
	}

	for i := range users {
		h.prepareUserResponse(c, &users[i])
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

	user, err := h.providers.Users.GetUser(c.Request.Context(), id)
	if err == nil && user == nil {
		err = errNilResource
	}
	if err != nil {
		providerError(c, err, userResourceType)
		return
	}

	h.prepareUserResponse(c, user)
	scimJSON(c, http.StatusOK, user)
}

func (h *Handler) createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		scimJSON(c, http.StatusBadRequest, scimError("400", err.Error(), ScimTypeInvalidSyntax))
		return
	}

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	if strings.TrimSpace(user.UserName) == "" {
		scimJSON(c, http.StatusBadRequest, scimError("400", "userName is required", ScimTypeInvalidValue))
		return
	}

	if user.CommonAttributes == nil {
		user.CommonAttributes = &CommonAttributes{}
	}
	user.ID = ""
	user.Meta = nil

	user.Schemas = []Schema{SchemaCoreUser}

	created, err := h.providers.Users.CreateUser(c.Request.Context(), user)
	if err == nil && created == nil {
		err = errNilResource
	}
	if err != nil {
		providerError(c, err, userResourceType)
		return
	}

	h.prepareUserResponse(c, created)
	c.Header("Location", created.Meta.Location)
	scimJSON(c, http.StatusCreated, created)
}

func (h *Handler) prepareUserResponse(c *gin.Context, user *User) {
	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1: password has
	// "returned" set to "never", whatever the provider hands back.
	user.Password = ""
	// https://www.rfc-editor.org/rfc/rfc7643#section-3: schemas MUST be a
	// non-empty array; providers rarely store it.
	if len(user.Schemas) == 0 {
		user.Schemas = []Schema{SchemaCoreUser}
	}
	if user.CommonAttributes == nil {
		user.CommonAttributes = &CommonAttributes{}
	}
	if user.Meta == nil {
		user.Meta = &Meta{}
	}
	setMeta(user.Meta, userResourceType, h.resourceURL(c, userEndpoint+"/"+user.ID))
}
