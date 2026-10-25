const {test}=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');
function harness(){
 const listeners={},timers=new Map();let time=0,seq=0;
 const document={readyState:'loading',querySelectorAll:()=>[],addEventListener(t,f){(listeners[t]??=[]).push(f)}};
 runInNewContext(readFileSync(`${__dirname}/behaviors/tooltips.js`,'utf8'),{document,window:{addEventListener(){}},setTimeout(fn,ms){const id=++seq;timers.set(id,{fn,at:time+ms});return id},clearTimeout(id){timers.delete(id)}});
 function demo(){const attributes=new Set(),d={matches:()=>false,querySelector:()=>({}),contains:e=>e?.parent===d,setAttribute:n=>attributes.add(n),removeAttribute:n=>attributes.delete(n),get visible(){return attributes.has('data-tooltip-visible')}};d.target={parent:d,closest:()=>d};return d}
 return{demo,send(type,d,more={}){for(const f of listeners[type]||[])f({target:d.target,relatedTarget:null,pointerId:1,isPrimary:true,pointerType:'mouse',clientX:0,clientY:0,...more})},advance(ms){time+=ms;for(const [id,t]of [...timers])if(t.at<=time){timers.delete(id);t.fn()}}};
}
test('plain tooltips retain 1.5s exit grace and cancel dismissal on re-entry',()=>{const h=harness(),d=h.demo();h.send('pointerover',d);assert.equal(d.visible,true);h.send('pointerout',d);h.advance(1499);assert.equal(d.visible,true);h.send('pointerover',d);h.advance(10);assert.equal(d.visible,true);h.send('pointerout',d);h.advance(1500);assert.equal(d.visible,false)});
test('touch hold opens after 500ms; taps and vertical gestures do not open',()=>{const h=harness(),d=h.demo();h.send('pointerdown',d,{pointerType:'touch'});h.send('focusin',d);h.advance(499);assert.equal(d.visible,false);h.advance(1);assert.equal(d.visible,true);h.send('pointerup',d,{pointerType:'touch'});h.advance(1500);assert.equal(d.visible,false);h.send('pointerdown',d,{pointerType:'touch'});h.send('pointerup',d,{pointerType:'touch'});h.advance(600);assert.equal(d.visible,false);h.send('pointerdown',d,{pointerType:'touch'});h.send('pointermove',d,{pointerType:'touch',clientY:20});h.advance(600);assert.equal(d.visible,false)});
test('opening another tooltip closes the previous one; Escape dismisses keyboard disclosure',()=>{const h=harness(),a=h.demo(),b=h.demo();h.send('focusin',a);assert.equal(a.visible,true);h.send('focusin',b);assert.equal(a.visible,false);assert.equal(b.visible,true);h.send('keydown',b,{key:'Escape'});assert.equal(b.visible,false)});

test('rich placement fits corners without covering the trigger',()=>{
 const out={exports:{}};runInNewContext(readFileSync(`${__dirname}/behaviors/tooltips.js`,'utf8'),{module:out});
 const place=out.exports.tooltipPosition,box={width:180,height:100},viewport={width:390,height:600};
 for(const anchor of [{left:30,right:70,top:40,bottom:80},{left:330,right:370,top:520,bottom:560},{left:100,right:140,top:70,bottom:110}]){
  const pos=place(anchor,box,viewport);assert.ok(pos);assert.ok(pos.left>=8&&pos.top>=8&&pos.left+180<=382&&pos.top+100<=592);assert.ok(pos.left>=anchor.right||pos.left+180<=anchor.left||pos.top>=anchor.bottom||pos.top+100<=anchor.top);
 }
 assert.equal(place({left:100,right:140,top:40,bottom:80},{width:380,height:200},{width:390,height:220}),null);
});
test('hover or keyboard focus inside the target keeps a transient tooltip open',()=>{const h=harness(),d=h.demo();h.send('focusin',d);d.matches=()=>true;h.send('pointerout',d);h.advance(1600);assert.equal(d.visible,true);d.matches=()=>false;h.send('focusout',d);h.advance(1500);assert.equal(d.visible,false)});
