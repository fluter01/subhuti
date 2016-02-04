// Copyright 2016 Alex Fluter

package bot

import (
	"github.com/mvdan/xurls"
	"testing"
)

func TestURL1(t *testing.T) {
	s := "https://golang.org/ref/spec#For_statements"

	var b bool
	var m []string
	b = xurls.Strict.MatchString(s)
	if !b {
		t.Log("not match")
		t.Fail()
	}

	m = xurls.Strict.FindAllString(s, -1)
	for i, e := range m {
		t.Logf("%d: %s", i, e)
	}
	if m == nil {
		t.Log("does not contain url")
		t.Fail()
	}
}
