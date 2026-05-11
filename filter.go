package scim

import (
	"fmt"
	"reflect"
	"strings"
)

// Filter https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
type Filter struct {
	// https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	Field string

	// https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	Operator FilterOperator

	// https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	Value string
}

// FilterOperator https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
type FilterOperator string

// https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
const (
	// FilterOperatorEqual https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	FilterOperatorEqual FilterOperator = "eq"

	// FilterOperatorNotEqual https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	FilterOperatorNotEqual FilterOperator = "ne"

	// FilterOperatorContains https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	FilterOperatorContains FilterOperator = "co"

	// FilterOperatorStartsWith https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	FilterOperatorStartsWith FilterOperator = "sw"

	// FilterOperatorEndsWith https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	FilterOperatorEndsWith FilterOperator = "ew"

	// FilterOperatorPresent https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	FilterOperatorPresent FilterOperator = "pr"

	// FilterOperatorGreaterThan https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	FilterOperatorGreaterThan FilterOperator = "gt"

	// FilterOperatorGreaterThanOrEqual https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	FilterOperatorGreaterThanOrEqual FilterOperator = "ge"

	// FilterOperatorLessThan https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	FilterOperatorLessThan FilterOperator = "lt"

	// FilterOperatorLessThanOrEqual https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	FilterOperatorLessThanOrEqual FilterOperator = "le"
)

// TokenType https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
type TokenType string

const (
	// TokenTypeAttrName https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	TokenTypeAttrName TokenType = "ATTR_NAME"

	// TokenTypeOperator https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	TokenTypeOperator TokenType = "OPERATOR"

	// TokenTypeLogical https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	TokenTypeLogical TokenType = "LOGICAL"

	// TokenTypeString https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	TokenTypeString TokenType = "STRING"

	// TokenTypeNumber https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	TokenTypeNumber TokenType = "NUMBER"

	// TokenTypeBool https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	TokenTypeBool TokenType = "BOOL"

	// TokenTypeNull https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	TokenTypeNull TokenType = "NULL"

	// TokenTypeLParen https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	TokenTypeLParen TokenType = "("

	// TokenTypeRParen https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	TokenTypeRParen TokenType = ")"

	// TokenTypeLBracket https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	TokenTypeLBracket TokenType = "["

	// TokenTypeRBracket https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	TokenTypeRBracket TokenType = "]"
)

type Token struct {
	Type     TokenType
	Value    string
	Operator FilterOperator
}

func parse(tokens []Token) (Filter, error) {
	switch len(tokens) {
	case 2:
		if tokens[1].Operator == FilterOperatorPresent {
			return Filter{Field: tokens[0].Value, Operator: FilterOperatorPresent}, nil
		}
		return Filter{}, fmt.Errorf("invalid 2-token filter: %s %s", tokens[0].Value, tokens[1].Value)
	case 3:
		if tokens[1].Type != TokenTypeOperator {
			return Filter{}, fmt.Errorf("expected operator at position 2, got %s", tokens[1].Type)
		}
		return Filter{
			Field:    tokens[0].Value,
			Operator: tokens[1].Operator,
			Value:    tokens[2].Value,
		}, nil
	default:
		return Filter{}, fmt.Errorf("compound filters not supported (got %d tokens)", len(tokens))
	}
}

func getAttribute(resource any, field string) string {
	v := reflect.ValueOf(resource)
	t := reflect.TypeOf(resource)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ""
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if f.Anonymous {
			fv := v.Field(i)
			if fv.Kind() == reflect.Ptr && fv.IsNil() {
				continue
			}
			if val := getAttribute(fv.Interface(), field); val != "" {
				return val
			}
			continue
		}

		tag := f.Tag.Get("scim")
		if tag == "" {
			continue
		}
		attr, _ := parseScimTag(tag)
		if attr == field {
			return fmt.Sprintf("%v", v.Field(i).Interface())
		}
	}
	return ""
}

