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
