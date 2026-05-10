package scim

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// https://www.rfc-editor.org/rfc/rfc7644#section-3.1
const contentTypeSCIM = "application/scim+json"

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
