package partial

import (
	"net/http"
)

// FlowStep represents a single step in a pageflow.
type FlowStep struct {
	Name     string
	Partial  *Partial
	Validate func(r *http.Request, data map[string]any) error // nil for info-only steps
}

// PageFlow manages a multi-step flow.
type PageFlow struct {
	Steps   []FlowStep
	Current int
}

// FlowSessionData holds all data and validation info for a flow, to be stored in session.
type FlowSessionData struct {
	StepData  map[string]map[string]any // stepName -> data
	Validated map[string]bool           // stepName -> validated
	Current   string                    // current step name
}

// NewPageFlow creates a new PageFlow with the given steps.
func NewPageFlow(steps []FlowStep) *PageFlow {
	return &PageFlow{
		Steps:   steps,
		Current: 0,
	}
}

// GetCurrentStep returns the current FlowStep.
func (f *PageFlow) GetCurrentStep() *FlowStep {
	if f.Current < 0 || f.Current >= len(f.Steps) {
		return nil
	}
	return &f.Steps[f.Current]
}

// Next advances to the next step if possible.
func (f *PageFlow) Next() bool {
	if f.Current < len(f.Steps)-1 {
		f.Current++
		return true
	}
	return false
}

// Prev goes back to the previous step if possible.
func (f *PageFlow) Prev() bool {
	if f.Current > 0 {
		f.Current--
		return true
	}
	return false
}

// FindStep returns the index of a step by name, or -1 if not found.
func (f *PageFlow) FindStep(name string) int {
	for i, step := range f.Steps {
		if step.Name == name {
			return i
		}
	}
	return -1
}

// AllPreviousValidated checks if all previous steps are validated.
func (f *PageFlow) AllPreviousValidated(session *FlowSessionData) bool {
	curStep := f.GetCurrentStep()
	if curStep == nil {
		return false
	}
	curIdx := f.FindStep(curStep.Name)
	for i := 0; i < curIdx; i++ {
		if !session.Validated[f.Steps[i].Name] {
			return false
		}
	}
	return true
}

// SetStepValidated marks a step as validated in the session.
func (session *FlowSessionData) SetStepValidated(stepName string, valid bool) {
	if session.Validated == nil {
		session.Validated = make(map[string]bool)
	}
	session.Validated[stepName] = valid
}

// SetStepData sets the data for a step in the session.
func (session *FlowSessionData) SetStepData(stepName string, data map[string]any) {
	if session.StepData == nil {
		session.StepData = make(map[string]map[string]any)
	}
	session.StepData[stepName] = data
}

// GetStepData gets the data for a step from the session.
func (session *FlowSessionData) GetStepData(stepName string) map[string]any {
	if session.StepData == nil {
		return nil
	}
	return session.StepData[stepName]
}

// GetAllData returns all step data as a single merged map.
func (session *FlowSessionData) GetAllData() map[string]any {
	merged := make(map[string]any)
	for _, data := range session.StepData {
		for k, v := range data {
			merged[k] = v
		}
	}
	return merged
}
