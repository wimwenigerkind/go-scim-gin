package scim

import "time"

type Schema string

const (
	// SchemaCoreUser https://www.rfc-editor.org/rfc/rfc7643#section-4.1
	SchemaCoreUser Schema = "urn:ietf:params:scim:schemas:core:2.0:User"

	// SchemaListResponse https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2
	SchemaListResponse Schema = "urn:ietf:params:scim:api:messages:2.0:ListResponse"

	// SchemaError https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	SchemaError Schema = "urn:ietf:params:scim:api:messages:2.0:Error"
)

// Meta https://www.rfc-editor.org/rfc/rfc7643#section-3.1
type Meta struct {
	// https://www.rfc-editor.org/rfc/rfc7643#section-3.1
	ResourceType string `json:"resourceType,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-3.1
	Created *time.Time `json:"created,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-3.1
	LastModified *time.Time `json:"lastModified,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-3.1
	Location string `json:"location,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-3.1
	Version string `json:"version,omitempty"`
}

type CommonAttributes struct {
	// https://www.rfc-editor.org/rfc/rfc7643#section-3.1
	ID string `json:"id,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-3.1
	ExternalID string `json:"externalId,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-3.1
	Meta *Meta `json:"meta,omitempty"`
}

// MultiValuedAttributes https://www.rfc-editor.org/rfc/rfc7643#section-2.4
type MultiValuedAttributes struct {
	// https://www.rfc-editor.org/rfc/rfc7643#section-2.4
	Type string `json:"type,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-2.4
	Primary bool `json:"primary,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-2.4
	Display string `json:"display,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-2.4
	Value string `json:"value,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-2.4
	Ref string `json:"$ref,omitempty"`
}

// Name https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
type Name struct {
	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	Formatted string `json:"formatted,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	FamilyName string `json:"familyName,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	GivenName string `json:"givenName,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	MiddleName string `json:"middleName,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	HonorificPrefix string `json:"honorificPrefix,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	HonorificSuffix string `json:"honorificSuffix,omitempty"`
}

type Address struct {
	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.2
	*MultiValuedAttributes

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.2
	Formatted string `json:"formatted,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.2
	StreetAddress string `json:"streetAddress,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.2
	Locality string `json:"locality,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.2
	Region string `json:"region,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.2
	PostalCode string `json:"postalCode,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.2
	Country string `json:"country,omitempty"`
}

// User https://www.rfc-editor.org/rfc/rfc7643#section-4.1
type User struct {
	// https://www.rfc-editor.org/rfc/rfc7643#section-3.1
	*CommonAttributes

	// https://www.rfc-editor.org/rfc/rfc7643#section-3
	Schemas []Schema `json:"schemas"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	UserName string `json:"userName"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	Name *Name `json:"name,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	DisplayName string `json:"displayName,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	NickName string `json:"nickName,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	ProfileURL string `json:"profileUrl,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	Title string `json:"title,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	UserType string `json:"userType,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	PreferredLanguage string `json:"preferredLanguage,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	Locale string `json:"locale,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	Timezone string `json:"timezone,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	Active bool `json:"active"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.1
	Password string `json:"password,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.2
	Emails []MultiValuedAttributes `json:"emails,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.2
	PhoneNumbers []MultiValuedAttributes `json:"phoneNumbers,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.2
	Ims []MultiValuedAttributes `json:"ims,omitempty"`

	// https://www.rfc-editor.org/rfc/rfc7643#section-4.1.2
	Addresses []Address `json:"addresses,omitempty"`
}

type ListResponse struct {
	Schemas      []Schema `json:"schemas"`
	TotalResults int      `json:"totalResults"`
	StartIndex   int      `json:"startIndex"`
	ItemsPerPage int      `json:"itemsPerPage"`
	Resources    []User   `json:"Resources"`
}
