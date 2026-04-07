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
	CopilotPresetPerformance: "Maximum quality: claude-opus-4.5 for all phases",
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

// copilotModelCycleOrder defines the cycling order when pressing Enter on a phase row.
var copilotModelCycleOrder = []model.CopilotModelID{
	model.CopilotModelDefault,
	model.CopilotModelSonnet45,
	model.CopilotModelOpus45,
	model.CopilotModelHaiku45,
	model.CopilotModelSonnet46,
	model.CopilotModelGPT41,
	model.CopilotModelGPT41Mini,
}

// CopilotModelPickerState holds navigation state for the Copilot CLI model picker screen.
type CopilotModelPickerState struct {
	Preset            CopilotModelPreset
	CustomAssignments map[string]model.CopilotModelID
	InCustomMode      bool
}

// NewCopilotModelPickerState returns the initial picker state: default preset selected.
func NewCopilotModelPickerState() CopilotModelPickerState {
	return CopilotModelPickerState{
		Preset:            CopilotPresetDefault,
		CustomAssignments: model.CopilotModelPresetDefault(),
		InCustomMode:      false,
	}
}

// HandleCopilotModelPickerNav processes a key press on the Copilot model picker screen.
// Returns (handled bool, assignments map) — assignments non-nil means the user confirmed.
func HandleCopilotModelPickerNav(key string, state *CopilotModelPickerState, cursor int) (bool, map[string]model.CopilotModelID) {
	if !state.InCustomMode {
		return handleCopilotPresetNav(key, state, cursor)
	}
	return handleCopilotCustomNav(key, state, cursor)
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
			phase := phases[cursor]
			current := state.CustomAssignments[phase]
			state.CustomAssignments[phase] = nextCopilotModel(current)
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

func nextCopilotModel(current model.CopilotModelID) model.CopilotModelID {
	for i, m := range copilotModelCycleOrder {
		if m == current {
			return copilotModelCycleOrder[(i+1)%len(copilotModelCycleOrder)]
		}
	}
	return model.CopilotModelSonnet45
}

// RenderCopilotModelPicker renders the Copilot CLI model picker screen.
func RenderCopilotModelPicker(state CopilotModelPickerState, cursor int) string {
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
	b.WriteString(styles.SubtextStyle.Render("Press enter on a phase to cycle through available models"))
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
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: cycle model / confirm • esc: back to presets"))

	return b.String()
}

func copilotModelTag(m model.CopilotModelID) string {
	switch m {
	case model.CopilotModelOpus45:
		return styles.WarningStyle.Render("[opus-4.5]")
	case model.CopilotModelHaiku45:
		return styles.SubtextStyle.Render("[haiku-4.5]")
	case model.CopilotModelSonnet46:
		return styles.SuccessStyle.Render("[sonnet-4.6]")
	case model.CopilotModelGPT41:
		return styles.HeadingStyle.Render("[gpt-4.1]")
	case model.CopilotModelGPT41Mini:
		return styles.SubtextStyle.Render("[gpt-4.1-mini]")
	case model.CopilotModelSonnet45:
		return styles.SuccessStyle.Render("[sonnet-4.5]")
	default: // CopilotModelDefault ("") or any unrecognised value
		return styles.SubtextStyle.Render("[inherit]")
	}
}

// CopilotModelPickerOptionCount returns the number of navigable options for the screen.
func CopilotModelPickerOptionCount(state CopilotModelPickerState) int {
	if state.InCustomMode {
		return len(model.CopilotPhases()) + 2 // phases + Confirm + Back
	}
	return len(copilotPresetOrder) + 1 // presets + Back
}
