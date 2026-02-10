package main

import (
	"context"
	"fmt"
	"net/http"

	partial "github.com/partial-coffee/go-partial"
)

// Example: Table with sortable columns
// The table header uses parent data to determine which column is currently sorted
func main() {
	svc := partial.NewService(&partial.Config{})

	fsys := &partial.InMemoryFS{
		Files: map[string]string{
			"table.html": `<table class="data-table">
  <thead>
    <tr>{{ child "header" }}</tr>
  </thead>
  <tbody>
    {{ range .Data.Rows }}
      {{ child "row" "Name" .Name "Age" .Age "Email" .Email }}
    {{ end }}
  </tbody>
</table>`,
			"header.html": `{{ range .Parent.Columns }}
      <th class="{{ if eq . $.Parent.SortColumn }}sorted{{ end }}">
        {{ . }}
        {{ if eq . $.Parent.SortColumn }}
          {{ if eq $.Parent.SortOrder "asc" }}↑{{ else }}↓{{ end }}
        {{ end }}
      </th>
    {{ end }}`,
			"row.html": `<tr>
      <td>{{ .Data.Name }}</td>
      <td>{{ .Data.Age }}</td>
      <td>{{ .Data.Email }}</td>
    </tr>`,
		},
	}

	// Create header partial
	headerPartial := partial.New("header.html").ID("header")

	// Create row partial
	rowPartial := partial.New("row.html").ID("row")

	// Create table partial with configuration
	tablePartial := partial.New("table.html").ID("table")
	tablePartial.SetData(map[string]any{
		"Columns":    []string{"Name", "Age", "Email"},
		"SortColumn": "Name",
		"SortOrder":  "asc",
		"Rows": []map[string]any{
			{"Name": "Alice", "Age": 30, "Email": "alice@example.com"},
			{"Name": "Bob", "Age": 25, "Email": "bob@example.com"},
			{"Name": "Charlie", "Age": 35, "Email": "charlie@example.com"},
		},
	})
	tablePartial.With(headerPartial)
	tablePartial.With(rowPartial)

	// Render
	layout := svc.NewLayout().FS(fsys).Set(tablePartial)
	req, _ := http.NewRequest("GET", "/", nil)
	html, err := layout.RenderWithRequest(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(html)
	fmt.Println("\nNotice how the header partial uses .Parent.Columns, .Parent.SortColumn,")
	fmt.Println("and .Parent.SortOrder without them being explicitly passed!")
}
