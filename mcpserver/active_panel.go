package mcpserver

import "sync"

// Tracks the last session an agent actively opened or sent input to, separately
// from viewer polling. The scope is the same signed owner used by terminal tools.
// Never update this from terminal_output/terminal_list, as the monitor calls them.
var activePanelSessions sync.Map

func setActiveAgentSession(owner, sessionID string) {
	if sessionID != "" {
		activePanelSessions.Store(owner, sessionID)
	}
}

func activeAgentSession(owner string) string {
	value, ok := activePanelSessions.Load(owner)
	if !ok {
		return ""
	}
	return value.(string)
}
