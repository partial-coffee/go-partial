package partial

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPageFlow_BasicNavigationAndValidation(t *testing.T) {
	// Define three steps: info, form, confirm
	steps := []FlowStep{
		{
			Name:     "info",
			Partial:  nil,
			Validate: nil, // info-only
		},
		{
			Name:    "form",
			Partial: nil,
			Validate: func(r *http.Request, data map[string]any) error {
				if data["field"] == "ok" {
					return nil
				}
				return errors.New("invalid")
			},
		},
		{
			Name:     "confirm",
			Partial:  nil,
			Validate: nil, // info-only
		},
	}
	flow := NewPageFlow(steps)
	session := &FlowSessionData{}

	// Step 1: info (no validation needed)
	if !flow.AllPreviousValidated(session) {
		t.Error("First step should not require previous validation")
	}
	session.SetStepValidated("info", true)

	// Move to step 2
	flow.Next()
	if !flow.AllPreviousValidated(session) {
		t.Error("Second step should see first as validated")
	}

	// Try invalid data for step 2
	badData := map[string]any{"field": "bad"}
	err := steps[1].Validate(nil, badData)
	if err == nil {
		t.Error("Expected validation error for bad data")
	}
	session.SetStepValidated("form", false)

	// Try valid data for step 2
	goodData := map[string]any{"field": "ok"}
	err = steps[1].Validate(nil, goodData)
	if err != nil {
		t.Errorf("Expected no error for good data, got: %v", err)
	}
	session.SetStepValidated("form", true)
	session.SetStepData("form", goodData)

	// Move to step 3
	flow.Next()
	if !flow.AllPreviousValidated(session) {
		t.Error("Third step should see previous as validated")
	}

	// Check data aggregation
	session.SetStepData("info", map[string]any{"foo": 1})
	all := session.GetAllData()
	if all["foo"] != 1 || all["field"] != "ok" {
		t.Errorf("Aggregated data incorrect: %#v", all)
	}
}

func TestPageFlow_StepFromURL(t *testing.T) {
	steps := []FlowStep{
		{Name: "one"},
		{Name: "two"},
		{Name: "three"},
	}
	flow := NewPageFlow(steps)

	// Simulate URL query: ?step=two
	stepName := "two"
	idx := flow.FindStep(stepName)
	if idx == -1 {
		t.Fatalf("Step %q not found", stepName)
	}
	flow.Current = idx
	if flow.GetCurrentStep().Name != "two" {
		t.Errorf("Expected current step to be 'two', got %q", flow.GetCurrentStep().Name)
	}

	// Simulate URL query: ?step=three
	stepName = "three"
	idx = flow.FindStep(stepName)
	if idx == -1 {
		t.Fatalf("Step %q not found", stepName)
	}
	flow.Current = idx
	if flow.GetCurrentStep().Name != "three" {
		t.Errorf("Expected current step to be 'three', got %q", flow.GetCurrentStep().Name)
	}

	// Simulate URL query: ?step=invalid
	stepName = "invalid"
	idx = flow.FindStep(stepName)
	if idx != -1 {
		t.Errorf("Expected -1 for invalid step, got %d", idx)
	}
}

func TestPageFlow_RenderIntegration(t *testing.T) {
	// Define three steps with simple partials
	infoPartial := New().ID("info").SetData(map[string]any{"msg": "Welcome!"})
	formPartial := New().ID("form").SetData(map[string]any{"prompt": "Enter value:"})
	confirmPartial := New().ID("confirm").SetData(map[string]any{"done": true})

	steps := []FlowStep{
		{
			Name:     "info",
			Partial:  infoPartial,
			Validate: nil, // info-only
		},
		{
			Name:    "form",
			Partial: formPartial,
			Validate: func(r *http.Request, data map[string]any) error {
				if data["field"] == "ok" {
					return nil
				}
				return errors.New("invalid")
			},
		},
		{
			Name:     "confirm",
			Partial:  confirmPartial,
			Validate: nil, // info-only
		},
	}
	flow := NewPageFlow(steps)

	// Simulate main page rendering the current step's partial
	renderStep := func(stepName string) map[string]any {
		idx := flow.FindStep(stepName)
		if idx == -1 {
			t.Fatalf("Step %q not found", stepName)
		}
		flow.Current = idx
		step := flow.GetCurrentStep()
		if step.Partial == nil {
			t.Fatalf("Step %q has no partial", stepName)
		}
		return step.Partial.data
	}

	// Render info step
	infoData := renderStep("info")
	if infoData["msg"] != "Welcome!" {
		t.Errorf("Expected info step to have msg 'Welcome!', got %#v", infoData)
	}

	// Render form step
	formData := renderStep("form")
	if formData["prompt"] != "Enter value:" {
		t.Errorf("Expected form step to have prompt 'Enter value:', got %#v", formData)
	}

	// Render confirm step
	confirmData := renderStep("confirm")
	if confirmData["done"] != true {
		t.Errorf("Expected confirm step to have done true, got %#v", confirmData)
	}
}

