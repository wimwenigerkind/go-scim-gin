package scim

import (
	"reflect"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		want   []Token
	}{
		{
			name:   "string comparison",
			filter: `userName eq "bjensen"`,
			want: []Token{
				{Type: TokenTypeAttrName, Value: "userName"},
				{Type: TokenTypeOperator, Value: "eq", Operator: FilterOperatorEqual},
				{Type: TokenTypeString, Value: "bjensen"},
			},
		},
		{
			name:   "operator is case-insensitive",
			filter: `userName EQ "bjensen"`,
			want: []Token{
				{Type: TokenTypeAttrName, Value: "userName"},
				{Type: TokenTypeOperator, Value: "EQ", Operator: FilterOperatorEqual},
				{Type: TokenTypeString, Value: "bjensen"},
			},
		},
		{
			name:   "quoted value keeps spaces",
			filter: `displayName co "Babs Jensen"`,
			want: []Token{
				{Type: TokenTypeAttrName, Value: "displayName"},
				{Type: TokenTypeOperator, Value: "co", Operator: FilterOperatorContains},
				{Type: TokenTypeString, Value: "Babs Jensen"},
			},
		},
		{
			name:   "number, bool and null literals",
			filter: `a gt 1.5 and b eq true or c eq null`,
			want: []Token{
				{Type: TokenTypeAttrName, Value: "a"},
				{Type: TokenTypeOperator, Value: "gt", Operator: FilterOperatorGreaterThan},
				{Type: TokenTypeNumber, Value: "1.5"},
				{Type: TokenTypeLogical, Value: "and"},
				{Type: TokenTypeAttrName, Value: "b"},
				{Type: TokenTypeOperator, Value: "eq", Operator: FilterOperatorEqual},
				{Type: TokenTypeBool, Value: "true"},
				{Type: TokenTypeLogical, Value: "or"},
				{Type: TokenTypeAttrName, Value: "c"},
				{Type: TokenTypeOperator, Value: "eq", Operator: FilterOperatorEqual},
				{Type: TokenTypeNull, Value: "null"},
			},
		},
		{
			name:   "brackets and parentheses",
			filter: `emails[type eq "work"]`,
			want: []Token{
				{Type: TokenTypeAttrName, Value: "emails"},
				{Type: TokenTypeLBracket, Value: "["},
				{Type: TokenTypeAttrName, Value: "type"},
				{Type: TokenTypeOperator, Value: "eq", Operator: FilterOperatorEqual},
				{Type: TokenTypeString, Value: "work"},
				{Type: TokenTypeRBracket, Value: "]"},
			},
		},
		{
			name:   "attribute path with schema urn",
			filter: `urn:ietf:params:scim:schemas:core:2.0:User:userName pr`,
			want: []Token{
				{Type: TokenTypeAttrName, Value: "urn:ietf:params:scim:schemas:core:2.0:User:userName"},
				{Type: TokenTypeOperator, Value: "pr", Operator: FilterOperatorPresent},
			},
		},
		{
			name:   "escaped quote and backslash",
			filter: `x eq "a\"b\\"`,
			want: []Token{
				{Type: TokenTypeAttrName, Value: "x"},
				{Type: TokenTypeOperator, Value: "eq", Operator: FilterOperatorEqual},
				{Type: TokenTypeString, Value: `a"b\`},
			},
		},
		{
			name:   "unicode escape",
			filter: `x eq "caf\u00e9"`,
			want: []Token{
				{Type: TokenTypeAttrName, Value: "x"},
				{Type: TokenTypeOperator, Value: "eq", Operator: FilterOperatorEqual},
				{Type: TokenTypeString, Value: "café"},
			},
		},
		{
			name:   "negative number with exponent",
			filter: `x gt -1.5e3`,
			want: []Token{
				{Type: TokenTypeAttrName, Value: "x"},
				{Type: TokenTypeOperator, Value: "gt", Operator: FilterOperatorGreaterThan},
				{Type: TokenTypeNumber, Value: "-1.5e3"},
			},
		},
		{
			name:   "empty filter",
			filter: "",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenize(tt.filter)
			if err != nil {
				t.Fatalf("tokenize(%q) unexpected error: %v", tt.filter, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tokenize(%q)\n got  %+v\n want %+v", tt.filter, got, tt.want)
			}
		})
	}
}

func TestTokenizeErrors(t *testing.T) {
	tests := []struct {
		name    string
		filter  string
		wantErr string
	}{
		{"unterminated string", `userName eq "bjen`, "unterminated string starting at position 12"},
		{"lone quote", `"`, "unterminated string starting at position 0"},
		{"unexpected character", `userName eq "a" & title pr`, `unexpected character '&' at position 16`},
		{"tab is not whitespace", "userName\teq \"a\"", `unexpected character '\t' at position 8`},
		{"invalid escape", `x eq "\x"`, `invalid string at position 5: `},
		{"leading zero", `x eq 01`, `invalid number "01" at position 5`},
		{"two decimal points", `x eq 1.2.3`, `invalid number "1.2.3" at position 5`},
		{"attribute starting with underscore", `_x pr`, `unexpected character '_' at position 0`},
		{"attribute starting with dot", `.x pr`, `unexpected character '.' at position 0`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenize(tt.filter)
			// Prefix match, so the encoding/json message behind ": " can change.
			if err == nil || !strings.HasPrefix(err.Error(), tt.wantErr) {
				t.Fatalf("tokenize(%q) = (%+v, %v), want error starting with %q", tt.filter, got, err, tt.wantErr)
			}
			if got != nil {
				t.Errorf("tokenize(%q) returned tokens %+v alongside the error", tt.filter, got)
			}
		})
	}
}

func TestParseFilter(t *testing.T) {
	tests := []struct {
		name    string
		filter  string
		want    Filter
		wantErr string
	}{
		{
			name:   "equal",
			filter: `userName eq "bjensen"`,
			want:   Filter{Field: "userName", Operator: FilterOperatorEqual, Value: "bjensen"},
		},
		{
			name:   "present",
			filter: `title pr`,
			want:   Filter{Field: "title", Operator: FilterOperatorPresent},
		},
		{
			name:   "uppercase operator is normalized",
			filter: `userName SW "b"`,
			want:   Filter{Field: "userName", Operator: FilterOperatorStartsWith, Value: "b"},
		},
		{
			name:   "number value",
			filter: `age gt 21`,
			want:   Filter{Field: "age", Operator: FilterOperatorGreaterThan, Value: "21"},
		},
		{
			name:    "two tokens without pr",
			filter:  `userName eq`,
			wantErr: "invalid 2-token filter",
		},
		{
			name:    "operator in wrong position",
			filter:  `userName "bjensen" eq`,
			wantErr: "expected operator",
		},
		{
			name:    "compound filter",
			filter:  `userName eq "a" and title pr`,
			wantErr: "compound filters not supported",
		},
		{
			name:    "empty filter",
			filter:  "",
			wantErr: "empty filter",
		},
		{
			name:    "whitespace-only filter",
			filter:  "   ",
			wantErr: "empty filter",
		},
		{
			name:    "attribute without operator",
			filter:  "userName",
			wantErr: "missing operator after userName",
		},
		{
			name:    "tokenizer error is returned",
			filter:  `userName eq "bjen`,
			wantErr: "unterminated string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFilter(tt.filter)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("parseFilter(%q) error = %v, want error containing %q", tt.filter, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFilter(%q) unexpected error: %v", tt.filter, err)
			}
			if got != tt.want {
				t.Errorf("parseFilter(%q) = %+v, want %+v", tt.filter, got, tt.want)
			}
		})
	}
}

