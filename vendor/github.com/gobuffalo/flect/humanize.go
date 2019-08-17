package flect

import (
	"strings"
)

// Humanize returns first letter of sentence capitalized
// employee_salary = Employee salary
// employee_id = employee ID
// employee_mobile_number = Employee mobile number
func Humanize(s string) string {
	return New(s).Humanize().String()
}

// Humanize First letter of sentence capitalized
func (i Ident) Humanize() Ident {
	if len(i.Original) == 0 {
		return New("")
	}

	var parts []string
	for index, part := range i.Parts {
		if index == 0 {
			part = strings.Title(i.Parts[0])
		}

		parts = xappend(parts, part)
	}

	return New(strings.Join(parts, " "))
}
