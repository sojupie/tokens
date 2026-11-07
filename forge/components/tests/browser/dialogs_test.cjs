const test=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');const out={exports:{}};runInNewContext(readFileSync(`${__dirname}/../../generated/dialogs.js`,'utf8'),{module:out});const {scrollEdges}=out.exports;
test('dialog elevation follows content hidden at each scroll edge',()=>{assert.deepEqual({...scrollEdges(0,100,300)},{start:false,end:true});assert.deepEqual({...scrollEdges(50,100,300)},{start:true,end:true});assert.deepEqual({...scrollEdges(200,100,300)},{start:true,end:false});assert.deepEqual({...scrollEdges(0,100,100)},{start:false,end:false});assert.equal(scrollEdges(199.5,100,300).end,false);});

const {getEventListeners}=require('node:events');
function dialogs({opened=false,motion=false}={}){
 const handlers={},observers=[],animations=[],notifications=[];
 const body=Object.assign(new EventTarget(),{scrollTop:37,clientHeight:100,scrollHeight:300});
 const attrs=new Set(['data-dialog-surface']);
 const panel=Object.assign(new EventTarget(),{
  id:'example',open:opened,inert:false,dataset:{variant:'basic'},value:'Existing draft',returnValue:'',
  querySelector:()=>body,querySelectorAll:()=>[],matches:()=>true,
  hasAttribute:name=>attrs.has(name),
  showModal(){this.open=true;},
  close(result=''){this.open=false;this.returnValue=result;this.dispatchEvent(new Event('close'));},
  animate(){let finish,fail;const animation={finished:new Promise((resolve,reject)=>{finish=resolve;fail=reject}),cancel(){fail(new Error('cancelled'))},finish(){finish()}};animations.push(animation);return animation;},
 });
 const opener={isConnected:true,expanded:'false',focused:0,setAttribute(name,value){this.expanded=value},focus(){this.focused++}};
 for(const name of ['opened','closed'])panel.addEventListener('md-dialog-'+name,()=>notifications.push(name));
 const document={activeElement:opener,querySelectorAll:()=>[panel],getElementById:()=>panel,addEventListener(name,fn){(handlers[name]??=[]).push(fn)}};
 const context={document,matchMedia:()=>({matches:!motion,addEventListener(){}}),AbortController,CustomEvent,
  MaterialMotion:{timing:()=>({duration:100})},getComputedStyle:()=>({opacity:'1',transform:'none'}),
  ResizeObserver:class{constructor(fn){this.fn=fn;observers.push(this)}observe(){this.connected=true}disconnect(){this.connected=false}},
 };
 runInNewContext(readFileSync(`${__dirname}/../../generated/dialogs.js`,'utf8'),context);
 return {api:context.MaterialDialogs,panel,body,opener,observers,animations,notifications,
  emit(name,root=panel){for(const fn of handlers[name]||[])fn({target:root,detail:{elt:root}})},
  scrollListeners:()=>getEventListeners(body,'scroll').length};
}

test('closed dialogs need no instance listeners or observers before first opening',async()=>{
 const h=dialogs();h.api.initialize(h.panel);
 assert.equal(h.observers.length,0);assert.equal(h.scrollListeners(),0);
 await h.api.open(h.panel,h.opener);
 assert.equal(h.panel.open,true);assert.equal(h.opener.expanded,'true');
 assert.equal(h.scrollListeners(),1);assert.equal(h.observers.length,1);
 assert.equal(h.body.scrollTop,37);assert.equal(h.panel.value,'Existing draft');
 assert.equal(h.panel.dataset.topScrolled,'true');
 await h.api.close(h.panel);
 assert.equal(h.panel.open,false);assert.equal(h.opener.expanded,'false');
 assert.equal(h.scrollListeners(),0);assert.equal(h.observers[0].connected,false);
});

test('initialization adopts open markup and remains idempotent across host events',()=>{
 const h=dialogs({opened:true});
 for(let i=0;i<3;i++){
  h.api.initialize(h.panel);h.emit('htmx:afterSwap');h.emit('tokens:assets-ready');
 }
 assert.equal(h.observers.length,1);assert.equal(h.scrollListeners(),1);
 assert.equal(h.panel.open,true);assert.equal(h.body.scrollTop,37);assert.equal(h.panel.value,'Existing draft');
 h.api.dispose(h.panel);h.api.dispose(h.panel);
 assert.equal(h.scrollListeners(),0);assert.equal(h.observers[0].connected,false);
 assert.equal(h.panel.open,false);assert.deepEqual(h.notifications,[],'disposal must not invoke application dismissal actions');
});

test('reopening reattaches one listener and restores focus once per close',async()=>{
 const h=dialogs();
 for(let i=0;i<3;i++){
  await h.api.open(h.panel,h.opener);h.api.initialize(h.panel);
  assert.equal(h.scrollListeners(),1);
  h.panel.dispatchEvent(new Event('cancel',{cancelable:true}));
  assert.equal(h.scrollListeners(),0);assert.equal(h.opener.focused,i+1);
 }
 assert.deepEqual(h.notifications,['opened','closed','opened','closed','opened','closed']);
});

test('cleanup cancels pending motion and allows the same node to be activated again',async()=>{
 const h=dialogs({motion:true});
 const opening=h.api.open(h.panel,h.opener);
 h.emit('htmx:beforeCleanupElement');await opening;
 assert.equal(h.scrollListeners(),0);assert.equal(h.panel.open,false);assert.equal(h.panel.inert,false);
 assert.equal(h.opener.expanded,'false');assert.equal(h.opener.focused,0);
 assert.equal(h.observers[0].connected,false);assert.deepEqual(h.notifications,[]);
 const reopening=h.api.open(h.panel,h.opener);h.animations.at(-1).finish();await reopening;
 assert.equal(h.scrollListeners(),1);assert.deepEqual(h.notifications,['opened']);
 h.api.dispose(h.panel);
});

test('a vetoed opening allocates no observers and stale native close events do not reset a reopened dialog',async()=>{
 const h=dialogs(),veto=event=>event.preventDefault();
 h.panel.addEventListener('md-dialog-beforeopen',veto);await h.api.open(h.panel);
 assert.equal(h.panel.open,false);assert.equal(h.observers.length,0);
 h.panel.removeEventListener('md-dialog-beforeopen',veto);
 await h.api.open(h.panel,h.opener);await h.api.close(h.panel);await h.api.open(h.panel,h.opener);
 h.panel.dispatchEvent(new Event('close'));
 assert.equal(h.opener.expanded,'true');assert.equal(h.scrollListeners(),1);
 assert.deepEqual(h.notifications,['opened','closed','opened']);
 h.api.dispose(h.panel);
});
