// Run against the real Catalog page through the local conformance proxy.
export async function checkCarousel(container) {
  const checks=[],assert=(ok,message)=>{if(!ok)throw Error(message);checks.push(message);};
  const frame=document.createElement('iframe');container.replaceChildren(frame);
  const loaded=new Promise(resolve=>frame.onload=resolve);
  frame.src='/md3/components/carousel?preview.compare=variants';
  await loaded;const w=frame.contentWindow,d=frame.contentDocument;await w.TokenAssets?.ready;
  frame.scrollIntoView();
  const wait=ms=>new Promise(resolve=>setTimeout(resolve,ms));
  async function settle(root){for(let i=0;i<100;i++){await wait(20);if(root.dataset.carouselSettled!==undefined)return;}throw Error('Carousel did not settle');}
  const results=[];
  for(const width of [960,390])for(const direction of ['ltr','rtl']) {
    d.documentElement.dir=direction;frame.style.width=width+'px';
    for(const root of d.querySelectorAll('.md-carousel'))root.style.inlineSize=(root.clientWidth-1)+'px';
    await wait(40);for(const root of d.querySelectorAll('.md-carousel'))root.style.removeProperty('inline-size');await wait(40);
    for(const root of d.querySelectorAll('.md-carousel')) {
      const viewport=root.querySelector('.md-carousel-viewport');if(!viewport)continue;
      const name=`${width}/${direction}/${root.dataset.variant}`;
      root.scrollIntoView({block:'center'});
      const track=root.querySelector('.md-carousel-track');
      assert(w.getComputedStyle(track).position==='sticky',`${name}: visual track uses native sticky positioning`);
      const previous=root.querySelector('[data-carousel-step="-1"]'),next=root.querySelector('[data-carousel-step="1"]');
      root.querySelector('.md-carousel-item').dispatchEvent(new w.KeyboardEvent('keydown',{key:'Home',bubbles:true,cancelable:true}));await settle(root);
      const geometry=()=>[...root.querySelectorAll('.md-carousel-item')].map(button=>button.style.width+' '+button.style.transform).join('|');
      const atStart=geometry();
      Object.defineProperty(viewport,'scrollLeft',{configurable:true,get:()=>direction==='rtl'?20:-20});
      viewport.dispatchEvent(new w.Event('scroll'));
      assert(geometry()===atStart,`${name}: outward overscroll does not fold into forward item motion`);
      delete viewport.scrollLeft;
      const art=root.querySelector('.md-carousel-art');
      assert(w.getComputedStyle(art).insetInlineStart==='0px'&&w.getComputedStyle(art).translate==='none',`${name}: artwork uses explicit fractional centering`);
      let recording=true;const samples=[];
      const sample=time=>{samples.push({time,x:Math.abs(viewport.scrollLeft),left:new w.DOMMatrixReadOnly(w.getComputedStyle(root.querySelectorAll('.md-carousel-item')[1]).transform).m41});if(recording)w.requestAnimationFrame(sample);};
      w.requestAnimationFrame(sample);next.click();await settle(root);
      const settledX=viewport.scrollLeft,settledTransform=root.querySelectorAll('.md-carousel-item')[1].style.transform;
      await wait(160);recording=false;await wait(30);
      assert(viewport.scrollLeft===settledX&&root.querySelectorAll('.md-carousel-item')[1].style.transform===settledTransform,`${name}: restoring native snapping has no late correction`);
      assert(new Set(samples.map(s=>s.x)).size<=2,`${name}: visual animation does not scroll the sticky layer each frame`);
      const moving=samples.filter((s,i,a)=>i&&s.left!==a[i-1].left);
      assert(moving.length>2,`${name}: animated geometry`);
      assert(moving.some(s=>Math.abs(s.left-Math.round(s.left))>.01),`${name}: fractional transforms during settling`);
      assert(root.dataset.carouselSettled==='1',`${name}: next settles once on item 2`);
      for(let i=1;i<samples.length;i++)assert(samples[i].x>=samples[i-1].x-1,`${name}: no scroll reversal`);
      previous.click();await settle(root);assert(root.dataset.carouselSettled==='0'&&previous.disabled,`${name}: previous returns to start`);
      const last=root.querySelectorAll('.md-carousel-item')[5];
      last.dispatchEvent(new w.KeyboardEvent('keydown',{key:'End',bubbles:true,cancelable:true}));await settle(root);
      assert(next.disabled&&d.activeElement===last,`${name}: End reaches last item and transfers focus`);
      const end=(viewport.scrollWidth-viewport.clientWidth)*(direction==='rtl'?-1:1);
      Object.defineProperty(viewport,'scrollLeft',{configurable:true,get:()=>end});viewport.dispatchEvent(new w.Event('scroll'));
      const atBoundary=geometry(),beyondEnd=end+(direction==='rtl'?-20:20);
      Object.defineProperty(viewport,'scrollLeft',{configurable:true,get:()=>beyondEnd});viewport.dispatchEvent(new w.Event('scroll'));
      assert(geometry()===atBoundary,`${name}: end overscroll keeps boundary geometry stable`);delete viewport.scrollLeft;viewport.dispatchEvent(new w.Event('scroll'));
      const boxes=[...root.querySelectorAll('.md-carousel-item')].filter(b=>w.getComputedStyle(b).visibility!=='hidden').map(b=>b.getBoundingClientRect()).sort((a,b)=>a.left-b.left);
      for(let i=1;i<boxes.length;i++)assert(boxes[i].left-boxes[i-1].right>=7.9,`${name}: item gaps preserved`);
      previous.click();
      viewport.dispatchEvent(new w.PointerEvent('pointerdown',{pointerType:'touch',button:0,bubbles:true}));await wait(50);
      const held=viewport.scrollLeft;await wait(100);assert(viewport.scrollLeft===held,`${name}: touch interrupts settling`);
      results.push({name,movingFrames:moving.length});
    }
  }
  // The explicit artwork transform must also account for the outlined mask border.
  const outlinedLoaded=new Promise(resolve=>frame.onload=resolve);
  frame.src='/md3/components/carousel?preview.compare=interactive&preview.control=outlined';
  await outlinedLoaded;await frame.contentWindow.TokenAssets?.ready;
  const outlined=frame.contentDocument.querySelector('.md-carousel');outlined.scrollIntoView();
  const centered=()=>{
    for(const button of outlined.querySelectorAll('.md-carousel-item')){
      if(frame.contentWindow.getComputedStyle(button).visibility==='hidden')continue;
      const b=button.getBoundingClientRect(),a=button.querySelector('.md-carousel-art').getBoundingClientRect();
      if(b.width<4)continue; // A clipped sliver can be narrower than the two borders.
      assert(Math.abs((b.left+b.right-a.left-a.right)/2)<.04,'Outlined artwork stays centered in its fractional mask');
    }
  };
  centered();outlined.querySelector('[data-carousel-step="1"]').click();
  await new Promise((resolve,reject)=>{const sample=()=>{try{centered();if(outlined.dataset.carouselSettled!==undefined)resolve();else frame.contentWindow.requestAnimationFrame(sample);}catch(error){reject(error);}};frame.contentWindow.requestAnimationFrame(sample);});
  // A reduced-motion frame exercises the same real script with its media query set.
  const html=await fetch('/md3/components/carousel?preview.compare=interactive').then(r=>r.text());
  const reducedLoaded=new Promise(resolve=>frame.onload=resolve);
  frame.srcdoc=html.replace('<head>',`<head><base href="${location.origin}/md3/"><script>const media=window.matchMedia.bind(window);window.matchMedia=q=>q.includes('prefers-reduced-motion')?{matches:true,addEventListener(){}}:media(q);<\/script>`);
  await reducedLoaded;await frame.contentWindow.TokenAssets?.ready;
  const root=frame.contentDocument.querySelector('.md-carousel');root.scrollIntoView();
  root.querySelector('[data-carousel-step="1"]').click();
  assert(root.dataset.carouselSettled==='1','Reduced motion settles immediately');
  return {checks:checks.length,results};
}
