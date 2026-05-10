package scim

import "errors"

// ErrNotFound https://www.rfc-editor.org/rfc/rfc7644#section-3.12
var ErrNotFound = errors.New("scim: resource not found")
