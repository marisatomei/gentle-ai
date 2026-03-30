package system

import (
	"os"
	"os/exec"
	"path/filepath"
)

// ConfigState records the filesystem presence of an agent's global config directory.
// All known registry agents are always represented — Exists=false for absent dirs.
// This contract is consumed by the TUI detection screen and install/validate flows.
type ConfigState struct {
	Agent       string
	Path        string
	Exists      bool
	IsDirectory bool
	// Binary is the CLI binary name to probe via PATH (optional).
	// When set, Exists is determined by exec.LookPath instead of os.Stat on Path.
	Binary string
}

// knownAgentConfigDirs enumerates every agent's GlobalConfigDir as a
// (agentID, path) pair for the given homeDir. This is a compatibility shim
// that mirrors the adapter registry's full set without importing the agents
// package (which would create an import cycle: system ← agents ← system).
//
// When a new agent is added to the registry, its entry must also be added here
// until the import cycle is resolved and ScanConfigs can delegate directly to
// agents.DiscoverInstalled.
func knownAgentConfigDirs(homeDir string) []ConfigState {
	return []ConfigState{
		{Agent: "claude-code", Path: filepath.Join(homeDir, ".claude")},
		{Agent: "opencode", Path: filepath.Join(homeDir, ".config", "opencode")},
		{Agent: "gemini-cli", Path: filepath.Join(homeDir, ".gemini")},
		{Agent: "cursor", Path: filepath.Join(homeDir, ".cursor")},
		{Agent: "vscode-copilot", Path: vscodeCopilotGlobalConfigDir(homeDir)},
		{Agent: "codex", Path: filepath.Join(homeDir, ".codex")},
		{Agent: "antigravity", Path: filepath.Join(homeDir, ".gemini", "antigravity")},
		{Agent: "windsurf", Path: filepath.Join(homeDir, ".codeium", "windsurf")},
		{Agent: "copilot-cli", Path: filepath.Join(homeDir, ".copilot"), Binary: "copilot"},
	}
}

// vscodeCopilotGlobalConfigDir returns ~/.copilot, the GlobalConfigDir used by
// the vscode-copilot adapter across all platforms. The vscode adapter's
// SystemPromptDir and SettingsPath are OS-dependent, but GlobalConfigDir is not.
func vscodeCopilotGlobalConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".copilot")
}

// lookPathFunc is the function used to probe for binaries on PATH.
// Package-level var for testability.
var lookPathFunc = exec.LookPath

// ScanConfigs returns the presence state of every known managed agent's global
// config directory. All agents are always represented in the result; Exists and
// IsDirectory reflect the actual filesystem state at call time.
//
// For agents with a Binary field set, detection uses exec.LookPath on the binary
// name instead of checking the config directory. This distinguishes agents like
// copilot-cli (binary "copilot") from vscode-copilot (which shares ~/.copilot).
//
// This is a compatibility shim: it preserves the ConfigState contract for TUI
// and validation callers while the canonical discovery (agents.DiscoverInstalled)
// is used by sync and upgrade flows. Full delegation is deferred until the
// system ← agents import cycle is resolved (follow-up change).
func ScanConfigs(homeDir string) []ConfigState {
	states := knownAgentConfigDirs(homeDir)

	for idx := range states {
		if states[idx].Binary != "" {
			// Detect by binary presence on PATH.
			_, err := lookPathFunc(states[idx].Binary)
			states[idx].Exists = err == nil
			continue
		}

		info, err := os.Stat(states[idx].Path)
		if err != nil {
			continue
		}

		states[idx].Exists = true
		states[idx].IsDirectory = info.IsDir()
	}

	return states
}
