package query

import (
	"fmt"
	"testing"
)

func TestQueryBuilder(t *testing.T) {
	testCase := func() *Query {
		var mini int32 = 1
		aaa := NewTerms("aaa", []string{})
		b := NewBool()
		b.AppendFilter(NewBool().
			AppendShould(aaa).
			WithMinimumShouldMatch(mini))

		return NewQuery().WithBool(b)
	}

	b := NewBuilder().
		WithQuery(testCase())

	fmt.Printf("aaaaaa: %+v\n", b)
	_, err := b.Bytes()
	if err != nil {
		t.Fatalf("err jsoniter.Marshal: %v", err)
	}
}
