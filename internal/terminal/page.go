package terminal

// terminalPageHTML is the browser terminal page. __SESSION_ID__ is replaced with the session ID.
// xterm.js renders the raw bytes received over SSE, including cursor movement, line editing, colors,
// cursor positioning, and alternate-screen apps such as vim/top/less. Take Control enables keyboard input.
// xterm onData sends keystrokes back to the PTY over WebSocket and keeps terminal size synchronized.
// xterm assets are loaded from a CDN; model-side output cleaning is handled separately in the backend.
const terminalPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="icon" type="image/png" href="/favicon.png">
<title>terminal-mcp terminal __SESSION_ID__</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@xterm/xterm@6.0.0/css/xterm.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@fontsource/jetbrains-mono@5/index.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@fontsource/ibm-plex-mono@5/index.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@fontsource/fira-code@5/index.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@fontsource/hack@5/index.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@fontsource/source-code-pro@5/index.css">
<style>
  :root{
    --bg:#1e1e2e; --bar:#181825; --termbg:#1e1e2e; --border:#313244;
    --text:#cdd6f4; --muted:#7f849c; --accent:#89b4fa; --glow:rgba(137,180,250,0.16);
    --green:#a6e3a1; --amber:#f9e2af; --purple:#cba6f7; --danger:#f38ba8;
    --h:54px; --radius:12px; --ctlh:32px;
    --font-ui:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;
    --font-mono:"JetBrains Mono",ui-monospace,Menlo,Consolas,monospace;
  }
  *{box-sizing:border-box}
  html,body{margin:0;height:100%;background:var(--bg);color:var(--text);font-family:var(--font-ui);
    font-size:13px;-webkit-font-smoothing:antialiased;overflow:hidden}
  body::before{content:"";position:fixed;inset:0;z-index:0;pointer-events:none;
    background:radial-gradient(1200px 560px at 18% -12%, var(--glow), transparent 60%)}

  #bar{position:fixed;top:0;left:0;right:0;height:var(--h);z-index:20;display:flex;align-items:center;
    gap:10px;padding:0 16px;background:var(--bar);border-bottom:1px solid var(--border)}
  #bar::after{content:"";position:absolute;left:0;right:0;bottom:-1px;height:1px;opacity:.55;
    background:linear-gradient(90deg,transparent,var(--accent),transparent)}
  .brand{display:inline-flex;align-items:center;gap:8px;color:var(--accent);flex:0 0 auto}
  .brand svg{width:18px;height:18px}
  .brand b{color:var(--text);font-weight:600;font-size:13px;letter-spacing:.2px}
  .brand b span{color:var(--muted);font-weight:500}

  .chip{height:var(--ctlh);display:inline-flex;align-items:center;gap:7px;padding:0 12px;flex:0 0 auto;
    background:rgba(255,255,255,0.05);border:1px solid var(--border);border-radius:9px;
    color:var(--text);font-size:12px;white-space:nowrap}
  .mono{font-family:var(--font-mono)}
  #dot{width:8px;height:8px;border-radius:50%;background:var(--muted);flex:0 0 auto;transition:background .2s}
  #state{font-size:12px;color:var(--text);text-transform:capitalize}

  .sid{display:inline-flex;align-items:center;gap:3px;min-width:0;flex:0 1 auto;color:var(--muted);
    font-family:var(--font-mono);font-size:12px}
  .sid .val{max-width:210px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
  .session-name{height:var(--ctlh);width:180px;max-width:20vw;padding:0 10px;border:1px solid var(--border);border-radius:9px;background:rgba(255,255,255,0.04);color:var(--text);font:12px/1 var(--font-ui)}
  .session-name::placeholder{color:var(--muted)}
  .session-name:focus{outline:2px solid var(--accent);outline-offset:1px;border-color:var(--accent)}
  .iconbtn{display:inline-flex;align-items:center;justify-content:center;width:28px;height:28px;
    border:none;background:transparent;color:var(--muted);cursor:pointer;border-radius:7px;transition:.15s}
  .iconbtn:hover{background:rgba(255,255,255,0.08);color:var(--text)}
  .iconbtn:focus-visible{outline:2px solid var(--accent);outline-offset:1px}
  .iconbtn svg{width:15px;height:15px}
  .spacer{margin-left:auto}

  .held-chip{display:none;align-items:center;gap:7px;height:var(--ctlh);padding:0 12px;border-radius:9px;flex:0 0 auto;
    white-space:nowrap;font-size:12px;font-weight:600;color:#fff;background:rgba(255,84,112,0.16);border:1px solid rgba(255,84,112,0.5)}
  .held-chip svg{width:8px;height:8px;fill:var(--danger);animation:blink 1.15s steps(1) infinite}
  body.held .held-chip{display:inline-flex}

  select.chip{appearance:none;-webkit-appearance:none;cursor:pointer;padding-right:30px;font-family:var(--font-mono);
    background-image:url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12' fill='none' stroke='%238a90a6' stroke-width='1.5' stroke-linecap='round'><path d='M2.5 4.5l3.5 3 3.5-3'/></svg>");
    background-repeat:no-repeat;background-position:right 10px center}
  select.chip:hover{border-color:var(--accent)}
  select.chip:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
  select.chip option{background:var(--bar);color:var(--text);font-family:var(--font-ui)}

  .stepper{height:var(--ctlh);display:inline-flex;align-items:center;overflow:hidden;flex:0 0 auto;
    background:rgba(255,255,255,0.05);border:1px solid var(--border);border-radius:9px}
  .stepper button{width:30px;height:100%;border:none;background:transparent;color:var(--text);cursor:pointer;
    font-size:16px;line-height:1;display:inline-flex;align-items:center;justify-content:center;transition:.15s}
  .stepper button:hover{background:rgba(255,255,255,0.09)}
  .stepper button:focus-visible{outline:2px solid var(--accent);outline-offset:-2px}
  #fsize{min-width:30px;text-align:center;font:12px/1 var(--font-mono);color:var(--muted)}

  #size{padding:0 11px;height:var(--ctlh);display:inline-flex;align-items:center;color:var(--muted);flex:0 0 auto;
    font:11px/1 var(--font-mono);border:1px solid var(--border);border-radius:9px;background:rgba(255,255,255,0.03)}

  #takeover{height:var(--ctlh);display:inline-flex;align-items:center;gap:7px;padding:0 14px;cursor:pointer;
    flex:0 0 auto;white-space:nowrap;font-size:12px;font-weight:600;color:#0b0f18;background:var(--accent);
    border:1px solid transparent;border-radius:9px;
    box-shadow:0 1px 3px rgba(0,0,0,0.35);transition:filter .15s,background .15s,color .15s}
  #takeover:hover{filter:brightness(1.08)}
  #takeover:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
  #takeover svg{width:13px;height:13px;flex:0 0 auto}
  body.held #takeover{background:var(--danger);color:#fff}
  /* Another operator has control: keep the button disabled to prevent a second takeover. */
  #takeover:disabled{cursor:not-allowed;background:rgba(255,255,255,0.06);color:var(--muted);
    border-color:var(--border);box-shadow:none;filter:none}

  #term{position:fixed;top:calc(var(--h) + 14px);left:16px;right:16px;bottom:16px;z-index:10;
    padding:12px 14px;background:var(--termbg);border:1px solid var(--border);border-radius:var(--radius);
    box-shadow:0 12px 40px rgba(0,0,0,0.4), inset 0 0 0 1px rgba(255,255,255,0.02);overflow:hidden}
  /* Mount xterm in a padding-free inner container so FitAddon calculates rows from the true content height. */
  #termbody{width:100%;height:100%}

  .s-loading{background:var(--purple)!important;box-shadow:0 0 8px var(--purple)}
  .s-running{background:var(--accent)!important;box-shadow:0 0 8px var(--accent);animation:pulse 1.4s ease-in-out infinite}
  .s-idle{background:var(--green)!important;box-shadow:0 0 8px var(--green)}
  .s-exited{background:var(--amber)!important}
  .s-dead{background:var(--danger)!important}
  .s-disconnected{background:var(--muted)!important;box-shadow:none}
  .s-reconnecting{background:var(--amber)!important;animation:pulse 1s ease-in-out infinite}
  @keyframes pulse{0%,100%{opacity:1}50%{opacity:.35}}
  @keyframes blink{0%,100%{opacity:1}50%{opacity:.2}}
  @media (prefers-reduced-motion:reduce){*{animation:none!important}}

  /* On narrow screens, hide secondary controls while keeping status and takeover actions fully visible. */
  @media (max-width:900px){ #size{display:none} }
  @media (max-width:800px){ #fontsel{display:none} }
  @media (max-width:680px){ #themesel{display:none} .stepper{display:none} .session-name{display:none} }
  @media (max-width:520px){ .sid{display:none} .brand b{display:none} }
</style>
</head>
<body>
<div id="bar">
  <span class="brand" aria-hidden="true">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 5l7 7-7 7"/><path d="M13 19h7"/></svg>
    <b>terminal<span>-mcp</span></b>
  </span>
  <span class="chip"><span id="dot"></span><span id="state">…</span></span>
  <input id="sessionName" class="session-name" type="text" maxlength="120" placeholder="Unnamed session" aria-label="Session name" title="Edit session name">
  <span class="sid" title="__SESSION_ID__">
    <span class="val">__SESSION_ID__</span>
    <button class="iconbtn" id="copyid" title="Copy session ID" aria-label="Copy session ID">
      <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4"><rect x="5.5" y="5.5" width="8" height="8" rx="1.5"/><path d="M10.5 5.5V3.5A1.5 1.5 0 0 0 9 2H3.5A1.5 1.5 0 0 0 2 3.5V9a1.5 1.5 0 0 0 1.5 1.5h2"/></svg>
    </button>
  </span>
  <span class="spacer"></span>
  <span class="held-chip"><svg viewBox="0 0 16 16"><circle cx="8" cy="8" r="6"/></svg>Human control active</span>
  <select id="themesel" class="chip" title="Color theme" aria-label="Color theme">
    <option value="catppuccin-mocha">Catppuccin Mocha</option>
    <option value="tokyo-night">Tokyo Night</option>
    <option value="one-dark">One Dark</option>
    <option value="gruvbox-dark">Gruvbox Dark</option>
    <option value="night-owl">Night Owl</option>
    <option value="nord">Nord</option>
    <option value="ayu-mirage">Ayu Mirage</option>
  </select>
  <select id="fontsel" class="chip" title="Terminal font" aria-label="Terminal font">
    <option value="jetbrains">JetBrains Mono</option>
    <option value="ibmplex">IBM Plex Mono</option>
    <option value="fira">Fira Code</option>
    <option value="hack">Hack</option>
    <option value="source">Source Code Pro</option>
    <option value="consolas">Consolas</option>
    <option value="cascadia">Cascadia Code</option>
    <option value="menlo">Menlo / SF Mono</option>
    <option value="dejavu">DejaVu Sans Mono</option>
  </select>
  <span class="stepper">
    <button id="fdec" title="Decrease font size" aria-label="Decrease font size">&#8722;</button>
    <span id="fsize">14</span>
    <button id="finc" title="Increase font size" aria-label="Increase font size">+</button>
  </span>
  <span id="size" title="Terminal columns × rows"></span>
  <button id="takeover" aria-label="Toggle human control">
    <svg viewBox="0 0 16 16" fill="currentColor"><path d="M5 3.5v9l7-4.5z"/></svg>Take Control
  </button>
</div>
<div id="term"><div id="termbody"></div></div>
<script src="https://cdn.jsdelivr.net/npm/@xterm/xterm@6.0.0/lib/xterm.js"></script>
<script src="https://cdn.jsdelivr.net/npm/@xterm/addon-fit@0.11.0/lib/addon-fit.js"></script>
<script src="https://cdn.jsdelivr.net/npm/@xterm/addon-webgl@0.19.0/lib/addon-webgl.js"></script>
<script>
(function(){
  var id = "__SESSION_ID__";
  // This page is .../terminal/<id>; stream/takeover/ws endpoints live under the same path.
  // Derive the base from location.pathname so mounting prefixes such as /view remain transparent.
  var base = location.pathname.replace(/\/+$/, "");
  var nameInput = document.getElementById("sessionName");
  var savedSessionName = "";
  function loadSessionName(){
    fetch(base+"/name", {cache:"no-store"}).then(function(r){ if(!r.ok) throw new Error(); return r.json(); }).then(function(j){
      savedSessionName = j.name || "";
      if(document.activeElement !== nameInput) nameInput.value = savedSessionName;
    }).catch(function(){});
  }
  function saveSessionName(){
    fetch(base+"/name", {method:"PATCH", headers:{"Content-Type":"application/json"}, body:JSON.stringify({name:nameInput.value})})
      .then(function(r){ if(!r.ok) throw new Error(); return r.json(); })
      .then(function(j){ savedSessionName=j.name||""; nameInput.value=savedSessionName; })
      .catch(function(){ nameInput.value=savedSessionName; });
  }
  if(nameInput){
    nameInput.addEventListener("change", saveSessionName);
    nameInput.addEventListener("keydown", function(e){
      if(e.key === "Enter"){ e.preventDefault(); nameInput.blur(); }
      if(e.key === "Escape"){ nameInput.value=savedSessionName; nameInput.blur(); }
    });
    loadSessionName();
    setInterval(loadSessionName, 3000);
  }
  var term = new Terminal({
    cursorBlink: true, convertEol: false, scrollback: 5000, lineHeight: 1.2, letterSpacing: 0,
    fontFamily: '"JetBrains Mono",ui-monospace,Menlo,Consolas,monospace', fontSize: 14
  });
  var fit = new FitAddon.FitAddon();
  term.loadAddon(fit);
  term.open(document.getElementById("termbody"));
  // Renderer: DOM by default for crisp native browser text rendering.
  // If default_renderer=webgl, use GPU rendering and fall back to DOM if unavailable or context is lost.
  if("__DEFAULT_RENDERER__" === "webgl"){
    try{
      if(window.WebglAddon){
        var webgl = new WebglAddon.WebglAddon();
        webgl.onContextLoss(function(){ try{ webgl.dispose(); }catch(e){} });
        term.loadAddon(webgl);
      }
    }catch(e){}
  }
  var sizeEl = document.getElementById("size");
  function showSize(){ if(sizeEl) sizeEl.textContent = term.cols + "×" + term.rows; }
  function refit(){ try{ fit.fit(); }catch(e){} showSize(); }
  // Fit only after initial layout and fonts settle so the PTY receives the correct dimensions.
  refit();
  setTimeout(refit, 0);
  setTimeout(refit, 200);

  // Ctrl/Cmd+C copies when text is selected; otherwise it still sends ^C. Ctrl/Cmd+V pastes.
  function fallbackCopy(t){
    try{ var ta=document.createElement("textarea"); ta.value=t; ta.style.position="fixed"; ta.style.opacity="0";
      document.body.appendChild(ta); ta.focus(); ta.select(); document.execCommand("copy"); document.body.removeChild(ta);
    }catch(e){}
  }
  function copyText(t){
    if(navigator.clipboard && navigator.clipboard.writeText){ navigator.clipboard.writeText(t).catch(function(){ fallbackCopy(t); }); }
    else { fallbackCopy(t); }
  }
  term.attachCustomKeyEventHandler(function(e){
    if(e.type !== "keydown") return true;
    var mod = e.ctrlKey || e.metaKey;
    if(mod && !e.shiftKey && (e.key === "c" || e.key === "C") && term.hasSelection()){
      copyText(term.getSelection());
      return false; // Selection exists: copy instead of sending ^C.
    }
    if(mod && !e.shiftKey && (e.key === "v" || e.key === "V")){
      if(navigator.clipboard && navigator.clipboard.readText){
        navigator.clipboard.readText().then(function(t){
          if(held && ws && ws.readyState === 1) ws.send(JSON.stringify({t:"in", d:t}));
        }).catch(function(){});
        return false;
      }
      return true; // Without the Clipboard API, fall back to the browser/native paste event.
    }
    return true;
  });

  function b64ToBytes(b64){
    var bin = atob(b64);
    var a = new Uint8Array(bin.length);
    for(var i=0;i<bin.length;i++) a[i] = bin.charCodeAt(i);
    return a;
  }
  var dot = document.getElementById("dot");
  var stateEl = document.getElementById("state");
  var STATE_LABEL = { disconnected: "Disconnected", reconnecting: "Reconnecting" };
  var finished = false; // Final disconnected state received; do not reconnect.
  function setState(st){
    stateEl.textContent = STATE_LABEL[st] || st;
    dot.className = "s-" + st;
    if(st === "disconnected"){
      finished = true;
      var tk = document.getElementById("takeover");
      if(tk) tk.style.display = "none";        // Historical sessions cannot be taken over.
      document.body.classList.remove("held");
      try{ es.close(); }catch(e){}             // Stop SSE to prevent reconnects.
    }
  }

  // ---- Human takeover: toggle control + WebSocket input ----
  // Browser signature: a stable token stored in localStorage for single-operator ownership checks.
  // Only one operator may write to a session at a time; others must wait until control is released.
  function browserToken(){
    var k="terminal_mcp_owner", t="";
    try{ t=localStorage.getItem(k)||""; }catch(e){}
    if(!t){
      t=(window.crypto&&crypto.randomUUID)?crypto.randomUUID():(Date.now()+"-"+Math.random().toString(16).slice(2));
      try{ localStorage.setItem(k,t); }catch(e){}
    }
    return t;
  }
  var owner = browserToken();
  var held = false, ws = null, holdMode = "free"; // free | mine | other
  var btn = document.getElementById("takeover");
  var ICON_PLAY = '<svg viewBox="0 0 16 16" fill="currentColor"><path d="M5 3.5v9l7-4.5z"/></svg>';
  var ICON_STOP = '<svg viewBox="0 0 16 16" fill="currentColor"><rect x="4" y="4" width="8" height="8" rx="1.5"/></svg>';
  var ICON_LOCK = '<svg viewBox="0 0 16 16" fill="currentColor"><path d="M4.5 7V5.2a3.5 3.5 0 0 1 7 0V7H12a1.5 1.5 0 0 1 1.5 1.5v4A1.5 1.5 0 0 1 12 14H4a1.5 1.5 0 0 1-1.5-1.5v-4A1.5 1.5 0 0 1 4 7h.5zm1.5 0h4V5.2a2 2 0 1 0-4 0V7z"/></svg>';
  var copyBtn = document.getElementById("copyid");
  if(copyBtn) copyBtn.addEventListener("click", function(){ copyText(id); });
  function sendResize(){
    if(ws && ws.readyState === 1 && term.cols > 0 && term.rows > 0)
      ws.send(JSON.stringify({t:"resize", cols:term.cols, rows:term.rows}));
  }
  term.onResize(function(){ showSize(); sendResize(); }); // Keep PTY size synchronized with xterm.

  // ---- Font, font size, and theme controls; preferences persist in localStorage ----
  var FONTS = {
    jetbrains: '"JetBrains Mono", ui-monospace, Menlo, Consolas, monospace',
    ibmplex:   '"IBM Plex Mono", ui-monospace, Menlo, Consolas, monospace',
    fira:      '"Fira Code", ui-monospace, Menlo, Consolas, monospace',
    hack:      '"Hack", ui-monospace, Menlo, Consolas, monospace',
    source:    '"Source Code Pro", ui-monospace, Menlo, Consolas, monospace',
    consolas:  'Consolas, "Cascadia Mono", "DejaVu Sans Mono", ui-monospace, monospace',
    cascadia:  '"Cascadia Code", "Cascadia Mono", Consolas, monospace',
    menlo:     '"SF Mono", Menlo, Monaco, ui-monospace, monospace',
    dejavu:    '"DejaVu Sans Mono", ui-monospace, monospace'
  };
  // Theme contains UI chrome colors and terminal colors; opaque backgrounds keep text rendering crisp.
  function ansi(a){ return {black:a[0],red:a[1],green:a[2],yellow:a[3],blue:a[4],magenta:a[5],cyan:a[6],white:a[7],
    brightBlack:a[8],brightRed:a[9],brightGreen:a[10],brightYellow:a[11],brightBlue:a[12],brightMagenta:a[13],brightCyan:a[14],brightWhite:a[15]}; }
  function mkTheme(bg,fg,cur,sel,a){ var t=ansi(a); t.background=bg; t.foreground=fg; t.cursor=cur; t.cursorAccent=bg; t.selectionBackground=sel; return t; }
  function hexToRgba(h,al){ h=(h||"").replace("#",""); if(h.length===3) h=h[0]+h[0]+h[1]+h[1]+h[2]+h[2];
    var n=parseInt(h,16); return "rgba("+((n>>16)&255)+","+((n>>8)&255)+","+(n&255)+","+al+")"; }
  var THEMES = {
    "catppuccin-mocha": { ui:{bg:"#1e1e2e",bar:"#181825",border:"#313244",text:"#cdd6f4",muted:"#7f849c",accent:"#89b4fa"},
      term: mkTheme("#1e1e2e","#cdd6f4","#f5e0dc","#414356",
        ["#45475a","#f38ba8","#a6e3a1","#f9e2af","#89b4fa","#f5c2e7","#94e2d5","#bac2de","#585b70","#f38ba8","#a6e3a1","#f9e2af","#89b4fa","#f5c2e7","#94e2d5","#a6adc8"]) },
    "tokyo-night": { ui:{bg:"#1a1b26",bar:"#16161e",border:"#2a2e42",text:"#c0caf5",muted:"#6b7394",accent:"#7aa2f7"},
      term: mkTheme("#1a1b26","#c0caf5","#c0caf5","#283457",
        ["#15161e","#f7768e","#9ece6a","#e0af68","#7aa2f7","#bb9af7","#7dcfff","#a9b1d6","#414868","#f7768e","#9ece6a","#e0af68","#7aa2f7","#bb9af7","#7dcfff","#c0caf5"]) },
    "one-dark": { ui:{bg:"#282c34",bar:"#21252b",border:"#3a3f4b",text:"#abb2bf",muted:"#6b7280",accent:"#61afef"},
      term: mkTheme("#282c34","#abb2bf","#528bff","#3e4451",
        ["#282c34","#e06c75","#98c379","#e5c07b","#61afef","#c678dd","#56b6c2","#abb2bf","#5c6370","#e06c75","#98c379","#e5c07b","#61afef","#c678dd","#56b6c2","#ffffff"]) },
    "gruvbox-dark": { ui:{bg:"#282828",bar:"#1d2021",border:"#3c3836",text:"#ebdbb2",muted:"#a89984",accent:"#fabd2f"},
      term: mkTheme("#282828","#ebdbb2","#ebdbb2","#3c3836",
        ["#282828","#cc241d","#98971a","#d79921","#458588","#b16286","#689d6a","#a89984","#928374","#fb4934","#b8bb26","#fabd2f","#83a598","#d3869b","#8ec07c","#ebdbb2"]) },
    "night-owl": { ui:{bg:"#011627",bar:"#010e1a",border:"#122d42",text:"#d6deeb",muted:"#637777",accent:"#82aaff"},
      term: mkTheme("#011627","#d6deeb","#80a4c2","#1d3b53",
        ["#011627","#ef5350","#22da6e","#addb67","#82aaff","#c792ea","#21c7a8","#ffffff","#575656","#ef5350","#22da6e","#ffeb95","#82aaff","#c792ea","#7fdbca","#ffffff"]) },
    "nord": { ui:{bg:"#2e3440",bar:"#272c36",border:"#3b4252",text:"#d8dee9",muted:"#7b88a1",accent:"#88c0d0"},
      term: mkTheme("#2e3440","#d8dee9","#d8dee9","#434c5e",
        ["#3b4252","#bf616a","#a3be8c","#ebcb8b","#81a1c1","#b48ead","#88c0d0","#e5e9f0","#4c566a","#bf616a","#a3be8c","#ebcb8b","#81a1c1","#b48ead","#8fbcbb","#eceff4"]) },
    "ayu-mirage": { ui:{bg:"#1f2430",bar:"#191e2a",border:"#2d3640",text:"#cbccc6",muted:"#707a8c",accent:"#ffcc66"},
      term: mkTheme("#1f2430","#cbccc6","#ffcc66","#34455a",
        ["#191e2a","#f28779","#d5ff80","#ffd173","#73d0ff","#dfbfff","#95e6cb","#c7c7c7","#686868","#f28779","#d5ff80","#ffd173","#73d0ff","#dfbfff","#95e6cb","#ffffff"]) }
  };
  var LS_FONT="terminal_mcp_font", LS_SIZE="terminal_mcp_fsize", LS_THEME="terminal_mcp_theme";
  var fontsel=document.getElementById("fontsel"), themesel=document.getElementById("themesel"), fsizeEl=document.getElementById("fsize");
  function applyFont(key){
    term.options.fontFamily = FONTS[key] || FONTS.jetbrains;
    try{ localStorage.setItem(LS_FONT, key); }catch(e){}
    refit(); sendResize();
  }
  function applyFontSize(px){
    px = Math.max(10, Math.min(24, px|0));
    term.options.fontSize = px;
    if(fsizeEl) fsizeEl.textContent = px;
    try{ localStorage.setItem(LS_SIZE, String(px)); }catch(e){}
    refit(); sendResize();
  }
  function applyTheme(key){
    var th = THEMES[key] || THEMES["tokyo-night"];
    term.options.theme = th.term;
    var rs = document.documentElement.style;
    rs.setProperty("--bg", th.ui.bg); rs.setProperty("--bar", th.ui.bar);
    rs.setProperty("--border", th.ui.border); rs.setProperty("--text", th.ui.text);
    rs.setProperty("--muted", th.ui.muted); rs.setProperty("--accent", th.ui.accent);
    rs.setProperty("--termbg", th.term.background);
    rs.setProperty("--glow", hexToRgba(th.ui.accent, 0.16));
    if(themesel) themesel.value = key;
    try{ localStorage.setItem(LS_THEME, key); }catch(e){}
    refit();
  }
  // Defaults come from server config, with localStorage preferences taking precedence.
  var savedFont="__DEFAULT_FONT__", savedSize=parseInt("__DEFAULT_FSIZE__",10)||14, savedTheme="__DEFAULT_THEME__";
  try{
    savedFont = localStorage.getItem(LS_FONT) || savedFont;
    var ss = parseInt(localStorage.getItem(LS_SIZE),10); if(!isNaN(ss)) savedSize = ss;
    savedTheme = localStorage.getItem(LS_THEME) || savedTheme;
  }catch(e){}
  if(fontsel){ fontsel.value=savedFont; fontsel.addEventListener("change", function(){ applyFont(fontsel.value); }); }
  if(themesel){ themesel.value=savedTheme; themesel.addEventListener("change", function(){ applyTheme(themesel.value); }); }
  var fdec=document.getElementById("fdec"), finc=document.getElementById("finc");
  if(fdec) fdec.addEventListener("click", function(){ applyFontSize((term.options.fontSize||14)-1); });
  if(finc) finc.addEventListener("click", function(){ applyFontSize((term.options.fontSize||14)+1); });
  applyTheme(savedTheme);
  // Apply webfonts only after they are ready so canvas measurements do not use a stale fallback font.
  if(document.fonts && document.fonts.ready){
    document.fonts.ready.then(function(){ applyFont(savedFont); applyFontSize(savedSize); });
  } else {
    applyFont(savedFont); applyFontSize(savedSize);
  }
  function setTakeover(on){
    fetch(base+"/takeover",
      {method:"POST", headers:{"Content-Type":"application/json"},
       body:JSON.stringify({on:on, owner:owner, cols:term.cols, rows:term.rows})})
     .then(function(r){ return r.json(); })
     .then(function(j){ applyState(!!j.held, !!j.mine); }) // Both 200 and 409 return {held,mine}.
     .catch(function(){});
  }
  // applyState maps authoritative server state to mine / other / free UI modes.
  // holdMode prevents polling from repeatedly reopening WebSocket/focus and interrupting local work.
  function applyState(isHeld, mine){
    var mode = (isHeld && mine) ? "mine" : (isHeld ? "other" : "free");
    if(mode === holdMode) return;
    holdMode = mode;
    held = (mode === "mine");
    if(mode === "mine"){
      btn.disabled = false; btn.innerHTML = ICON_STOP + "Release Control"; btn.title = "Release human control";
      document.body.classList.remove("locked"); document.body.classList.add("held");
      openWS(); term.focus();
    } else if(mode === "other"){
      closeWS();
      btn.disabled = true; btn.innerHTML = ICON_LOCK + "Controlled by another user";
      btn.title = "This session is controlled by another user. You can take control after they release it.";
      document.body.classList.remove("held"); document.body.classList.add("locked");
    } else {
      closeWS();
      btn.disabled = false; btn.innerHTML = ICON_PLAY + "Take Control"; btn.title = "";
      document.body.classList.remove("held"); document.body.classList.remove("locked");
    }
  }
  function openWS(){
    if(ws) return;
    var proto = location.protocol === "https:" ? "wss:" : "ws:";
    ws = new WebSocket(proto+"//"+location.host+base+"/ws?owner="+encodeURIComponent(owner));
    ws.onopen = function(){ sendResize(); };
    ws.onclose = function(){ ws = null; };
  }
  function closeWS(){ if(ws){ try{ ws.close(); }catch(e){} ws = null; } }
  btn.addEventListener("click", function(){ if(!btn.disabled) setTakeover(!held); });

  // Forward input to the PTY only while this browser holds control; otherwise xterm stays read-only.
  term.onData(function(d){ if(held && ws && ws.readyState === 1) ws.send(JSON.stringify({t:"in", d:d})); });
  window.addEventListener("resize", refit);

  // Takeover is session-wide server state; sync on load and poll so all windows stay consistent.
  function syncTakeover(){
    if(finished) return;
    fetch(base+"/takeover?owner="+encodeURIComponent(owner))
     .then(function(r){ return r.json(); })
     .then(function(j){ if(!finished) applyState(!!j.held, !!j.mine); })
     .catch(function(){});
  }
  syncTakeover();
  setInterval(syncTakeover, 2500);

  // ---- Output: feed raw SSE bytes directly into xterm ----
  var es = new EventSource(base + "/stream");
  es.addEventListener("data", function(e){ term.write(b64ToBytes(e.data)); });
  es.addEventListener("state", function(e){ setState(e.data); });
  es.onerror = function(){ if(!finished) setState("reconnecting"); };
})();
</script>
</body>
</html>
`
