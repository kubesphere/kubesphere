package idutils

import (
	"fmt"
	"sort"
	"testing"
)

func TestGetUuid(t *testing.T) {
	fmt.Println(GetUuid(""))
}

func TestGetUuid36(t *testing.T) {
	fmt.Println(GetUuid36(""))
}

func TestGetManyUuid(t *testing.T) {
	var strSlice []string
	for i := 0; i < 10000; i++ {
		testId := GetUuid("")
		strSlice = append(strSlice, testId)
	}
	sort.Strings(strSlice)
}
