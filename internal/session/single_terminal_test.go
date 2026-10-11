package session

import (
	"context"
	"testing"
	"time"

	"github.com/fzxbl/terminal-mcp/internal/config"
)

func waitSessionIdle(t *testing.T, id string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if Status(id).State == "idle" {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("session %s did not become idle", id)
}

func TestSingleTerminalModeEnforcesOneAndReusesIt(t *testing.T) {
	t.Setenv("TERMINAL_MCP_UI_STATE_DIR", t.TempDir())
	config.Load("")
	InitStore(4)
	SetSelfAddr("")
	defer func() {
		_ = SetSingleTerminalMode(false)
		Shutdown()
	}()

	firstResult, err := OpenNamed("local", "", "", "owner-a", "First")
	first := firstResult["session_id"]
	if err != nil {
		t.Fatal(err)
	}
	waitSessionIdle(t, first)
	time.Sleep(10 * time.Millisecond)
	secondResult, err := OpenNamed("local", "", "", "owner-a", "Second")
	second := secondResult["session_id"]
	if err != nil {
		t.Fatal(err)
	}
	waitSessionIdle(t, second)

	if err := SetSingleTerminalMode(true); err != nil {
		t.Fatal(err)
	}
	all := List()
	if len(all) != 1 {
		t.Fatalf("force-one left %d sessions: %#v", len(all), all)
	}
	if all[0]["session_id"] != second {
		t.Fatalf("force-one kept %s, want newest %s", all[0]["session_id"], second)
	}

	if _, unauthorized := OpenNamed("local", "", "", "owner-a", "Other"); unauthorized == nil {
		t.Fatal("existing terminal capability must be required for reuse")
	}
	reopened, err := OpenNamed("local", "", "", "owner-a", "Persistent dev terminal", secondResult["session_key"])
	if err != nil {
		t.Fatal(err)
	}
	if reopened["session_id"] != second {
		t.Fatalf("terminal_open created/reused %s, want %s", reopened["session_id"], second)
	}
	if got := len(List()); got != 1 {
		t.Fatalf("terminal_open under force-one left %d sessions", got)
	}
}

func TestSingleTerminalGuardRestartsKilledShellWithSameID(t *testing.T) {
	t.Setenv("TERMINAL_MCP_UI_STATE_DIR", t.TempDir())
	config.Load("")
	InitStore(2)
	SetSelfAddr("")
	defer func() {
		_ = SetSingleTerminalMode(false)
		Shutdown()
	}()

	if err := SetSingleTerminalMode(true); err != nil {
		t.Fatal(err)
	}
	created, err := OpenNamed("local", "", "", "owner-a", "Persistent")
	if err != nil {
		t.Fatal(err)
	}
	all := List()
	if len(all) != 1 {
		t.Fatalf("want one persistent session, got %#v", all)
	}
	id := created["session_id"]
	waitSessionIdle(t, id)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startSingleTerminalGuard(ctx)

	sess := theStore.get(id)
	if sess == nil || sess.getProc() == nil {
		t.Fatal("missing persistent session process")
	}
	sess.getProc().Close()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if Status(id).State == "idle" {
			env := Send(id, "echo recovered-same-session", 3000)
			if env.Output != "recovered-same-session" {
				t.Fatalf("recovered output=%q state=%q error=%q", env.Output, env.State, env.Error)
			}
			all = List()
			if len(all) != 1 || all[0]["session_id"] != id {
				t.Fatalf("session id changed after recovery: %#v", all)
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("killed persistent session %s was not recovered", id)
}

func TestSingleTerminalModePersistsAcrossStoreReinit(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TERMINAL_MCP_UI_STATE_DIR", dir)
	config.Load("")
	InitStore(2)
	SetSelfAddr("")
	if err := SetSingleTerminalMode(true); err != nil {
		t.Fatal(err)
	}
	Shutdown()

	InitStore(2)
	if !SingleTerminalMode() {
		t.Fatal("single-terminal mode did not reload from persisted UI state")
	}
	if err := SetSingleTerminalMode(false); err != nil {
		t.Fatal(err)
	}
	Shutdown()
}

func TestForceOneSkipsIdleReap(t *testing.T) {
	t.Setenv("TERMINAL_MCP_UI_STATE_DIR", t.TempDir())
	config.Load("")
	InitStore(2)
	SetSelfAddr("")
	defer func() {
		_ = SetSingleTerminalMode(false)
		Shutdown()
	}()
	if err := SetSingleTerminalMode(true); err != nil {
		t.Fatal(err)
	}
	created, err := OpenNamed("local", "", "", "owner-a", "Persistent")
	if err != nil {
		t.Fatal(err)
	}
	all := List()
	if len(all) != 1 {
		t.Fatalf("want one session, got %#v", all)
	}
	id := created["session_id"]
	theStore.gcIdle(0)
	if _, ok := SessionName(id); !ok {
		t.Fatal("force-one persistent session was reaped by idle GC")
	}
}
