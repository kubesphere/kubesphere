package resources

import (
	"testing"

	"github.com/xenolf/lego/log"
)

func TestConditions(t *testing.T) {
	log.Println(parseToConditions("label~xxx"))
}
