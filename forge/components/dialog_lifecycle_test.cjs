const test=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');
function harness(){
 const documentHandlers={},panelHandlers={},events=[],animations=[];
 const opener={isConnected:true,expanded:'false',focuses:0,setAttribute(k,v){if(k==='aria-expanded')this.expanded=v},focus(){this.focuses++}};
 const panel={open:false,inert:false,returnValue:'',listeners:panelHandlers,
  addEventListener(k,fn,options){panelHandlers[k]=fn;options?.signal.addEventListener('abort',()=>{if(panelHandlers[k]===fn)delete panelHandlers[k]},{once:true})},dispatchEvent(event){events.push(event);return !(event.type==='md-dialog-beforeclose'&&this.veto)},
  showModal(){this.open=true},close(value=''){this.open=false;this.returnValue=value;panelHandlers.close?.()},
  animate(){let finish,reject;const finished=new Promise((resolve,no)=>{finish=resolve;reject=no});const a={finished,finish,cancel(){reject(new Error('cancelled'))}};animations.push(a);return a}
 };
 const context={AbortController,document:{activeElement:opener,querySelectorAll(){return [panel]},addEventListener(k,fn){documentHandlers[k]=fn}},matchMedia:()=>({matches:false}),getComputedStyle:()=>({opacity:'0.5'}),MaterialMotion:{timing:()=>({duration:100})},CustomEvent:class {constructor(type,options){this.type=type;Object.assign(this,options)}}};
 runInNewContext(readFileSync(`${__dirname}/static/dialogs.js`,'utf8'),context);
 return {panel,opener,events,animations,api:context.MaterialDialogs,documentHandlers};
}
test('dialog reopening cancels a pending close and restores focus only after actual close',async()=>{
 const h=harness();let pending=h.api.open(h.panel,h.opener);h.animations.at(-1).finish();await pending;
 const closing=h.api.close(h.panel,'cancel');assert.equal(h.panel.inert,true);
 pending=h.api.open(h.panel,h.opener);h.animations.at(-1).finish();await Promise.all([closing,pending]);
 assert.equal(h.panel.open,true);assert.equal(h.panel.inert,false);assert.equal(h.opener.focuses,0);
 pending=h.api.close(h.panel,'done','action');h.animations.at(-1).finish();await pending;
 assert.equal(h.panel.open,false);assert.equal(h.panel.inert,false);assert.equal(h.panel.returnValue,'done');assert.equal(h.opener.expanded,'false');assert.equal(h.opener.focuses,1);
});
test('application can veto dismissal; cleanup accepts text nodes and cancels pending work',async()=>{
 const h=harness();let pending=h.api.open(h.panel,h.opener);h.animations.at(-1).finish();await pending;
 h.panel.veto=true;await h.api.close(h.panel);assert.equal(h.panel.open,true);assert.equal(h.panel.inert,false);
 h.documentHandlers['htmx:beforeCleanupElement']({detail:{elt:{nodeType:3}}});
 h.panel.veto=false;pending=h.api.close(h.panel);h.documentHandlers['htmx:beforeCleanupElement']({detail:{elt:{querySelectorAll:()=>[h.panel]}}});await pending;assert.equal(h.panel.open,false);
});
