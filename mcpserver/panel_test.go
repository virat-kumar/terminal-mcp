package mcpserver

import (
	"bytes"
	"encoding/json"
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
	resources := request("resources/list", map[string]any{})["resources"].([]any)
	if len(resources) == 0 || resources[0].(map[string]any)["uri"] != panelURI {
		t.Fatalf("panel resource missing: %v", resources)
	}
	contents := request("resources/read", map[string]any{"uri": panelURI})["contents"].([]any)
	html := contents[0].(map[string]any)["text"].(string)
	for _, needle := range []string{"Ubuntu Terminal Monitor", "ui/initialize", "appCapabilities", "terminal_list", "terminal_output", "mode:'tail'"} {
		if !bytes.Contains([]byte(html), []byte(needle)) {
			t.Errorf("panel HTML missing %s", needle)
		}
	}
}
