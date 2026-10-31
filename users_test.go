package scim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const testToken = "secret"

type fakeUserProvider struct {
	users []User
	err   error

	gotFilter  *Filter
	listCalled bool
	gotCreate  *User
	createdID  string
}

func (p *fakeUserProvider) ListUsers(_ context.Context, filter *Filter) ([]User, error) {
	p.listCalled = true
	p.gotFilter = filter
	return p.users, p.err
}

func (p *fakeUserProvider) GetUser(_ context.Context, id string) (*User, error) {
	if p.err != nil {
		return nil, p.err
	}
	for i := range p.users {
		if p.users[i].ID == id {
			return &p.users[i], nil
		}
	}
	return nil, ErrNotFound
}

func (p *fakeUserProvider) CreateUser(_ context.Context, user User) (*User, error) {
	p.gotCreate = &user
	if p.err != nil {
		return nil, p.err
	}
	created := user
	created.CommonAttributes = &CommonAttributes{ID: p.createdID, ExternalID: user.ExternalID}
	return &created, nil
}

func newTestRouter(provider UserProvider) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(Providers{Users: provider}, testToken).RegisterRoutes(r)
	return r
}

func doRequest(r *gin.Engine, method, target, body string, header http.Header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+testToken)
	for k, v := range header {
		req.Header[k] = v
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decode[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	return v
}

func assertSCIMResponse(t *testing.T, w *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()
	if w.Code != wantStatus {
		t.Fatalf("status = %d, want %d, body %s", w.Code, wantStatus, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, contentTypeSCIM) {
		t.Errorf("Content-Type = %q, want %q", ct, contentTypeSCIM)
	}
}

func assertSCIMError(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, wantScimType ScimType) {
	t.Helper()
	assertSCIMResponse(t, w, wantStatus)
	body := decode[map[string]any](t, w)
	if got := body["status"]; got != fmt.Sprint(wantStatus) {
		t.Errorf("status field = %v, want %q", got, fmt.Sprint(wantStatus))
	}
	if got, _ := body["schemas"].([]any); len(got) != 1 || got[0] != string(SchemaError) {
		t.Errorf("schemas = %v, want [%s]", body["schemas"], SchemaError)
	}
	got, present := body["scimType"]
	if wantScimType == "" {
		if present {
			t.Errorf("scimType = %v, want absent", got)
		}
	} else if got != string(wantScimType) {
		t.Errorf("scimType = %v, want %q", got, wantScimType)
	}
}

func TestAuth(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   int
	}{
		{"missing header", "", http.StatusUnauthorized},
		{"wrong token", "Bearer wrong", http.StatusUnauthorized},
		{"token without scheme", testToken, http.StatusUnauthorized},
		{"valid token", "Bearer " + testToken, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter(&fakeUserProvider{})
			req := httptest.NewRequest(http.MethodGet, "/scim/v2/Users", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if tt.want == http.StatusOK {
				assertSCIMResponse(t, w, http.StatusOK)
				return
			}
			assertSCIMError(t, w, http.StatusUnauthorized, "")
			if got := w.Header().Get("WWW-Authenticate"); got != `Bearer realm="scim"` {
				t.Errorf("WWW-Authenticate = %q", got)
			}
		})
	}
}

func TestRegisterRoutesOnGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	provider := &fakeUserProvider{users: []User{{CommonAttributes: &CommonAttributes{ID: "1"}}}}
	NewHandler(Providers{Users: provider}, testToken).RegisterRoutes(r.Group("/api"))

	w := doRequest(r, http.MethodGet, "/api/scim/v2/Users/1", "", nil)
	assertSCIMResponse(t, w, http.StatusOK)
	got := decode[User](t, w)
	if want := "http://example.com/api/scim/v2/Users/1"; got.Meta.Location != want {
		t.Errorf("meta.location = %q, want %q", got.Meta.Location, want)
	}
}

func TestWithBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	provider := &fakeUserProvider{createdID: "42"}
	NewHandler(Providers{Users: provider}, testToken,
		WithBaseURL("https://portal.example.com/api/scim/v2/")).RegisterRoutes(r)

	w := doRequest(r, http.MethodPost, "/scim/v2/Users", `{"userName":"alice"}`,
		http.Header{"X-Forwarded-Proto": {"http"}})
	assertSCIMResponse(t, w, http.StatusCreated)

	want := "https://portal.example.com/api/scim/v2/Users/42"
	if got := w.Header().Get("Location"); got != want {
		t.Errorf("Location header = %q, want %q", got, want)
	}
	if got := decode[User](t, w).Meta.Location; got != want {
		t.Errorf("meta.location = %q, want %q", got, want)
	}
}

func TestUserRoutesWithoutProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(Providers{}, testToken).RegisterRoutes(r)

	w := doRequest(r, http.MethodGet, "/scim/v2/Users", "", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestListUsers(t *testing.T) {
	t.Run("empty result is an empty array", func(t *testing.T) {
		r := newTestRouter(&fakeUserProvider{})
		w := doRequest(r, http.MethodGet, "/scim/v2/Users", "", nil)

		assertSCIMResponse(t, w, http.StatusOK)
		body := decode[map[string]any](t, w)
		if res, ok := body["Resources"].([]any); !ok || len(res) != 0 {
			t.Errorf("Resources = %v, want []", body["Resources"])
		}
		if body["totalResults"] != float64(0) {
			t.Errorf("totalResults = %v, want 0", body["totalResults"])
		}
	})

	t.Run("users get list envelope and meta", func(t *testing.T) {
		provider := &fakeUserProvider{users: []User{
			{CommonAttributes: &CommonAttributes{ID: "1"}, UserName: "alice"},
			{UserName: "bob"},
		}}
		r := newTestRouter(provider)
		w := doRequest(r, http.MethodGet, "/scim/v2/Users", "", nil)

		assertSCIMResponse(t, w, http.StatusOK)
		got := decode[ListResponse](t, w)
		if !reflect.DeepEqual(got.Schemas, []Schema{SchemaListResponse}) {
			t.Errorf("schemas = %v", got.Schemas)
		}
		if got.TotalResults != 2 || got.ItemsPerPage != 2 || got.StartIndex != 1 {
			t.Errorf("totalResults/itemsPerPage/startIndex = %d/%d/%d, want 2/2/1",
				got.TotalResults, got.ItemsPerPage, got.StartIndex)
		}
		if loc := got.Resources[0].Meta.Location; loc != "http://example.com/scim/v2/Users/1" {
			t.Errorf("meta.location = %q", loc)
		}
		if rt := got.Resources[0].Meta.ResourceType; rt != userResourceType {
			t.Errorf("meta.resourceType = %q", rt)
		}
		if got.Resources[1].Meta == nil {
			t.Error("user without CommonAttributes got no meta")
		}
	})

	t.Run("filter is passed to the provider", func(t *testing.T) {
		provider := &fakeUserProvider{}
		r := newTestRouter(provider)
		target := "/scim/v2/Users?filter=" + url.QueryEscape(`userName eq "alice"`)
		w := doRequest(r, http.MethodGet, target, "", nil)

		assertSCIMResponse(t, w, http.StatusOK)
		want := &Filter{Field: "userName", Operator: FilterOperatorEqual, Value: "alice"}
		if !reflect.DeepEqual(provider.gotFilter, want) {
			t.Errorf("filter = %+v, want %+v", provider.gotFilter, want)
		}
	})

	t.Run("no filter passes nil", func(t *testing.T) {
		provider := &fakeUserProvider{}
		r := newTestRouter(provider)
		doRequest(r, http.MethodGet, "/scim/v2/Users", "", nil)

		if provider.gotFilter != nil {
			t.Errorf("filter = %+v, want nil", provider.gotFilter)
		}
	})

	for _, raw := range []string{" ", `userName`, `userName eq`, `userName eq "bjen`, `userName eq "a" & title pr`} {
		t.Run("invalid filter "+raw, func(t *testing.T) {
			provider := &fakeUserProvider{}
			r := newTestRouter(provider)
			target := "/scim/v2/Users?filter=" + url.QueryEscape(raw)
			w := doRequest(r, http.MethodGet, target, "", nil)

			assertSCIMError(t, w, http.StatusBadRequest, ScimTypeInvalidFilter)
			if provider.listCalled {
				t.Error("provider was called despite invalid filter")
			}
		})
	}

	t.Run("provider error", func(t *testing.T) {
		r := newTestRouter(&fakeUserProvider{err: errors.New("boom")})
		w := doRequest(r, http.MethodGet, "/scim/v2/Users", "", nil)

		assertSCIMError(t, w, http.StatusInternalServerError, "")
	})
}

func TestGetUser(t *testing.T) {
	alice := User{CommonAttributes: &CommonAttributes{ID: "1"}, UserName: "alice"}

	t.Run("found", func(t *testing.T) {
		r := newTestRouter(&fakeUserProvider{users: []User{alice}})
		w := doRequest(r, http.MethodGet, "/scim/v2/Users/1", "", nil)

		assertSCIMResponse(t, w, http.StatusOK)
		got := decode[User](t, w)
		if got.UserName != "alice" || got.ID != "1" {
			t.Errorf("user = %+v", got)
		}
		if got.Meta.Location != "http://example.com/scim/v2/Users/1" {
			t.Errorf("meta.location = %q", got.Meta.Location)
		}
	})

	t.Run("location honours X-Forwarded-Proto", func(t *testing.T) {
		r := newTestRouter(&fakeUserProvider{users: []User{alice}})
		w := doRequest(r, http.MethodGet, "/scim/v2/Users/1", "",
			http.Header{"X-Forwarded-Proto": {"https"}})

		got := decode[User](t, w)
		if got.Meta.Location != "https://example.com/scim/v2/Users/1" {
			t.Errorf("meta.location = %q", got.Meta.Location)
		}
	})

	t.Run("not found", func(t *testing.T) {
		r := newTestRouter(&fakeUserProvider{})
		w := doRequest(r, http.MethodGet, "/scim/v2/Users/missing", "", nil)

		assertSCIMError(t, w, http.StatusNotFound, "")
	})

	t.Run("wrapped not found", func(t *testing.T) {
		r := newTestRouter(&fakeUserProvider{err: fmt.Errorf("lookup: %w", ErrNotFound)})
		w := doRequest(r, http.MethodGet, "/scim/v2/Users/1", "", nil)

		assertSCIMError(t, w, http.StatusNotFound, "")
	})

	t.Run("provider error", func(t *testing.T) {
		r := newTestRouter(&fakeUserProvider{err: errors.New("boom")})
		w := doRequest(r, http.MethodGet, "/scim/v2/Users/1", "", nil)

		assertSCIMError(t, w, http.StatusInternalServerError, "")
	})
}

func TestCreateUser(t *testing.T) {
	t.Run("created", func(t *testing.T) {
		provider := &fakeUserProvider{createdID: "42"}
		r := newTestRouter(provider)
		body := `{
			"schemas": ["something:else"],
			"id": "client-chosen",
			"externalId": "ext-1",
			"meta": {"location": "http://evil.example/x"},
			"userName": "alice"
		}`
		w := doRequest(r, http.MethodPost, "/scim/v2/Users", body, nil)

		assertSCIMResponse(t, w, http.StatusCreated)

		in := provider.gotCreate
		if in.ID != "" {
			t.Errorf("provider got id %q, want it cleared", in.ID)
		}
		if in.Meta != nil {
			t.Errorf("provider got meta %+v, want it cleared", in.Meta)
		}
		if in.ExternalID != "ext-1" {
			t.Errorf("provider got externalId %q, want ext-1", in.ExternalID)
		}
		if !reflect.DeepEqual(in.Schemas, []Schema{SchemaCoreUser}) {
			t.Errorf("provider got schemas %v, want [%s]", in.Schemas, SchemaCoreUser)
		}

		got := decode[User](t, w)
		wantLocation := "http://example.com/scim/v2/Users/42"
		if got.Meta.Location != wantLocation {
			t.Errorf("meta.location = %q, want %q", got.Meta.Location, wantLocation)
		}
		if loc := w.Header().Get("Location"); loc != wantLocation {
			t.Errorf("Location header = %q, want %q", loc, wantLocation)
		}
	})

	t.Run("body without common attributes", func(t *testing.T) {
		provider := &fakeUserProvider{createdID: "1"}
		r := newTestRouter(provider)
		w := doRequest(r, http.MethodPost, "/scim/v2/Users", `{"userName":"bob"}`, nil)

		assertSCIMResponse(t, w, http.StatusCreated)
		if provider.gotCreate.CommonAttributes == nil {
			t.Error("provider got nil CommonAttributes")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		provider := &fakeUserProvider{}
		r := newTestRouter(provider)
		w := doRequest(r, http.MethodPost, "/scim/v2/Users", `{"userName":`, nil)

		assertSCIMError(t, w, http.StatusBadRequest, ScimTypeInvalidSyntax)
		if provider.gotCreate != nil {
			t.Error("provider was called despite invalid body")
		}
	})

	t.Run("provider error", func(t *testing.T) {
		r := newTestRouter(&fakeUserProvider{err: errors.New("boom")})
		w := doRequest(r, http.MethodPost, "/scim/v2/Users", `{"userName":"alice"}`, nil)

		assertSCIMError(t, w, http.StatusInternalServerError, "")
	})
}

func TestPasswordIsNeverReturned(t *testing.T) {
	stored := User{CommonAttributes: &CommonAttributes{ID: "1"}, UserName: "alice", Password: "hunter2"}

	tests := []struct {
		name   string
		method string
		target string
		body   string
		want   int
	}{
		{"list", http.MethodGet, "/scim/v2/Users", "", http.StatusOK},
		{"get", http.MethodGet, "/scim/v2/Users/1", "", http.StatusOK},
		{"create", http.MethodPost, "/scim/v2/Users", `{"userName":"bob","password":"hunter2"}`, http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &fakeUserProvider{users: []User{stored}, createdID: "2"}
			r := newTestRouter(provider)
			w := doRequest(r, tt.method, tt.target, tt.body, nil)

			assertSCIMResponse(t, w, tt.want)
			if strings.Contains(w.Body.String(), "hunter2") {
				t.Errorf("response contains the password: %s", w.Body.String())
			}
		})
	}

	t.Run("provider still receives the password on create", func(t *testing.T) {
		provider := &fakeUserProvider{createdID: "2"}
		r := newTestRouter(provider)
		doRequest(r, http.MethodPost, "/scim/v2/Users", `{"userName":"bob","password":"hunter2"}`, nil)

		if provider.gotCreate.Password != "hunter2" {
			t.Errorf("provider got password %q, want hunter2", provider.gotCreate.Password)
		}
	})
}

func TestSchemasAreAlwaysReturned(t *testing.T) {
	core := []Schema{SchemaCoreUser}

	t.Run("get fills missing schemas", func(t *testing.T) {
		r := newTestRouter(&fakeUserProvider{users: []User{{CommonAttributes: &CommonAttributes{ID: "1"}}}})
		got := decode[User](t, doRequest(r, http.MethodGet, "/scim/v2/Users/1", "", nil))
		if !reflect.DeepEqual(got.Schemas, core) {
			t.Errorf("schemas = %v, want %v", got.Schemas, core)
		}
	})

	t.Run("list fills missing schemas", func(t *testing.T) {
		r := newTestRouter(&fakeUserProvider{users: []User{{UserName: "a"}}})
		got := decode[ListResponse](t, doRequest(r, http.MethodGet, "/scim/v2/Users", "", nil))
		if !reflect.DeepEqual(got.Resources[0].Schemas, core) {
			t.Errorf("schemas = %v, want %v", got.Resources[0].Schemas, core)
		}
	})

	t.Run("provider schemas are kept", func(t *testing.T) {
		own := []Schema{SchemaCoreUser, "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"}
		r := newTestRouter(&fakeUserProvider{users: []User{{CommonAttributes: &CommonAttributes{ID: "1"}, Schemas: own}}})
		got := decode[User](t, doRequest(r, http.MethodGet, "/scim/v2/Users/1", "", nil))
		if !reflect.DeepEqual(got.Schemas, own) {
			t.Errorf("schemas = %v, want %v", got.Schemas, own)
		}
	})
}

func TestProviderErrorTextIsNotExposed(t *testing.T) {
	const internal = "pq: duplicate key violates constraint users_pkey"

	for _, tt := range []struct{ method, target, body string }{
		{http.MethodGet, "/scim/v2/Users", ""},
		{http.MethodGet, "/scim/v2/Users/1", ""},
		{http.MethodPost, "/scim/v2/Users", `{"userName":"a"}`},
	} {
		t.Run(tt.method+" "+tt.target, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			var logged []string
			r.Use(func(c *gin.Context) {
				c.Next()
				logged = c.Errors.Errors()
			})
			NewHandler(Providers{Users: &fakeUserProvider{err: errors.New(internal)}}, testToken).RegisterRoutes(r)

			w := doRequest(r, tt.method, tt.target, tt.body, nil)

			assertSCIMError(t, w, http.StatusInternalServerError, "")
			if strings.Contains(w.Body.String(), internal) {
				t.Errorf("response exposes the provider error: %s", w.Body.String())
			}
			if len(logged) != 1 || logged[0] != internal {
				t.Errorf("gin context errors = %v, want [%q]", logged, internal)
			}
		})
	}
}
