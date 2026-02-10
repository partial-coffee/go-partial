# Parent Data Injection

## Overview

When a child partial is added to a parent partial using the `With()` method, the parent's data is automatically injected into the child's template context under the `.Parent` key. This allows child partials to access their parent's data during rendering.

## Usage

### Basic Example

```go
// Create parent partial with data
parentPartial := partial.New("templates/parent.html").ID("parent")
parentPartial.SetData(map[string]any{
    "Title": "My List",
    "Count": 5,
})

// Create child partial
childPartial := partial.New("templates/child.html").ID("child")
childPartial.SetData(map[string]any{
    "ItemName": "Item 1",
})

// Add child to parent
parentPartial.With(childPartial)
```

### Template Access

In the child template, you can access:
- `.Data` - The child's own data
- `.Parent` - The parent's data (when rendered as a child)
- `.Global` - Merged data from all ancestors
- `.Service` - Service-level data
- `.Layout` - Layout-level data

**Parent template (`templates/parent.html`):**
```html
<div class="parent">
    <h1>{{ .Data.Title }}</h1>
    {{ child "child" }}
</div>
```

**Child template (`templates/child.html`):**
```html
<div class="child">
    <p>Item: {{ .Data.ItemName }}</p>
    <p>From: {{ .Parent.Title }}</p>
    <p>Count: {{ .Parent.Count }}</p>
</div>
```

### Nested Children

The `.Parent` field always contains the immediate parent's data, allowing for multi-level hierarchies:

```go
grandparent := partial.New("templates/grandparent.html").ID("grandparent")
grandparent.SetData(map[string]any{"Level": "Grandparent"})

parent := partial.New("templates/parent.html").ID("parent")
parent.SetData(map[string]any{"Level": "Parent"})

child := partial.New("templates/child.html").ID("child")
child.SetData(map[string]any{"Level": "Child"})

parent.With(child)
grandparent.With(parent)
```

**Child template:**
```html
<div>
    My level: {{ .Data.Level }}
    Parent level: {{ .Parent.Level }}
    <!-- Note: To access grandparent, use .Global which contains merged ancestor data -->
</div>
```

### Passing Additional Data with child Function

When using the `child` template function, you can pass additional data that will be merged into the child's `.Data`:

**Parent template:**
```html
{{ range .Data.Items }}
    {{ child "item" "ItemName" . "Index" $index }}
{{ end }}
```

**Child template:**
```html
<div>
    Item {{ .Data.Index }}: {{ .Data.ItemName }}
    From list: {{ .Parent.ListTitle }}
</div>
```

### Checking for Parent

If a partial might be rendered without a parent, you can check if `.Parent` exists:

```html
<div>
    {{ if .Parent }}
        Parent: {{ .Parent.Title }}
    {{ else }}
        No parent
    {{ end }}
</div>
```

## Data Structure

The complete data structure available in templates:

```go
type Data struct {
    Ctx      context.Context    // Request context
    URL      *url.URL          // Request URL
    Request  *http.Request     // HTTP request
    Data     map[string]any    // This partial's data
    Service  map[string]any    // Service-level data
    Layout   map[string]any    // Layout-level data
    Global   map[string]any    // Merged data from all ancestors
    Parent   map[string]any    // Immediate parent's data
    Loc      Localizer         // Localization helper
    Csrf     CsrfToken         // CSRF token
    BasePath string            // Base path
}
```

## Use Cases

### 1. Shared Context
Children can access parent configuration without passing data explicitly:

```go
tablePartial.SetData(map[string]any{
    "Columns": []string{"Name", "Age", "Email"},
    "Sortable": true,
})
```

In row partials, you can access `{{ .Parent.Sortable }}` without passing it to each row.

### 2. Conditional Rendering
Children can adapt their rendering based on parent context:

```html
{{ if .Parent.CompactMode }}
    <div class="compact">{{ .Data.Title }}</div>
{{ else }}
    <div class="full">
        <h2>{{ .Data.Title }}</h2>
        <p>{{ .Data.Description }}</p>
    </div>
{{ end }}
```

### 3. Theming and Styling
Parent components can set theme data that children inherit:

```go
containerPartial.SetData(map[string]any{
    "Theme": "dark",
    "Size": "large",
})
```

Children can then use `{{ .Parent.Theme }}` and `{{ .Parent.Size }}` to apply appropriate styles.

