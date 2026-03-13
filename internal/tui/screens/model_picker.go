package screens

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/opencode"
	"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

// ModelPickerState holds the available providers and models for the picker screen.
type ModelPickerState struct {
	Providers    map[string]opencode.Provider
	AvailableIDs []string                       // provider IDs with env vars set
	SDDModels    map[string][]opencode.Model    // provider ID → SDD-capable models
}

// NewModelPickerState initializes the picker state from the models cache.
func NewModelPickerState(cachePath string) ModelPickerState {
	providers, err := opencode.LoadModels(cachePath)
	if err != nil {
		return ModelPickerState{}
	}

	available := opencode.DetectAvailableProviders(providers)

	sddModels := make(map[string][]opencode.Model, len(available))
	for _, id := range available {
		sddModels[id] = opencode.FilterModelsForSDD(providers[id])
	}

	return ModelPickerState{
		Providers:    providers,
		AvailableIDs: available,
		SDDModels:    sddModels,
	}
}

// ModelPickerRows returns the row labels for the model picker screen.
// Row 0 is "Set all", rows 1-9 are the SDD phases.
func ModelPickerRows() []string {
	rows := make([]string, 0, 10)
	rows = append(rows, "Set all phases")
	rows = append(rows, opencode.SDDPhases()...)
	return rows
}

// flatModelList builds a flat list of (providerID, model) pairs from available providers.
type flatModelEntry struct {
	ProviderID string
	Model      opencode.Model
}

func buildFlatModelList(state ModelPickerState) []flatModelEntry {
	var entries []flatModelEntry
	for _, provID := range state.AvailableIDs {
		for _, m := range state.SDDModels[provID] {
			entries = append(entries, flatModelEntry{ProviderID: provID, Model: m})
		}
	}
	return entries
}

// CycleModelAssignment cycles to the next available model for the given phase (or all phases).
// Returns updated assignments map. cursor=0 means "set all", cursor 1-9 means individual phase.
func CycleModelAssignment(
	assignments map[string]model.ModelAssignment,
	state ModelPickerState,
	cursor int,
) map[string]model.ModelAssignment {
	if assignments == nil {
		assignments = make(map[string]model.ModelAssignment)
	}

	entries := buildFlatModelList(state)
	if len(entries) == 0 {
		return assignments
	}

	phases := opencode.SDDPhases()

	if cursor == 0 {
		// "Set all" — cycle based on the first phase's current assignment.
		current := assignments[phases[0]]
		next := nextEntry(entries, current)
		for _, phase := range phases {
			assignments[phase] = next
		}
		return assignments
	}

	// Individual phase.
	phaseIdx := cursor - 1
	if phaseIdx >= len(phases) {
		return assignments
	}

	phase := phases[phaseIdx]
	current := assignments[phase]
	assignments[phase] = nextEntry(entries, current)
	return assignments
}

// nextEntry finds the next model entry after the current assignment, cycling back to the first.
func nextEntry(entries []flatModelEntry, current model.ModelAssignment) model.ModelAssignment {
	if current.ProviderID == "" && current.ModelID == "" {
		// No assignment yet — pick first entry.
		e := entries[0]
		return model.ModelAssignment{ProviderID: e.ProviderID, ModelID: e.Model.ID}
	}

	for i, e := range entries {
		if e.ProviderID == current.ProviderID && e.Model.ID == current.ModelID {
			next := entries[(i+1)%len(entries)]
			return model.ModelAssignment{ProviderID: next.ProviderID, ModelID: next.Model.ID}
		}
	}

	// Current not found in list, reset to first.
	e := entries[0]
	return model.ModelAssignment{ProviderID: e.ProviderID, ModelID: e.Model.ID}
}

// RenderModelPicker renders the model picker screen.
func RenderModelPicker(
	assignments map[string]model.ModelAssignment,
	state ModelPickerState,
	cursor int,
) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Assign Models to SDD Phases"))
	b.WriteString("\n\n")

	if len(state.AvailableIDs) == 0 {
		b.WriteString(styles.WarningStyle.Render("No providers detected."))
		b.WriteString("\n")
		b.WriteString(styles.SubtextStyle.Render("Set API key env vars (e.g. OPENAI_API_KEY, ANTHROPIC_API_KEY) or switch to single mode."))
		b.WriteString("\n\n")
		b.WriteString(renderOptions([]string{"← Back to SDD mode"}, cursor))
		b.WriteString("\n")
		b.WriteString(styles.HelpStyle.Render("enter/esc: go back"))
		return b.String()
	}

	b.WriteString(styles.SubtextStyle.Render("Press enter to cycle models. j/k to navigate."))
	b.WriteString("\n\n")

	rows := ModelPickerRows()
	for idx, row := range rows {
		focused := idx == cursor

		var label string
		if idx == 0 {
			label = row
		} else {
			phase := opencode.SDDPhases()[idx-1]
			assignment, ok := assignments[phase]
			if ok && assignment.ProviderID != "" {
				provName := assignment.ProviderID
				if p, exists := state.Providers[assignment.ProviderID]; exists && p.Name != "" {
					provName = p.Name
				}
				modelName := assignment.ModelID
				if p, exists := state.Providers[assignment.ProviderID]; exists {
					if m, mOK := p.Models[assignment.ModelID]; mOK && m.Name != "" {
						modelName = m.Name
						if m.Cost.Input > 0 {
							modelName += fmt.Sprintf(" ($%.2f/$%.2f)", m.Cost.Input, m.Cost.Output)
						}
					}
				}
				label = fmt.Sprintf("%-14s %s / %s", row, provName, modelName)
			} else {
				label = fmt.Sprintf("%-14s (default)", row)
			}
		}

		if focused {
			b.WriteString(styles.SelectedStyle.Render(styles.Cursor+label) + "\n")
		} else {
			b.WriteString(styles.UnselectedStyle.Render("  "+label) + "\n")
		}
	}

	b.WriteString("\n")
	actionIdx := cursor - len(rows)
	b.WriteString(renderOptions([]string{"Continue", "← Back"}, actionIdx))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: cycle model / confirm • esc: back"))

	return b.String()
}
