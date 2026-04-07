package screens

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

// CopilotModelPreset represents a named preset for Copilot CLI model assignments.
type CopilotModelPreset string

const (
	CopilotPresetDefault     CopilotModelPreset = "default"
	CopilotPresetPerformance CopilotModelPreset = "performance"
	CopilotPresetEconomy     CopilotModelPreset = "economy"
	CopilotPresetCustom      CopilotModelPreset = "custom"
)

var copilotPresetDescriptions = map[CopilotModelPreset]string{
	CopilotPresetDefault:     "Inherit the model from your active Copilot CLI session",
	CopilotPresetPerformance: "Maximum quality: claude-opus-4.6 for all phases",
	CopilotPresetEconomy:     "Cost-optimised: haiku for most phases, sonnet for spec & verify",
	CopilotPresetCustom:      "Pick the model for each SDD phase individually",
}

var copilotPresetOrder = []CopilotModelPreset{
	CopilotPresetDefault,
	CopilotPresetPerformance,
	CopilotPresetEconomy,
	CopilotPresetCustom,
}

// copilotPhaseLabels maps phase key to human-readable label.
var copilotPhaseLabels = map[string]string{
	"sdd-init":    "Init",
	"sdd-explore": "Explore",
	"sdd-propose": "Propose",
	"sdd-spec":    "Spec",
	"sdd-design":  "Design",
	"sdd-tasks":   "Tasks",
	"sdd-apply":   "Apply",
	"sdd-verify":  "Verify",
	"sdd-archive": "Archive",
}

// CopilotModelPickerState holds navigation state for the Copilot CLI model picker screen.
type CopilotModelPickerState struct {
	Preset            CopilotModelPreset
	CustomAssignments map[string]model.CopilotModelID
	InCustomMode      bool
	// Phase picker sub-mode: selecting a model for a single phase from a scrollable list.
	PhasePickerMode   bool
	PhasePickerPhase  string // which phase is being configured
	PhasePickerReturn int    // cursor position to restore in the phase list on exit
}

// NewCopilotModelPickerState returns the initial picker state: default preset selected.
func NewCopilotModelPickerState() CopilotModelPickerState {
	return CopilotModelPickerState{
		Preset:            CopilotPresetDefault,
		CustomAssignments: model.CopilotModelPresetDefault(),
		InCustomMode:      false,
	}
}

// CopilotModelIndexFor returns the index of id in CopilotAllModels, or 0 (inherit) if not found.
func CopilotModelIndexFor(id model.CopilotModelID) int {
	for i, entry := range model.CopilotAllModels() {
		if entry.ID == id {
			return i
		}
	}
	return 0
}

// HandleCopilotModelPickerNav processes a key press on the Copilot model picker screen.
// Returns (handled bool, assignments map) — assignments non-nil means the user confirmed.
func HandleCopilotModelPickerNav(key string, state *CopilotModelPickerState, cursor int) (bool, map[string]model.CopilotModelID) {
	if state.PhasePickerMode {
		return handleCopilotPhasePickerNav(key, state, cursor)
	}
	if state.InCustomMode {
		return handleCopilotCustomNav(key, state, cursor)
	}
	return handleCopilotPresetNav(key, state, cursor)
}

func handleCopilotPresetNav(key string, state *CopilotModelPickerState, cursor int) (bool, map[string]model.CopilotModelID) {
	if key != "enter" {
		return false, nil
	}

	if cursor >= len(copilotPresetOrder) {
		// Back option — caller handles screen transition.
		return false, nil
	}

	selected := copilotPresetOrder[cursor]
	state.Preset = selected

	if selected == CopilotPresetCustom {
		state.InCustomMode = true
		if state.CustomAssignments == nil {
			state.CustomAssignments = model.CopilotModelPresetDefault()
		}
		return true, nil
	}

	var assignments map[string]model.CopilotModelID
	switch selected {
	case CopilotPresetPerformance:
		assignments = model.CopilotModelPresetPerformance()
	case CopilotPresetEconomy:
		assignments = model.CopilotModelPresetEconomy()
	default: // CopilotPresetDefault
		assignments = model.CopilotModelPresetDefault()
	}
	state.CustomAssignments = assignments
	return true, assignments
}

func handleCopilotCustomNav(key string, state *CopilotModelPickerState, cursor int) (bool, map[string]model.CopilotModelID) {
	phases := model.CopilotPhases()
	switch key {
	case "esc":
		state.InCustomMode = false
		return true, nil
	case "enter":
		if cursor < len(phases) {
			// Enter phase picker sub-mode for the selected phase.
			phase := phases[cursor]
			state.PhasePickerPhase = phase
			state.PhasePickerReturn = cursor
			state.PhasePickerMode = true
			return true, nil
		}
		if cursor == len(phases) { // "Confirm"
			return true, state.CustomAssignments
		}
		// "Back" row — exit custom mode.
		state.InCustomMode = false
		return true, nil
	}
	return false, nil
}

