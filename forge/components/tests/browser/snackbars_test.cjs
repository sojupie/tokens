const {test}=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');
function harness({action=true,inline=false,obstacles=[],host=true}={}){
 const listeners={},starts=[];let focus;
 const opener={getClientRects:()=>[],focus(){focus=this},hasAttribute:a=>a==='data-snackbar-open'},close={hasAttribute:a=>a==='data-snackbar-close'};
 const actionButton={focus(){focus=this}};
 const positions={},barEvents={},frames=new Map(),observed=new Set();let nextFrame=0,mutationCallback,resizeCallback,watching=false,observations=0;
 const bar={style:{removeProperty:n=>delete positions[n],setProperty:(n,v)=>positions[n]=v},matches(){return !!this.open},getBoundingClientRect:()=>({left:400,right:700,top:724,bottom:780,height:56}),dataset:{snackbarDuration:'6000'},addEventListener:(type,callback)=>(barEvents[type]??=[]).push(callback),showPopover(){this.open=true},hidePopover(){this.open=false;for(const callback of barEvents.toggle??[])callback()},contains:e=>e===close,querySelector:s=>s==='[data-snackbar-action]'?(action?actionButton:null):{textContent:'Message sent'}};
 const output=inline?{textContent:''}:null,demo={querySelector:s=>s==='[data-snackbar-inline]'?output:bar};
 for(const b of [opener,close]){b.closest=s=>s==='.md-snackbar'?(b===close?bar:null):s==='[data-snackbar-demo]'&&host?demo:null;b.dispatchEvent=()=>{}};
 const document={body:{},querySelector:()=>bar.open&&action?actionButton:null,activeElement:opener,querySelectorAll:s=>s.includes('.md-fab')?obstacles:[bar],addEventListener:(t,f)=>(listeners[t]??=[]).push(f)};
 const runtime={document,window:{addEventListener(){}},getComputedStyle:()=>({visibility:'visible',getPropertyValue:()=> '16px'}),performance:{now:()=>0},setTimeout(fn,delay){starts.push(delay);return starts.length},clearTimeout(){},CustomEvent:class{},
 requestAnimationFrame(callback){frames.set(++nextFrame,callback);return nextFrame},cancelAnimationFrame(id){frames.delete(id)},
 ResizeObserver:class{constructor(callback){resizeCallback=callback}observe(target){observed.add(target);observations++}unobserve(target){observed.delete(target)}},
 MutationObserver:class{constructor(callback){mutationCallback=callback}observe(){watching=true}disconnect(){watching=false}}};
 runInNewContext(readFileSync(`${__dirname}/../../generated/runtime/snackbars.js`,'utf8'),runtime);
 return{show:()=>runtime.MaterialSnackbars.show(bar),bar,positions,output,starts,opener,actionButton,
 layout(records=[{type:'childList'}]){if(watching)mutationCallback(records)},resize(){if(observed.size)resizeCallback()},flush(){const pending=[...frames.values()];frames.clear();for(const callback of pending)callback()},get pending(){return frames.size},get observations(){return observations},get watched(){return watching||observed.size>0},
 key(key,options={}){const e={key,preventDefault(){this.defaultPrevented=true},...options};for(const f of listeners.keydown)f(e);return !!e.defaultPrevented},get focus(){return focus},click(button=opener){for(const f of listeners.click)f({target:{closest:()=>button}})},close(){document.activeElement=close;this.click(close)}};
}
test('Actionable snackbars persist instead of scheduling dismissal',()=>{const h=harness();h.click();assert.equal(h.starts.at(-1),undefined);assert.equal(h.bar.open,true);h.close();assert.equal(h.bar.open,false);assert.equal(h.focus,h.opener)});
test('Actionless snackbars require inline feedback before auto-dismiss',()=>{const persistent=harness({action:false});persistent.click();assert.equal(persistent.starts.at(-1),undefined);
const timed=harness({action:false,inline:true});timed.click();assert.equal(timed.starts.at(-1),6000);assert.equal(timed.output.textContent,'Message sent');});

test('F6 focuses an available action and leaves other shortcuts alone',()=>{
 const h=harness();assert.equal(h.key('F6'),false);h.click();assert.equal(h.key('F6'),true);assert.equal(h.focus,h.actionButton);assert.equal(h.key('F6',{shiftKey:true}),false);assert.equal(h.key('F6',{ctrlKey:true}),false);
});

test('snackbar clears overlapping controls and resets placement before each open',()=>{
 const obstacle={getClientRects:()=>[{}],getBoundingClientRect:()=>({left:400,right:700,top:740,bottom:800})};
 const h=harness({obstacles:[obstacle]});h.click();assert.equal(h.positions['--_snackbar-avoid-offset'],'56px');h.close();obstacle.getClientRects=()=>[];h.click();assert.equal(h.positions['--_snackbar-avoid-offset'],'0px');
});

test('programmatic status uses shared placement and restores the actual opener',()=>{const h=harness();h.show();assert.equal(h.bar.open,true);assert.equal(h.starts.at(-1),undefined);h.close();assert.equal(h.focus,h.opener)});

test('programmatic snackbar dismisses without a Catalog demo wrapper',()=>{const h=harness({host:false});h.show();assert.equal(h.bar.open,true);h.close();assert.equal(h.bar.open,false);assert.equal(h.focus,h.opener)});

test('open snackbars follow inserted and resized controls, coalesce work and stop watching after dismissal',()=>{
 const obstacles=[],h=harness({obstacles});h.show();assert.equal(h.watched,true);
 let top=740;const obstacle={getClientRects:()=>[{}],getBoundingClientRect:()=>({left:400,right:700,top,bottom:800})};obstacles.push(obstacle);
 h.layout();h.layout();assert.equal(h.pending,1);h.flush();assert.equal(h.positions['--_snackbar-avoid-offset'],'56px');
 const observations=h.observations;top=680;h.resize();h.flush();assert.equal(h.positions['--_snackbar-avoid-offset'],'116px');assert.equal(h.observations,observations);
 h.layout([{type:'attributes',attributeName:'style',target:h.bar}]);assert.equal(h.pending,0,'Placement style writes do not schedule another frame.');
 obstacles.pop();h.layout();h.flush();assert.equal(h.positions['--_snackbar-avoid-offset'],'0px');
 h.layout();h.close();assert.equal(h.pending,0);assert.equal(h.watched,false);h.layout();h.resize();assert.equal(h.pending,0);
});
