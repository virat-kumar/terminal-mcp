package terminal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fzxbl/terminal-mcp/internal/session"
)

func TestSessionIndexListsNewestFirstAndLinksToTerminal(t *testing.T) {
	first := openLocalReady(t)
	defer session.Close(first)
	time.Sleep(10 * time.Millisecond)
	second := openLocalReady(t)
	defer session.Close(second)

	srv := httptest.NewServer(SessionIndexHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	page := string(body)
	if !strings.Contains(page, "Terminal Sessions") {
		t.Fatalf("page missing English heading")
	}
	if !strings.Contains(page, "/view/terminal/"+first) || !strings.Contains(page, "/view/terminal/"+second) {
		t.Fatalf("page missing terminal links")
	}
	if strings.Index(page, second) > strings.Index(page, first) {
		t.Fatalf("newest session is not listed first")
	}
}

func TestSessionIndexRejectsUnknownPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rec := httptest.NewRecorder()
	SessionIndexHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestFormatAge(t *testing.T) {
	for _, tc := range []struct {
		seconds int
		want    string
	}{{4, "4s ago"}, {120, "2m ago"}, {7200, "2h ago"}, {172800, "2d ago"}} {
		if got := formatAge(tc.seconds); got != tc.want {
			t.Fatalf("formatAge(%d)=%q want %q", tc.seconds, got, tc.want)
		}
	}
}

func TestSessionAPIListsAndRenames(t *testing.T) {
	id := openLocalReady(t)
	defer session.Close(id)
	if _, ok := session.Rename(id, "Initial name"); !ok {
		t.Fatal("rename setup failed")
	}

	h := SessionAPIHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/sessions", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"name":"Initial name"`) {
		t.Fatalf("GET sessions status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/sessions", strings.NewReader(`{"session_id":"`+id+`","name":"Edited in UI"}`))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"name":"Edited in UI"`) {
		t.Fatalf("PATCH sessions status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got, _ := session.SessionName(id); got != "Edited in UI" {
		t.Fatalf("renamed name=%q", got)
	}
}

func TestSessionIndexContainsLiveUpdatesAndEditableNames(t *testing.T) {
	rec := httptest.NewRecorder()
	SessionIndexHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	page := rec.Body.String()
	for _, want := range []string{"name-input", "/api/sessions", "setInterval(sync,2000)", "rankByActivity", "list=rankByActivity(list)", "Session List", "Force one terminal", "themeToggle", "Light mode", "terminal_mcp_index_theme"} {
		if !strings.Contains(page, want) {
			t.Fatalf("index missing %q", want)
		}
	}
}

func TestFaviconHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	FaviconHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/favicon.png", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("favicon status=%d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("favicon content-type=%q", ct)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("favicon body is empty")
	}
}

func TestSessionAPITogglesSingleTerminalModeWithoutNewRoute(t *testing.T) {
	t.Setenv("TERMINAL_MCP_UI_STATE_DIR", t.TempDir())
	session.InitStore(3)
	defer func() {
		_ = session.SetSingleTerminalMode(false)
		session.Shutdown()
	}()

	first := openLocalReady(t)
	time.Sleep(10 * time.Millisecond)
	second := openLocalReady(t)
	if first == second {
		t.Fatal("expected distinct baseline sessions")
	}

	h := SessionAPIHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/sessions", strings.NewReader(`{"single_terminal_mode":true}`))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("toggle status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"single_terminal_mode":true`) {
		t.Fatalf("toggle response missing enabled state: %s", rec.Body.String())
	}
	if got := len(session.List("")); got != 1 {
		t.Fatalf("toggle did not enforce one session, got %d", got)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/sessions", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"single_terminal_mode":true`) {
		t.Fatalf("GET state status=%d body=%s", rec.Code, rec.Body.String())
	}
}
