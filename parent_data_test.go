package partial

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParentDataInjection(t *testing.T) {
	svc := NewService(&Config{})

	var handleRequest = func(w http.ResponseWriter, r *http.Request) {
		// Define in-memory templates for parent and child
		fsys := &InMemoryFS{
			Files: map[string]string{
				"templates/parent.html": `<div>Parent: {{ .Data.ParentName }} {{ child "child" }}</div>`,
				"templates/child.html":  `<div>Child: {{ .Data.ChildName }}, Parent's Name: {{ .Parent.ParentName }}</div>`,
			},
		}

		// Create the child partial
		childPartial := New("templates/child.html").ID("child")
		childPartial.SetData(map[string]any{
			"ChildName": "Child1",
		})

		// Create the parent partial
		parentPartial := New("templates/parent.html").ID("parent")
		parentPartial.SetData(map[string]any{
			"ParentName": "Parent1",
		})
		parentPartial.With(childPartial)

		// Render the parent partial
		out, err := svc.NewLayout().FS(fsys).Set(parentPartial).RenderWithRequest(r.Context(), r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(out))
	}

	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handleRequest(response, request)

	expected := `<div>Parent: Parent1 <div>Child: Child1, Parent's Name: Parent1</div></div>`

	if strings.TrimSpace(response.Body.String()) != expected {
		t.Errorf("expected %s, got %s", expected, response.Body.String())
	}
}

func TestParentDataWithNestedChildren(t *testing.T) {
	svc := NewService(&Config{})

	var handleRequest = func(w http.ResponseWriter, r *http.Request) {
		// Define in-memory templates for grandparent, parent, and child
		fsys := &InMemoryFS{
			Files: map[string]string{
				"templates/grandparent.html": `<div>GP: {{ .Data.Name }} {{ child "parent" }}</div>`,
				"templates/parent.html":      `<div>P: {{ .Data.Name }}, GP: {{ .Parent.Name }} {{ child "child" }}</div>`,
				"templates/child.html":       `<div>C: {{ .Data.Name }}, P: {{ .Parent.Name }}</div>`,
			},
		}

		// Create the child partial
		childPartial := New("templates/child.html").ID("child")
		childPartial.SetData(map[string]any{
			"Name": "Child",
		})

		// Create the parent partial
		parentPartial := New("templates/parent.html").ID("parent")
		parentPartial.SetData(map[string]any{
			"Name": "Parent",
		})
		parentPartial.With(childPartial)

		// Create the grandparent partial
		grandparentPartial := New("templates/grandparent.html").ID("grandparent")
		grandparentPartial.SetData(map[string]any{
			"Name": "Grandparent",
		})
		grandparentPartial.With(parentPartial)

		// Render the grandparent partial
		out, err := svc.NewLayout().FS(fsys).Set(grandparentPartial).RenderWithRequest(r.Context(), r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(out))
	}

	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handleRequest(response, request)

	expected := `<div>GP: Grandparent <div>P: Parent, GP: Grandparent <div>C: Child, P: Parent</div></div></div>`

	if strings.TrimSpace(response.Body.String()) != expected {
		t.Errorf("expected %s, got %s", expected, response.Body.String())
	}
}

func TestParentDataWithChildFunc(t *testing.T) {
	svc := NewService(&Config{})

	var handleRequest = func(w http.ResponseWriter, r *http.Request) {
		// Define in-memory templates
		fsys := &InMemoryFS{
			Files: map[string]string{
				"templates/parent.html": `<div>{{ .Data.Title }} {{ child "item" "ItemID" "123" }}</div>`,
				"templates/item.html":   `<div>Item: {{ .Data.ItemID }}, Title: {{ .Parent.Title }}</div>`,
			},
		}

		// Create the item partial
		itemPartial := New("templates/item.html").ID("item")

		// Create the parent partial
		parentPartial := New("templates/parent.html").ID("parent")
		parentPartial.SetData(map[string]any{
			"Title": "My List",
		})
		parentPartial.With(itemPartial)

		// Render the parent partial
		out, err := svc.NewLayout().FS(fsys).Set(parentPartial).RenderWithRequest(r.Context(), r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(out))
	}

	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handleRequest(response, request)

	expected := `<div>My List <div>Item: 123, Title: My List</div></div>`

	if strings.TrimSpace(response.Body.String()) != expected {
		t.Errorf("expected %s, got %s", expected, response.Body.String())
	}
}

func TestParentDataIsNilWhenNoParent(t *testing.T) {
	svc := NewService(&Config{})

	var handleRequest = func(w http.ResponseWriter, r *http.Request) {
		// Define in-memory template
		fsys := &InMemoryFS{
			Files: map[string]string{
				"templates/root.html": `<div>Root{{ if .Parent }}, Parent: {{ .Parent.Name }}{{ end }}</div>`,
			},
		}

		// Create a root partial without a parent
		rootPartial := New("templates/root.html").ID("root")
		rootPartial.SetData(map[string]any{
			"Name": "Root",
		})

		// Render the root partial
		out, err := svc.NewLayout().FS(fsys).Set(rootPartial).RenderWithRequest(r.Context(), r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(out))
	}

	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handleRequest(response, request)

	expected := `<div>Root</div>`

	if strings.TrimSpace(response.Body.String()) != expected {
		t.Errorf("expected %s, got %s", expected, response.Body.String())
	}
}
