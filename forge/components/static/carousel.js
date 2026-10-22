(() => {
  const clamp=(v,a,b)=>Math.max(a,Math.min(b,v));
  // Keylines map logical, equally spaced item centers to visible masked centers.
  // Media keeps its full size as the button's visible window changes width.
  function layout(width,preferred,gap,variant,count) {
    if(variant==='uncontained'||variant==='full-screen'||width<176||count<3) {
      const large=variant==='full-screen'?width:Math.min(preferred,width);
      return {large,gap,width,variant:'uncontained',keylines:[]};
    }
    const small=56;
    let large,big=1;
    if(variant==='hero')large=width-2*(small+gap);
    else {
      let best=Infinity;
      for(let n=1;n<=Math.min(count-2,8);n++) {
        const candidate=(width-1.5*small-(n+1)*gap)/(n+.5);
        if(candidate<small*2)continue;
        const cost=Math.abs(candidate-preferred);
        if(cost<best){large=candidate;big=n;best=cost;}
      }
      if(!large)return layout(width,preferred,gap,'uncontained',count);
    }
    const keylines=[];
    if(variant==='hero') {
      keylines.push({i:-1,x:small/2,size:small},{i:0,x:width/2,size:large},{i:1,x:width-small/2,size:small});
    } else {
      keylines.push({i:-1,x:-small/2-gap,size:small});
      for(let i=0;i<big;i++)keylines.push({i,x:i*(large+gap)+large/2,size:large});
      const medium=(large+small)/2,left=big*(large+gap);
      keylines.push({i:big,x:left+medium/2,size:medium},{i:big+1,x:width-small/2,size:small});
    }
    return {width,large,gap,small,variant,keylines};
  }
  function keylineAt(lines,t,gap) {
    if(t<lines[0].i){const p=lines[0];return{x:p.x+(t-p.i)*(p.size+gap),size:p.size};}
    if(t>lines.at(-1).i){const p=lines.at(-1);return{x:p.x+(t-p.i)*(p.size+gap),size:p.size};}
    const right=lines.findIndex(p=>p.i>=t),b=lines[right],a=lines[Math.max(0,right-1)],f=b.i===a.i?0:(t-a.i)/(b.i-a.i);
    return{x:a.x+(b.x-a.x)*f,size:a.size+(b.size-a.size)*f};
  }
  function itemFrame(plan,index,position,count) {
    const {width,large,gap,variant}=plan,step=large+gap,t=index-position;
    if(variant==='uncontained')return{left:t*step,size:large};
    let p=keylineAt(plan.keylines,t,gap);
    if(variant==='hero') {
      // Edge items align to the edge; interior focal items are centered.
      p.x+=(plan.small+gap)*(clamp(position-(count-2),0,1)-clamp(1-position,0,1));
    } else {
      const span=Math.min(plan.keylines.at(-1).i,count-1),mix=clamp((position-(count-1-span))/span,0,1);
      if(mix>0) {
        // Interpolate fixed endpoint layouts. Moving both layouts during the blend
        // makes edges reverse direction and item widths repeatedly shrink and grow.
        const start=keylineAt(plan.keylines,index-(count-1-span),gap);
        const end=keylineAt(plan.keylines,count-1-index,gap);
        p={x:start.x*(1-mix)+(width-end.x)*mix,size:start.size*(1-mix)+end.size*mix};
      }
    }
    return{left:p.x-p.size/2,size:p.size};
  }
  if(typeof module!=='undefined')module.exports={layout,itemFrame};
  if(typeof document==='undefined')return;
  const roots=new WeakMap(),reduced=matchMedia('(prefers-reduced-motion: reduce)');
  const items=root=>[...root.querySelectorAll('.md-carousel-item')];
  function offset(state){
    // Safari can report positions beyond either end during rubber-banding.
    // Taking abs() would fold a negative LTR overscroll into forward movement.
    return clamp((state.rtl?-1:1)*state.viewport.scrollLeft,0,state.max??Infinity);
  }
  function cancelMotion(state) {
    if(state.motion){
      // Hand the visible position back to native scrolling before user input or resize.
      state.viewport.scrollTo({left:(state.rtl?-1:1)*state.motion.position,behavior:'instant'});
      state.motion.clock.cancel();cancelAnimationFrame(state.motion.frame);state.motion=null;
    }
  }
  function go(state,index,focus=false) {
    index=clamp(index,0,state.buttons.length-1);
    const from=state.motion?.position??offset(state);cancelMotion(state);
    const requested=state.targets[index];
    state.pendingFocus=focus&&!state.buttons[index].disabled?index:undefined;
    state.viewport.style.scrollSnapType='none';
    delete state.root.dataset.carouselSettled;
    // Resolve the real snap offset through the browser once, including its scroll
    // precision. Keep fractional keylines; rounding every item changes the layout.
    state.viewport.scrollTo({left:(state.rtl?-1:1)*requested,behavior:'instant'});
    const target=offset(state);
    if(reduced.matches||Math.abs(from-target)<.01){settled(state);return;}
    state.viewport.scrollTo({left:(state.rtl?-1:1)*from,behavior:'instant'});
    // The sticky visual track does not need native scrolling on every animation
    // frame. One token clock drives fractional geometry; sync at completion/input.
    const style=getComputedStyle(state.root),clock=new Animation(new KeyframeEffect(null,[],{
      duration:parseFloat(style.getPropertyValue('--md-sys-motion-duration-medium2')),
      easing:style.getPropertyValue('--md-sys-motion-easing-standard').trim(),fill:'both',
    }),document.timeline);
    const motion=state.motion={clock,position:from,frame:0};clock.play();
    function tick(){
      const progress=clock.effect.getComputedTiming().progress??0;
      motion.position=from+(target-from)*progress;
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
    if(state.pendingFocus!==undefined){const button=state.buttons[state.pendingFocus];state.pendingFocus=undefined;button.focus({preventScroll:true});}
    state.root.dataset.carouselSettled=state.root.dataset.carouselIndex;
    state.root.querySelector('output').textContent=`Item ${Number(state.root.dataset.carouselIndex)+1} of ${state.buttons.length}`;
  }
  function paint(state,scroll=offset(state)) {
    if(state.lastPainted===scroll)return;
    state.lastPainted=scroll;
    const {root,buttons,plan}=state,position=scroll/(plan.large+plan.gap),width=plan.width;
    const max=state.max;
    buttons.forEach((button,index)=>{
      const frame=itemFrame(plan,index,position,buttons.length),left=Math.max(0,frame.left),right=Math.min(width,frame.left+frame.size),size=Math.max(0,right-left);
      // Outside items remain in the native scroll track without extending its bounds.
      button.style.visibility=size<1?'hidden':'visible';
      button.style.width=size+'px';
      button.style.transform=`translateX(${state.rtl?-left:left}px)`;
      // Position artwork from the same fractional mask width, without percentage
      // layout on a changing box. Static thumbnails retain their CSS centering.
      const art=state.art[index],shift=(size-plan.large)/2;
      if(art)art.style.transform=`translateX(${state.rtl?-shift:shift}px)`;
    });
    const index=clamp(Math.round(position),0,buttons.length-1);
    root.dataset.carouselIndex=String(index);
    root.querySelector('[data-carousel-step="-1"]').disabled=scroll<1;
    root.querySelector('[data-carousel-step="1"]').disabled=scroll>=max-1;
  }
  function resize(root) {
    const state=roots.get(root);if(!state)return;cancelMotion(state);
    const style=getComputedStyle(root),prior=state.plan?offset(state)/(state.plan.large+state.plan.gap):0;
    state.rtl=style.direction==='rtl';
    const viewportStyle=getComputedStyle(state.viewport),width=state.viewport.clientWidth-parseFloat(viewportStyle.paddingInlineStart)-parseFloat(viewportStyle.paddingInlineEnd);
    state.plan=layout(width,parseFloat(style.getPropertyValue('--_carousel-preferred')),parseFloat(style.getPropertyValue('--_carousel-gap')),root.dataset.variant,state.buttons.length);
    root.style.setProperty('--_carousel-width',width+'px');
    root.style.setProperty('--_carousel-logical',state.plan.large+'px');
    root.style.setProperty('--_carousel-end',(root.dataset.variant==='uncontained'?0:Math.max(0,state.plan.width-state.plan.large))+'px');
    state.max=state.viewport.scrollWidth-state.viewport.clientWidth;
    const snaps=[...root.querySelectorAll('.md-carousel-snap')],first=snaps[0].getBoundingClientRect().left;
    state.targets=snaps.map(snap=>clamp(Math.abs(snap.getBoundingClientRect().left-first),0,state.max));
    state.lastPainted=undefined;
    state.viewport.scrollLeft=(state.rtl?-1:1)*prior*(state.plan.large+state.plan.gap);paint(state);
  }
  const observer=new ResizeObserver(entries=>entries.forEach(e=>resize(e.target)));
  function init(scope=document) {
    for(const root of scope.querySelectorAll('.md-carousel')) {
      // Landing-page thumbnails are static compositions without a scrolling track.
      if(!root.querySelector('.md-carousel-scroll-track'))continue;
      if(roots.has(root))continue;
      const viewport=root.querySelector('.md-carousel-viewport'),buttons=items(root),state={root,viewport,buttons,art:buttons.map(button=>button.querySelector('.md-carousel-art'))};roots.set(root,state);
      root.dataset.carouselReady='';
      viewport.addEventListener('scroll',()=>{if(!state.motion&&!(drag?.state===state&&drag.moved))paint(state);},{passive:true});
      viewport.addEventListener('wheel',()=>{cancelMotion(state);state.pendingFocus=undefined;},{passive:true});
      viewport.addEventListener('scrollend',()=>settled(state));
      observer.observe(root);resize(root);
    }
  }
  document.addEventListener('click',event=>{
    const root=event.target.closest?.('.md-carousel');if(!root)return;const state=roots.get(root);if(!state)return;
    if(state.suppress&&event.detail!==0){event.preventDefault();event.stopImmediatePropagation();state.suppress=false;return;}
    const step=event.target.closest('[data-carousel-step]');if(step){go(state,Number(root.dataset.carouselIndex)+Number(step.dataset.carouselStep));return;}
    const button=event.target.closest('.md-carousel-item');if(button&&!button.disabled)root.querySelector('output').textContent=`Item ${Number(button.dataset.carouselIndex)+1} selected`;
  },true);
  document.addEventListener('keydown',event=>{
    const root=event.target.closest?.('.md-carousel');if(!root||event.target.closest('[data-carousel-step]'))return;const state=roots.get(root);if(!state)return;
    const button=event.target.closest('.md-carousel-item'),index=state.pendingFocus??(button?Number(button.dataset.carouselIndex):Number(root.dataset.carouselIndex));
    let target;
    if(event.key==='ArrowRight')target=index+(state.rtl?-1:1);
    if(event.key==='ArrowLeft')target=index+(state.rtl?1:-1);
    if(event.key==='Home')target=0;if(event.key==='End')target=state.buttons.length-1;
    if(target===undefined)return;event.preventDefault();
    const direction=target>=index?1:-1;
    while(target>=0&&target<state.buttons.length&&state.buttons[target].disabled)target+=direction;
    go(state,target,true);
  });
  let drag;
  document.addEventListener('pointerdown',event=>{
    const root=event.target.closest?.('.md-carousel');if(!root||!event.target.closest('.md-carousel-viewport'))return;
    const state=roots.get(root);if(!state)return;cancelMotion(state);state.pendingFocus=undefined;state.suppress=false;
    if(event.pointerType!=='mouse'||event.button!==0)return;
    drag={state,id:event.pointerId,x:event.clientX,start:state.viewport.scrollLeft,button:event.target.closest('.md-carousel-item'),moved:false,lastX:event.clientX,lastTime:event.timeStamp,velocity:0};
  });
  document.addEventListener('pointermove',event=>{
    if(!drag||event.pointerId!==drag.id)return;const dx=event.clientX-drag.x;
    if(!drag.moved&&Math.abs(dx)<8)return;
    if(!drag.moved){drag.moved=true;drag.state.root.dataset.carouselDragging='';drag.state.viewport.setPointerCapture(event.pointerId);drag.button?.dispatchEvent(new Event('material:cancel-feedback',{bubbles:true}));}
    const dt=event.timeStamp-drag.lastTime;
    if(dt>0)drag.velocity=(event.clientX-drag.lastX)/dt;
    drag.lastX=event.clientX;drag.lastTime=event.timeStamp;
    event.preventDefault();drag.state.viewport.scrollLeft=drag.start-dx;
    paint(drag.state);
  });
  function finish(event,cancel=false) {
    if(!drag||event.pointerId!==drag.id)return;
    const {state,moved,id,velocity,lastTime}=drag;drag=null;
    if(state.viewport.hasPointerCapture(id))state.viewport.releasePointerCapture(id);
    if(moved){
      state.suppress=true;
      // Keep snapping off through the smooth settle; restoring it first can snap immediately.
      state.viewport.style.scrollSnapType='none';
      delete state.root.dataset.carouselDragging;
      const momentum=!cancel&&event.timeStamp-lastTime<80?velocity*(state.rtl?1:-1)*120:0;
      go(state,Math.round((offset(state)+momentum)/(state.plan.large+state.plan.gap)));
    }
  }
  document.addEventListener('pointerup',finish);document.addEventListener('pointercancel',e=>finish(e,true));
  window.addEventListener('blur',()=>{if(drag)finish({pointerId:drag.id},true);});
  document.addEventListener('htmx:beforeCleanupElement',event=>{
    const node=event.detail?.elt;if(!node)return;
    const affected=[...(node.matches?.('.md-carousel')?[node]:[]),...node.querySelectorAll('.md-carousel')];
    for(const root of affected){observer.unobserve(root);const state=roots.get(root);if(state)cancelMotion(state);if(drag?.state===state)drag=null;roots.delete(root);}
  });
  reduced.addEventListener?.('change',()=>{if(reduced.matches)for(const root of document.querySelectorAll('.md-carousel')){const state=roots.get(root);if(state?.motion){state.motion.clock.finish();}}});
  document.addEventListener('htmx:afterSwap',()=>init());
  if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',()=>init());else init();
  document.addEventListener('tokens:assets-ready',()=>init());
})();
