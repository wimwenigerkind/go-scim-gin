package scim

import "errors"

// Providers return these, wrapped or not, to get the matching SCIM error
// response. Any other error becomes a 500 without its text.
var (
	// ErrNotFound https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ErrNotFound = errors.New("scim: resource not found")

	// ErrConflict signals a uniqueness violation, such as a duplicate
	// userName. https://www.rfc-editor.org/rfc/rfc7644#section-3.3
	ErrConflict = errors.New("scim: resource already exists")
)
