package scim

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// https://www.rfc-editor.org/rfc/rfc7644#section-3.1
const contentTypeSCIM = "application/scim+json"

// ScimType https://www.rfc-editor.org/rfc/rfc7644#section-3.12
type ScimType string

const (
	// ScimTypeInvalidFilter https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ScimTypeInvalidFilter ScimType = "invalidFilter"

	// ScimTypeTooMany https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ScimTypeTooMany ScimType = "tooMany"

	// ScimTypeUniqueness https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ScimTypeUniqueness ScimType = "uniqueness"

	// ScimTypeMutability https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ScimTypeMutability ScimType = "mutability"

	// ScimTypeInvalidSyntax https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ScimTypeInvalidSyntax ScimType = "invalidSyntax"

	// ScimTypeInvalidPath https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ScimTypeInvalidPath ScimType = "invalidPath"

	// ScimTypeNoTarget https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ScimTypeNoTarget ScimType = "noTarget"

	// ScimTypeInvalidValue https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ScimTypeInvalidValue ScimType = "invalidValue"

	// ScimTypeInvalidVers https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ScimTypeInvalidVers ScimType = "invalidVers"

	// ScimTypeSensitive https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ScimTypeSensitive ScimType = "sensitive"
)

// https://www.rfc-editor.org/rfc/rfc7644#section-3.12
func scimError(status, detail string, scimType ...ScimType) gin.H {
	h := gin.H{
		"schemas": []Schema{SchemaError},
		"status":  status,
		"detail":  detail,
	}
	if len(scimType) > 0 && scimType[0] != "" {
		h["scimType"] = scimType[0]
	}
	return h
}

// https://www.rfc-editor.org/rfc/rfc7644#section-3.1
func scimJSON(c *gin.Context, status int, body any) {
	c.Header("Content-Type", contentTypeSCIM)
	c.JSON(status, body)
}

func absoluteURL(c *gin.Context, path string) string {
	scheme := "http"
	if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + path
}

// https://www.rfc-editor.org/rfc/rfc7643#section-3.1
func setMeta(meta *Meta, resourceType, location string) {
	meta.ResourceType = resourceType
	meta.Location = location
}

// providerError maps the sentinel errors a provider may return to SCIM error
// responses. Anything else becomes a 500 whose text stays on the server: it
// is attached to the gin context for the caller's logging middleware, since
// provider errors tend to carry database details.
func providerError(c *gin.Context, err error, resourceType string) {
	switch {
	case errors.Is(err, ErrNotFound):
		scimJSON(c, http.StatusNotFound, scimError("404", resourceType+" not found"))
	case errors.Is(err, ErrConflict):
		scimJSON(c, http.StatusConflict, scimError("409", resourceType+" already exists", ScimTypeUniqueness))
	case errors.Is(err, ErrTooMany):
		scimJSON(c, http.StatusBadRequest, scimError("400", "Too many results, narrow the filter", ScimTypeTooMany))
	case errors.Is(err, ErrUnsupportedFilter):
		scimJSON(c, http.StatusBadRequest, scimError("400", err.Error(), ScimTypeInvalidFilter))
	default:
		_ = c.Error(err)
		scimJSON(c, http.StatusInternalServerError, scimError("500", "Internal server error"))
	}
}
