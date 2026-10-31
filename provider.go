package scim

import "context"

// Providers bundles all resource providers a SCIM server can expose.
type Providers struct {
	Users UserProvider
	// Groups GroupProvider // RFC 7643 §4.2
}

// UserProvider https://www.rfc-editor.org/rfc/rfc7644#section-3.2
//
// Methods return ErrNotFound, ErrConflict, ErrTooMany or
// ErrUnsupportedFilter, wrapped or not, to get the matching SCIM error. Any
// other error becomes a 500 whose text is not sent to the client.
type UserProvider interface {
	// ListUsers https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2
	// ListUsers(ctx context.Context, filter *Filter, pagination Pagination) ([]User, int, error)
	//
	// userName is not case-exact (RFC 7643 §4.1.1): compare its values
	// case-insensitively, or an IdP's lookup before create misses an
	// existing user and creates it twice.
	ListUsers(ctx context.Context, filter *Filter) ([]User, error)

	// GetUser https://www.rfc-editor.org/rfc/rfc7644#section-3.4.1
	GetUser(ctx context.Context, id string) (*User, error)

	// CreateUser https://www.rfc-editor.org/rfc/rfc7644#section-3.3
	// Return ErrConflict for a duplicate userName.
	CreateUser(ctx context.Context, user User) (*User, error)

	// ReplaceUser https://www.rfc-editor.org/rfc/rfc7644#section-3.5.1
	// ReplaceUser(ctx context.Context, id string, user User) (*User, error)

	// PatchUser https://www.rfc-editor.org/rfc/rfc7644#section-3.5.2
	// PatchUser(ctx context.Context, id string, ops []Operation) (*User, error)

	// DeleteUser https://www.rfc-editor.org/rfc/rfc7644#section-3.6
	// DeleteUser(ctx context.Context, id string) error
}
