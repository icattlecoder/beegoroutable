package beegoroutable

import (
	"reflect"
	"testing"
)

type Controller struct {
}

func (c *Controller) Handler() {
}

func TestMappingMethods(t *testing.T) {

	c := Controller{}
	str := MappingMethods(GET(c.Handler), POST(c.Handler), DELETE(c.Handler), PUT(c.Handler))
	expect := "get:Handler;post:Handler;delete:Handler;put:Handler"
	if !reflect.DeepEqual(expect, str) {
		t.Fatalf("expect :%s, but got:%s ", expect, str)
	}
}
