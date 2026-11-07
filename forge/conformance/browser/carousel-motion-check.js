const variants=['multi-browse','uncontained','uncontained-multi-aspect-ratio','hero','center-aligned-hero','full-screen'];
const host=document.querySelector('#samples'),out=document.querySelector('#result'),wait=ms=>new Promise(r=>setTimeout(r,ms));
let frame;
async function load(variant){
 host.replaceChildren();frame=document.createElement('iframe');host.append(frame);
 await new Promise(r=>{frame.onload=r;frame.src='/components/carousel?preview.compare=interactive&preview.variant='+variant+'&preview.action=none'});
 const w=frame.contentWindow,d=frame.contentDocument;await w.TokenAssets.ready;
 const root=d.querySelector('#component-preview .md-carousel'),view=root.querySelector('.md-carousel-viewport');
 for(let n=root;n&&n!==d.documentElement;n=n.parentElement){if(n!==root&&!n.classList.contains('tokens-carousel-frame'))Object.assign(n.style,{display:'block',width:'100%',maxWidth:'none',minWidth:'0',padding:'0',margin:'0',border:'0',minHeight:'0'});if(n.parentElement)for(const s of n.parentElement.children)if(s!==n&&!['STYLE','LINK','SCRIPT','TEMPLATE'].includes(s.tagName))s.style.display='none';}
 await wait(100);return{w,d,root,view,items:[...root.querySelectorAll('.md-carousel-item')]};
}
function measure(c,i=0){const b=c.items[i],r=b.getBoundingClientRect(),art=b.querySelector('img').getBoundingClientRect();return{left:r.left,top:r.top,width:r.width,height:r.height,imageLeft:art.left,imageTop:art.top,imageWidth:art.width,imageHeight:art.height};}
function pointer(c,type,button,x=100,y=100,pointerType='mouse'){button.dispatchEvent(new c.w.PointerEvent(type,{bubbles:true,pointerId:71,pointerType,isPrimary:true,button:0,clientX:x,clientY:y}));}
document.querySelector('#run').onclick=async()=>{
 const tests=[],records=[],check=(name,pass,detail)=>tests.push({name,pass:!!pass,detail});out.textContent='Running';document.querySelector('#run').disabled=true;
 try{for(const variant of variants){
  const c=await load(variant),before=measure(c),input=variant==='uncontained'?'touch':'mouse';pointer(c,'pointerdown',c.items[0],100,100,input);await wait(400);const held=measure(c);
  check(variant+': press widens mask',Math.abs(held.width-before.width-10)<.1,{before,held});
  check(variant+': press preserves image scale/position',Math.abs(held.imageWidth-before.imageWidth)<.1&&Math.abs(held.imageLeft-before.imageLeft)<.1,{before,held});
  check(variant+': pressed mask remains covered by image',held.imageLeft<=held.left+.1&&held.imageLeft+held.imageWidth>=held.left+held.width-.1,{held});
  pointer(c,'pointerup',c.items[0],100,100,input);c.items[0].click();await wait(400);const released=measure(c);
  check(variant+': release restores mask',Math.abs(released.width-before.width)<.1,{before,released});
  const index=1,a=measure(c,index),vertical=variant==='full-screen';
  const free=c.d.createElement('style');free.textContent='.md-carousel-viewport{scroll-snap-type:none!important}';c.d.head.append(free);c.view.scrollTo(vertical?{top:180,behavior:'instant'}:{left:80,behavior:'instant'});await wait(120);const b=measure(c,index);
  const maskDelta=vertical?b.top-a.top:b.left-a.left,imageDelta=vertical?b.imageTop-a.imageTop:b.imageLeft-a.imageLeft;
  check(variant+': image moves independently of mask edge',Math.abs(imageDelta-maskDelta)>.1,{maskDelta,imageDelta,a,b});
  check(variant+': scroll preserves image size',Math.abs(b.imageWidth-a.imageWidth)<.1&&Math.abs(b.imageHeight-a.imageHeight)<.1,{a,b});records.push({variant,before,held,released,scroll:b});
 }
 for(const variant of ['uncontained','uncontained-multi-aspect-ratio']){
  const c=await load(variant);c.d.documentElement.dir='rtl';frame.style.width='361px';await wait(100);
  c.view.scrollLeft=-80;await wait(80);
  const visible=c.items.filter(i=>c.w.getComputedStyle(i).visibility!=='hidden').map(i=>i.getBoundingClientRect());
  check(variant+': RTL pan and gaps',c.view.scrollLeft<0&&visible.slice(1).every((r,i)=>Math.abs(visible[i].left-r.right-8)<.1),{scroll:c.view.scrollLeft,rects:visible.map(r=>({left:r.left,right:r.right}))});
 }
 // Leave a live uncontained viewport for a real pointer drag. Synthetic pointer
 // capture does not represent a browser gesture.
 const c=await load('uncontained');
 let nativePress;
 c.root.addEventListener('material:press',event=>{
  const index=Number(event.target.dataset.carouselIndex);
  if(event.detail.pressed){nativePress={index,before:measure(c,index)};}
  else if(nativePress){const record=nativePress;record.held=measure(c,index);nativePress=null;setTimeout(()=>{
   record.released=measure(c,index);const result=JSON.parse(out.textContent);result.nativePress=record;out.textContent=JSON.stringify(result,null,2);
  },420);}
 });
 c.view.addEventListener('pointerup',()=>{
  const released=c.view.scrollLeft;
  setTimeout(()=>{const coast=c.view.scrollLeft;setTimeout(()=>{
   const result=JSON.parse(out.textContent);result.nativeDrag={released,coast,end:c.view.scrollLeft};out.textContent=JSON.stringify(result,null,2);
  },1500);},90);
 });
 }catch(error){tests.push({name:'Completion',pass:false,detail:error.stack});}
 out.textContent=JSON.stringify({browser:navigator.userAgent,passed:tests.filter(t=>t.pass).length,failed:tests.filter(t=>!t.pass).length,tests,records},null,2);document.querySelector('#run').disabled=false;
};
