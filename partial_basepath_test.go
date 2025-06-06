package partial_test

import (
	. "github.com/partial-coffee/go-partial"
	"testing"
)

func TestGetBasePath_Simple(t *testing.T) {
	p := New()
	p.SetBasePath("/foo")
	if got := p.GetBasePath(); got != "/foo" {
		t.Errorf("expected /foo, got %q", got)
	}
}

func TestGetBasePath_ParentFallback(t *testing.T) {
	parent := New()
	parent.SetBasePath("/parent")
	child := New().SetParent(parent)
	if got := child.GetBasePath(); got != "/parent" {
		t.Errorf("expected /parent, got %q", got)
	}
}

func TestGetBasePath_ParentChain(t *testing.T) {
	grandparent := New()
	grandparent.SetBasePath("/grand")
	parent := New().SetParent(grandparent)
	child := New().SetParent(parent)
	if got := child.GetBasePath(); got != "/grand" {
		t.Errorf("expected /grand, got %q", got)
	}
}

func TestGetBasePath_Override(t *testing.T) {
	parent := New()
	parent.SetBasePath("/parent")
	child := New().SetParent(parent)
	child.SetBasePath("/child")
	if got := child.GetBasePath(); got != "/child" {
		t.Errorf("expected /child, got %q", got)
	}
}

func TestGetBasePath_Empty(t *testing.T) {
	p := New()
	if got := p.GetBasePath(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