type columnBase struct {
	ExternalID string `scim:"externalId"`
}

type columnModel struct {
	columnBase
	ID          string `scim:"id"`
	UserName    string `scim:"userName"`
	DisplayName string `scim:"displayName,column:name"`
	Internal    string
}

type columnModelWithPointerEmbed struct {
	*columnBase
	UserName string `scim:"userName"`
}

func TestFilterToColumn(t *testing.T) {
	tests := []struct {
		name      string
		modelType reflect.Type
		attr      string
		want      string
		wantOK    bool
	}{
		{"derived from field name", reflect.TypeFor[columnModel](), "userName", "user_name", true},
		{"acronym field name", reflect.TypeFor[columnModel](), "id", "id", true},
		{"explicit column override", reflect.TypeFor[columnModel](), "displayName", "name", true},
		{"embedded struct", reflect.TypeFor[columnModel](), "externalId", "external_id", true},
		{"embedded pointer struct", reflect.TypeFor[columnModelWithPointerEmbed](), "externalId", "external_id", true},
		{"pointer to model", reflect.TypeFor[*columnModel](), "userName", "user_name", true},
		{"unknown attribute", reflect.TypeFor[columnModel](), "nickName", "", false},
		{"untagged field is not matched", reflect.TypeFor[columnModel](), "Internal", "", false},
		{"non-struct type", reflect.TypeFor[string](), "userName", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FilterToColumn(tt.modelType, tt.attr)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("FilterToColumn(%v, %q) = (%q, %v), want (%q, %v)",
					tt.modelType, tt.attr, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := map[string]string{
		"UserName":    "user_name",
		"ExternalID":  "external_id",
		"ID":          "id",
		"HTTPServer":  "http_server",
		"DisplayName": "display_name",
		"name":        "name",
	}

	for in, want := range tests {
		if got := toSnakeCase(in); got != want {
			t.Errorf("toSnakeCase(%q) = %q, want %q", in, got, want)
		}
	}
}
