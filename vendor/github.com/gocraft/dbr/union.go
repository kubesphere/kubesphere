package dbr

type union struct {
	builder []Builder
	all     bool
}

func Union(builder ...Builder) interface {
	Builder
	As(string) Builder
} {
	return &union{
		builder: builder,
	}
}

func UnionAll(builder ...Builder) interface {
	Builder
	As(string) Builder
} {
	return &union{
		builder: builder,
		all:     true,
	}
}

func (u *union) Build(d Dialect, buf Buffer) error {
	for i, b := range u.builder {
		if i > 0 {
			buf.WriteString(" UNION ")
			if u.all {
				buf.WriteString("ALL ")
			}
		}
		buf.WriteString(placeholder)
		buf.WriteValue(b)
	}
	return nil
}

func (u *union) As(alias string) Builder {
	return as(u, alias)
}
