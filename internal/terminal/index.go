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
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="icon" type="image/png" href="/favicon.png">
<title>terminal-mcp sessions</title>
<style>
:root{color-scheme:dark;--bg:#11111b;--panel:#181825;--panel2:#1e1e2e;--border:#313244;--text:#cdd6f4;--muted:#a6adc8;--accent:#89b4fa;--green:#a6e3a1;--yellow:#f9e2af;--red:#f38ba8;--purple:#cba6f7}
*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--text);font:14px/1.5 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif}main{max-width:1320px;margin:0 auto;padding:40px 24px 64px}header{display:flex;align-items:flex-start;gap:16px;margin-bottom:24px}header>div{flex:1}h1{font-size:26px;margin:0 0 4px;font-weight:650;letter-spacing:-.02em}p{margin:0;color:var(--muted)}
.live{display:inline-flex;align-items:center;gap:8px;border:1px solid var(--border);background:var(--panel2);padding:9px 13px;border-radius:9px;color:var(--muted);font-size:12px}.live i{width:8px;height:8px;border-radius:50%;background:var(--green);box-shadow:0 0 8px var(--green)}
.panel{background:var(--panel);border:1px solid var(--border);border-radius:14px;overflow:hidden;box-shadow:0 14px 40px rgba(0,0,0,.24)}table{width:100%;border-collapse:collapse}th,td{padding:12px 14px;text-align:left;border-bottom:1px solid var(--border);vertical-align:middle}th{font-size:11px;text-transform:uppercase;letter-spacing:.08em;color:var(--muted);background:rgba(255,255,255,.02)}tr:last-child td{border-bottom:0}tbody tr:hover{background:rgba(137,180,250,.045)}
.session{font-family:"SFMono-Regular",Consolas,"Liberation Mono",monospace;font-size:12px}.session a{color:var(--accent);text-decoration:none}.session a:hover{text-decoration:underline}.name-input{width:190px;max-width:28vw;padding:7px 9px;border:1px solid transparent;border-radius:7px;background:transparent;color:var(--text);font:inherit}.name-input:hover{border-color:var(--border);background:rgba(255,255,255,.03)}.name-input:focus{outline:none;border-color:var(--accent);background:var(--panel2)}
.status{display:inline-flex;align-items:center;gap:7px;text-transform:capitalize}.dot{width:8px;height:8px;border-radius:50%;background:var(--muted)}.ready .dot,.idle .dot{background:var(--green)}.running .dot{background:var(--accent)}.loading .dot{background:var(--purple)}.dead .dot,.closed .dot{background:var(--red)}.held{display:inline-block;margin-left:8px;padding:2px 7px;border-radius:999px;background:rgba(249,226,175,.12);border:1px solid rgba(249,226,175,.32);color:var(--yellow);font-size:11px;font-weight:600}.muted{color:var(--muted)}.empty{padding:44px 20px;text-align:center;color:var(--muted)}
@media(max-width:900px){main{padding:24px 12px 48px}.panel{overflow-x:auto}table{min-width:980px}header{padding:0 4px}}
</style></head><body><main>
<header><div><h1>Terminal Sessions</h1><p>Session List</p></div><span class="live"><i></i><span id="liveText">Live</span></span></header>
<div class="panel"><table id="sessionTable"{{if not .Sessions}} style="display:none"{{end}}><thead><tr><th>Name</th><th>Session</th><th>Status</th><th>Mode</th><th>Host</th><th>Last activity</th><th>Created</th></tr></thead><tbody id="sessionBody">
{{range .Sessions}}<tr data-session-id="{{.SessionID}}"><td><input class="name-input" value="{{.Name}}" placeholder="Unnamed session" maxlength="120" aria-label="Session name"></td><td class="session"><a href="{{.Href}}" title="Open terminal">{{.SessionID}}</a>{{if .Held}}<span class="held">Human control</span>{{end}}</td><td><span class="status {{.Status}}"><span class="dot"></span>{{.Status}}</span></td><td>{{.Mode}}</td><td>{{if .Host}}{{.Host}}{{else}}<span class="muted">local</span>{{end}}</td><td>{{.Idle}}</td><td>{{.Created}}</td></tr>{{end}}
</tbody></table><div id="emptyState" class="empty"{{if .Sessions}} style="display:none"{{end}}>No terminal sessions are active on this instance yet.</div></div>
<script>
(function(){
  var body=document.getElementById('sessionBody'), table=document.getElementById('sessionTable'), empty=document.getElementById('emptyState'), live=document.getElementById('liveText');
  function escAge(n){n=Math.max(0,parseInt(n||'0',10)||0);if(n<60)return n+'s ago';if(n<3600)return Math.floor(n/60)+'m ago';if(n<86400)return Math.floor(n/3600)+'h ago';return Math.floor(n/86400)+'d ago'}
  function fmtCreated(v){var d=new Date(v);return isNaN(d)?v:d.toLocaleString()}
  function el(tag,cls,txt){var x=document.createElement(tag);if(cls)x.className=cls;if(txt!==undefined)x.textContent=txt;return x}
  function bindNameInput(input,id){
    if(input.dataset.bound==='1')return;input.dataset.bound='1';input.dataset.saved=input.value||'';
    input.addEventListener('change',function(){rename(id,input.value,input)});
    input.addEventListener('keydown',function(e){if(e.key==='Enter'){e.preventDefault();input.blur()}if(e.key==='Escape'){input.value=input.dataset.saved||'';input.blur()}});
  }
  function makeRow(s){
    var tr=el('tr');tr.dataset.sessionId=s.session_id;
    var tdName=el('td'), input=el('input','name-input');input.placeholder='Unnamed session';input.maxLength=120;input.setAttribute('aria-label','Session name');input.value=s.name||'';tdName.appendChild(input);tr.appendChild(tdName);
    var tdSession=el('td','session'), a=el('a','',s.session_id);a.href='/view/terminal/'+encodeURIComponent(s.session_id);a.title='Open terminal';tdSession.appendChild(a);if(s.held==='true')tdSession.appendChild(el('span','held','Human control'));tr.appendChild(tdSession);
    var tdStatus=el('td'), st=el('span','status '+s.status);st.appendChild(el('span','dot'));st.appendChild(document.createTextNode(s.status));tdStatus.appendChild(st);tr.appendChild(tdStatus);
    tr.appendChild(el('td','',s.mode||''));var h=el('td');h.appendChild(s.host?document.createTextNode(s.host):el('span','muted','local'));tr.appendChild(h);tr.appendChild(el('td','',escAge(s.idle_seconds)));tr.appendChild(el('td','',fmtCreated(s.created_at)));
    bindNameInput(input,s.session_id);return tr
  }
  function rename(id,name,input){fetch('/api/sessions',{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({session_id:id,name:name})}).then(function(r){if(!r.ok)throw new Error();return r.json()}).then(function(j){input.value=j.name||'';input.dataset.saved=input.value}).catch(function(){input.value=input.dataset.saved||''})}
  function render(list){
    var focused=document.activeElement, focusedID=focused&&focused.classList&&focused.classList.contains('name-input')?focused.closest('tr').dataset.sessionId:'';
    var current={};Array.from(body.querySelectorAll('tr')).forEach(function(r){current[r.dataset.sessionId]=r});var seen={};
    list.forEach(function(s){var r=current[s.session_id];if(!r){r=makeRow(s)}else{var inp=r.querySelector('.name-input');bindNameInput(inp,s.session_id);if(s.session_id!==focusedID){inp.value=s.name||'';inp.dataset.saved=inp.value}r.querySelector('.status').className='status '+s.status;r.querySelector('.status').lastChild.textContent=s.status;r.cells[3].textContent=s.mode||'';r.cells[4].textContent=s.host||'local';r.cells[5].textContent=escAge(s.idle_seconds);r.cells[6].textContent=fmtCreated(s.created_at);var held=r.querySelector('.held');if(s.held==='true'&&!held)r.cells[1].appendChild(el('span','held','Human control'));if(s.held!=='true'&&held)held.remove()}body.appendChild(r);seen[s.session_id]=1});
    Object.keys(current).forEach(function(id){if(!seen[id])current[id].remove()});table.style.display=list.length?'':'none';empty.style.display=list.length?'none':'';
  }
  Array.from(body.querySelectorAll('tr')).forEach(function(r){var inp=r.querySelector('.name-input');if(inp)bindNameInput(inp,r.dataset.sessionId)});
  function sync(){fetch('/api/sessions',{cache:'no-store'}).then(function(r){if(!r.ok)throw new Error();return r.json()}).then(function(j){render(j.sessions||[]);live.textContent='Live'}).catch(function(){live.textContent='Reconnecting…'})}
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
			_ = json.NewEncoder(w).Encode(map[string]any{"sessions": sortedSessions()})
		case http.MethodPatch:
			var in struct {
				SessionID string `json:"session_id"`
				Name      string `json:"name"`
			}
			if json.NewDecoder(r.Body).Decode(&in) != nil || in.SessionID == "" {
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
