package terminal

import (
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/fzxbl/terminal-mcp/internal/session"
)

type sessionIndexRow struct {
	SessionID string
	Href      string
	Mode      string
	Host      string
	Status    string
	Idle      string
	Created   string
	Held      bool
}

type sessionIndexData struct {
	Sessions []sessionIndexRow
}

var sessionIndexTemplate = template.Must(template.New("session-index").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>terminal-mcp sessions</title>
<style>
  :root{color-scheme:dark;--bg:#11111b;--panel:#181825;--panel2:#1e1e2e;--border:#313244;--text:#cdd6f4;--muted:#a6adc8;--accent:#89b4fa;--green:#a6e3a1;--yellow:#f9e2af;--red:#f38ba8;--purple:#cba6f7}
  *{box-sizing:border-box} body{margin:0;background:var(--bg);color:var(--text);font:14px/1.5 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif}
  main{max-width:1200px;margin:0 auto;padding:40px 24px 64px} header{display:flex;align-items:flex-start;gap:16px;margin-bottom:24px} header>div{flex:1}
  h1{font-size:26px;margin:0 0 4px;font-weight:650;letter-spacing:-.02em} p{margin:0;color:var(--muted)}
  .refresh{border:1px solid var(--border);background:var(--panel2);color:var(--text);padding:9px 14px;border-radius:9px;text-decoration:none;font-weight:600}.refresh:hover{border-color:var(--accent);color:var(--accent)}
  .panel{background:var(--panel);border:1px solid var(--border);border-radius:14px;overflow:hidden;box-shadow:0 14px 40px rgba(0,0,0,.24)}
  table{width:100%;border-collapse:collapse} th,td{padding:13px 16px;text-align:left;border-bottom:1px solid var(--border);vertical-align:middle} th{font-size:11px;text-transform:uppercase;letter-spacing:.08em;color:var(--muted);background:rgba(255,255,255,.02)} tr:last-child td{border-bottom:0} tbody tr:hover{background:rgba(137,180,250,.045)}
  .session{font-family:"SFMono-Regular",Consolas,"Liberation Mono",monospace;font-size:12px}.session a{color:var(--accent);text-decoration:none}.session a:hover{text-decoration:underline}
  .status{display:inline-flex;align-items:center;gap:7px;text-transform:capitalize}.dot{width:8px;height:8px;border-radius:50%;background:var(--muted)}.ready .dot,.idle .dot{background:var(--green)}.running .dot{background:var(--accent)}.loading .dot{background:var(--purple)}.dead .dot,.closed .dot{background:var(--red)}
  .held{display:inline-block;margin-left:8px;padding:2px 7px;border-radius:999px;background:rgba(249,226,175,.12);border:1px solid rgba(249,226,175,.32);color:var(--yellow);font-size:11px;font-weight:600}
  .muted{color:var(--muted)} .empty{padding:44px 20px;text-align:center;color:var(--muted)}
  @media(max-width:780px){main{padding:24px 12px 48px}.panel{overflow-x:auto}table{min-width:800px}header{padding:0 4px}}
</style>
</head>
<body><main>
<header><div><h1>Terminal Sessions</h1><p>Newest sessions are shown first. Click a session ID to open its live terminal view.</p></div><a class="refresh" href="/">Refresh</a></header>
<div class="panel">
{{if .Sessions}}
<table><thead><tr><th>Session</th><th>Status</th><th>Mode</th><th>Host</th><th>Last activity</th><th>Created</th></tr></thead><tbody>
{{range .Sessions}}<tr>
<td class="session"><a href="{{.Href}}" title="Open terminal">{{.SessionID}}</a>{{if .Held}}<span class="held">Human control</span>{{end}}</td>
<td><span class="status {{.Status}}"><span class="dot"></span>{{.Status}}</span></td>
<td>{{.Mode}}</td><td>{{if .Host}}{{.Host}}{{else}}<span class="muted">local</span>{{end}}</td><td>{{.Idle}}</td><td>{{.Created}}</td>
</tr>{{end}}
</tbody></table>
{{else}}<div class="empty">No terminal sessions are active on this instance yet.</div>{{end}}
</div>
</main></body></html>`))

// SessionIndexHandler serves a simple English session browser at GET /.
// It intentionally lists all sessions in this instance because the standalone
// server is expected to be bound to loopback or placed behind a trusted proxy.
func SessionIndexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		raw := session.List("")
		sort.SliceStable(raw, func(i, j int) bool {
			return raw[i]["created_at"] > raw[j]["created_at"]
		})

		rows := make([]sessionIndexRow, 0, len(raw))
		for _, s := range raw {
			id := s["session_id"]
			created := s["created_at"]
			if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
				created = t.Local().Format("Jan 02, 2006 03:04:05 PM")
			}
			idleSec, _ := strconv.Atoi(s["idle_seconds"])
			rows = append(rows, sessionIndexRow{
				SessionID: id,
				Href:      "/view/terminal/" + url.PathEscape(id),
				Mode:      s["mode"],
				Host:      s["host"],
				Status:    s["status"],
				Idle:      formatAge(idleSec),
				Created:   created,
				Held:      s["held"] == "true",
			})
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodHead {
			return
		}
		if err := sessionIndexTemplate.Execute(w, sessionIndexData{Sessions: rows}); err != nil {
			http.Error(w, "failed to render session index", http.StatusInternalServerError)
		}
	})
}

func formatAge(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	switch {
	case seconds < 60:
		return strconv.Itoa(seconds) + "s ago"
	case seconds < 3600:
		return strconv.Itoa(seconds/60) + "m ago"
	case seconds < 86400:
		return strconv.Itoa(seconds/3600) + "h ago"
	default:
		return strconv.Itoa(seconds/86400) + "d ago"
	}
}
