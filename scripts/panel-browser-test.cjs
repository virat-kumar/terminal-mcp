// Run: NODE_PATH=/path/to/playwright/node_modules node scripts/panel-browser-test.cjs
// No connection to a live ChatGPT session; tests the exact MCP Apps HTML in an iframe.
const fs = require('node:fs');
const assert = require('node:assert/strict');
const path = require('node:path');
const { chromium } = require('playwright');
(async () => {
  const raw = fs.readFileSync(path.join(__dirname, '../mcpserver/panel.go'), 'utf8');
  const html = raw.split('const panelHTML = `')[1].split('`\n\nfunc panelMeta')[0];
  assert.ok(html.includes('id="follow"'));
  const browser = await chromium.launch({
    headless:true, executablePath:process.env.BROWSER_BIN||'/usr/bin/google-chrome',
    args:['--no-sandbox','--disable-dev-shm-usage']
  });
  try {
    const page = await browser.newPage({viewport:{width:1000,height:760}});
    const errors = [];
    page.on('pageerror', e=>errors.push(e.message));
    await page.setContent('<iframe id="monitor" style="width:950px;height:700px;border:0"></iframe>');
    await page.evaluate((markup)=>{
      window.mock = {
        sessions:[{session_id:'s1',name:'Existing session'}],
        active:'s1', fail:false, calls:[],
        outputs:{s1:Array.from({length:200},(_,i)=>'LINE '+i).join('\n')}
      };
      window.addEventListener('message',event=>{
        if (event.source!==document.querySelector('#monitor').contentWindow) return;
        const m=event.data; if(m.jsonrpc!=='2.0'||!m.id)return;
        const reply={jsonrpc:'2.0',id:m.id};
        if(m.method==='ui/initialize'){
          reply.result={protocolVersion:'2026-01-26',hostInfo:{name:'mock'},hostCapabilities:{}};
        } else if(m.method==='tools/call'){
          const name=m.params.name;
          window.mock.calls.push(name);
          if(name==='terminal_list') {
            if(window.mock.fail){ reply.error={message:'network test failure'}; }
            else {reply.result={structuredContent:{sessions:window.mock.sessions,active_session_id:window.mock.active}};}
          } else if(name==='terminal_output'){
            reply.result={structuredContent:{output:window.mock.outputs[m.params.arguments.session_id]||''}};
          } else {reply.error={message:'unexpected tool '+name};}
        } else {reply.error={message:'unrecognized '+m.method};}
        event.source.postMessage(reply,'*');
      });
      document.querySelector('#monitor').srcdoc=markup;
    },html);
    const frame=page.frameLocator('#monitor');
    await frame.locator('#sessions').waitFor();
    await frame.locator('#status').getByText('1 session').waitFor();
    await frame.locator('#output').getByText('LINE 199',{exact:false}).waitFor();
    assert.equal(await frame.locator('#sessions').inputValue(),'s1','selects agent session');
    assert.equal(await frame.locator('#follow').getAttribute('aria-pressed'),'true','follow defaults on');
    const scroll = await frame.locator('#output').evaluate(e=>({top:e.scrollTop,max:e.scrollHeight-e.clientHeight}));
    assert.ok(scroll.max>0&&scroll.top>=scroll.max-4,'initial output scrolled to end');
    console.log('PASS auto-select active terminal and follow at bottom');

    await page.evaluate(()=>{window.mock.outputs.s1+='\nADDED_BY_REFRESH';});
    await frame.locator('#refresh').click();
    await frame.locator('#output').getByText('ADDED_BY_REFRESH',{exact:false}).waitFor();
    console.log('PASS Refresh reloads live output');

    await frame.locator('#follow').click();
    assert.equal(await frame.locator('#follow').getAttribute('aria-pressed'),'false');
    await frame.locator('#output').evaluate(e=>e.scrollTop=0);
    await page.evaluate(()=>{window.mock.outputs.s1+='\nANOTHER_APPEND';});
    await frame.locator('#refresh').click();
    await frame.locator('#output').getByText('ANOTHER_APPEND',{exact:false}).waitFor();
    const unfollowed=await frame.locator('#output').evaluate(e=>e.scrollTop);
    assert.ok(unfollowed<=4,'Follow Off must preserve manual position');
    await frame.locator('#follow').click();
    const resumed=await frame.locator('#output').evaluate(e=>({top:e.scrollTop,max:e.scrollHeight-e.clientHeight}));
    assert.ok(resumed.top>=resumed.max-4,'Follow On scrolls to bottom immediately');
    console.log('PASS Follow On/Off and default auto-scroll');

    await page.evaluate(()=>{window.mock.sessions.push({session_id:'s2',name:'New agent terminal'});window.mock.active='s2';window.mock.outputs.s2='NEW LIVE TERMINAL';});
    await frame.locator('#refresh').click();
    await frame.locator('#output').getByText('NEW LIVE TERMINAL').waitFor();
    assert.equal(await frame.locator('#sessions').inputValue(),'s2');
    await frame.locator('#sessions').selectOption('s1');
    assert.equal(await frame.locator('#sessions').inputValue(),'s1');
    await page.evaluate(()=>{window.mock.sessions.push({session_id:'s3',name:'Third agent terminal'});window.mock.active='s3';window.mock.outputs.s3='THIRD LIVE TERMINAL';});
    await frame.locator('#refresh').click();
    await frame.locator('#output').getByText('THIRD LIVE TERMINAL').waitFor();
    assert.equal(await frame.locator('#sessions').inputValue(),'s3','switches on new agent session even after manual selection');
    console.log('PASS auto-selects each newly created agent terminal');

    await page.evaluate(()=>{window.mock.fail=true;});
    await frame.locator('#refresh').click();
    await frame.locator('#output').getByText('Refresh failed:',{exact:false}).waitFor();
    await page.evaluate(()=>{window.mock.fail=false;});
    await frame.locator('#refresh').click();
    await frame.locator('#output').getByText('THIRD LIVE TERMINAL').waitFor();
    const calls=await page.evaluate(()=>window.mock.calls);
    assert.ok(!calls.some(c=>['terminal_send','terminal_control','terminal_close'].includes(c)),'monitor must stay read-only');
    assert.deepEqual(errors,[],'no browser JS errors');
    console.log('PASS failure/recovery, refresh button, read-only guarantee, browser console');
  }finally{
    await browser.close();
  }
})().catch(e=>{console.error(e);process.exit(1)});
