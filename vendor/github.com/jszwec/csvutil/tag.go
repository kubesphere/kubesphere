package csvutil

import (
	"reflect"
	"strings"
)

type tag struct {
	name      string
	prefix    string
	empty     bool
	omitEmpty bool
	ignore    bool
	inline    bool
}

func parseTag(tagname string, field reflect.StructField) (t tag) {
	tags := strings.Split(field.Tag.Get(tagname), ",")
	if len(tags) == 1 && tags[0] == "" {
		t.name = field.Name
		t.empty = true
		return
	}

	switch tags[0] {
	case "-":
		t.ignore = true
		return
	case "":
		t.name = field.Name
	default:
		t.name = tags[0]
	}

	for _, tagOpt := range tags[1:] {
		switch tagOpt {
		case "omitempty":
			t.omitEmpty = true
		case "inline":
			if walkType(field.Type).Kind() == reflect.Struct {
				t.inline = true
				t.prefix = tags[0]
			}
		}
	}
	return
}
