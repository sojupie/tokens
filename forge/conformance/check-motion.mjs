// Serve this fixture and /md3 through one local origin; see the development guide.
export async function checkMotion(container) {
  const checks = [], wait = ms => new Promise(resolve => setTimeout(resolve, ms));
  const assert = (condition, message) => { if (!condition) throw new Error(message); checks.push(message); };
  let frame;
  async function load(slug, {reduced = false, fallback = false, query = ''} = {}) {
    frame?.remove();frame = document.createElement('iframe');container.append(frame);
    const bootstrap = `
      window.motionErrors=[];addEventListener('error',event=>motionErrors.push(event.message));
      const originalMedia=window.matchMedia.bind(window),queries=[];
      window.motionReduced=${reduced};
      window.matchMedia=query=>{if(!query.includes('prefers-reduced-motion'))return originalMedia(query);const media=new EventTarget();Object.defineProperties(media,{matches:{get:()=>window.motionReduced},media:{value:query}});queries.push(media);return media;};
      window.setMotionReduced=value=>{window.motionReduced=value;for(const media of queries)media.dispatchEvent(Object.assign(new Event('change'),{matches:value}));for(const sheet of document.styleSheets){const visit=rules=>{for(const rule of rules){if(rule.media?.mediaText.includes('prefers-reduced-motion')){rule.motionMedia=true;}if(rule.motionMedia)rule.media.mediaText=value?'all':'not all';else if(rule.cssRules)visit(rule.cssRules);}};try{visit(sheet.cssRules);}catch{}}};
      const originalSupports=CSS.supports.bind(CSS);CSS.supports=(...args)=>args[0]==='overlay'&&${fallback}?false:originalSupports(...args);
      addEventListener('load',()=>{setMotionReduced(window.motionReduced);${fallback ? "const style=document.createElement('style');style.textContent=':is(.md-menu,.md-field-menu,.md-fab-menu-panel)[popover]{transition-property:opacity,scale,border-radius!important}';document.head.append(style);" : ''}});
    `;
    const html = await fetch(slug ? `/md3/components/${slug}?preview.compare=interactive${query}` : '/md3/').then(r=>{if(!r.ok)throw Error(`Page ${slug}: ${r.status}`);return r.text();});
    const loaded = new Promise(resolve=>frame.onload=resolve);
    frame.srcdoc = html.replace('<head>', `<head><base href="${location.origin}/md3/"><script>${bootstrap}<\/script>`);
    await loaded;await wait(30);
    return {w:frame.contentWindow,d:frame.contentDocument};
  }
  for (const fallback of [false,true]) {
    const {w,d} = await load('menus',{fallback});
    const menu=d.querySelector('.tokens-token-set-menu'),trigger=d.querySelector('[popovertarget="tokens-token-set-menu"]');
    const root=w.getComputedStyle(d.documentElement);
    assert(root.getPropertyValue('--md-sys-motion-spring-fast-spatial-damping').trim()==='.6'||Number(root.getPropertyValue('--md-sys-motion-spring-fast-spatial-damping'))===.6,'Expressive spatial damping reaches the rendered page');
    for(const theme of ['light','dark'])for(const contrast of ['default','medium','high']){d.documentElement.dataset.theme=theme;d.documentElement.dataset.contrast=contrast;assert(Number(w.getComputedStyle(d.documentElement).getPropertyValue('--md-sys-motion-spring-fast-spatial-stiffness'))===800,`${theme}/${contrast} retains the Expressive scheme`);}
    trigger.click();await wait(40);
    assert(menu.matches(':popover-open'),'Menu opens');
    assert(w.getComputedStyle(menu).transitionTimingFunction.includes('linear('),'Menu uses the sampled spring curve');
    menu.hidePopover();
    assert(!menu.matches(':popover-open'),'Menu closes without reopening');
    if(fallback){const copy=d.querySelector('[data-menu-exit-copy]');assert(copy?.inert&&copy.getAttribute('aria-hidden')==='true','Fallback exit is inert and hidden from assistive technology');assert(copy.getAnimations().length>=2,'Fallback retains separate geometry and opacity animations');}
    trigger.click();await wait(20);assert(!d.querySelector('[data-menu-exit-copy]'),'Reopening removes the previous exit copy');
    menu.hidePopover();await wait(550);assert(!d.querySelector('[data-menu-exit-copy]'),'Exit copies clean up after completion');
    assert(!w.motionErrors.length,'Menu lifecycle has no script errors');
  }
  {
    const {w,d}=await load('fab-menu',{fallback:true}),trigger=d.querySelector('.md-fab-menu-opener'),panel=d.querySelector('.md-fab-menu-panel');
    trigger.click();await wait(850);
    const before=panel.querySelector('.md-fab-menu-item'),beforeStyle=w.getComputedStyle(before),paint={background:beforeStyle.backgroundColor,color:beforeStyle.color,height:before.getBoundingClientRect().height,width:before.getBoundingClientRect().width};
    const start=w.performance.now();panel.hidePopover();
    assert(w.performance.now()-start<50,'FAB fallback exit setup stays below the long-task threshold');
    const copy=d.querySelector('[data-menu-exit-copy]');assert(copy?.inert,'FAB exit receives a retained surface without overlay support');
    assert([...copy.querySelectorAll('.md-fab-menu-item')].some(item=>item.getAnimations().length),'FAB fallback mirrors the item spring animations');
    const copied=copy.querySelector('.md-fab-menu-item'),copiedStyle=w.getComputedStyle(copied);
    assert(copiedStyle.backgroundColor===paint.background&&copiedStyle.color===paint.color,'FAB exit copy retains its resolved colors');
    assert(Math.abs(copied.getBoundingClientRect().height-paint.height)<.1&&Math.abs(copied.getBoundingClientRect().width-paint.width)<.1,'FAB exit copy retains its starting geometry');
    await wait(600);trigger.click();await wait(650);
    const actionStart=w.performance.now();panel.querySelector('[data-fab-menu-action]').click();
    assert(w.performance.now()-actionStart<50&&!panel.matches(':popover-open'),'FAB item selection dismisses without blocking');
    await wait(1000);assert(!d.querySelector('[data-menu-exit-copy]'),'FAB exit copy cleans up');assert(!w.motionErrors.length,'FAB lifecycle has no script errors');
  }
  for(const slug of ['date-pickers','time-pickers']) {
    const {w,d}=await load(slug,{query:slug==='date-pickers'?'&preview.variant=modal':''});
    const date=slug==='date-pickers',opener=d.querySelector(date?'[data-date-open]':'[data-time-open]'),panel=d.querySelector(date?'.md-date-panel':'.md-time-panel');
    opener.click();await wait(30);const before=Number(w.getComputedStyle(panel).opacity);
    panel.dispatchEvent(new w.Event('cancel',{cancelable:true}));
    const exit=panel.getAnimations().find(a=>a.effect.getKeyframes().some(k=>'opacity'in k));
    const from=Number(exit.effect.getKeyframes()[0].opacity);
    assert(before<1&&from<1,`${slug}: interrupted entry does not restart exit at full opacity`);
    assert(panel.inert,`${slug}: closing picker does not accept input`);
    await wait(250);assert(!panel.open&&!panel.inert,`${slug}: dismissal finishes and clears inert`);
    opener.click();await wait(25);panel.dispatchEvent(new w.Event('cancel',{cancelable:true}));w.setMotionReduced(true);await wait(30);
    assert(!panel.open&&!panel.inert,`${slug}: changing reduced motion finishes an in-flight dismissal`);
    if(date){for(const icon of d.querySelectorAll('.md-date-menu-button > svg'))assert(w.getComputedStyle(icon).transitionDuration==='0s','Reduced motion disables the date-picker arrow');}
    assert(!w.motionErrors.length,`${slug}: lifecycle has no script errors`);
  }
  {
    const {w,d}=await load('search',{reduced:true});const input=d.querySelector('.md-search-input');input.focus();await wait(30);
    for(const icon of d.querySelectorAll('.md-search-leading-slot > *'))assert(w.getComputedStyle(icon).transitionDuration==='0s','Reduced motion disables search icon rotation and fading');
    assert(!w.motionErrors.length,'Reduced-motion search has no script errors');
  }
  {
    const {w,d}=await load('bottom-sheets',{query:'&preview.control=modal'}),opener=d.querySelector('[data-sheet-open]');opener.click();await wait(40);
    const sheet=d.querySelector('.md-sheet');assert(sheet.open||!sheet.hidden,'Sheet opens');
    assert(sheet.getAnimations().some(a=>a.effect.getTiming().easing.includes('linear(')),'Sheet geometry uses the default spatial spring');
    sheet.dispatchEvent(new w.Event('cancel',{cancelable:true}));await wait(600);
    assert(!sheet.open,'Sheet exit completes');assert(!w.motionErrors.length,'Sheet lifecycle has no script errors');
  }
  {
    const {w,d}=await load('navigation-rail'),frame=d.querySelector('.md-rail-frame'),rail=frame.querySelector('.md-navigation-rail');
    await d.fonts.ready;
    for(const direction of ['ltr','rtl'])for(const expanded of [true,false]){
      rail.style.direction=direction;rail.querySelector('[data-navigation-expand]').click();
      assert(rail.dataset.variant===(expanded?'expanded':'collapsed'),`Rail ${direction}: switches to the requested variant`);
      const nodes=[...frame.querySelectorAll('.md-nav-indicator,.md-nav-icon,.md-nav-label,.md-navigation-slot,.md-rail-body,.md-rail-screen,.md-rail-screen h3,.md-rail-screen article:first-of-type')].filter(n=>n.getClientRects().length);
      const animations=frame.getAnimations({subtree:true});
      const directory=frame.querySelector('.md-rail-body');
      for(const a of animations){a.pause();a.currentTime=0;}
      const directoryWidth=directory.getBoundingClientRect().width;
      for(const a of animations)a.currentTime=Number(a.effect.getComputedTiming().endTime)/2;
      assert(Math.abs(directory.getBoundingClientRect().width-directoryWidth)<.1,'Rail directory text keeps one layout width throughout motion');
      assert(rail.querySelectorAll('.md-nav-label').length>0&&[...rail.querySelectorAll('.md-nav-label')].every(n=>n.getAnimations().every(a=>!a.effect.getKeyframes().some(k=>'transform' in k))),'Rail labels move without animated transform layers');
      for(const a of animations){a.pause();a.currentTime=Number(a.effect.getComputedTiming().endTime)-.001;}
      const before=nodes.map(n=>n.getBoundingClientRect());
      animations.forEach(a=>a.finish());await Promise.all(animations.map(a=>a.finished.catch(()=>{})));await new Promise(w.requestAnimationFrame);
      const jump=Math.max(...nodes.map((n,i)=>{const r=n.getBoundingClientRect();return Math.max(Math.abs(r.x-before[i].x),Math.abs(r.y-before[i].y));}));
      assert(jump<.1,`Rail ${direction}/${expanded?'expanded':'collapsed'}: final layout stays within 0.1px of the animation endpoint`);
      assert(!frame.hasAttribute('data-transitioning'),'Rail clears the temporary transition layout');
    }
    assert(!w.motionErrors.length,'Rail transitions have no script errors');
  }
  {
    const {w,d}=await load('navigation-rail',{query:'&preview.layout=modal'}),frame=d.querySelector('.md-rail-frame'),rail=frame.querySelector('.md-navigation-rail'),scrim=frame.querySelector('.md-navigation-scrim'),menu=rail.querySelector('[data-navigation-expand]');
    for(const close of ['scrim','menu','escape']){
      menu.click();await wait(600);
      assert(!scrim.hidden&&Number(w.getComputedStyle(scrim).opacity)===1,'Modal rail reopens with a visible scrim');
      if(close==='escape')menu.dispatchEvent(new w.KeyboardEvent('keydown',{key:'Escape',bubbles:true}));else(close==='scrim'?scrim:menu).click();
      const fade=scrim.getAnimations().find(a=>a.effect.getKeyframes().some(k=>'opacity' in k));
      assert(!!fade,`Modal rail ${close}: scrim has an exit fade`);
      await fade.finished;
      assert(frame.hasAttribute('data-transitioning'),`Modal rail ${close}: opacity settles before spatial motion`);
      for(let i=0;i<3;i++){await new Promise(w.requestAnimationFrame);assert(scrim.hidden||Number(w.getComputedStyle(scrim).opacity)<.001,`Modal rail ${close}: scrim stays transparent after the fade`);}
      await wait(500);
      assert(scrim.hidden&&frame.dataset.modalOpen==='false',`Modal rail ${close}: cleanup hides the scrim`);
    }
    assert(!w.motionErrors.length,'Modal rail scrim lifecycle has no script errors');
  }
  {
    const {w,d}=await load(''),art=d.querySelector('[data-artwork]'),toggle=d.querySelector('.tokens-art-pause input');
    const animations=art.getAnimations({subtree:true});
    assert(animations.length===22,'Hero starts its layout, stripe and corner animations');
    toggle.click();assert(animations.every(a=>a.playState==='paused'),'Hero pause control pauses every track');
    toggle.click();w.setMotionReduced(true);assert(animations.every(a=>a.playState==='paused'),'Hero reduced motion pauses every track');
    w.setMotionReduced(false);assert(animations.every(a=>a.playState==='running'),'Hero resumes after reduced motion clears');
    animations.forEach(a=>a.pause());
    for(const width of [960,390]){
      frame.style.width=width+'px';await new Promise(w.requestAnimationFrame);
      let maxGap=0,turned=false,endpoint=false;
      for(let time=0;time<=8300;time+=50){
        animations.forEach(a=>a.currentTime=time);
        for(const node of d.querySelectorAll('.tokens-art-sections')){
          const style=w.getComputedStyle(node),tile=w.getComputedStyle(node.closest('.tokens-art-tile'));
          const angle=parseFloat(style.getPropertyValue('--_art-angle')),r=angle*Math.PI/180;
          const tw=parseFloat(tile.width),th=parseFloat(tile.height),rw=parseFloat(style.width),rh=parseFloat(style.height);
          assertFinite(rw,rh);
          maxGap=Math.max(maxGap,(Math.abs(tw*Math.cos(r))+Math.abs(th*Math.sin(r))-rw)/2,(Math.abs(tw*Math.sin(r))+Math.abs(th*Math.cos(r))-rh)/2);
          if(angle<0||angle>90)throw Error('Hero stripe turn overshoots');
          turned ||= angle>1&&angle<89;
          if(angle===90){endpoint=true;if(Math.abs(rw-th)>.1||Math.abs(rh-tw)>.1)throw Error('Stripe thirds lose their proportions at a quarter turn');}
        }
      }
      assert(maxGap<.1&&turned&&endpoint,`Hero ${width}px: full-loop stripe coverage and endpoint proportions`);
      assert([...d.querySelectorAll('.tokens-art-surface')].every(n=>w.getComputedStyle(n).transform==='none'),`Hero ${width}px: tile shells never rotate or contract for a turn`);
    }
    assert(!w.motionErrors.length,'Hero has no script errors');
    function assertFinite(...values){if(!values.every(value=>Number.isFinite(value)&&value>0))throw Error('Invalid stripe dimensions');}
  }
  const response=await fetch('/md3/downloads/dtcg/light-default/report.json').then(r=>r.json());
  assert(response.context['Motion scheme']==='Standard','Public download context remains Standard');
  frame.remove();return {passed:checks.length,checks};
}
