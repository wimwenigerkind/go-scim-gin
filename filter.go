package scim

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Filter https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
type Filter struct {
	// https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	// Field keeps the client's spelling, but attribute names are case
	// insensitive: compare it with strings.EqualFold or use FilterToColumn.
	Field string

	// https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	Operator FilterOperator

	// https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2
	// Value holds the decoded literal; ValueType says which JSON type it had,
	// so `active eq true` and `active eq "true"` can be told apart.
	Value string

	ValueType FilterValueType
}

// FilterValueType is the JSON type of a filter's comparison value.
type FilterValueType string

const (
	FilterValueString  FilterValueType = "string"
	FilterValueNumber  FilterValueType = "number"
	FilterValueBoolean FilterValueType = "boolean"
	FilterValueNull    FilterValueType = "null"
)

var filterValueTypes = map[TokenType]FilterValueType{
	TokenTypeString: FilterValueString,
	TokenTypeNumber: FilterValueNumber,
	TokenTypeBool:   FilterValueBoolean,
	TokenTypeNull:   FilterValueNull,
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

// parseFilter parses raw for a resource of the given schema. A fully
// qualified attribute of that schema is reduced to its short name, since
// https://www.rfc-editor.org/rfc/rfc7644#section-3.10 allows both forms.
func parseFilter(raw string, schema Schema) (Filter, error) {
	tokens, err := tokenize(raw)
	if err != nil {
		return Filter{}, err
	}
	f, err := parse(tokens)
	if err != nil {
		return Filter{}, err
	}
	prefix := string(schema) + ":"
	if len(f.Field) > len(prefix) && strings.EqualFold(f.Field[:len(prefix)], prefix) {
		f.Field = f.Field[len(prefix):]
	}
	return f, nil
}

// parse accepts the two single-expression forms of the ABNF:
// attrPath SP "pr" and attrPath SP compareOp SP compValue.
func parse(tokens []Token) (Filter, error) {
	if len(tokens) == 0 {
		return Filter{}, fmt.Errorf("empty filter")
	}
	if tokens[0].Type != TokenTypeAttrName {
		return Filter{}, fmt.Errorf("expected attribute name, got %q", tokens[0].Value)
	}
	field := tokens[0].Value

	switch len(tokens) {
	case 1:
		return Filter{}, fmt.Errorf("missing operator after %s", field)
	case 2:
		if tokens[1].Operator != FilterOperatorPresent {
			return Filter{}, fmt.Errorf("expected pr or an operator with a value after %s, got %q", field, tokens[1].Value)
		}
		return Filter{Field: field, Operator: FilterOperatorPresent}, nil
	case 3:
		op := tokens[1].Operator
		if tokens[1].Type != TokenTypeOperator || op == FilterOperatorPresent {
			return Filter{}, fmt.Errorf("expected comparison operator after %s, got %q", field, tokens[1].Value)
		}
		valueType, ok := filterValueTypes[tokens[2].Type]
		if !ok {
			return Filter{}, fmt.Errorf("expected a string, number, true, false or null after %s, got %q", op, tokens[2].Value)
		}
		// https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2: gt, ge, lt
		// and le on boolean values SHALL fail with invalidFilter.
		switch op {
		case FilterOperatorGreaterThan, FilterOperatorGreaterThanOrEqual,
			FilterOperatorLessThan, FilterOperatorLessThanOrEqual:
			if valueType == FilterValueBoolean || valueType == FilterValueNull {
				return Filter{}, fmt.Errorf("operator %s cannot compare %s values", op, valueType)
			}
		}
		return Filter{Field: field, Operator: op, Value: tokens[2].Value, ValueType: valueType}, nil
	default:
		return Filter{}, fmt.Errorf("compound filters not supported (got %d tokens)", len(tokens))
	}
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
		if !strings.EqualFold(attr, scimAttr) {
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

func tokenize(filter string) ([]Token, error) {
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

		// https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2: string and
		// number values follow the JSON grammar, so encoding/json decodes and
		// validates them once their end is found.
		if c == '"' {
			start := i
			for i++; i < len(filter) && filter[i] != '"'; i++ {
				if filter[i] == '\\' {
					i++
				}
			}
			if i >= len(filter) {
				return nil, fmt.Errorf("unterminated string starting at position %d", start)
			}
			var value string
			if err := json.Unmarshal([]byte(filter[start:i+1]), &value); err != nil {
				return nil, fmt.Errorf("invalid string at position %d: %w", start, err)
			}
			tokens = append(tokens, Token{Type: TokenTypeString, Value: value})
			continue
		}

		if c == '-' || (c >= '0' && c <= '9') {
			start := i
			for i < len(filter) && strings.IndexByte("0123456789.eE+-", filter[i]) >= 0 {
				i++
			}
			number := filter[start:i]
			if !json.Valid([]byte(number)) {
				return nil, fmt.Errorf("invalid number %q at position %d", number, start)
			}
			tokens = append(tokens, Token{Type: TokenTypeNumber, Value: number})
			i--
			continue
		}

		// https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2.2:
		// ATTRNAME = ALPHA *(nameChar)
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			start := i
			for i < len(filter) && isWordChar(filter[i]) {
				i++
			}
			tokens = append(tokens, classify(filter[start:i]))
			i--
			continue
		}

		return nil, fmt.Errorf("unexpected character %q at position %d", c, i)
	}

	return tokens, nil
}
