package main

import (
	"context"
	"fmt"
	"net/http"

	partial "github.com/partial-coffee/go-partial"
)

func main() {
	// Create a new service
	svc := partial.NewService(&partial.Config{})

	// Define in-memory templates
	fsys := &partial.InMemoryFS{
		Files: map[string]string{
			"templates/parent.html": `<div class="parent">
  <h1>{{ .Data.Title }}</h1>
  <div class="items">
    {{ range .Data.Items }}
      {{ child "item" "ItemName" . }}
    {{ end }}
  </div>
</div>`,
			"templates/item.html": `<div class="item">
  <p>Item: {{ .Data.ItemName }}</p>
  <p>From list: {{ .Parent.Title }}</p>
</div>`,
		},
	}

	// Create the item partial
	itemPartial := partial.New("templates/item.html").ID("item")

	// Create the parent partial with data
	parentPartial := partial.New("templates/parent.html").ID("parent")
	parentPartial.SetData(map[string]any{
		"Title": "My Shopping List",
		"Items": []string{"Apples", "Bananas", "Oranges"},
	})
	parentPartial.With(itemPartial)

	// Render the partial
	layout := svc.NewLayout().FS(fsys).Set(parentPartial)
	req, _ := http.NewRequest("GET", "/", nil)
	html, err := layout.RenderWithRequest(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(html)
}
