(() => {
  const clamp=(v,a,b)=>Math.max(a,Math.min(b,v));
  // Keylines map logical, equally spaced item centers to visible masked centers.
  // Media keeps its full size as the button's visible window changes width.
  function layout(width,preferred,gap,variant,count,maximum=Infinity) {
    if(variant.startsWith('uncontained')||variant==='full-screen'||width<3*40+2*gap||count<3) {
      const large=variant==='full-screen'?width:Math.min(preferred,width,maximum);
      return {large,gap,width,variant:'uncontained',keylines:[]};
    }
    let small=clamp(width*.12,40,56);
    if(variant==='center-aligned-hero'&&Number.isFinite(maximum))small=clamp(Math.max(small,(width-maximum-2*gap)/2),40,56);
    let large,big=1;
    if(variant==='center-aligned-hero')large=width-2*small-2*gap;
    else if(variant==='hero'){big=clamp(Math.max(Math.round((width-small-gap)/(preferred+gap)),Math.ceil((width-small)/(maximum+gap))),1,count-1);large=(width-small-big*gap)/big;}
    else {
      let best=Infinity;
      for(let n=1;n<=Math.min(count-2,8);n++) {
        const candidate=(width-1.5*small-(n+1)*gap)/(n+.5);
        if(candidate<=small||candidate>maximum)continue;
        const cost=Math.abs(candidate-preferred);
        if(cost<best){large=candidate;big=n;best=cost;}
      }
      if(!large)return layout(width,preferred,gap,'uncontained',count,maximum);
    }
    if(large<=small||large>maximum)return layout(width,preferred,gap,'uncontained',count,maximum);
    const keylines=[];
    if(variant==='center-aligned-hero') {
      const half=(big-1)/2;
      keylines.push({i:-half-1,x:small/2,size:small});
      for(let i=-half;i<=half;i++)keylines.push({i,x:width/2+i*(large+gap),size:large});
      keylines.push({i:half+1,x:width-small/2,size:small});
    } else {
      keylines.push({i:-1,x:-small/2-gap,size:small});
      for(let i=0;i<big;i++)keylines.push({i,x:i*(large+gap)+large/2,size:large});
      const left=big*(large+gap);
      if(variant==='hero')keylines.push({i:big,x:width-small/2,size:small});
      else {const medium=(large+small)/2;keylines.push({i:big,x:left+medium/2,size:medium},{i:big+1,x:width-small/2,size:small});}
    }
    return {width,large,gap,small,big,variant,keylines};
  }
  function keylineAt(lines,t,gap) {
    if(t<lines[0].i){const p=lines[0];return{x:p.x+(t-p.i)*(p.size+gap),size:p.size};}
    if(t>lines.at(-1).i){const p=lines.at(-1);return{x:p.x+(t-p.i)*(p.size+gap),size:p.size};}
    const right=lines.findIndex(p=>p.i>=t),b=lines[right],a=lines[Math.max(0,right-1)],f=b.i===a.i?0:(t-a.i)/(b.i-a.i);
    return{x:a.x+(b.x-a.x)*f,size:a.size+(b.size-a.size)*f};
  }
  // Each integer stop has a complete arrangement. Near the ends, move the
  // preview slots around the focal range instead of blending two distant layouts.
  function restingFrame(plan,index,position,count) {
    const {width,large,gap,small,big,variant}=plan;
    if(variant==='uncontained')return{left:(index-position)*(large+gap),size:large};
    let sizes,first;
    if(variant==='center-aligned-hero') {
      if(position===0){sizes=[large,small,small];first=0;}
      else if(position===count-1){sizes=[small,small,large];first=count-3;}
      else {sizes=[small,large,small];first=position-1;}
    } else {
      sizes=Array(big).fill(large);
      if(variant==='multi-browse')sizes.push((large+small)/2);
      sizes.push(small);
      first=Math.min(position,count-sizes.length);
      const shift=position-first;
      for(let n=0;n<shift;n++)sizes.unshift(sizes.splice(big+n,1)[0]);
    }
    const lines=[{i:first-1,x:-small/2-gap,size:small}];let x=0;
    sizes.forEach((size,k)=>{lines.push({i:first+k,x:x+size/2,size});x+=size+gap;});
    lines.push({i:first+sizes.length,x:width+gap+small/2,size:small});
    const frame=keylineAt(lines,index,gap);
    return{left:frame.x-frame.size/2,size:frame.size};
  }
  function itemFrame(plan,index,position,count) {
    const max=plan.variant==='center-aligned-hero'?count-1:count-(plan.big||1);
    position=clamp(position,0,max);
    const start=Math.floor(position),fraction=position-start;
    const a=restingFrame(plan,index,start,count);
    if(!fraction)return a;
    const b=restingFrame(plan,index,start+1,count);
    return{left:a.left+(b.left-a.left)*fraction,size:a.size+(b.size-a.size)*fraction};
  }
  // Uncontained items retain their measured media width, not a constant visible
  // mask. The trailing medium keyline follows Material's uncontained strategy.
  function uncontainedFrames(width,widths,gap,scroll,max) {
    if(!widths.length)return [];
    function arrangement(first,media=widths){
      const frames=Array(media.length);let left=0,preview=media[first],cut=false;
      for(let i=first;i<media.length;i++){
        let size=media[i];
        if(left+size>width){
          const remaining=Math.max(0,width-left);
          let medium=Math.max(gap,remaining*1.5);
          if(medium>size*.85)medium=Math.max(size*.85,remaining*1.2);
          size=Math.min(size,medium);if(!cut){preview=size;cut=true;}
        }
        frames[i]={left,size,media:media[i]};left+=size+gap;
      }
      left=-gap;
      for(let i=first-1;i>=0;i--){const size=Math.min(media[i],preview/2);left-=size;frames[i]={left,size,media:media[i]};left-=gap;}
      return frames;
    }
    const stops=[0];let start=0;
    for(const size of widths){start+=size+gap;if(start<max)stops.push(start);}
    if(max>0)stops.push(max);
    scroll=clamp(scroll,0,max);
    const end=stops.findIndex(v=>v>=scroll),last=Math.max(0,end),first=Math.max(0,last-1);
    function at(stop){
      if(stop===stops.length-1&&max>0){
        return arrangement(0,[...widths].reverse()).reverse().map(f=>({...f,left:width-f.left-f.size}));
      }
      return arrangement(stop);
    }
    const a=at(first),b=at(last),fraction=stops[last]===stops[first]?0:(scroll-stops[first])/(stops[last]-stops[first]);
    return a.map((f,i)=>({left:f.left+(b[i].left-f.left)*fraction,size:f.size+(b[i].size-f.size)*fraction,media:f.media}));
  }
  function scrollTargets(rects,rtl,max) {
    const edge=rtl?'right':'left',first=rects[0][edge];
    return rects.map(rect=>clamp(Math.abs(rect[edge]-first),0,max));
  }
  if(typeof module!=='undefined')module.exports={layout,itemFrame,scrollTargets,uncontainedFrames};
  if(typeof document==='undefined')return;
  const roots=new WeakMap(),reduced=matchMedia('(prefers-reduced-motion: reduce)');
  const items=root=>[...root.querySelectorAll('.md-carousel-item')];
  function scrollTo(state,position,behavior='instant'){
    state.viewport.scrollTo(state.vertical?{top:position,behavior}:{left:(state.rtl?-1:1)*position,behavior});
  }
  function offset(state){
    // Safari can report positions beyond either end during rubber-banding.
    // Taking abs() would fold a negative LTR overscroll into forward movement.
    return clamp(state.vertical?state.viewport.scrollTop:(state.rtl?-1:1)*state.viewport.scrollLeft,0,state.max??Infinity);
  }
  function cancelMotion(state) {
    if(state.motion){
      // Hand the visible position back to native scrolling before user input or resize.
      scrollTo(state,state.motion.position);
      state.motion.clock.cancel();cancelAnimationFrame(state.motion.frame);state.motion=null;
    }
  }
  function go(state,index,focus=false) {
    if(!state.buttons.length)return;
    index=clamp(index,0,state.buttons.length-1);
    const from=state.motion?.position??offset(state);cancelMotion(state);
    const requested=state.targets[index];state.requestedIndex=index;
    state.pendingFocus=focus&&!state.buttons[index].disabled?index:undefined;
    if(reduced.matches){
      state.viewport.style.removeProperty('scroll-snap-type');
      scrollTo(state,requested);settled(state);return;
    }
    animateTo(state,requested);
  }
  function animateTo(state,requested,duration,easing,decay) {
    const from=state.motion?.position??offset(state);cancelMotion(state);
    state.viewport.style.scrollSnapType='none';
    delete state.root.dataset.carouselSettled;
    // Resolve the real snap offset through the browser once, including its scroll
    // precision. Keep fractional keylines; rounding every item changes the layout.
    scrollTo(state,requested);
    const target=offset(state);
    if(reduced.matches||Math.abs(from-target)<.01){settled(state);return;}
    scrollTo(state,from);
    // The sticky visual track does not need native scrolling on every animation
    // frame. One token clock drives fractional geometry; sync at completion/input.
    const style=getComputedStyle(state.root),clock=new Animation(new KeyframeEffect(null,[],{
      duration:duration??parseFloat(style.getPropertyValue('--md-sys-motion-duration-medium2')),
      easing:easing??style.getPropertyValue('--md-sys-motion-easing-standard').trim(),fill:'both',
    }),document.timeline);
    const motion=state.motion={clock,position:from,frame:0};clock.play();
    function tick(){
      let progress=clock.effect.getComputedTiming().progress??0;
      if(decay)progress=(1-Math.exp(-progress*duration/decay))/(1-Math.exp(-duration/decay));
      motion.position=from+(target-from)*progress;
      if(state.flow)scrollTo(state,motion.position);
      paint(state,motion.position);
      if(clock.playState==='finished'){cancelMotion(state);settled(state);}
      else motion.frame=requestAnimationFrame(tick);
    }
    motion.frame=requestAnimationFrame(tick);
  }
  function settled(state) {
    if(state.motion||drag?.state===state&&drag.moved)return;
    state.viewport.style.removeProperty('scroll-snap-type');
    paint(state);
    if(state.requestedIndex!==undefined){state.root.dataset.carouselIndex=String(state.requestedIndex);state.requestedIndex=undefined;}
    if(state.pendingFocus!==undefined){const index=state.pendingFocus,button=state.buttons[index];state.pendingFocus=undefined;state.root.dataset.carouselIndex=String(index);button.focus({preventScroll:true});}
    state.root.dataset.carouselSettled=state.root.dataset.carouselIndex;

  }
  function paint(state,scroll=offset(state)) {
    if(state.lastPainted===scroll)return;
    state.lastPainted=scroll;
    const {root,buttons,plan}=state,position=scroll/(plan.large+plan.gap),width=plan.width;
    const max=state.max;
    const uncontained=state.flow&&!state.vertical&&!reduced.matches;
    const frames=uncontained?uncontainedFrames(width,state.widths,plan.gap,scroll,state.max):null;
    if(!state.flow||uncontained||state.vertical)buttons.forEach((button,index)=>{
      const frame=frames?.[index]??(state.vertical?{left:0,size:width,media:width}:itemFrame(plan,index,position,buttons.length));
      const left=frame.left,size=frame.size;
      // Fade brief labels before their mask becomes too narrow to read. These
      // renderer thresholds are not prescribed motion/token values.
      const slot=state.slots?.[index];
      slot?.style?.setProperty('--_carousel-caption-left',left+'px');
      slot?.style?.setProperty('--_carousel-caption-width',size+'px');
      slot?.style?.setProperty('--_carousel-text-opacity',clamp((size-80)/64,0,1));
      const press=reduced.matches?0:state.press?.button===button?state.press.value:button.dataset?.componentState==='pressed'?5:0;
      const raw=frames?state.starts[index]-scroll:0;
      button.style.visibility=left+size<=0||left>=width?'hidden':'visible';
      button.style.width=(size+2*press)+'px';
      button.style.transform=`translateX(${(state.rtl?-1:1)*(left-raw-press)}px)`;
      const art=state.art[index],media=(frame.media??plan.large)+(reduced.matches?0:10),shift=(size-media)/2+press;
      if(art){
        // Reserve the press reveal at rest: pressing must never expose an empty strip.
        art.style.width=media+'px';
        const top=index*(plan.large+plan.gap)-scroll;
        const verticalShift=state.vertical&&!reduced.matches?(Math.max(0,-top)+Math.min(plan.large,plan.large-top)-plan.large)/2:0;
        art.style.transform=state.vertical?`translate(${state.rtl?-shift:shift}px,${verticalShift}px)`:`translateX(${state.rtl?-shift:shift}px)`;
      }
    });
    let index=state.flow?state.targets.reduce((best,target,i)=>Math.abs(target-scroll)<Math.abs(state.targets[best]-scroll)?i:best,0):clamp(Math.round(position),0,buttons.length-1);
    root.dataset.carouselIndex=String(index);
    const progress=root.querySelector('.md-carousel-position');if(progress)progress.value=max>0?scroll/max:0;
    const previous=root.querySelector('[data-carousel-step="-1"]'),next=root.querySelector('[data-carousel-step="1"]');
    if(previous)previous.disabled=scroll<1;
    if(next)next.disabled=scroll>=max-1;
  }
  function resize(root) {
    const state=roots.get(root);if(!state)return;cancelMotion(state);
    const style=getComputedStyle(root),prior=state.plan?offset(state)/(state.plan.large+state.plan.gap):0,priorIndex=Number(root.dataset.carouselIndex)||0;
    state.rtl=style.direction==='rtl';state.vertical=root.dataset.variant==='full-screen';
    const viewportStyle=getComputedStyle(state.viewport),width=state.viewport.clientWidth-parseFloat(viewportStyle.paddingInlineStart)-parseFloat(viewportStyle.paddingInlineEnd);
    const preferred=parseFloat(style.getPropertyValue('--_carousel-preferred')),gap=parseFloat(style.getPropertyValue('--_carousel-gap'));
    const maximum=parseFloat(style.getPropertyValue('--_carousel-maximum'))||Infinity;
    state.plan=layout(width,preferred,gap,root.dataset.variant,state.buttons.length,maximum);
    state.flow=reduced.matches||state.plan.variant==='uncontained';
    root.dataset.carouselLayout=state.flow?'flow':'masked';
    if(reduced.matches&&!state.vertical){
      const count=root.dataset.variant==='full-screen'?1:Math.max(1,Math.floor((width+gap)/(Math.min(preferred,width)+gap)));
      state.plan.large=root.dataset.variant.includes('hero')?Math.min(preferred,width):(width-(count-1)*gap)/count;
    }
    if(state.vertical){state.plan.large=state.viewport.clientHeight;root.style.setProperty('--_carousel-page-height',state.plan.large+'px');}
    root.style.setProperty('--_carousel-width',width+'px');
    root.style.setProperty('--_carousel-logical',state.plan.large+'px');
    root.style.setProperty('--_carousel-end',(state.flow?0:Math.max(0,width-state.plan.large*(state.plan.big||1)-((state.plan.big||1)-1)*gap))+'px');
    if(state.flow)for(let i=0;i<state.buttons.length;i++){
      const button=state.buttons[i];
      for(const name of ['width','transform','visibility'])button.style[name]='';
      if(state.art[i]){state.art[i].style.transform='';state.art[i].style.width='';}
    }
    const slots=state.slots=[...root.querySelectorAll('.md-carousel-slot')];
    state.starts=[];let start=0;
    state.widths=slots.map(slot=>{state.starts.push(start);const size=slot.getBoundingClientRect().width;start+=size+gap;return size;});
    state.max=state.vertical?state.viewport.scrollHeight-state.viewport.clientHeight:state.viewport.scrollWidth-state.viewport.clientWidth;
    const snaps=[...root.querySelectorAll(state.flow?'.md-carousel-slot':'.md-carousel-snap')];
    state.targets=state.vertical?snaps.map((snap,i)=>Math.min(i*(state.plan.large+gap),state.max)):scrollTargets(snaps.map(snap=>snap.getBoundingClientRect()),state.rtl,state.max);
    state.lastPainted=undefined;
    scrollTo(state,state.flow?state.targets[priorIndex]:Math.min(prior*(state.plan.large+state.plan.gap),state.max));paint(state);
  }
  const observer=new ResizeObserver(entries=>entries.forEach(e=>resize(e.target)));
  function init(scope=document) {
    for(const root of scope.querySelectorAll('.md-carousel')) {
      // Landing-page thumbnails are static compositions without a scrolling track.
      if(!root.querySelector('.md-carousel-scroll-track'))continue;
      if(roots.has(root))continue;
      const viewport=root.querySelector('.md-carousel-viewport'),buttons=items(root),state={root,viewport,buttons,art:buttons.map(button=>button.querySelector('.md-carousel-art'))};roots.set(root,state);
      root.dataset.carouselReady='';
      viewport.addEventListener('scroll',()=>{if(!state.motion&&!(drag?.state===state&&drag.moved)){state.requestedIndex=undefined;paint(state);}},{passive:true});
      viewport.addEventListener('wheel',()=>{cancelMotion(state);state.pendingFocus=undefined;},{passive:true});
      viewport.addEventListener('scrollend',()=>settled(state));
      observer.observe(root);resize(root);
    }
  }
  document.addEventListener('click',event=>{
    const root=event.target.closest?.('.md-carousel');if(!root)return;const state=roots.get(root);if(!state)return;
    if(state.suppress&&event.detail!==0){event.preventDefault();event.stopImmediatePropagation();state.suppress=false;return;}
    const step=event.target.closest('[data-carousel-step]');if(step){
      const anchor=state.requestedIndex===undefined?offset(state):state.targets[state.requestedIndex];
      const forward=Number(step.dataset.carouselStep)>0;
      let index=forward?state.targets.findIndex(target=>target>anchor+1):state.targets.findLastIndex(target=>target<anchor-1);
      if(index>=0)go(state,index);return;
    }
  },true);
  document.addEventListener('keydown',event=>{
    const root=event.target.closest?.('.md-carousel');if(!root||!event.target.closest('.md-carousel-viewport'))return;const state=roots.get(root);if(!state)return;
    const button=event.target.closest('.md-carousel-item'),index=state.pendingFocus??(button?Number(button.dataset.carouselIndex):Number(root.dataset.carouselIndex));
    if(!state.vertical&&(event.key==='ArrowUp'||event.key==='ArrowDown')){
      const forward=event.key==='ArrowDown';
      const outside=[...document.querySelectorAll('a[href],button,input,select,textarea,[tabindex]')].filter(el=>el.tabIndex>=0&&!el.disabled&&!el.closest('[inert]')&&el.getClientRects().length&&getComputedStyle(el).visibility==='visible'&&!state.viewport.contains(el)&&(state.viewport.compareDocumentPosition(el)&(forward?Node.DOCUMENT_POSITION_FOLLOWING:Node.DOCUMENT_POSITION_PRECEDING)));
      const next=forward?outside[0]:outside.at(-1);
      if(next){event.preventDefault();next.focus();}
      return;
    }
    let target;
    if(event.key==='Tab'&&button)target=index+(event.shiftKey?-1:1);
    if(state.vertical){if(event.key==='ArrowDown')target=index+1;if(event.key==='ArrowUp')target=index-1;}
    else {if(event.key==='ArrowRight')target=index+(state.rtl?-1:1);if(event.key==='ArrowLeft')target=index+(state.rtl?1:-1);}
    if(event.key==='Home')target=0;if(event.key==='End')target=state.buttons.length-1;
    if(target===undefined)return;
    if(event.key==='Tab'&&(target<0||target>=state.buttons.length))return;
    const direction=target>=index?1:-1;
    while(target>=0&&target<state.buttons.length&&state.buttons[target].disabled)target+=direction;
    if(event.key==='Tab'&&(target<0||target>=state.buttons.length))return;
    event.preventDefault();
    if(target>=0&&target<state.buttons.length)go(state,target,true);
  });
  document.addEventListener('material:press',event=>{
    const button=event.target,root=button.closest?.('.md-carousel');
    if(!button.matches?.('.md-carousel-item')||!root)return;
    const state=roots.get(root);if(!state)return;
    const from=state.press?.button===button?state.press.value:0;
    state.press?.animation?.cancel();cancelAnimationFrame(state.press?.frame);
    // The bundled Motion Lab reference expands the mask by 10px per side at
    // 2x scale. Keep media and captured corner tokens unchanged.
    const to=event.detail.pressed&&!reduced.matches?5:0;
    const press=state.press={button,value:from,frame:0};
    if(reduced.matches){press.value=to;state.lastPainted=undefined;paint(state);return;}
    const style=getComputedStyle(root);
    const animation=press.animation=new Animation(new KeyframeEffect(null,[],{
      duration:parseFloat(style.getPropertyValue('--_carousel-press-duration')),
      easing:style.getPropertyValue('--md-sys-motion-easing-standard').trim(),fill:'both'
    }),document.timeline);animation.play();
    function tick(){
      press.value=from+(to-from)*(animation.effect.getComputedTiming().progress??0);
      state.lastPainted=undefined;paint(state,state.motion?.position??offset(state));
      if(animation.playState==='finished'){animation.cancel();press.animation=null;}
      else press.frame=requestAnimationFrame(tick);
    }
    press.frame=requestAnimationFrame(tick);
  });
  let drag;
  document.addEventListener('pointerdown',event=>{
    const root=event.target.closest?.('.md-carousel');if(!root||!event.target.closest('.md-carousel-viewport'))return;
    const state=roots.get(root);if(!state)return;cancelMotion(state);state.pendingFocus=undefined;state.suppress=false;
    if(event.pointerType!=='mouse'||event.button!==0)return;
    drag={state,id:event.pointerId,x:state.vertical?event.clientY:event.clientX,start:state.vertical?state.viewport.scrollTop:state.viewport.scrollLeft,origin:Number(root.dataset.carouselIndex),button:event.target.closest('.md-carousel-item'),moved:false,lastX:state.vertical?event.clientY:event.clientX,lastTime:event.timeStamp,velocity:0};
  });
  document.addEventListener('pointermove',event=>{
    if(!drag||event.pointerId!==drag.id)return;const pointer=drag.state.vertical?event.clientY:event.clientX,dx=pointer-drag.x;
    if(!drag.moved&&Math.abs(dx)<8)return;
    if(!drag.moved){drag.moved=true;drag.state.root.dataset.carouselDragging='';drag.state.viewport.style.scrollSnapType='none';drag.state.viewport.setPointerCapture(event.pointerId);drag.button?.dispatchEvent(new Event('material:cancel-feedback',{bubbles:true}));}
    const dt=event.timeStamp-drag.lastTime;
    if(dt>0)drag.velocity=(pointer-drag.lastX)/dt;
    drag.lastX=pointer;drag.lastTime=event.timeStamp;
    event.preventDefault();if(drag.state.vertical)drag.state.viewport.scrollTop=drag.start-dx;else drag.state.viewport.scrollLeft=drag.start-dx;
    paint(drag.state);
  });
  function finish(event,cancel=false) {
    if(!drag||event.pointerId!==drag.id)return;
    const {state,moved,id,velocity,lastTime,origin}=drag;drag=null;
    if(state.viewport.hasPointerCapture(id))state.viewport.releasePointerCapture(id);
    if(moved){
      state.suppress=true;
      if(state.root.dataset.variant.startsWith('uncontained')&&state.root.dataset.scrolling!=='snap'){
        delete state.root.dataset.carouselDragging;
        const speed=!cancel&&event.timeStamp-lastTime<80?velocity*(state.rtl?1:-1):0;
        if(!reduced.matches&&Math.abs(speed)>.02){
          // Mouse dragging has no browser inertia. A 250ms decay constant follows
          // the measured reference tail; this is renderer physics, not a token override.
          const decay=250,duration=Math.min(1500,decay*Math.log(Math.abs(speed)/.02));
          const distance=speed*decay*(1-Math.exp(-duration/decay));
          animateTo(state,offset(state)+distance,duration,'linear',decay);
        }
        else settled(state);
        return;
      }
      // Keep snapping off through the smooth settle; restoring it first can snap immediately.
      state.viewport.style.scrollSnapType='none';
      delete state.root.dataset.carouselDragging;
      const momentum=!cancel&&event.timeStamp-lastTime<80?velocity*(state.rtl?1:-1)*120:0;
      const position=offset(state)+momentum;
      let index=state.flow?state.targets.reduce((best,target,i)=>Math.abs(target-position)<Math.abs(state.targets[best]-position)?i:best,0):Math.round(position/(state.plan.large+state.plan.gap));
      if(state.vertical||state.root.dataset.variant.includes('hero'))index=clamp(index,origin-1,origin+1);
      go(state,index);
    }
  }
  document.addEventListener('pointerup',finish);document.addEventListener('pointercancel',e=>finish(e,true));
  window.addEventListener('blur',()=>{if(drag)finish({pointerId:drag.id},true);});
  document.addEventListener('htmx:beforeCleanupElement',event=>{
    const node=(event.detail?.elt || event.target);if(!node)return;
    const affected=[...(node.matches?.('.md-carousel')?[node]:[]),...node.querySelectorAll('.md-carousel')];
    for(const root of affected){observer.unobserve(root);const state=roots.get(root);if(state){cancelMotion(state);state.press?.animation?.cancel();cancelAnimationFrame(state.press?.frame);}if(drag?.state===state)drag=null;roots.delete(root);}
  });
  reduced.addEventListener?.('change',()=>{for(const root of document.querySelectorAll('.md-carousel'))resize(root);});
  document.addEventListener('htmx:afterSwap',()=>init());
  if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',()=>init());else init();
  document.addEventListener('tokens:assets-ready',()=>init());
})();
