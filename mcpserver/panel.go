package mcpserver

import (
	"context"

	"github.com/fzxbl/terminal-mcp/internal/session"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const panelURI = "ui://terminal-mcp/monitor-v3.html"

// The app is intentionally read-only. It never sends input or takeover commands.
const panelHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>
body{font:14px system-ui;margin:0;background:#111820;color:#e4edf5}
header{padding:12px;background:#202d3a;display:flex;justify-content:space-between;align-items:center}
main{padding:12px} .controls{display:flex;flex-wrap:wrap;gap:6px;align-items:center}
select,button{background:#314254;color:white;border:1px solid #5b6b7b;padding:7px;border-radius:5px;font:inherit}
select{max-width:min(72vw,400px)} button{cursor:pointer}
button:disabled{opacity:.65;cursor:wait}button[aria-pressed="true"]{border-color:#b5f0c2}
pre{background:#080d12;color:#b5f0c2;border-radius:6px;padding:14px;white-space:pre;overflow:auto;min-height:180px;height:58vh;max-height:72vh;box-sizing:border-box}
.hint{font-size:12px;color:#aec0cd}
</style></head><body>
<header><strong>Ubuntu Terminal Monitor</strong><span id="status">Connecting...</span></header>
<main><div class="controls"><label for="sessions">Session</label>
<select id="sessions"><option>Loading...</option></select>
<button id="refresh" type="button">Refresh</button>
<button id="follow" type="button" aria-pressed="true">Follow: On</button></div>
<p class="hint">Read-only monitor. No commands are sent from this panel.</p>
<pre id="output" role="log" aria-live="off">Select a session to monitor.</pre></main>
<script>
const origin='*';
let counter=1, waiting=new Map(), active='', agentActive='', manuallySelected=false;
let known=new Set(), initializedList=false, follow=true, outputRequest=0, listing=null;
let preferred='', lastOutput='', sessionKeys={};
const select=document.getElementById('sessions'), out=document.getElementById('output');
const status=document.getElementById('status'), refreshButton=document.getElementById('refresh');
const followButton=document.getElementById('follow');
function send(method,params){
  return new Promise((resolve,reject)=>{
    const id=counter++;
    waiting.set(id,{resolve,reject});
    parent.postMessage({jsonrpc:'2.0',id,method,params},origin);
    setTimeout(()=>{if(waiting.has(id)){waiting.delete(id);reject(Error('Host did not respond'))}},10000);
  });
}
async function tool(name,args={}){
  const r=await send('tools/call',{name,arguments:args});
  if(r.error)throw Error(r.error.message);
  if(r.result?.isError)throw Error(r.result?.content?.find(c=>c.type==='text')?.text||'Tool error');
  return r.result?.structuredContent||JSON.parse(r.result?.content?.find(c=>c.type==='text')?.text||'{}');
}
function scrollToEnd(){if(follow){out.scrollTop=out.scrollHeight;out.scrollLeft=0}}
function setFollow(value){
  follow=value;
  followButton.textContent='Follow: '+(follow?'On':'Off');
  followButton.setAttribute('aria-pressed',String(follow));
  scrollToEnd();
}
function showSession(id){
  if(id===active)return false;
  active=id;
  select.value=active;
  lastOutput='';
  out.textContent=active?'Loading terminal...':'No active sessions.';
  outputRequest++;
  scrollToEnd();
  return true;
}
function initialToolSession(params){
  const obj=params?.structuredContent||params?.result?.structuredContent||params;
  if(obj&&typeof obj.session_id==='string'&&obj.session_id){
    if(obj.session_key)sessionKeys[obj.session_id]=obj.session_key;
    preferred=obj.session_id;manuallySelected=false;refresh();
  }
}
window.addEventListener('message',e=>{
  if(e.source!==parent)return;
  const m=e.data;
  if(!m||m.jsonrpc!=='2.0')return;
  if(m.id&&waiting.has(m.id)){
    const p=waiting.get(m.id);waiting.delete(m.id);
    m.error?p.reject(Error(m.error.message)):p.resolve(m);
  }else if(m.method==='ui/notifications/tool-result'){
    initialToolSession(m.params);
  }
});
async function poll(force=false){
  const sid=active;
  if(!sid){out.textContent='No active sessions.';return}
  const request=++outputRequest;
  try{
    if(!sessionKeys[sid]){
      out.textContent='Read-only output requires the session capability from terminal_open or terminal_monitor.';
      return;
    }
    const r=await tool('terminal_output',{session_id:sid,session_key:sessionKeys[sid],mode:'tail'});
    if(request!==outputRequest||sid!==active)return;
    if(r.error)throw Error(r.error);
    const next=r.output||'(No output)';
    if(force||next!==lastOutput){
      const wasAtEnd=out.scrollTop+out.clientHeight>=out.scrollHeight-20;
      out.textContent=next;
      lastOutput=next;
      // Follow defaults on; when disabled preserve user's scroll position.
      if(follow)scrollToEnd();
      else if(wasAtEnd)out.scrollTop=out.scrollHeight;
    }
  }catch(e){
    if(request===outputRequest&&sid===active)out.textContent=e.message;
  }
}
async function refresh(){
  if(listing)return listing;
  listing=(async()=>{
    refreshButton.disabled=true;
    try{
      const r=await tool('terminal_list');
      const sessions=r.sessions||[];
      if(r.session_id&&r.session_key)sessionKeys[r.session_id]=r.session_key;
      const ids=new Set(sessions.map(s=>s.session_id));
      const fresh=sessions.filter(s=>!known.has(s.session_id));
      const suggested=r.active_session_id||'';
      const hadList=initializedList;
      agentActive=suggested;
      select.replaceChildren();
      for(const s of sessions){
        const option=document.createElement('option');
        option.value=s.session_id;
        option.textContent=s.name||s.session_id;
        select.appendChild(option);
      }
      let target=active;
      if(preferred&&ids.has(preferred)){
        target=preferred;preferred='';manuallySelected=false;
      }else if(!hadList){
        target=(suggested&&ids.has(suggested)?suggested:sessions[0]?.session_id)||'';
      }else if(fresh.length&&suggested&&fresh.some(s=>s.session_id===suggested)){
        // A newly opened agent terminal takes precedence, even over a manual selection.
        target=suggested;manuallySelected=false;
      }else if(!manuallySelected&&suggested&&ids.has(suggested)){
        target=suggested;
      }else if(!ids.has(target)){
        target=(suggested&&ids.has(suggested)?suggested:sessions[0]?.session_id)||'';
        manuallySelected=false;
      }
      showSession(target);
      select.value=active;
      known=ids;initializedList=true;
      status.textContent=sessions.length+' session'+(sessions.length===1?'':'s');
      await poll(true);
    }catch(e){
      status.textContent='Unavailable';
      out.textContent='Refresh failed: '+e.message;
    }finally{
      refreshButton.disabled=false;
    }
  })();
  try{await listing}finally{listing=null}
}
select.addEventListener('change',e=>{manuallySelected=true;showSession(e.target.value);poll(true)});
refreshButton.addEventListener('click',()=>refresh());
followButton.addEventListener('click',()=>setFollow(!follow));
send('ui/initialize',{protocolVersion:'2026-01-26',
  appInfo:{name:'terminal-monitor',version:'1.3.0'},
  appCapabilities:{availableDisplayModes:['inline','fullscreen']}})
.then(()=>{
  parent.postMessage({jsonrpc:'2.0',method:'ui/notifications/initialized'},origin);
  refresh();
  setInterval(()=>{if(!listing)refresh()},3000);
  setInterval(()=>{if(!listing)poll()},1800);
}).catch(e=>status.textContent=e.message);
</script></body></html>`

func panelMeta() mcp.Meta {
	return mcp.Meta{
		"ui":                    map[string]any{"resourceUri": panelURI},
		"openai/outputTemplate": panelURI,
		"openai/ui":             map[string]any{"entrypoints": []map[string]any{{"type": "thread"}}},
	}
}

func registerPanel(server *mcp.Server) {
	// Older ChatGPT plugin registrations may retain v1 or v2 resource URIs.
	// Always serve the current Follow-enabled UI under these compatibility URIs.
	for _, resourceURI := range []string{panelURI,
		"ui://terminal-mcp/monitor-v1.html",
		"ui://terminal-mcp/monitor-v2.html"} {
		uri := resourceURI
		server.AddResource(&mcp.Resource{
			URI: uri, Name: "terminal-monitor", Title: "Terminal Monitor",
			MIMEType:    "text/html;profile=mcp-app",
			Description: "Read-only terminal monitor with Follow and Refresh",
		}, func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
				URI: uri, MIMEType: "text/html;profile=mcp-app", Text: panelHTML,
			}}}, nil
		})
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "terminal_monitor",
		Description: "Show the read-only live terminal monitor in ChatGPT. Automatically selects the agent's active terminal and follows its output at the bottom by default; Refresh reloads both sessions and output. New agent-created terminals are automatically selected.",
		Meta:        panelMeta(),
	}, func(_ context.Context, req *mcp.CallToolRequest, in monitorInput) (*mcp.CallToolResult, monitorOutput, error) {
		owner := mcpUser(reqHeader(req))
		out := monitorOutput{Sessions: session.List(), ActiveSessionID: activeAgentSession(owner)}
		if in.SessionID != "" && session.Authorize(in.SessionID, in.SessionKey) {
			out.SessionID, out.SessionKey = in.SessionID, in.SessionKey
			out.ActiveSessionID = in.SessionID
		}
		return nil, out, nil
	})
}
