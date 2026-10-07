package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fzxbl/terminal-mcp/internal/config"
)

var singleTerminalMode atomic.Bool
var singleTerminalMu sync.Mutex

type singleTerminalState struct {
	Enabled bool `json:"enabled"`
}

func singleTerminalStatePath() string {
	if dir := os.Getenv("TERMINAL_MCP_UI_STATE_DIR"); dir != "" {
		return filepath.Join(dir, "ui-state.json")
	}
	return filepath.Join(config.Get().DataDir, "ui-state.json")
}

func loadSingleTerminalMode() {
	singleTerminalMode.Store(false)
	b, err := os.ReadFile(singleTerminalStatePath())
	if err != nil {
		return
	}
	var st singleTerminalState
	if json.Unmarshal(b, &st) == nil {
		singleTerminalMode.Store(st.Enabled)
	}
}

func persistSingleTerminalMode(on bool) error {
	path := singleTerminalStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(singleTerminalState{Enabled: on}, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func SingleTerminalMode() bool { return singleTerminalMode.Load() }

// SetSingleTerminalMode toggles the optional persistent one-terminal policy.
// Enabling keeps the newest live session and closes any extras. If no session
// exists, a default local shell is created immediately.
func SetSingleTerminalMode(on bool) error {
	singleTerminalMu.Lock()
	defer singleTerminalMu.Unlock()

	old := singleTerminalMode.Load()
	singleTerminalMode.Store(on)
	if err := persistSingleTerminalMode(on); err != nil {
		singleTerminalMode.Store(old)
		return err
	}
	if on {
		_, err := ensureSingleTerminalLocked()
		return err
	}
	return nil
}

func canonicalSessionLocked() *Session {
	if theStore == nil {
		return nil
	}
	all := theStore.list()
	if len(all) == 0 {
		return nil
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })
	keep := all[0]
	for _, extra := range all[1:] {
		if p := extra.getProc(); p != nil {
			p.Close()
		}
		extra.setStatus("closed", "")
		theStore.remove(extra.ID)
	}
	return keep
}

func ensureSingleTerminalLocked() (*Session, error) {
	if !singleTerminalMode.Load() {
		return nil, nil
	}
	sess := canonicalSessionLocked()
	if sess == nil {
		res, err := openNamedRaw("local", "", "", "", "Persistent terminal")
		if err != nil {
			return nil, err
		}
		sess = theStore.get(res["session_id"])
		if sess == nil {
			return nil, fmt.Errorf("persistent terminal was created but not found")
		}
		return sess, nil
	}

	state, _, _ := computeState(sess)
	if state == "dead" {
		env := hardRestart(sess)
		if env.Error != "" || env.State == "dead" {
			return nil, fmt.Errorf("restart persistent terminal %s: %s", sess.ID, env.Error)
		}
	}
	return sess, nil
}

func ensureSingleTerminal() error {
	if !singleTerminalMode.Load() {
		return nil
	}
	singleTerminalMu.Lock()
	defer singleTerminalMu.Unlock()
	_, err := ensureSingleTerminalLocked()
	return err
}

func startSingleTerminalGuard(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		if singleTerminalMode.Load() {
			_ = ensureSingleTerminal()
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if singleTerminalMode.Load() {
					_ = ensureSingleTerminal()
				}
			}
		}
	}()
}