func evaluate(filter Filter, resource any) bool {
	val := getAttribute(resource, filter.Field)

	switch filter.Operator {
	case FilterOperatorEqual:
		return val == filter.Value
	case FilterOperatorNotEqual:
		return val != filter.Value
	case FilterOperatorContains:
		return strings.Contains(val, filter.Value)
	case FilterOperatorStartsWith:
		return strings.HasPrefix(val, filter.Value)
	case FilterOperatorEndsWith:
		return strings.HasSuffix(val, filter.Value)
	case FilterOperatorPresent:
		return val != ""
	}

	return false
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == '.' || c == ':'
}

func classify(word string) Token {
	switch FilterOperator(strings.ToLower(word)) {
	case FilterOperatorEqual,
		FilterOperatorNotEqual,
		FilterOperatorContains,
		FilterOperatorStartsWith,
		FilterOperatorEndsWith,
		FilterOperatorPresent,
		FilterOperatorGreaterThan,
		FilterOperatorGreaterThanOrEqual,
		FilterOperatorLessThan,
		FilterOperatorLessThanOrEqual:
		return Token{
			Type:     TokenTypeOperator,
			Value:    word,
			Operator: FilterOperator(strings.ToLower(word)),
		}
	case "and", "or", "not":
		return Token{Type: TokenTypeLogical, Value: strings.ToLower(word)}
	case "true", "false":
		return Token{Type: TokenTypeBool, Value: strings.ToLower(word)}
	case "null":
		return Token{Type: TokenTypeNull, Value: "null"}
	default:
		return Token{Type: TokenTypeAttrName, Value: word}
	}
}

// FilterToColumn
// Tag format:
//
//	`scim:"<attr>"`              column = snake_case of the Go field name
//	`scim:"<attr>,column:<col>"` column = <col> (explicit override)
func FilterToColumn(modelType reflect.Type, scimAttr string) (string, bool) {
	if modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return "", false
	}

	for i := 0; i < modelType.NumField(); i++ {
		f := modelType.Field(i)

		if f.Anonymous {
			ft := f.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if col, ok := FilterToColumn(ft, scimAttr); ok {
				return col, true
			}
			continue
		}

		tag := f.Tag.Get("scim")
		if tag == "" {
			continue
		}
		attr, column := parseScimTag(tag)
		if attr != scimAttr {
			continue
		}
		if column != "" {
			return column, true
		}
		return toSnakeCase(f.Name), true
	}
	return "", false
}

func parseScimTag(tag string) (attr, column string) {
	parts := strings.Split(tag, ",")
	attr = parts[0]
	for _, p := range parts[1:] {
		if strings.HasPrefix(p, "column:") {
			column = strings.TrimPrefix(p, "column:")
		}
	}
	return attr, column
}

func toSnakeCase(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := runes[i-1]
				prevLower := prev >= 'a' && prev <= 'z'
				nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
				if prevLower || nextLower {
					b.WriteRune('_')
				}
			}
			b.WriteRune(r + ('a' - 'A'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func tokenize(filter string) []Token {
	var tokens []Token

	for i := 0; i < len(filter); i++ {
		c := filter[i]

		if c == ' ' {
			continue
		}

		switch c {
		case '(':
			tokens = append(tokens, Token{Type: TokenTypeLParen, Value: "("})
			continue
		case ')':
			tokens = append(tokens, Token{Type: TokenTypeRParen, Value: ")"})
			continue
		case '[':
			tokens = append(tokens, Token{Type: TokenTypeLBracket, Value: "["})
			continue
		case ']':
			tokens = append(tokens, Token{Type: TokenTypeRBracket, Value: "]"})
			continue
		}

		if c == '"' {
			i++
			start := i
			for i < len(filter) && filter[i] != '"' {
				i++
			}
			tokens = append(tokens, Token{Type: TokenTypeString, Value: filter[start:i]})
			continue
		}

		if c >= '0' && c <= '9' {
			start := i
			for i < len(filter) && (filter[i] >= '0' && filter[i] <= '9' || filter[i] == '.') {
				i++
			}
			tokens = append(tokens, Token{Type: TokenTypeNumber, Value: filter[start:i]})
			i--
			continue
		}

		if isWordChar(c) {
			start := i
			for i < len(filter) && isWordChar(filter[i]) {
				i++
			}
			tokens = append(tokens, classify(filter[start:i]))
			i--
			continue
		}
	}

	return tokens
}
