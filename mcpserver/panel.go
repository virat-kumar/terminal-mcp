package mcpserver

import (
	"context"
	"fmt"
	"github.com/fzxbl/terminal-mcp/internal/session"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const panelURI = "ui://terminal-mcp/monitor-v1.html"
const panelHTML = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><style>body{font:14px system-ui;margin:0;background:#111820;color:#e4edf5}header{padding:12px;background:#202d3a;display:flex;justify-content:space-between;align-items:center}main{padding:12px}select,button{background:#314254;color:white;border:1px solid #5b6b7b;padding:7px;border-radius:5px}pre{background:#080d12;color:#b5f0c2;border-radius:6px;padding:14px;white-space:pre-wrap;overflow:auto;min-height:180px;max-height:65vh}.hint{font-size:12px;color:#aec0cd}</style></head><body><header><strong>Ubuntu Terminal Monitor</strong><span id="status">Connecting…</span></header><main><label>Session <select id="sessions"><option>Loading…</option></select></label> <button id="refresh">Refresh</button><p class="hint">Read-only monitor. No commands are sent from this panel.</p><pre id="output">Select a session to monitor.</pre></main><script>
const origin='*';let counter=1,waiting=new Map(),active='',busy=false;
function send(method,params){return new Promise((resolve,reject)=>{let id=counter++;waiting.set(id,{resolve,reject});parent.postMessage({jsonrpc:'2.0',id,method,params},origin);setTimeout(()=>{if(waiting.has(id)){waiting.delete(id);reject(Error('Host did not respond'))}},10000)})}
async function tool(name,args={}){const r=await send('tools/call',{name,arguments:args});if(r.error)throw Error(r.error.message);return r.result?.structuredContent||JSON.parse(r.result?.content?.find(c=>c.type==='text')?.text||'{}')}
window.addEventListener('message',e=>{if(e.source!==parent)return;const m=e.data;if(!m||m.jsonrpc!=='2.0')return;if(m.id&&waiting.has(m.id)){const p=waiting.get(m.id);waiting.delete(m.id);m.error?p.reject(Error(m.error.message)):p.resolve(m)} });
async function refresh(){try{const r=await tool('terminal_list');const sel=document.getElementById('sessions');const old=active;sel.textContent='';for(const s of r.sessions||[]){let o=document.createElement('option');o.value=s.session_id;o.textContent=s.name||s.session_id;sel.appendChild(o)}active=Array.from(sel.options).some(o=>o.value===old)?old:(sel.options[0]?.value||'');sel.value=active;document.getElementById('status').textContent=(r.sessions||[]).length+' sessions';await poll()}catch(e){document.getElementById('status').textContent='Unavailable';document.getElementById('output').textContent=e.message}}
async function poll(){if(!active||busy)return;busy=true;try{const r=await tool('terminal_output',{session_id:active,mode:'tail'});document.getElementById('output').textContent=r.output||'(No output)';}catch(e){document.getElementById('output').textContent=e.message}finally{busy=false}}
document.getElementById('sessions').addEventListener('change',e=>{active=e.target.value;poll()});document.getElementById('refresh').onclick=refresh;
send('ui/initialize',{protocolVersion:'2026-01-26',clientInfo:{name:'terminal-monitor',version:'1.0.0'},appCapabilities:{availableDisplayModes:['inline','fullscreen']}}).then(()=>{parent.postMessage({jsonrpc:'2.0',method:'ui/notifications/initialized'},origin);refresh();setInterval(poll,2500);setInterval(refresh,15000)}).catch(e=>document.getElementById('status').textContent=e.message);
</script></body></html>`

func registerPanel(server *mcp.Server) {
	server.AddResource(&mcp.Resource{URI: panelURI, Name: "terminal-monitor", Title: "Terminal Monitor", MIMEType: "text/html;profile=mcp-app", Description: "Read-only interactive terminal monitor"}, func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: panelURI, MIMEType: "text/html;profile=mcp-app", Text: panelHTML}}}, nil
	})
	mcp.AddTool(server, &mcp.Tool{Name: "terminal_monitor", Description: "Display an interactive read-only monitor for existing terminal sessions in ChatGPT.", Meta: mcp.Meta{"ui": map[string]any{"resourceUri": panelURI}, "openai/outputTemplate": panelURI}}, func(_ context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, listOutput, error) {
		owner, ok := ownerSig(req)
		if !ok {
			return nil, listOutput{}, fmt.Errorf("missing required identity header")
		}
		return nil, listOutput{Sessions: session.List(owner)}, nil
	})
}
