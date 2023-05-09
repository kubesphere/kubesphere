package restfulspec

import "strings"

// DefaultNameHandler GoRestfulDefinition -> GoRestfulDefinition (not changed)
func DefaultNameHandler(name string) string {
	return name
}

// LowerSnakeCasedNameHandler GoRestfulDefinition -> go_restful_definition
func LowerSnakeCasedNameHandler(name string) string {
	definitionName := make([]byte, 0, len(name)+1)
	for i := 0; i < len(name); i++ {
		c := name[i]
		if isUpper(c) {
			if i > 0 {
				definitionName = append(definitionName, '_')
			}
			c += 'a' - 'A'
		}
		definitionName = append(definitionName, c)
	}

	return string(definitionName)
}

// LowerCamelCasedNameHandler GoRestfulDefinition -> goRestfulDefinition
func LowerCamelCasedNameHandler(name string) string {
	definitionName := make([]byte, 0, len(name)+1)
	for i := 0; i < len(name); i++ {
		c := name[i]
		if isUpper(c) && i == 0 {
			c += 'a' - 'A'
		}
		definitionName = append(definitionName, c)
	}

	return string(definitionName)
}

// GoLowerCamelCasedNameHandler HTTPRestfulDefinition -> httpRestfulDefinition
func GoLowerCamelCasedNameHandler(name string) string {
	var i = 0
	// for continuous Upper letters, check whether is it a common Initialisms
	for ; i < len(name) && isUpper(name[i]); i++ {
	}
	if len(name) != i && i > 1 {
		i-- // for continuous Upper letters, the last Upper is should not be check, eg: S for HTTPStatus
	}
	for ; i > 1; i-- {
		if _, ok := commonInitialisms[name[:i]]; ok {
			break
		}
	}

	return strings.ToLower(name[:i]) + name[i:]
}

// commonInitialisms is a set of common initialisms. (from https://github.com/golang/lint/blob/master/lint.go)
// Only add entries that are highly unlikely to be non-initialisms.
// For instance, "ID" is fine (Freudian code is rare), but "AND" is not.
var commonInitialisms = map[string]bool{
	"ACL":   true,
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XMPP":  true,
	"XSRF":  true,
	"XSS":   true,
}

func isUpper(r uint8) bool {
	return 'A' <= r && r <= 'Z'
}
