package scim

import (
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
func setMeta(c *gin.Context, meta *Meta, resourceType, endpointBase, id string) {
	meta.ResourceType = resourceType
	meta.Location = absoluteURL(c, endpointBase+"/"+id)
}
