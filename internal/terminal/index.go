package terminal

import (
	"encoding/json"
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
	Name      string
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
<html lang="en" data-theme="dark"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="icon" type="image/png" href="/favicon.png">
<title>terminal-mcp sessions</title>
<script>try{var t=localStorage.getItem('terminal_mcp_index_theme');if(t==='light'||t==='dark')document.documentElement.dataset.theme=t}catch(e){}</script>
<style>
:root{color-scheme:dark;--bg:#11111b;--panel:#181825;--panel2:#1e1e2e;--border:#313244;--text:#cdd6f4;--muted:#a6adc8;--accent:#89b4fa;--green:#a6e3a1;--yellow:#f9e2af;--red:#f38ba8;--purple:#cba6f7;--hover:rgba(137,180,250,.045);--input-hover:rgba(255,255,255,.03);--thead:rgba(255,255,255,.02);--shadow:rgba(0,0,0,.24)}
html[data-theme="light"]{color-scheme:light;--bg:#f6f7fb;--panel:#ffffff;--panel2:#f0f3f9;--border:#d9dfeb;--text:#202534;--muted:#667085;--accent:#2563eb;--green:#16803b;--yellow:#9a6700;--red:#c33b52;--purple:#7c3aed;--hover:rgba(37,99,235,.055);--input-hover:rgba(32,37,52,.035);--thead:rgba(32,37,52,.025);--shadow:rgba(22,34,54,.10)}
*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--text);font:14px/1.5 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;transition:background .15s,color .15s}main{max-width:1320px;margin:0 auto;padding:40px 24px 64px}header{display:flex;align-items:flex-start;gap:16px;margin-bottom:24px}header>.title{flex:1}h1{font-size:26px;margin:0 0 4px;font-weight:650;letter-spacing:-.02em}p{margin:0;color:var(--muted)}
.controls{display:flex;align-items:center;justify-content:flex-end;gap:9px;flex-wrap:wrap}.live,.control{display:inline-flex;align-items:center;gap:8px;border:1px solid var(--border);background:var(--panel2);padding:9px 12px;border-radius:9px;color:var(--muted);font-size:12px;min-height:38px}.live i{width:8px;height:8px;border-radius:50%;background:var(--green);box-shadow:0 0 8px var(--green)}.theme-btn{border:1px solid var(--border);background:var(--panel2);color:var(--text);border-radius:9px;padding:9px 12px;min-height:38px;cursor:pointer;font:600 12px/1 inherit}.theme-btn:hover{border-color:var(--accent)}
.switch{position:relative;display:inline-block;width:36px;height:20px}.switch input{opacity:0;width:0;height:0}.slider{position:absolute;cursor:pointer;inset:0;background:var(--border);border-radius:999px;transition:.18s}.slider:before{content:"";position:absolute;width:14px;height:14px;left:3px;top:3px;background:#fff;border-radius:50%;transition:.18s;box-shadow:0 1px 3px rgba(0,0,0,.3)}.switch input:checked+.slider{background:var(--accent)}.switch input:checked+.slider:before{transform:translateX(16px)}.switch input:disabled+.slider{opacity:.55;cursor:wait}.control strong{font-weight:650;color:var(--text)}
.panel{background:var(--panel);border:1px solid var(--border);border-radius:14px;overflow:hidden;box-shadow:0 14px 40px var(--shadow)}table{width:100%;border-collapse:collapse}th,td{padding:12px 14px;text-align:left;border-bottom:1px solid var(--border);vertical-align:middle}th{font-size:11px;text-transform:uppercase;letter-spacing:.08em;color:var(--muted);background:var(--thead)}tr:last-child td{border-bottom:0}tbody tr:hover{background:var(--hover)}
.session{font-family:"SFMono-Regular",Consolas,"Liberation Mono",monospace;font-size:12px}.session a{color:var(--accent);text-decoration:none}.session a:hover{text-decoration:underline}.name-input{width:190px;max-width:28vw;padding:7px 9px;border:1px solid transparent;border-radius:7px;background:transparent;color:var(--text);font:inherit}.name-input:hover{border-color:var(--border);background:var(--input-hover)}.name-input:focus{outline:none;border-color:var(--accent);background:var(--panel2)}
.status{display:inline-flex;align-items:center;gap:7px;text-transform:capitalize}.dot{width:8px;height:8px;border-radius:50%;background:var(--muted)}.ready .dot,.idle .dot{background:var(--green)}.running .dot{background:var(--accent)}.loading .dot{background:var(--purple)}.dead .dot,.closed .dot{background:var(--red)}.held{display:inline-block;margin-left:8px;padding:2px 7px;border-radius:999px;background:rgba(249,226,175,.12);border:1px solid rgba(249,226,175,.32);color:var(--yellow);font-size:11px;font-weight:600}.muted{color:var(--muted)}.empty{padding:44px 20px;text-align:center;color:var(--muted)}
@media(max-width:900px){main{padding:24px 12px 48px}.panel{overflow-x:auto}table{min-width:980px}header{padding:0 4px;flex-direction:column}.controls{justify-content:flex-start}}
</style></head><body><main>
<header><div class="title"><h1>Terminal Sessions</h1><p>Session List</p></div><div class="controls"><label class="control" title="Keep exactly one persistent terminal. New terminal_open calls reuse it, and its shell is restarted automatically if it dies."><span><strong>Force one terminal</strong></span><span class="switch"><input id="oneTerminal" type="checkbox" aria-label="Force one terminal"><span class="slider"></span></span></label><button id="themeToggle" class="theme-btn" type="button">Light mode</button><span class="live"><i></i><span id="liveText">Live</span></span></div></header>
<div class="panel"><table id="sessionTable"{{if not .Sessions}} style="display:none"{{end}}><thead><tr><th>Name</th><th>Session</th><th>Status</th><th>Mode</th><th>Host</th><th>Last activity</th><th>Created</th></tr></thead><tbody id="sessionBody">
{{range .Sessions}}<tr data-session-id="{{.SessionID}}"><td><input class="name-input" value="{{.Name}}" placeholder="Unnamed session" maxlength="120" aria-label="Session name"></td><td class="session"><a href="{{.Href}}" title="Open terminal">{{.SessionID}}</a>{{if .Held}}<span class="held">Human control</span>{{end}}</td><td><span class="status {{.Status}}"><span class="dot"></span>{{.Status}}</span></td><td>{{.Mode}}</td><td>{{if .Host}}{{.Host}}{{else}}<span class="muted">local</span>{{end}}</td><td>{{.Idle}}</td><td>{{.Created}}</td></tr>{{end}}
</tbody></table><div id="emptyState" class="empty"{{if .Sessions}} style="display:none"{{end}}>No terminal sessions are active on this instance yet.</div></div>
<script>
(function(){
  var body=document.getElementById('sessionBody'), table=document.getElementById('sessionTable'), empty=document.getElementById('emptyState'), live=document.getElementById('liveText'), one=document.getElementById('oneTerminal'), theme=document.getElementById('themeToggle');
  function currentTheme(){return document.documentElement.dataset.theme==='light'?'light':'dark'}
  function applyTheme(v){document.documentElement.dataset.theme=v;theme.textContent=v==='light'?'Dark mode':'Light mode';try{localStorage.setItem('terminal_mcp_index_theme',v)}catch(e){}}
  theme.addEventListener('click',function(){applyTheme(currentTheme()==='light'?'dark':'light')});applyTheme(currentTheme());
  function escAge(n){n=Math.max(0,parseInt(n||'0',10)||0);if(n<60)return n+'s ago';if(n<3600)return Math.floor(n/60)+'m ago';if(n<86400)return Math.floor(n/3600)+'h ago';return Math.floor(n/86400)+'d ago'}
  function fmtCreated(v){var d=new Date(v);return isNaN(d)?v:d.toLocaleString()}
  function el(tag,cls,txt){var x=document.createElement(tag);if(cls)x.className=cls;if(txt!==undefined)x.textContent=txt;return x}
  function rankByActivity(list){return list.slice().sort(function(a,b){var ai=parseInt(a.idle_seconds||'0',10)||0,bi=parseInt(b.idle_seconds||'0',10)||0;if(ai!==bi)return ai-bi;return String(b.created_at||'').localeCompare(String(a.created_at||''))})}
  function bindNameInput(input,id){if(input.dataset.bound==='1')return;input.dataset.bound='1';input.dataset.saved=input.value||'';input.addEventListener('change',function(){rename(id,input.value,input)});input.addEventListener('keydown',function(e){if(e.key==='Enter'){e.preventDefault();input.blur()}if(e.key==='Escape'){input.value=input.dataset.saved||'';input.blur()}})}
  function makeRow(s){var tr=el('tr');tr.dataset.sessionId=s.session_id;var tdName=el('td'), input=el('input','name-input');input.placeholder='Unnamed session';input.maxLength=120;input.setAttribute('aria-label','Session name');input.value=s.name||'';tdName.appendChild(input);tr.appendChild(tdName);var tdSession=el('td','session'), a=el('a','',s.session_id);a.href='/view/terminal/'+encodeURIComponent(s.session_id);a.title='Open terminal';tdSession.appendChild(a);if(s.held==='true')tdSession.appendChild(el('span','held','Human control'));tr.appendChild(tdSession);var tdStatus=el('td'), st=el('span','status '+s.status);st.appendChild(el('span','dot'));st.appendChild(document.createTextNode(s.status));tdStatus.appendChild(st);tr.appendChild(tdStatus);tr.appendChild(el('td','',s.mode||''));var h=el('td');h.appendChild(s.host?document.createTextNode(s.host):el('span','muted','local'));tr.appendChild(h);tr.appendChild(el('td','',escAge(s.idle_seconds)));tr.appendChild(el('td','',fmtCreated(s.created_at)));bindNameInput(input,s.session_id);return tr}
  function rename(id,name,input){fetch('/api/sessions',{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({session_id:id,name:name})}).then(function(r){if(!r.ok)throw new Error();return r.json()}).then(function(j){input.value=j.name||'';input.dataset.saved=input.value}).catch(function(){input.value=input.dataset.saved||''})}
  function render(list){list=rankByActivity(list);var focused=document.activeElement, focusedID=focused&&focused.classList&&focused.classList.contains('name-input')?focused.closest('tr').dataset.sessionId:'';var current={};Array.from(body.querySelectorAll('tr')).forEach(function(r){current[r.dataset.sessionId]=r});var seen={};list.forEach(function(s){var r=current[s.session_id];if(!r){r=makeRow(s)}else{var inp=r.querySelector('.name-input');bindNameInput(inp,s.session_id);if(s.session_id!==focusedID){inp.value=s.name||'';inp.dataset.saved=inp.value}r.querySelector('.status').className='status '+s.status;r.querySelector('.status').lastChild.textContent=s.status;r.cells[3].textContent=s.mode||'';r.cells[4].textContent=s.host||'local';r.cells[5].textContent=escAge(s.idle_seconds);r.cells[6].textContent=fmtCreated(s.created_at);var held=r.querySelector('.held');if(s.held==='true'&&!held)r.cells[1].appendChild(el('span','held','Human control'));if(s.held!=='true'&&held)held.remove()}body.appendChild(r);seen[s.session_id]=1});Object.keys(current).forEach(function(id){if(!seen[id])current[id].remove()});table.style.display=list.length?'':'none';empty.style.display=list.length?'none':''}
  one.addEventListener('change',function(){var desired=one.checked;one.disabled=true;live.textContent='Applying…';fetch('/api/sessions',{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({single_terminal_mode:desired})}).then(function(r){if(!r.ok)throw new Error();return r.json()}).then(function(j){one.checked=!!j.single_terminal_mode;render(j.sessions||[]);live.textContent='Live'}).catch(function(){one.checked=!desired;live.textContent='Setting failed'}).then(function(){one.disabled=false})});
  Array.from(body.querySelectorAll('tr')).forEach(function(r){var inp=r.querySelector('.name-input');if(inp)bindNameInput(inp,r.dataset.sessionId)});
  function sync(){fetch('/api/sessions',{cache:'no-store'}).then(function(r){if(!r.ok)throw new Error();return r.json()}).then(function(j){render(j.sessions||[]);if(!one.disabled)one.checked=!!j.single_terminal_mode;live.textContent='Live'}).catch(function(){live.textContent='Reconnecting…'})}
  sync();setInterval(sync,2000);
})();
</script></main></body></html>`))

func sortedSessions() []map[string]string {
	raw := session.List("")
	sort.SliceStable(raw, func(i, j int) bool { return raw[i]["created_at"] > raw[j]["created_at"] })
	return raw
}

func indexRows(raw []map[string]string) []sessionIndexRow {
	rows := make([]sessionIndexRow, 0, len(raw))
	for _, s := range raw {
		created := s["created_at"]
		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			created = t.Local().Format("Jan 02, 2006 03:04:05 PM")
		}
		idleSec, _ := strconv.Atoi(s["idle_seconds"])
		rows = append(rows, sessionIndexRow{SessionID: s["session_id"], Name: s["name"], Href: "/view/terminal/" + url.PathEscape(s["session_id"]), Mode: s["mode"], Host: s["host"], Status: s["status"], Idle: formatAge(idleSec), Created: created, Held: s["held"] == "true"})
	}
	return rows
}

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
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodHead {
			return
		}
		if err := sessionIndexTemplate.Execute(w, sessionIndexData{Sessions: indexRows(sortedSessions())}); err != nil {
			http.Error(w, "failed to render session index", http.StatusInternalServerError)
		}
	})
}

func SessionAPIHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"sessions": sortedSessions(), "single_terminal_mode": session.SingleTerminalMode()})
		case http.MethodPatch:
			var in struct {
				SessionID          string `json:"session_id"`
				Name               string `json:"name"`
				SingleTerminalMode *bool  `json:"single_terminal_mode"`
			}
			if json.NewDecoder(r.Body).Decode(&in) != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if in.SingleTerminalMode != nil {
				if err := session.SetSingleTerminalMode(*in.SingleTerminalMode); err != nil {
					http.Error(w, "failed to update single-terminal mode", http.StatusInternalServerError)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"sessions": sortedSessions(), "single_terminal_mode": session.SingleTerminalMode()})
				return
			}
			if in.SessionID == "" {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			name, ok := session.Rename(in.SessionID, in.Name)
			if !ok {
				http.Error(w, "session not found", http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"session_id": in.SessionID, "name": name})
		default:
			w.Header().Set("Allow", "GET, PATCH")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func formatAge(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	if seconds < 60 {
		return strconv.Itoa(seconds) + "s ago"
	}
	if seconds < 3600 {
		return strconv.Itoa(seconds/60) + "m ago"
	}
	if seconds < 86400 {
		return strconv.Itoa(seconds/3600) + "h ago"
	}
	return strconv.Itoa(seconds/86400) + "d ago"
}
