package expressions

type labelReplaceFn func(expr, ns string) (string, error)

var ReplaceNamespaceFns = make(map[string]labelReplaceFn)

func register(name string, fn labelReplaceFn) {
	ReplaceNamespaceFns[name] = fn
}
