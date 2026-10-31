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

	// ErrTooMany signals that a query would return more results than the
	// provider is willing to process.
	// https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2
	ErrTooMany = errors.New("scim: too many results")

	// ErrUnsupportedFilter signals an attribute and operator combination the
	// provider cannot filter on. Its message is returned to the client, so
	// wrap it with the attribute name rather than a database error.
	// https://www.rfc-editor.org/rfc/rfc7644#section-3.12
	ErrUnsupportedFilter = errors.New("scim: unsupported filter")
)

// errNilResource replaces a provider's (nil, nil) result, which would
// otherwise panic in the handler.
var errNilResource = errors.New("scim: provider returned neither a resource nor an error")
