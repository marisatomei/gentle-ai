package model

type AgentID string

const (
	AgentClaudeCode    AgentID = "claude-code"
	AgentOpenCode      AgentID = "opencode"
	AgentGeminiCLI     AgentID = "gemini-cli"
	AgentCursor        AgentID = "cursor"
	AgentVSCodeCopilot AgentID = "vscode-copilot"
	AgentCodex         AgentID = "codex"
	AgentAntigravity   AgentID = "antigravity"
	AgentWindsurf      AgentID = "windsurf"
	AgentCopilotCLI    AgentID = "copilot-cli"
)

// SupportTier indicates how fully an agent supports the Gentleman AI ecosystem.
// All current agents receive the full SDD orchestrator, skill files, MCP config,
// and system prompt injection. The tier is kept as metadata for display purposes.
type SupportTier string

const (
	// TierFull — the agent receives all ecosystem features: SDD orchestrator,
	// skill files, MCP servers, system prompt, and sub-agent delegation.
	TierFull SupportTier = "full"
)

type ComponentID string

const (
	ComponentEngram     ComponentID = "engram"
	ComponentSDD        ComponentID = "sdd"
	ComponentSkills     ComponentID = "skills"
	ComponentContext7   ComponentID = "context7"
	ComponentPersona    ComponentID = "persona"
	ComponentPermission ComponentID = "permissions"
	ComponentGGA        ComponentID = "gga"
	ComponentTheme      ComponentID = "theme"
)

type SkillID string

const (
	SkillSDDInit       SkillID = "sdd-init"
	SkillSDDApply      SkillID = "sdd-apply"
	SkillSDDVerify     SkillID = "sdd-verify"
	SkillSDDExplore    SkillID = "sdd-explore"
	SkillSDDPropose    SkillID = "sdd-propose"
	SkillSDDSpec       SkillID = "sdd-spec"
	SkillSDDDesign     SkillID = "sdd-design"
	SkillSDDTasks      SkillID = "sdd-tasks"
	SkillSDDArchive    SkillID = "sdd-archive"
	SkillSDDOnboard    SkillID = "sdd-onboard"
	SkillGoTesting     SkillID = "go-testing"
	SkillCreator       SkillID = "skill-creator"
	SkillJudgmentDay   SkillID = "judgment-day"
	SkillBranchPR      SkillID = "branch-pr"
	SkillIssueCreation SkillID = "issue-creation"
	SkillSkillRegistry SkillID = "skill-registry"
)

type PersonaID string

const (
	PersonaGentleman PersonaID = "gentleman"
	PersonaNeutral   PersonaID = "neutral"
	PersonaCustom    PersonaID = "custom"
)

// SystemPromptStrategy defines how an agent's system prompt file is managed.
type SystemPromptStrategy int

const (
	// StrategyMarkdownSections uses <!-- gentle-ai:ID --> markers to inject sections
	// into an existing file without clobbering user content (Claude Code CLAUDE.md).
	StrategyMarkdownSections SystemPromptStrategy = iota
	// StrategyFileReplace replaces the entire system prompt file (OpenCode AGENTS.md).
	StrategyFileReplace
	// StrategyAppendToFile appends content to an existing system prompt file.
	StrategyAppendToFile
)

// MCPStrategy defines how MCP server configs are written for an agent.
type MCPStrategy int

const (
	// StrategySeparateMCPFiles writes one JSON file per server in a dedicated directory
	// (e.g., ~/.claude/mcp/context7.json).
	StrategySeparateMCPFiles MCPStrategy = iota
	// StrategyMergeIntoSettings merges mcpServers into a settings.json file
	// (e.g., OpenCode, Gemini CLI).
	StrategyMergeIntoSettings
	// StrategyMCPConfigFile writes to a dedicated mcp.json config file (e.g., Cursor ~/.cursor/mcp.json).
	StrategyMCPConfigFile
	// StrategyTOMLFile writes MCP config to a TOML file (e.g., Codex ~/.codex/config.toml).
	StrategyTOMLFile
)

type PresetID string

const (
	PresetFullGentleman PresetID = "full-gentleman"
	PresetEcosystemOnly PresetID = "ecosystem-only"
	PresetMinimal       PresetID = "minimal"
	PresetCustom        PresetID = "custom"
)

type SDDModeID string

const (
	SDDModeSingle SDDModeID = "single"
	SDDModeMulti  SDDModeID = "multi"
)

// CopilotModelID is the model identifier string used in Copilot CLI agent frontmatter.
// The empty string means "inherit from session" (omit the model field).
type CopilotModelID string

