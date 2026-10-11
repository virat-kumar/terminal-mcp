package mcpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPanelMCPDiscovery(t *testing.T) {
	Init("")
	handler := NewHTTPHandler(nil)
	request := func(method string, params map[string]any) map[string]any {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		if recorder.Code != 200 {
			t.Fatalf("%s: HTTP %d: %s", method, recorder.Code, recorder.Body.String())
		}
		var out map[string]any
		if err := json.Unmarshal(recorder.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		if out["error"] != nil {
			t.Fatalf("%s failed: %v", method, out["error"])
		}
		return out["result"].(map[string]any)
	}
	tools := request("tools/list", map[string]any{})["tools"].([]any)
	found := false
	for _, v := range tools {
		tool := v.(map[string]any)
		if tool["name"] == "terminal_monitor" {
			found = true
			meta := tool["_meta"].(map[string]any)
			if ext, ok := meta["openai/ui"].(map[string]any); !ok || len(ext["entrypoints"].([]any)) != 1 || ext["entrypoints"].([]any)[0].(map[string]any)["type"] != "thread" {
				t.Fatal("thread panel entrypoint missing")
			}
			if meta["openai/outputTemplate"] != panelURI {
				t.Fatalf("bad UI template: %v", meta)
			}
			if meta["ui"].(map[string]any)["resourceUri"] != panelURI {
				t.Fatal("bad app resource URI")
			}
		}
	}
	if !found {
		t.Fatal("terminal_monitor tool missing")
	}
	// Opening a terminal advertises the same embedded viewer, so a new terminal
	// can be presented without asking the user to launch the monitor again.
	for _, v := range tools {
		item := v.(map[string]any)
		if item["name"] == "terminal_open" {
			meta := item["_meta"].(map[string]any)
			if meta["openai/outputTemplate"] != panelURI {
				t.Fatalf("terminal_open must provide new-session live viewer: %v", meta)
			}
		}
	}
	resources := request("resources/list", map[string]any{})["resources"].([]any)
	if len(resources) == 0 || !bytes.Contains([]byte(fmt.Sprint(resources)), []byte(panelURI)) {
		t.Fatalf("panel resource missing: %v", resources)
	}
	for _, uri := range []string{"ui://terminal-mcp/monitor-v1.html", "ui://terminal-mcp/monitor-v3.html"} {
		legacy := request("resources/read", map[string]any{"uri": uri})["contents"].([]any)
		if legacy[0].(map[string]any)["uri"] != uri ||
			!bytes.Contains([]byte(legacy[0].(map[string]any)["text"].(string)), []byte("Follow: On")) {
			t.Fatalf("cached monitor %s does not serve Follow UI", uri)
		}
	}
	contents := request("resources/read", map[string]any{"uri": panelURI})["contents"].([]any)
	html := contents[0].(map[string]any)["text"].(string)
	for _, needle := range []string{"Ubuntu Terminal Monitor", "ui/initialize", "appInfo", "appCapabilities", "terminal_list", "terminal_output", "mode:'tail'", "id=\"follow\"", "Follow: On", "scrollToEnd()", "active_session_id", "id=\"refresh\"", "setInterval"} {
		if !bytes.Contains([]byte(html), []byte(needle)) {
			t.Errorf("panel HTML missing %s", needle)
		}
	}
}

// TestPanelAgentSelection verifies the active terminal ID is scoped by owner
// and that polling the output itself does not steal the agent's selection.
func TestPanelAgentSelection(t *testing.T) {
	setActiveAgentSession("test-owner-1", "s1")
	setActiveAgentSession("test-owner-2", "s2")
	if got := activeAgentSession("test-owner-1"); got != "s1" {
		t.Fatal(fmt.Sprintf("owner 1: got %q", got))
	}
	if got := activeAgentSession("test-owner-2"); got != "s2" {
		t.Fatal(fmt.Sprintf("owner 2: got %q", got))
	}
	setActiveAgentSession("test-owner-1", "s3")
	if activeAgentSession("test-owner-2") != "s2" {
		t.Fatal("owner 2 was overwritten")
	}
}