func handleCopilotPhasePickerNav(key string, state *CopilotModelPickerState, cursor int) (bool, map[string]model.CopilotModelID) {
	allModels := model.CopilotAllModels()
	switch key {
	case "esc":
		state.PhasePickerMode = false
		return true, nil
	case "enter":
		if cursor < len(allModels) {
			state.CustomAssignments[state.PhasePickerPhase] = allModels[cursor].ID
			state.PhasePickerMode = false
			return true, nil
		}
		// "← Back" row
		state.PhasePickerMode = false
		return true, nil
	}
	return false, nil
}

// RenderCopilotModelPicker renders the Copilot CLI model picker screen.
func RenderCopilotModelPicker(state CopilotModelPickerState, cursor int) string {
	if state.PhasePickerMode {
		return renderCopilotPhaseModelPicker(state, cursor)
	}
	if state.InCustomMode {
		return renderCopilotCustomPhaseList(state, cursor)
	}
	return renderCopilotPresetList(state, cursor)
}

func renderCopilotPresetList(state CopilotModelPickerState, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Copilot CLI Model Assignments"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Choose how models are assigned to each SDD agent:"))
	b.WriteString("\n\n")

	for idx, preset := range copilotPresetOrder {
		isSelected := preset == state.Preset
		focused := idx == cursor
		b.WriteString(renderRadio(string(preset), isSelected, focused))
		b.WriteString(styles.SubtextStyle.Render("    "+copilotPresetDescriptions[preset]) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"← Back"}, cursor-len(copilotPresetOrder)))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}

func renderCopilotCustomPhaseList(state CopilotModelPickerState, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Copilot CLI Custom Model Assignments"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Select a phase and press enter to choose its model"))
	b.WriteString("\n\n")

	phases := model.CopilotPhases()
	for idx, phase := range phases {
		focused := idx == cursor
		assigned := state.CustomAssignments[phase]
		label := fmt.Sprintf("%-20s %s", copilotPhaseLabels[phase], copilotModelTag(assigned))

		if focused {
			b.WriteString(styles.SelectedStyle.Render(styles.Cursor+label) + "\n")
		} else {
			b.WriteString(styles.UnselectedStyle.Render("  "+label) + "\n")
		}
	}

	b.WriteString("\n")
	actionCursor := cursor - len(phases)
	b.WriteString(renderOptions([]string{"Confirm", "← Back"}, actionCursor))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: pick model / confirm • esc: back to presets"))

	return b.String()
}

func renderCopilotPhaseModelPicker(state CopilotModelPickerState, cursor int) string {
	var b strings.Builder

	phaseLabel := copilotPhaseLabels[state.PhasePickerPhase]
	b.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Model for %s phase", phaseLabel)))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Choose a Copilot model for this phase:"))
	b.WriteString("\n\n")

	allModels := model.CopilotAllModels()
	current := state.CustomAssignments[state.PhasePickerPhase]

	for idx, entry := range allModels {
		focused := idx == cursor
		isSelected := entry.ID == current

		catTag := styles.SubtextStyle.Render(fmt.Sprintf("[%s]", entry.Category))
		label := fmt.Sprintf("%-28s %s", entry.Label, catTag)

		var marker string
		if isSelected {
			marker = "◉ "
		} else {
			marker = "○ "
		}

		if focused {
			b.WriteString(styles.SelectedStyle.Render(styles.Cursor+marker+label) + "\n")
		} else {
			b.WriteString(styles.UnselectedStyle.Render("  "+marker+label) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"← Back"}, cursor-len(allModels)))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: cancel"))

	return b.String()
}

func copilotModelTag(m model.CopilotModelID) string {
	entry := model.CopilotAllModels()
	for _, e := range entry {
		if e.ID == m {
			switch e.Category {
			case "Claude":
				return styles.SuccessStyle.Render(fmt.Sprintf("[%s]", e.Label))
			case "GPT":
				return styles.HeadingStyle.Render(fmt.Sprintf("[%s]", e.Label))
			default:
				return styles.SubtextStyle.Render("[inherit]")
			}
		}
	}
	if m == model.CopilotModelDefault {
		return styles.SubtextStyle.Render("[inherit]")
	}
	// Unknown model ID — show it as-is
	return styles.SubtextStyle.Render(fmt.Sprintf("[%s]", string(m)))
}

// CopilotModelPickerOptionCount returns the number of navigable options for the screen.
func CopilotModelPickerOptionCount(state CopilotModelPickerState) int {
	if state.PhasePickerMode {
		return len(model.CopilotAllModels()) + 1 // models + Back
	}
	if state.InCustomMode {
		return len(model.CopilotPhases()) + 2 // phases + Confirm + Back
	}
	return len(copilotPresetOrder) + 1 // presets + Back
}
