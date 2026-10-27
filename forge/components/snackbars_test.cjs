const {test}=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');
function harness({action=true,inline=false,obstacles=[]}={}){
 const listeners={},starts=[];let focus;
 const opener={getClientRects:()=>[],focus(){focus=this},hasAttribute:a=>a==='data-snackbar-open'},close={hasAttribute:a=>a==='data-snackbar-close'};
 const actionButton={focus(){focus=this}};
 const positions={};
 const bar={style:{removeProperty:n=>delete positions[n],setProperty:(n,v)=>positions[n]=v},matches(){return !!this.open},getBoundingClientRect:()=>({left:400,right:700,top:724,bottom:780,height:56}),dataset:{snackbarDuration:'6000'},addEventListener(){},showPopover(){this.open=true},hidePopover(){this.open=false},contains:e=>e===close,querySelector:s=>s==='[data-snackbar-action]'?(action?actionButton:null):{textContent:'Message sent'}};
 const output=inline?{textContent:''}:null,demo={querySelector:s=>s==='[data-snackbar-inline]'?output:bar};
 for(const b of [opener,close]){b.closest=()=>demo;b.dispatchEvent=()=>{}};
 const components={snackbarTimer:()=>({start:n=>starts.push(n),stop(){}})};
 const document={querySelector:()=>bar.open&&action?actionButton:null,activeElement:opener,querySelectorAll:s=>s.includes('.md-fab')?obstacles:[bar],addEventListener:(t,f)=>(listeners[t]??=[]).push(f)};
 const runtime={document,window:{addEventListener(){}},getComputedStyle:()=>({visibility:'visible',getPropertyValue:()=> '16px'}),MaterialComponents:Object.freeze(components),CustomEvent:class{}};
 runInNewContext(readFileSync(`${__dirname}/behaviors/snackbars.js`,'utf8'),runtime);
 return{show:()=>runtime.MaterialSnackbars.show(bar),bar,positions,output,starts,opener,actionButton,key(key,options={}){const e={key,preventDefault(){this.defaultPrevented=true},...options};for(const f of listeners.keydown)f(e);return !!e.defaultPrevented},get focus(){return focus},click(button=opener){for(const f of listeners.click)f({target:{closest:()=>button}})},close(){document.activeElement=close;this.click(close)}};
}
test('Actionable snackbars persist instead of scheduling dismissal',()=>{const h=harness();h.click();assert.equal(h.starts.at(-1),0);assert.equal(h.bar.open,true);h.close();assert.equal(h.bar.open,false);assert.equal(h.focus,h.opener)});
test('Actionless snackbars require inline feedback before auto-dismiss',()=>{const persistent=harness({action:false});persistent.click();assert.equal(persistent.starts.at(-1),0);
const timed=harness({action:false,inline:true});timed.click();assert.equal(timed.starts.at(-1),6000);assert.equal(timed.output.textContent,'Message sent');});

test('F6 focuses an available action and leaves other shortcuts alone',()=>{
 const h=harness();assert.equal(h.key('F6'),false);h.click();assert.equal(h.key('F6'),true);assert.equal(h.focus,h.actionButton);assert.equal(h.key('F6',{shiftKey:true}),false);assert.equal(h.key('F6',{ctrlKey:true}),false);
});

test('snackbar clears overlapping controls and resets placement before each open',()=>{
 const obstacle={getClientRects:()=>[{}],getBoundingClientRect:()=>({left:400,right:700,top:740,bottom:800})};
 const h=harness({obstacles:[obstacle]});h.click();assert.equal(h.positions['--_snackbar-avoid-offset'],'56px');h.close();obstacle.getClientRects=()=>[];h.click();assert.equal(h.positions['--_snackbar-avoid-offset'],'0px');
});

test('programmatic status uses shared placement and restores the actual opener',()=>{const h=harness();h.show();assert.equal(h.bar.open,true);assert.equal(h.starts.at(-1),0);h.close();assert.equal(h.focus,h.opener)});
