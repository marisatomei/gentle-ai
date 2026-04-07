package screens

import (
"fmt"
"strings"

"github.com/gentleman-programming/gentle-ai/internal/model"
"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

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

// copilotSetAllPhases is the sentinel phase key for the "Set all phases" row.
const copilotSetAllPhases = "__set_all__"

// CopilotModelPickerState holds navigation state for the Copilot CLI model picker screen.
type CopilotModelPickerState struct {
Assignments map[string]model.CopilotModelID
// Phase picker sub-mode: selecting a model for a single phase from a scrollable list.
PhasePickerMode   bool
PhasePickerPhase  string // "__set_all__" or a phase key
PhasePickerReturn int    // cursor position to restore in the phase list on exit
ModelScroll       int    // scroll offset inside the model list
}

// NewCopilotModelPickerState returns the initial picker state (all phases inherit).
func NewCopilotModelPickerState() CopilotModelPickerState {
return CopilotModelPickerState{
Assignments: model.CopilotModelPresetDefault(),
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
return handleCopilotPhaseListNav(key, state, cursor)
}

// copilotPhaseListRows returns [set-all, phase0, phase1, ..., phase8].
func copilotPhaseListRows() []string {
rows := []string{copilotSetAllPhases}
rows = append(rows, model.CopilotPhases()...)
return rows
}

func handleCopilotPhaseListNav(key string, state *CopilotModelPickerState, cursor int) (bool, map[string]model.CopilotModelID) {
rows := copilotPhaseListRows()
switch key {
case "enter":
if cursor < len(rows) {
phase := rows[cursor]
state.PhasePickerPhase = phase
state.PhasePickerReturn = cursor
state.PhasePickerMode = true
state.ModelScroll = 0
// Position cursor at the currently assigned model.
// For "set all", use the first phase's assignment as hint.
var current model.CopilotModelID
if phase == copilotSetAllPhases {
phases := model.CopilotPhases()
if len(phases) > 0 {
current = state.Assignments[phases[0]]
}
} else {
current = state.Assignments[phase]
}
_ = current // returned via CopilotModelIndexFor in model.go
return true, nil
}
if cursor == len(rows) { // Continue
return true, state.Assignments
}
// Back — handled by caller (returns false so model.go does the screen transition)
return false, nil
}
return false, nil
}

func handleCopilotPhasePickerNav(key string, state *CopilotModelPickerState, cursor int) (bool, map[string]model.CopilotModelID) {
allModels := model.CopilotAllModels()
switch key {
case "esc":
state.PhasePickerMode = false
return true, nil
case "up", "k":
if cursor > 0 {
newCursor := cursor - 1
if newCursor < state.ModelScroll {
state.ModelScroll = newCursor
}
}
return false, nil // let model.go handle cursor movement
case "down", "j":
max := len(allModels) // +1 for Back row, but cursor bounded by optionCount
if cursor < max {
newCursor := cursor + 1
if newCursor >= state.ModelScroll+maxVisibleItems {
state.ModelScroll = newCursor - maxVisibleItems + 1
}
}
return false, nil
case "enter":
if cursor < len(allModels) {
chosen := allModels[cursor].ID
if state.PhasePickerPhase == copilotSetAllPhases {
for _, p := range model.CopilotPhases() {
state.Assignments[p] = chosen
}
} else {
state.Assignments[state.PhasePickerPhase] = chosen
}
state.PhasePickerMode = false
return true, nil
}
// Back row
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
return renderCopilotPhaseList(state, cursor)
}

func renderCopilotPhaseList(state CopilotModelPickerState, cursor int) string {
var b strings.Builder

b.WriteString(styles.TitleStyle.Render("Copilot CLI Model Assignments"))
b.WriteString("\n\n")
b.WriteString(styles.SubtextStyle.Render("Current assignments (enter to change):"))
b.WriteString("\n\n")

rows := copilotPhaseListRows()
for idx, row := range rows {
focused := idx == cursor

var label string
if row == copilotSetAllPhases {
label = fmt.Sprintf("%-20s %s", "Set all phases", copilotAllPhasesTag(state.Assignments))
} else {
assigned := state.Assignments[row]
label = fmt.Sprintf("%-20s %s", copilotPhaseLabels[row], copilotModelTag(assigned))
}

if focused {
b.WriteString(styles.SelectedStyle.Render(styles.Cursor+label) + "\n")
} else {
b.WriteString(styles.UnselectedStyle.Render("  "+label) + "\n")
}
}

b.WriteString("\n")
actionCursor := cursor - len(rows)
b.WriteString(renderOptions([]string{"Continue", "← Back"}, actionCursor))
b.WriteString("\n")
b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: change model / confirm • esc: back"))

return b.String()
}

func renderCopilotPhaseModelPicker(state CopilotModelPickerState, cursor int) string {
var b strings.Builder

var title string
if state.PhasePickerPhase == copilotSetAllPhases {
title = "Model for all phases"
} else {
title = fmt.Sprintf("Model for %s phase", copilotPhaseLabels[state.PhasePickerPhase])
}
b.WriteString(styles.TitleStyle.Render(title))
b.WriteString("\n\n")
b.WriteString(styles.SubtextStyle.Render("Choose a Copilot model:"))
b.WriteString("\n\n")

allModels := model.CopilotAllModels()

var current model.CopilotModelID
if state.PhasePickerPhase == copilotSetAllPhases {
phases := model.CopilotPhases()
if len(phases) > 0 {
current = state.Assignments[phases[0]]
}
} else {
current = state.Assignments[state.PhasePickerPhase]
}

end := state.ModelScroll + maxVisibleItems
if end > len(allModels) {
end = len(allModels)
}

if state.ModelScroll > 0 {
b.WriteString(styles.SubtextStyle.Render("  ↑ more") + "\n")
}

for i := state.ModelScroll; i < end; i++ {
entry := allModels[i]
focused := i == cursor
isSelected := entry.ID == current

catTag := styles.SubtextStyle.Render(fmt.Sprintf("[%s]", entry.Category))
marker := "○ "
if isSelected {
marker = "◉ "
}
label := fmt.Sprintf("%s%-28s %s", marker, entry.Label, catTag)

if focused {
b.WriteString(styles.SelectedStyle.Render(styles.Cursor+label) + "\n")
} else {
b.WriteString(styles.UnselectedStyle.Render("  "+label) + "\n")
}
}

if end < len(allModels) {
b.WriteString(styles.SubtextStyle.Render("  ↓ more") + "\n")
}

b.WriteString("\n")
b.WriteString(renderOptions([]string{"← Back"}, cursor-len(allModels)))
b.WriteString("\n")
b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: cancel"))

return b.String()
}

// copilotAllPhasesTag returns a tag summarising all phase assignments.
// When all phases share the same model, shows that model.
// When they differ, shows "(mixed)".
func copilotAllPhasesTag(assignments map[string]model.CopilotModelID) string {
phases := model.CopilotPhases()
if len(phases) == 0 {
return copilotModelTag(model.CopilotModelDefault)
}
first := assignments[phases[0]]
for _, p := range phases[1:] {
if assignments[p] != first {
return styles.SubtextStyle.Render("(mixed)")
}
}
return copilotModelTag(first)
}

func copilotModelTag(m model.CopilotModelID) string {
for _, e := range model.CopilotAllModels() {
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
return styles.SubtextStyle.Render(fmt.Sprintf("[%s]", string(m)))
}

// CopilotModelPickerOptionCount returns the number of navigable options for the screen.
func CopilotModelPickerOptionCount(state CopilotModelPickerState) int {
if state.PhasePickerMode {
return len(model.CopilotAllModels()) + 1 // models + Back
}
// Phase list: 1 (Set all) + 9 phases + 2 (Continue + Back) = 12
return 1 + len(model.CopilotPhases()) + 2
}