func TestPageFlow_EndToEndHTTP(t *testing.T) {
	// Define simple HTML templates for each step
	infoPartial := New().ID("info").SetData(map[string]any{"msg": "Welcome!"})
	formPartial := New().ID("form").SetData(map[string]any{"prompt": "Enter value:"})
	confirmPartial := New().ID("confirm").SetData(map[string]any{"done": true})

	steps := []FlowStep{
		{
			Name:     "info",
			Partial:  infoPartial,
			Validate: nil, // info-only
		},
		{
			Name:    "form",
			Partial: formPartial,
			Validate: func(r *http.Request, data map[string]any) error {
				if data["field"] == "ok" {
					return nil
				}
				return errors.New("invalid")
			},
		},
		{
			Name:     "confirm",
			Partial:  confirmPartial,
			Validate: nil, // info-only
		},
	}
	flow := NewPageFlow(steps)

	// Use a map as a fake session store keyed by a constant session ID
	sessionStore := map[string]*FlowSessionData{}
	sessionID := "testsession"

	getSession := func(r *http.Request) *FlowSessionData {
		s, ok := sessionStore[sessionID]
		if !ok {
			s = &FlowSessionData{}
			sessionStore[sessionID] = s
		}
		return s
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stepName := r.URL.Query().Get("step")
		if stepName == "" {
			stepName = steps[0].Name
		}
		idx := flow.FindStep(stepName)
		if idx == -1 {
			http.Error(w, "step not found", http.StatusNotFound)
			return
		}
		flow.Current = idx
		step := flow.GetCurrentStep()
		session := getSession(r)

		// Simulate POST for validation
		if r.Method == http.MethodPost && step.Validate != nil {
			r.ParseForm()
			data := map[string]any{"field": r.FormValue("field")}
			err := step.Validate(r, data)
			if err == nil {
				session.SetStepValidated(step.Name, true)
				session.SetStepData(step.Name, data)
				http.Redirect(w, r, "/?step=confirm", http.StatusSeeOther)
				return
			} else {
				session.SetStepValidated(step.Name, false)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("validation failed"))
				return
			}
		}

		// Render the partial for the current step
		w.Header().Set("Content-Type", "text/html")
		if step.Name == "info" {
			w.Write([]byte("<h1>" + step.Partial.data["msg"].(string) + "</h1>"))
		} else if step.Name == "form" {
			w.Write([]byte("<form method='POST'><input name='field'><button>Submit</button></form>"))
		} else if step.Name == "confirm" {
			w.Write([]byte("<div>Done: true</div>"))
		}
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Step 1: GET info
	resp, err := http.Get(ts.URL + "?step=info")
	if err != nil {
		t.Fatalf("GET info failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	if cerr := resp.Body.Close(); cerr != nil {
		t.Errorf("Error closing response body: %v", cerr)
	}
	if string(body) != "<h1>Welcome!</h1>" {
		t.Errorf("Expected info page, got: %s", string(body))
	}

	// Step 2: GET form
	resp, err = http.Get(ts.URL + "?step=form")
	if err != nil {
		t.Fatalf("GET form failed: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	if cerr := resp.Body.Close(); cerr != nil {
		t.Errorf("Error closing response body: %v", cerr)
	}
	if string(body) != "<form method='POST'><input name='field'><button>Submit</button></form>" {
		t.Errorf("Expected form page, got: %s", string(body))
	}

	// Step 3: POST form with invalid data
	resp, err = http.PostForm(ts.URL+"?step=form", map[string][]string{"field": {"bad"}})
	if err != nil {
		t.Fatalf("POST form failed: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	if cerr := resp.Body.Close(); cerr != nil {
		t.Errorf("Error closing response body: %v", cerr)
	}
	if resp.StatusCode != http.StatusBadRequest || string(body) != "validation failed" {
		t.Errorf("Expected validation failure, got: %d %s", resp.StatusCode, string(body))
	}

	// Step 4: POST form with valid data
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err = client.PostForm(ts.URL+"?step=form", map[string][]string{"field": {"ok"}})
	if err != nil {
		t.Fatalf("POST form (valid) failed: %v", err)
	}
	if cerr := resp.Body.Close(); cerr != nil {
		t.Errorf("Error closing response body: %v", cerr)
	}
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected redirect after valid form, got: %d", resp.StatusCode)
	}

	// Step 5: GET confirm
	resp, err = http.Get(ts.URL + "?step=confirm")
	if err != nil {
		t.Fatalf("GET confirm failed: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	if cerr := resp.Body.Close(); cerr != nil {
		t.Errorf("Error closing response body: %v", cerr)
	}
	if string(body) != "<div>Done: true</div>" {
		t.Errorf("Expected confirm page, got: %s", string(body))
	}
}
