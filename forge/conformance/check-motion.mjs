import { checkComponentDemos } from './check-component-demos.mjs';

// Serve this fixture and /md3 through one local origin; see the development guide.
export async function checkMotion(container) {
  const checks = [], wait = ms => new Promise(resolve => setTimeout(resolve, ms));
  const assert = (condition, message) => { if (!condition) throw new Error(message); checks.push(message); };
  let frame;
  async function load(slug, {reduced = false, fallback = false, query = ''} = {}) {
    frame?.remove();frame = document.createElement('iframe');frame.style.width=Math.min(960,document.documentElement.clientWidth-48)+'px';container.append(frame);
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
    await loaded;frame.scrollIntoView({block:'start'});await wait(30);
    return {w:frame.contentWindow,d:frame.contentDocument};
  }
  await checkComponentDemos(load, assert);
  {
    const {w,d}=await load('navigation-rail'),input=d.querySelector('#preview-variant'),root=input.closest('[data-field-example]'),menu=root.querySelector('.md-field-menu'),option=menu.querySelector('[data-field-value="expanded"]');
    let changes=0;input.addEventListener('change',event=>{changes++;event.stopPropagation();});input.addEventListener('input',event=>event.stopPropagation());
    const pointer=(type,target)=>target.dispatchEvent(new w.PointerEvent(type,{bubbles:true,cancelable:true,pointerType:'touch',pointerId:7,button:0,isPrimary:true}));
    input.click();pointer('pointerdown',option);input.dispatchEvent(new w.FocusEvent('focusout',{bubbles:true,relatedTarget:null}));
    assert(menu.matches(':popover-open'),'Touch option survives trigger blur before release');
    pointer('pointerup',option);
    assert(root.querySelector('[data-field-value]').value==='expanded'&&changes===1,'Touch release commits the filter exactly once without a synthetic click');
    assert(!menu.matches(':popover-open')&&!option.hasAttribute('data-pressed'),'Touch release clears the pressed state and dismisses');
    option.dispatchEvent(new w.MouseEvent('click',{bubbles:true,detail:1}));assert(changes===1,'Compatibility click does not submit the filter twice');
    input.click();pointer('pointerdown',option);pointer('pointercancel',option);assert(changes===1&&!option.hasAttribute('data-pressed'),'Cancelled touch does not select and clears the pressed state');
    input.click();pointer('pointerdown',option);pointer('pointerup',d.body);assert(changes===1&&!menu.matches(':popover-open'),'Releasing outside cancels selection');
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
  {
    const {w,d}=await load('navigation-rail'),rail=d.querySelector('.md-navigation-rail'),item=rail.querySelector('[aria-current="page"]'),label=item.querySelector('.md-nav-label'),icon=item.querySelector('.md-nav-icon svg');
    const luminance=color=>{const c=color.match(/[\d.]+/g).slice(0,3).map(n=>{const v=Number(n)/255;return v<=.04045?v/12.92:((v+.055)/1.055)**2.4;});return c[0]*.2126+c[1]*.7152+c[2]*.0722;};
    for(const theme of ['light','dark'])for(const contrast of ['default','medium','high']){
      d.documentElement.dataset.theme=theme;d.documentElement.dataset.contrast=contrast;
      rail.dataset.orientation='horizontal';const foreground=w.getComputedStyle(label).color,background=w.getComputedStyle(item.querySelector('.md-nav-indicator')).backgroundColor,a=luminance(foreground),b=luminance(background);
      assert(foreground===w.getComputedStyle(icon).color&&(Math.max(a,b)+.05)/(Math.min(a,b)+.05)>=4.5,`Horizontal rail ${theme}/${contrast}: label matches the indicator foreground and exceeds 4.5:1`);
      rail.dataset.orientation='vertical';assert(w.getComputedStyle(label).color===w.getComputedStyle(item).color,`Vertical rail ${theme}/${contrast}: outside label retains its source color`);
    }
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
    const {w,d}=await load('navigation-rail'),rail=d.querySelector('.md-navigation-rail'),fab=rail.querySelector('.md-nav-fab');
    for(const expanded of [false,true]){
      if(expanded){rail.querySelector('[data-navigation-expand]').click();await wait(600);}
      const before=fab.getBoundingClientRect(),radius=w.getComputedStyle(fab).borderRadius;
      assert(fab.classList.contains(expanded?'md-extended-fab':'md-fab'),'Rail Create uses the matching shared FAB class');
      fab.dispatchEvent(new w.PointerEvent('pointerdown',{bubbles:true,pointerType:'mouse',pointerId:12,button:0,isPrimary:true,clientX:before.x+20,clientY:before.y+20}));
      const ripple=fab.querySelector('.md-button-ripple'),ink=ripple?.firstElementChild,style=ripple&&w.getComputedStyle(ripple);
      assert(!!ripple&&style.display!=='none'&&style.position==='absolute','Rail FAB ripple remains visible and outside flex layout');
      assert(Number(w.getComputedStyle(ink).opacity)>0&&w.getComputedStyle(ink).backgroundImage.includes('radial-gradient'),'Rail FAB ripple has a painted pressed state');
      const pressed=fab.getBoundingClientRect();assert(Math.abs(before.width-pressed.width)<.1&&Math.abs(before.x-pressed.x)<.1,'Pressing the rail FAB does not extend or shift it');
      assert(w.getComputedStyle(fab).borderRadius===radius,'Rail FAB retains its canonical fixed press shape');
      fab.dispatchEvent(new w.PointerEvent('pointerup',{bubbles:true,pointerType:'mouse',pointerId:12,button:0,isPrimary:true}));fab.click();await wait(700);
      assert(!fab.hasAttribute('data-pressed')&&!fab.querySelector('.md-button-ripple'),'Rail FAB releases and removes its ripple');
    }
    assert(!w.motionErrors.length,'Rail FAB feedback has no script errors');
  }
  {
    const {w,d}=await load('navigation-rail'),frame=d.querySelector('.md-rail-frame'),rail=frame.querySelector('.md-navigation-rail');
    await d.fonts.ready;
    function collapsedLabels(){
      for(const label of rail.querySelectorAll('.md-nav-label'))if(label.getClientRects().length){
        const style=w.getComputedStyle(label),r=label.getBoundingClientRect(),icon=label.closest('.md-nav-indicator').getBoundingClientRect();
        assert(style.position==='absolute'&&Math.abs((r.left+r.right-icon.left-icon.right)/2)<.1,'Collapsed rail labels remain absolutely positioned and centered under their indicators');
        assert(r.height<=parseFloat(style.lineHeight)+.1,'Collapsed rail labels fit on one line');
      }
    }
    collapsedLabels();
    for(const direction of ['ltr','rtl'])for(const expanded of [true,false]){
      rail.style.direction=direction;rail.querySelector('[data-navigation-expand]').click();
      assert(rail.dataset.variant===(expanded?'expanded':'collapsed'),`Rail ${direction}: switches to the requested variant`);
      const nodes=[...frame.querySelectorAll('.md-nav-indicator,.md-nav-icon,.md-nav-label,.md-navigation-slot,.md-rail-body,.md-rail-screen,.md-rail-screen h3,.md-rail-screen article:first-of-type')].filter(n=>n.getClientRects().length);
      const animations=frame.getAnimations({subtree:true});
      const directory=frame.querySelector('.md-rail-body');
      for(const a of animations){a.pause();a.currentTime=0;}
      const labels=[...rail.querySelectorAll('.md-nav-label')].filter(n=>n.getClientRects().length),copies=[...frame.querySelectorAll('[data-rail-text-exit]')];
      const oldPositions=copies.map(n=>n.getBoundingClientRect());
      assert(copies.length>0&&copies.every(n=>n.inert&&n.getAttribute('aria-hidden')==='true'),'Outgoing rail text is retained only as inert visual content');
      assert(labels.every(n=>Number(w.getComputedStyle(n).opacity)===0),'Incoming labels start invisible in their destination layout');
      const heading=rail.querySelector('.md-nav-section'),headingRect=heading.getBoundingClientRect();
      assert(heading.getAnimations().every(a=>a.effect.getKeyframes().every(k=>!['left','top','transform','scale','fontSize'].some(p=>p in k))),'Library heading animates opacity only');
      const labelPositions=labels.map(n=>n.getBoundingClientRect());
      const textAnimation=labels[0].getAnimations().find(a=>a.effect.getKeyframes().some(k=>'opacity' in k)),textFrames=textAnimation.effect.getKeyframes(),middle=textFrames.findLast(k=>Number(k.opacity)===0).offset*Number(textAnimation.effect.getTiming().duration);
      animations.forEach(a=>a.currentTime=middle);
      assert(labels.every(n=>Number(w.getComputedStyle(n).opacity)<.001)&&copies.filter(n=>!n.hasAttribute('data-rail-section-exit')).every(n=>Number(w.getComputedStyle(n).opacity)<.001),'Both label layouts are invisible at the shared halfway phase');
      const glyph=rail.querySelector(expanded?'[data-navigation-collapse-icon]':'[data-navigation-menu-icon]'),matrix=w.getComputedStyle(glyph).transform.match(/matrix\(([^)]+)\)/),angle=matrix?Math.abs(Math.atan2(...matrix[1].split(',').slice(0,2).reverse().map(Number))*180/Math.PI):0;
      assert(Math.abs(angle-45)<1,'Menu icon is halfway through its quarter turn when labels disappear');
      animations.forEach(a=>a.currentTime=Number(textAnimation.effect.getTiming().duration));
      const labelEnds=labels.map(n=>n.getBoundingClientRect());
      animations.forEach(a=>a.currentTime=0);
      for(const time of [60,160,280]){
        animations.forEach(a=>a.currentTime=time);
        if(expanded){
          const r=heading.getBoundingClientRect();assert(Math.abs(r.x-headingRect.x)<.1&&Math.abs(r.y-headingRect.y)<.1&&Math.abs(r.width-headingRect.width)<.1&&Math.abs(r.height-headingRect.height)<.1,'Library heading stays at its final position and size throughout its fade');
          assert(labels.every((n,i)=>{const r=n.getBoundingClientRect(),opacity=Number(w.getComputedStyle(n).opacity),distance=Math.abs(labelPositions[i].x-labelEnds[i].x);return Math.abs(r.y-labelEnds[i].y)<.1&&(opacity<.5||Math.abs(r.x-labelEnds[i].x)<=distance*.2+.1);}),`Entering labels approach horizontally and are nearly placed once visible (${time}ms)`);
        }else{
          const travel=Math.max(...labels.map((n,i)=>{const r=n.getBoundingClientRect();return Math.max(Math.abs(r.x-labelPositions[i].x),Math.abs(r.y-labelPositions[i].y));}));
          assert(travel<.1,`Collapsed labels fade in place (${time}ms, drift=${travel.toFixed(3)}px)`);
        }
        assert(copies.every((n,i)=>{const r=n.getBoundingClientRect();return Math.abs(r.x-oldPositions[i].x)<.1&&Math.abs(r.y-oldPositions[i].y)<.1;}),'Outgoing labels fade without travelling');
      }
      assert(copies.every(n=>Number(w.getComputedStyle(n).opacity)<.001),'Outgoing text finishes fading before the geometry settles');
      animations.forEach(a=>a.currentTime=0);
      const directoryWidth=directory.getBoundingClientRect().width;
      for(const a of animations)a.currentTime=Number(a.effect.getComputedTiming().endTime)/2;
      assert(Math.abs(directory.getBoundingClientRect().width-directoryWidth)<.1,'Rail directory text keeps one layout width throughout motion');
      assert(rail.querySelectorAll('.md-nav-label').length>0&&[...rail.querySelectorAll('.md-nav-label')].every(n=>n.getAnimations().every(a=>!a.effect.getKeyframes().some(k=>'transform' in k))),'Rail labels crossfade without animated transform layers');
      for(const a of animations){a.pause();a.currentTime=Number(a.effect.getComputedTiming().endTime)-.001;}
      const before=nodes.map(n=>n.getBoundingClientRect());
      animations.forEach(a=>a.finish());await Promise.all(animations.map(a=>a.finished.catch(()=>{})));await new Promise(w.requestAnimationFrame);
      const jump=Math.max(...nodes.map((n,i)=>{const r=n.getBoundingClientRect();return Math.max(Math.abs(r.x-before[i].x),Math.abs(r.y-before[i].y));}));
      assert(jump<.1,`Rail ${direction}/${expanded?'expanded':'collapsed'}: final layout stays within 0.1px of the animation endpoint`);
      assert(!frame.hasAttribute('data-transitioning')&&!frame.querySelector('[data-rail-text-exit]'),'Rail clears the temporary layout and outgoing text');
      if(!expanded)collapsedLabels();
    }
    const toggle=rail.querySelector('[data-navigation-expand]');toggle.click();await wait(60);toggle.click();await wait(60);toggle.click();await wait(600);
    assert(!frame.querySelector('[data-rail-text-exit]')&&!frame.hasAttribute('data-transitioning'),'Interrupted rail transitions remove outgoing text and finish cleanly');
    w.setMotionReduced(true);toggle.click();assert(!frame.querySelector('[data-rail-text-exit]')&&[...rail.querySelectorAll('.md-nav-label')].filter(n=>n.getClientRects().length).every(n=>Number(w.getComputedStyle(n).opacity)===1),'Reduced motion swaps label layouts immediately without visual copies');
    assert(!w.motionErrors.length,'Rail transitions have no script errors');
  }
  {
    const {w,d}=await load('navigation-rail',{query:'&preview.layout=modal'}),frame=d.querySelector('.md-rail-frame'),rail=frame.querySelector('.md-navigation-rail'),scrim=frame.querySelector('.md-navigation-scrim'),menu=rail.querySelector('[data-navigation-expand]');
    for(const close of ['scrim','menu','escape']){
      menu.click();await wait(600);
      assert(!scrim.hidden&&Number(w.getComputedStyle(scrim).opacity)===1,'Modal rail reopens with a visible scrim');
      if(close==='escape')menu.dispatchEvent(new w.KeyboardEvent('keydown',{key:'Escape',bubbles:true}));else(close==='scrim'?scrim:menu).click();
      assert(frame.querySelectorAll('[data-rail-text-exit]').length>0,'Modal dismissal retains outgoing text for its exit fade');
      const fade=scrim.getAnimations().find(a=>a.effect.getKeyframes().some(k=>'opacity' in k));
      assert(!!fade,`Modal rail ${close}: scrim has an exit fade`);
      await fade.finished;
      assert(frame.hasAttribute('data-transitioning'),`Modal rail ${close}: opacity settles before spatial motion`);
      for(let i=0;i<3;i++){await new Promise(w.requestAnimationFrame);assert(scrim.hidden||Number(w.getComputedStyle(scrim).opacity)<.001,`Modal rail ${close}: scrim stays transparent after the fade`);}
      await wait(500);
      assert(scrim.hidden&&frame.dataset.modalOpen==='false'&&!frame.querySelector('[data-rail-text-exit]'),`Modal rail ${close}: cleanup hides the scrim and removes outgoing text`);
    }
    assert(!w.motionErrors.length,'Modal rail scrim lifecycle has no script errors');
  }
  {
    const {w,d}=await load(''),art=d.querySelector('[data-artwork]'),toggle=d.querySelector('.tokens-art-pause input');
    art.scrollIntoView({block:'center'});await wait(50);
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
