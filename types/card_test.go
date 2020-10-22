package types

import (
	"fmt"
	"testing"
)

func TestCardStringer(t *testing.T) {
	c := Card(12345)

	if s := fmt.Sprintf("%v", c); s != "12345" {
		t.Errorf("Expected: %v, got:%v", "12345", s)
	}
}

func TestCardStringerWithPointer(t *testing.T) {
	c := Card(12345)
	p := &c

	if s := fmt.Sprintf("%v", p); s != "12345" {
		t.Errorf("Expected: %v, got:%v", "12345", s)
	}
}

func TestCardStringerWithNil(t *testing.T) {
	var c *Card

	if s := fmt.Sprintf("%v", c); s != "" {
		t.Errorf("Expected: %v, got:%v", "", s)
	}
}
