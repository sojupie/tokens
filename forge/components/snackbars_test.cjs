const {test}=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');
function harness({action=true,inline=false}={}){
 const listeners={},starts=[];let focus;
 const opener={focus(){focus=this},hasAttribute:a=>a==='data-snackbar-open'},close={hasAttribute:a=>a==='data-snackbar-close'};
 const actionButton={focus(){focus=this}};
 const bar={dataset:{snackbarDuration:'6000'},addEventListener(){},showPopover(){this.open=true},hidePopover(){this.open=false},contains:e=>e===close,querySelector:s=>s==='[data-snackbar-action]'?(action?actionButton:null):{textContent:'Message sent'}};
 const output=inline?{textContent:''}:null,demo={querySelector:s=>s==='[data-snackbar-inline]'?output:bar};
 for(const b of [opener,close]){b.closest=()=>demo;b.dispatchEvent=()=>{}};
 const document={querySelector:()=>bar.open&&action?actionButton:null,activeElement:opener,querySelectorAll:()=>[bar],addEventListener:(t,f)=>(listeners[t]??=[]).push(f)};
 runInNewContext(readFileSync(`${__dirname}/behaviors/snackbars.js`,'utf8'),{document,MaterialComponents:{snackbarTimer:()=>({start:n=>starts.push(n),stop(){}})},CustomEvent:class{}});
 return{bar,output,starts,opener,actionButton,key(key,options={}){const e={key,preventDefault(){this.defaultPrevented=true},...options};for(const f of listeners.keydown)f(e);return !!e.defaultPrevented},get focus(){return focus},click(button=opener){for(const f of listeners.click)f({target:{closest:()=>button}})},close(){document.activeElement=close;this.click(close)}};
}
test('Actionable snackbars persist instead of scheduling dismissal',()=>{const h=harness();h.click();assert.equal(h.starts.at(-1),0);assert.equal(h.bar.open,true);h.close();assert.equal(h.bar.open,false);assert.equal(h.focus,h.opener)});
test('Actionless snackbars require inline feedback before auto-dismiss',()=>{const persistent=harness({action:false});persistent.click();assert.equal(persistent.starts.at(-1),0);
const timed=harness({action:false,inline:true});timed.click();assert.equal(timed.starts.at(-1),6000);assert.equal(timed.output.textContent,'Message sent');});

test('F6 focuses an available action and leaves other shortcuts alone',()=>{
 const h=harness();assert.equal(h.key('F6'),false);h.click();assert.equal(h.key('F6'),true);assert.equal(h.focus,h.actionButton);assert.equal(h.key('F6',{shiftKey:true}),false);assert.equal(h.key('F6',{ctrlKey:true}),false);
});
