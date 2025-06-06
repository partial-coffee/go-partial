package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/partial-coffee/go-partial"
)

// In-memory session store for demonstration
var (
	sessionStore = make(map[string]*partial.FlowSessionData)
	sessionMu    sync.Mutex
)

func getSession(r *http.Request) *partial.FlowSessionData {
	// For demo, use a fixed session key
	sessionMu.Lock()
	defer sessionMu.Unlock()
	if sessionStore["demo"] == nil {
		sessionStore["demo"] = &partial.FlowSessionData{}
	}
	return sessionStore["demo"]
}

func main() {
	fsys := &partial.InMemoryFS{
		Files: map[string]string{
			"templates/layout.html":  `<html><head><title>Flow Example</title></head><body>{{ child "content" }}</body></html>`,
			"templates/welcome.html": `<h1>Welcome</h1><a href='/?step=info'>Next</a>`,
			"templates/info.html":    `<h1>Info</h1><a href='/?step=welcome'>Back</a> <a href='/?step=form'>Next</a>`,
			"templates/form.html":    `<h1>Form</h1><form method='post'><input name='field' placeholder='Type something'><button type='submit'>Submit</button></form>{{ if .Data.Error }}<div style='color:red'>{{ .Data.Error }}</div>{{ end }}<a href='/?step=info'>Back</a>`,
			"templates/confirm.html": `<h1>Confirm</h1><div>Done: true</div><a href='/?step=welcome'>Restart</a>`,
		},
	}

	steps := []partial.FlowStep{
		{
			Name:     "welcome",
			Partial:  partial.New("templates/welcome.html").ID("content"),
			Validate: nil,
		},
		{
			Name:     "info",
			Partial:  partial.New("templates/info.html").ID("content"),
			Validate: nil,
		},
		{
			Name:    "form",
			Partial: partial.New("templates/form.html").ID("content"),
			Validate: func(r *http.Request, data map[string]any) error {
				if r.Method == http.MethodPost {
					if r.FormValue("field") == "" {
						return fmt.Errorf("field required")
					}
				}
				return nil
			},
		},
		{
			Name:     "confirm",
			Partial:  partial.New("templates/confirm.html").ID("content"),
			Validate: nil,
		},
	}
	flow := partial.NewPageFlow(steps)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r)
		stepName := r.URL.Query().Get("step")
		if stepName == "" {
			stepName = steps[0].Name
		}
		idx := flow.FindStep(stepName)
		if idx == -1 {
			http.Error(w, "Step not found", http.StatusNotFound)
			return
		}
		flow.Current = idx
		curStep := flow.GetCurrentStep()

		var renderError string
		if r.Method == http.MethodPost && curStep.Validate != nil {
			err := curStep.Validate(r, nil)
			if err != nil {
				renderError = err.Error()
			} else {
				session.SetStepValidated(curStep.Name, true)
				if flow.Next() {
					http.Redirect(w, r, fmt.Sprintf("/?step=%s", flow.GetCurrentStep().Name), http.StatusSeeOther)
					return
				}
			}
		}

		// Prepare the layout and content partials
		layout := partial.New("templates/layout.html").ID("root")
		layout.With(curStep.Partial.SetData(map[string]any{"Error": renderError}))

		service := partial.NewService(&partial.Config{})
		out, err := service.NewLayout().FS(fsys).Set(layout).RenderWithRequest(r.Context(), r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(out))
	})

	log.Println("Example flow running at http://localhost:8123/")
	log.Fatal(http.ListenAndServe(":8123", nil))
}