const (
	CopilotModelDefault    CopilotModelID = ""
	CopilotModelOpus46     CopilotModelID = "claude-opus-4.6"
	CopilotModelOpus45     CopilotModelID = "claude-opus-4.5"
	CopilotModelSonnet46   CopilotModelID = "claude-sonnet-4.6"
	CopilotModelSonnet45   CopilotModelID = "claude-sonnet-4.5"
	CopilotModelSonnet4    CopilotModelID = "claude-sonnet-4"
	CopilotModelHaiku45    CopilotModelID = "claude-haiku-4.5"
	CopilotModelGPT54      CopilotModelID = "gpt-5.4"
	CopilotModelGPT53Codex CopilotModelID = "gpt-5.3-codex"
	CopilotModelGPT52Codex CopilotModelID = "gpt-5.2-codex"
	CopilotModelGPT52      CopilotModelID = "gpt-5.2"
	CopilotModelGPT51      CopilotModelID = "gpt-5.1"
	CopilotModelGPT54Mini  CopilotModelID = "gpt-5.4-mini"
	CopilotModelGPT5Mini   CopilotModelID = "gpt-5-mini"
	CopilotModelGPT41      CopilotModelID = "gpt-4.1"
	CopilotModelGPT41Mini  CopilotModelID = "gpt-4.1-mini"
)

// CopilotModelEntry describes a single model available in Copilot CLI.
type CopilotModelEntry struct {
	ID       CopilotModelID
	Label    string // model ID as displayed; "(inherit from session)" for empty ID
	Category string // "Claude", "GPT", or "—" for inherit
}

var copilotAllModels = []CopilotModelEntry{
	{CopilotModelDefault, "(inherit from session)", "—"},
	{CopilotModelOpus46, "claude-opus-4.6", "Claude"},
	{CopilotModelOpus45, "claude-opus-4.5", "Claude"},
	{CopilotModelSonnet46, "claude-sonnet-4.6", "Claude"},
	{CopilotModelSonnet45, "claude-sonnet-4.5", "Claude"},
	{CopilotModelSonnet4, "claude-sonnet-4", "Claude"},
	{CopilotModelHaiku45, "claude-haiku-4.5", "Claude"},
	{CopilotModelGPT54, "gpt-5.4", "GPT"},
	{CopilotModelGPT53Codex, "gpt-5.3-codex", "GPT"},
	{CopilotModelGPT52Codex, "gpt-5.2-codex", "GPT"},
	{CopilotModelGPT52, "gpt-5.2", "GPT"},
	{CopilotModelGPT51, "gpt-5.1", "GPT"},
	{CopilotModelGPT54Mini, "gpt-5.4-mini", "GPT"},
	{CopilotModelGPT5Mini, "gpt-5-mini", "GPT"},
	{CopilotModelGPT41, "gpt-4.1", "GPT"},
	{CopilotModelGPT41Mini, "gpt-4.1-mini", "GPT"},
}

// CopilotAllModels returns the ordered list of all Copilot CLI models available for selection.
func CopilotAllModels() []CopilotModelEntry {
	return copilotAllModels
}

// copilotPhases is the ordered list of SDD phase keys for Copilot CLI.
var copilotPhases = []string{
	"sdd-init", "sdd-explore", "sdd-propose", "sdd-spec",
	"sdd-design", "sdd-tasks", "sdd-apply", "sdd-verify", "sdd-archive",
}

// CopilotPhases returns the ordered list of SDD phase keys for Copilot CLI.
func CopilotPhases() []string {
	return copilotPhases
}

// CopilotModelPresetDefault returns the default assignments (all inherit from session).
func CopilotModelPresetDefault() map[string]CopilotModelID {
	return map[string]CopilotModelID{}
}

// CopilotModelPresetPerformance returns the performance preset (all claude-opus-4.6).
func CopilotModelPresetPerformance() map[string]CopilotModelID {
	out := map[string]CopilotModelID{}
	for _, p := range copilotPhases {
		out[p] = CopilotModelOpus46
	}
	return out
}

// CopilotModelPresetEconomy returns the economy preset.
// Most phases use haiku; spec and verify use sonnet for quality.
func CopilotModelPresetEconomy() map[string]CopilotModelID {
	out := map[string]CopilotModelID{}
	for _, p := range copilotPhases {
		out[p] = CopilotModelHaiku45
	}
	out["sdd-verify"] = CopilotModelSonnet45
	out["sdd-spec"] = CopilotModelSonnet45
	return out
}

// Profile represents a named SDD orchestrator configuration with model assignments.
// The default profile (Name="" or Name="default") maps to the base sdd-orchestrator.
// Named profiles generate sdd-orchestrator-{Name} + suffixed sub-agents.
type Profile struct {
	Name              string                     // e.g. "cheap", "premium"; empty = default
	OrchestratorModel ModelAssignment            // orchestrator model
	PhaseAssignments  map[string]ModelAssignment // key = phase name (e.g. "sdd-apply")
}
