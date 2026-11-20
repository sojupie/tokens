(() => {
 const wrap=(value,length)=>(value%length+length)%length,pad=value=>String(value).padStart(2,'0');
 const outerRing=101,innerRing=69,diameter=256;
 const inRange=(value,min,max)=>/^\d{1,2}$/.test(value)&&Number(value)>=min&&Number(value)<=max;
 function parseFields(hour,minute,is24,pm){
  if(!inRange(hour,is24?0:1,is24?23:12)||!inRange(minute,0,59))return null;
  const h=Number(hour);
  return{hour:is24?h:h%12+(pm?12:0),minute:Number(minute)};
 }
 function nearestAngle(previous,target){return previous+wrap(target-previous+180,360)-180;}
 function dialValue(x,y,size,part,is24,centerSize=0){
  const dx=x-size/2,dy=y-size/2,distance=Math.hypot(dx,dy);if(distance<=centerSize/2)return null;
  const angle=wrap(Math.atan2(dx,-dy)*180/Math.PI,360);
  if(part==='minute')return Math.round(angle/6)%60;
  // Compare distances to the nearest hour's two centers. Project onto its
  // spoke so the dividing line stays halfway between the visible targets.
  const hour=Math.round(angle/30);
  const projection=distance*Math.cos((angle-hour*30)*Math.PI/180);
  return hour%12+(is24&&projection<size*(outerRing+innerRing)/(2*diameter)?12:0);
 }
 if(typeof module!=='undefined')module.exports={parseFields,nearestAngle,dialValue,wrap};
 if(typeof document==='undefined')return;
 const states=new WeakMap();
 const reduced=matchMedia('(prefers-reduced-motion: reduce)');
 const hostSelector='.md-time-demo,[data-time-picker]';
 const panelSelector='[data-time-panel]';
 const hostFor=panel=>panel.closest(hostSelector)||panel;
 const stateFor=target=>{
  const root=target.closest?.(hostSelector+','+panelSelector);
  return root&&states.get(hostFor(root));
 };
 function panels(root){
  return [...(root.matches?.(panelSelector)?[root]:[]),...(root.querySelectorAll?.(panelSelector)??[])];
 }
 const hourDisplay=state=>state.is24?state.hour:state.hour%12||12;
 const timeText=state=>pad(state.hour)+':'+pad(state.minute);
 const pressed=(button,value)=>{const text=String(value);if(button.getAttribute('aria-pressed')!==text)button.setAttribute('aria-pressed',text);};
 function face(state){
  const faceKey=state.active+':'+state.is24;if(state.faceKey===faceKey)return;state.faceKey=faceKey;
  const nodes=[];
  for(let ring=0;ring<(state.is24&&state.active==='hour'?2:1);ring++)for(let i=0;i<12;i++){
   const tick=document.createElement('span'),radius=(ring?innerRing:outerRing)/diameter,theta=i*Math.PI/6;
   tick.className='md-time-tick';tick.style.setProperty('--_time-label-x',(50+Math.sin(theta)*radius*100)+'%');tick.style.setProperty('--_time-label-y',(50-Math.cos(theta)*radius*100)+'%');
   tick.textContent=state.active==='minute'?pad(i*5):ring?String(i+12):i?String(i):state.is24?'00':'12';nodes.push(tick);
  }
  state.dial.querySelector('.md-time-labels').replaceChildren(...nodes);
  state.dial.querySelector('.md-time-selected-labels').replaceChildren(...nodes.map(n=>n.cloneNode(true)));
  const choices=nodes.map((tick,i)=>{
   const choice=document.createElement('button');
   const value=state.active==='minute'?i*5:state.is24?i:i||12;
   choice.type='button';choice.className='md-time-choice';choice.tabIndex=-1;
   choice.dataset.timeChoice=String(value);choice.style.cssText=tick.style.cssText;
   choice.setAttribute('aria-label',state.active==='minute'?'Minute '+value+' of 60':state.is24?'Hour '+value:'Hour '+value+' of 12');
   return choice;
  });
  const intermediate=document.createElement('button');
  intermediate.type='button';intermediate.className='md-time-choice';intermediate.tabIndex=-1;
  intermediate.dataset.timeIntermediate='';intermediate.hidden=true;
  state.dial.querySelector('.md-time-choices').replaceChildren(...choices,intermediate);
 }
 function render(state,syncInputs=true){
  const {panel,dial}=state;panel.dataset.active=state.active;
  panel.querySelector('[data-time-title]').textContent=panel.dataset.mode==='input'?'Enter time':'Select time';
  const toggle=panel.querySelector('[data-time-mode]');
  toggle.setAttribute('aria-label',panel.dataset.mode==='input'?(state.returnMode==='scroll'?'Toggle scroll picker':'Toggle dial picker'):'Toggle input picker');
  toggle.hidden=panel.hasAttribute('data-time-inline')&&panel.dataset.mode==='input'&&state.returnMode!=='scroll';
  for(const part of ['hour','minute']){
   const value=part==='hour'?hourDisplay(state):state.minute,button=panel.querySelector(`[data-time-value="${part}"]`);
   button.textContent=pad(value);pressed(button,state.active===part);
   button.setAttribute('aria-label',(part==='hour'?'Hour':'Minute')+' '+pad(value));
   const rotor=panel.querySelector('[data-time-scroll="'+part+'"]');
   if(rotor)globalThis.MaterialScrollFields?.set(rotor,part==='hour'&&!state.is24?value-1:value);
   if(syncInputs){const input=panel.querySelector(`[data-time-input="${part}"]`);input.value=pad(value);input.setAttribute('aria-invalid','false');}
  }
  if(syncInputs)panel.querySelector('[data-time-error]').textContent='';
  const pm=state.hour>=12;for(const button of panel.querySelectorAll('[data-time-period]')){const selected=(button.dataset.timePeriod==='pm')===pm;button.dataset.selected=String(selected);button.setAttribute('aria-checked',String(selected));button.tabIndex=selected?0:-1;}
  panel.querySelector('.md-time-period-group').hidden=state.is24;
  const value=state.active==='hour'?(state.is24?state.hour:hourDisplay(state)):state.minute;
  dial.setAttribute('aria-label',state.active==='hour'?'Hours':'Minutes');
  const target=state.active==='hour'?state.hour%12*30:state.minute*6;
  const current=state.angle===undefined?target:state.gesture?state.angle:parseFloat(getComputedStyle(dial).getPropertyValue('--_time-angle'));
  state.angle=nearestAngle(current,target);dial.style.setProperty('--_time-angle',state.angle+'deg');
  dial.dataset.ring=state.active==='hour'&&state.is24&&pm?'inner':'outer';dial.dataset.dot=String(state.active==='minute'&&state.minute%5!==0);
  if(panel.dataset.mode==='clock'){
   face(state);
   for(const choice of dial.querySelectorAll('[data-time-choice]'))pressed(choice,Number(choice.dataset.timeChoice)===value);
  }
 }
 function focusField(state, part=state.active){
  const mode=state.panel.dataset.mode;
  const selector=mode==='scroll'?'[data-time-scroll="'+part+'"] [role=spinbutton]':
   '[data-time-'+(mode==='input'?'input':'value')+'="'+part+'"]';
  const field=state.panel.querySelector(selector);
  if(!field)return;
  if(mode==='input')globalThis.MaterialComponents.focusForEntry(field);
  else field.focus({preventScroll:true});
 }
 function resetScrollInteraction(panel){
  for(const field of panel.querySelectorAll('[data-time-scroll]')){
   globalThis.MaterialScrollFields?.stop(field);
   globalThis.MaterialScrollFields?.clearHover(field);
  }
 }
 function stopModeMotion(state){
  const {panel,dial}=state;
  for(const animation of state.modeAnimations||[])animation.cancel();
  state.modeAnimations=null;
  delete panel.dataset.timeResizing;
  delete panel.dataset.timeDialSwap;
  dial.inert=false;
  dial.style.removeProperty('--_time-swap-x');
  dial.style.removeProperty('--_time-swap-y');
 }
 function dialPose(dial){
  const rect=dial.getBoundingClientRect();
  if(!rect.width)return null;
  const style=getComputedStyle(dial);
  return {x:rect.x+rect.width/2,y:rect.y+rect.height/2,scale:parseFloat(style.scale)||1,opacity:Number(style.opacity)};
 }
 function animateDialSwap(state, before, after, outgoing, incoming, animations){
  const {panel,dial}=state;
  const anchor=incoming||outgoing;
  const frame=incoming?after:before;
  dial.style.setProperty('--_time-swap-x',anchor.x-frame.x-frame.width/2+'px');
  dial.style.setProperty('--_time-swap-y',anchor.y-frame.y-frame.height/2+'px');
  panel.dataset.timeDialSwap='';
  dial.inert=!incoming;
  // The face has independent size and opacity tracks. It enters after the
  // dialog starts to grow and leaves before the dialog finishes shrinking.
  const reversing=outgoing&&outgoing.opacity>0;
  const standard='cubic-bezier(0.2, 0, 0, 1)';
  animations.push(
   dial.animate([{scale:reversing?outgoing.scale:.6},{scale:incoming?1:.4}],{
    duration:incoming?320:100,delay:incoming&&!reversing?60:0,
    easing:incoming?'cubic-bezier(0, 0, 0, 1)':standard,fill:'both',
   }),
   dial.animate([{opacity:outgoing?.opacity??0},{opacity:incoming?1:0}],{
    duration:incoming?150:100,delay:incoming&&!reversing?50:0,easing:standard,fill:'both',
   }),
  );
 }
 function changeMode(state){
  const {panel,dial}=state;
  const before=reduced.matches?null:panel.getBoundingClientRect();
  const outgoing=before&&dialPose(dial);
  const mode=panel.dataset.mode;
  const landscapeDial=mode==='clock'&&panel.dataset.orientation==='landscape';
  const selectors=before&&panel.querySelector('.md-time-selectors');
  const selectorsBefore=selectors&&selectors.getBoundingClientRect();
  stopModeMotion(state);
  if(mode!=='input'){state.returnMode=mode;panel.dataset.returnMode=mode;}
  state.preferredMode=mode==='input'?state.returnMode:'input';panel.dataset.mode=state.preferredMode;
  resetScrollInteraction(panel);
  render(state);adapt(state);focusField(state);
  if(!before)return;
  const motion=globalThis.MaterialMotion;
  const after=panel.getBoundingClientRect();
  const modalSurface=panel.matches(':modal, [data-time-preview]');
  const incoming=modalSurface&&panel.dataset.mode==='clock'&&dialPose(dial);
  const dialSwap=modalSurface&&panel.dataset.timeStyle==='expressive'&&(outgoing||incoming);
  const scrollSwap=mode==='scroll'||panel.dataset.mode==='scroll';
  const landscapeSwap=dialSwap&&(landscapeDial||panel.dataset.orientation==='landscape');
  const selectorsAfter=landscapeSwap&&selectors.getBoundingClientRect();
  const animations=[];
  if(dialSwap)animateDialSwap(state,before,after,outgoing,incoming,animations);
  else if(!scrollSwap)animations.push(panel.querySelector('.md-time-body').animate([{opacity:0},{opacity:1}],motion.timing(panel,'fast-effects')));
  if(landscapeSwap||(modalSurface&&(before.width!==after.width||before.height!==after.height))){
   const size=rect=>({width:rect.width+'px',height:rect.height+'px'});
   panel.dataset.timeResizing='';
   // The reference clip matches the Expressive default spatial spring.
   const spring=dialSwap?motion.cssTiming(.8,380):motion.timing(panel,'fast-spatial');
   const timing={...spring,duration:parseFloat(spring.duration),fill:'both'};
   if(landscapeSwap){
    // The animated face is out of flow, so the landscape grid loses its
    // row height. Interpolate the selectors between the laid-out positions.
    const swapOffset=selectors.getBoundingClientRect().top-panel.getBoundingClientRect().top;
    const position=(fields,frame)=>({translate:'0 '+(fields.top-frame.top-swapOffset)+'px'});
    animations.push(selectors.animate([position(selectorsBefore,before),position(selectorsAfter,after)],timing));
   }
   animations.push(panel.animate([size(before),size(after)],timing));
  }
  state.modeAnimations=animations;
  Promise.allSettled(animations.map(animation=>animation.finished)).then(()=>{
   if(state.modeAnimations===animations)stopModeMotion(state);
  });
 }
 // Fit the real content, including text scaling and the on-screen keyboard.
 function adapt(state){
  const panel=state.panel;if(!panel.matches(':modal')||panel.hasAttribute('data-time-resizing'))return;
  const focused=document.activeElement;
  const previousMode=panel.dataset.mode;
  panel.dataset.mode=state.preferredMode;panel.dataset.orientation=state.preferredOrientation;
  const fits=()=>panel.scrollHeight<=panel.clientHeight+1&&panel.scrollWidth<=panel.clientWidth+1&&
   panel.getBoundingClientRect().bottom<=(window.visualViewport?.height??window.innerHeight)+1;
  if(panel.dataset.mode==='clock'&&!fits()){
   panel.dataset.orientation='landscape';
   if(!fits())panel.dataset.mode=panel.dataset.timeStyle==='expressive'?'scroll':'input';
  }
  render(state,panel.dataset.mode==='input'&&previousMode!=='input');
  if(panel.contains(focused)&&!focused.getClientRects().length)focusField(state);
 }
 function validate(state,focusInvalid=true,editing=null){
  if(state.panel.dataset.mode!=='input')return true;
  const hour=state.panel.querySelector('[data-time-input="hour"]');
  const minute=state.panel.querySelector('[data-time-input="minute"]');
  const error=state.panel.querySelector('[data-time-error]');
  const value=parseFields(hour.value,minute.value,state.is24,state.hour>=12);
  const finished=field=>field!==editing||!/^\d?$/.test(field.value);
  const invalidHour=finished(hour)&&!inRange(hour.value,state.is24?0:1,state.is24?23:12);
  const invalidMinute=finished(minute)&&!inRange(minute.value,0,59);
  hour.setAttribute('aria-invalid',String(invalidHour));
  minute.setAttribute('aria-invalid',String(invalidMinute));
  const message=invalidHour||invalidMinute?'Enter a valid time':'';
  if(error.textContent!==message)error.textContent=message;
  if(focusInvalid&&message){
   const field=invalidHour?hour:minute;
   field.focus({preventScroll:true});
   field.select();
  }
  // A valid first digit can still be the start of an invalid two-digit entry.
  if(value&&(!editing||finished(editing)))Object.assign(state,value);
  return Boolean(value);
 }
 function cancelGesture(state){const g=state.gesture;if(!g)return;state.gesture=null;delete state.dial.dataset.dragging;if(state.dial.hasPointerCapture(g.id))state.dial.releasePointerCapture(g.id);}
 function commit(state){
  state.committed={hour:state.hour,minute:state.minute};
  const value=timeText(state),output=state.root.querySelector('[data-time-output]'),input=state.root.querySelector('[data-time-value-input]');
  if(output)output.textContent=value;
  if(input&&input.value!==value){
   input.value=value;
   input.dispatchEvent(new Event('input',{bubbles:true}));
   input.dispatchEvent(new Event('change',{bubbles:true}));
  }
 }
 function setPart(state,part,value){
  if(part==='hour')state.hour=state.is24?value:value%12+(state.hour>=12?12:0);
  else state.minute=value;
 }
 function inlineChanged(state){
  if(state.panel.hasAttribute('data-time-inline')&&
    (state.hour!==state.committed.hour||state.minute!==state.committed.minute))commit(state);
 }
 function selectPeriod(state,period){
  state.hour=state.hour%12+(period==='pm'?12:0);
  if(validate(state,false))inlineChanged(state);
  render(state,false);
 }
 function close(state,save){
  if(state.panel.hasAttribute('data-time-inline')){
   if(save&&validate(state))commit(state);
   else if(!save){Object.assign(state,state.committed);render(state);}
   return;
  }
  globalThis.MaterialDialogs?.close(state.panel,save?'confirm':'','action');
 }
 function focusDialChoice(state, value){
  state.choice=value;
  const interim=state.dial.querySelector('[data-time-intermediate]');
  const visible=state.dial.querySelector('[data-time-choice="'+value+'"]:not([data-time-intermediate])');
  if(visible){interim.hidden=true;visible.focus({preventScroll:true});return;}
  const angle=value*Math.PI/30;
  interim.style.setProperty('--_time-label-x',(50+Math.sin(angle)*101/256*100)+'%');
  interim.style.setProperty('--_time-label-y',(50-Math.cos(angle)*101/256*100)+'%');
  interim.dataset.timeChoice=String(value);
  interim.setAttribute('aria-label','Minute '+value+' of 60');interim.hidden=false;
  pressed(interim,value===state.minute);interim.focus({preventScroll:true});
 }
 function confirmDialChoice(state){
  if(state.choice!==undefined)setPart(state,state.active,state.choice);
  state.choice=undefined;
  if(state.active==='hour'){
   state.active='minute';render(state);focusField(state);
  }else close(state,true);
 }
 function reservePreview(state){
  const {panel,root}=state;
  if(root===panel||!panel.hasAttribute('data-time-preview'))return;
  // Reserve the two in-flow modes without changing the size of either panel.
  const mode=panel.dataset.mode,height=panel.getBoundingClientRect().height;
  panel.dataset.mode=mode==='input'?state.returnMode:'input';
  const alternateHeight=panel.getBoundingClientRect().height;
  panel.dataset.mode=mode;
  root.style.setProperty('--_time-preview-block-size',Math.ceil(Math.max(height,alternateHeight))+'px');
 }
 function initialize(scope=document){for(const panel of panels(scope)){
  const root=hostFor(panel);
  if(states.has(root))continue;
  const binding=new AbortController();
  const [hour,minute]=(root.querySelector('[data-time-value-input]')?.value||'').split(':').map(Number);
  const hasValue=Number.isFinite(minute);
  const state={root,panel,dial:panel.querySelector('[data-time-dial]'),is24:panel.dataset.format==='24h',
   binding,
   hour:hasValue?hour:Number(panel.dataset.hour)%24,minute:hasValue?minute:Number(panel.dataset.minute)%60,active:panel.dataset.active,
   initialMode:panel.dataset.mode,preferredMode:panel.dataset.mode,preferredOrientation:panel.dataset.orientation,initialReturn:panel.dataset.returnMode||'clock',returnMode:panel.dataset.returnMode||'clock'};
  state.committed={hour:state.hour,minute:state.minute};states.set(root,state);render(state);reservePreview(state);
  panel.addEventListener('md-dialog-beforeopen',()=>{
   resetScrollInteraction(panel);
   state.active='hour';state.choice=undefined;state.preferredMode=state.initialMode;state.returnMode=state.initialReturn;
   panel.dataset.mode=state.initialMode;panel.dataset.returnMode=state.returnMode;render(state);
   requestAnimationFrame(()=>{if(panel.open&&!binding.signal.aborted){adapt(state);focusField(state,'hour');}});
  },{signal:binding.signal});
  panel.addEventListener('md-dialog-beforeclose',event=>{
   const save=event.detail.result==='confirm';
   if(save&&!validate(state)){event.preventDefault();return;}
   stopModeMotion(state);
   resetScrollInteraction(panel);
   cancelGesture(state);
   if(save)commit(state);
   else Object.assign(state,state.committed);
   render(state);
  },{signal:binding.signal});
 }}
 document.addEventListener('md-scroll-field-change',event=>{
  const part=event.target.dataset.timeScroll,state=stateFor(event.target);
  if(!part||!state||state.panel.dataset.mode!=='scroll')return;
  state.active=part;setPart(state,part,Number(event.detail.value));render(state);inlineChanged(state);
 });
 document.addEventListener('click',event=>{
  const state=stateFor(event.target);if(!state)return;
  const value=event.target.closest('[data-time-value]');
  if(value){state.active=value.dataset.timeValue;render(state,false);focusDialChoice(state,state.active==='minute'?0:state.is24?0:12);return;}
  const choice=event.target.closest('[data-time-choice]');
  if(choice&&event.detail===0){state.choice=Number(choice.dataset.timeChoice);confirmDialChoice(state);return;}
  const period=event.target.closest('[data-time-period]');
  if(period){selectPeriod(state,period.dataset.timePeriod);return;}
  if(event.target.closest('[data-time-mode]')){
   // Leaving direct input restores the last valid time if the draft is invalid.
   if(state.panel.dataset.mode==='input'&&validate(state,false))inlineChanged(state);
   cancelGesture(state);
   changeMode(state);return;
  }
  if(event.target.closest('[data-time-save]'))close(state,true);if(event.target.closest('[data-time-cancel]'))close(state,false);
 });
 document.addEventListener('input',event=>{
  const input=event.target;if(!input.matches?.('[data-time-input]'))return;const state=stateFor(input);if(!state)return;
  if(!validate(state,false,input)||input.value.length!==2)return;
  inlineChanged(state);
  render(state,false);
  if(input.dataset.timeInput==='hour')focusField(state,'minute');
 });
 document.addEventListener('focusout',event=>{
  const input=event.target;if(!input.matches?.('[data-time-input]'))return;
  const state=stateFor(input);
  if(state?.panel.dataset.mode==='input'&&(state.panel.open||state.panel.hasAttribute('data-time-inline'))&&validate(state,false))inlineChanged(state);
 });
 document.addEventListener('change',event=>{
  const input=event.target;if(!input.matches?.('[data-time-value-input]'))return;
  const state=stateFor(input);if(!state)return;
  const [hour,minute]=(input.value||input.defaultValue).split(':').map(Number);
  if(hour===state.hour&&minute===state.minute&&hour===state.committed.hour&&minute===state.committed.minute)return;
  if(Number.isFinite(hour)&&Number.isFinite(minute)){
   Object.assign(state,{hour,minute});state.committed={hour,minute};
   const output=state.root.querySelector('[data-time-output]');
   if(output)output.textContent=timeText(state);
   render(state);
  }
 });
 document.addEventListener('reset',event=>{
  queueMicrotask(()=>{
   if(event.defaultPrevented)return;
   for(const input of event.target.elements)
    if(input.matches('[data-time-value-input]'))input.dispatchEvent(new Event('change',{bubbles:true}));
  });
 });
 document.addEventListener('focusin',event=>{const input=event.target;if(!input.matches?.('[data-time-input]'))return;const state=stateFor(input);if(state){state.active=input.dataset.timeInput;render(state,false);}});
 document.addEventListener('keydown',event=>{
  const state=stateFor(event.target);if(!state)return;
  const period=event.target.closest?.('[data-time-period]');
  if(period){
   const next=['ArrowRight','ArrowDown','End'].includes(event.key)?'pm':['ArrowLeft','ArrowUp','Home'].includes(event.key)?'am':null;
   if(next){event.preventDefault();selectPeriod(state,next);state.panel.querySelector('[data-time-period="'+next+'"]').focus({preventScroll:true});}
   return;
  }
  const valueField=event.target.closest?.('[data-time-value]');
  if(valueField&&['ArrowLeft','ArrowRight'].includes(event.key)){
   event.preventDefault();focusField(state,valueField.dataset.timeValue==='hour'?'minute':'hour');return;
  }
  const input=event.target.closest?.('[data-time-input]'),dial=event.target.closest?.('[data-time-dial]');
  if(input&&event.key==='Enter'){event.preventDefault();close(state,true);return;}
  if(!dial)return;
  if(event.key==='Enter'||event.key===' '){event.preventDefault();confirmDialChoice(state);return;}
  const part=state.active,min=part==='hour'&&!state.is24?1:0,max=part==='minute'?59:state.is24?23:12;
  const current=state.choice??(part==='hour'?hourDisplay(state):state.minute);
  let value;
  if(['ArrowRight','ArrowDown'].includes(event.key))value=current+1;
  if(['ArrowLeft','ArrowUp'].includes(event.key))value=current-1;
  if(event.key==='PageUp')value=current+5;
  if(event.key==='PageDown')value=current-5;
  if(event.key==='Home')value=min;
  if(event.key==='End')value=max;
  if(value!==undefined){
   event.preventDefault();focusDialChoice(state,wrap(value-min,max-min+1)+min);
   if(!state.dial.contains(document.activeElement))state.dial.focus({preventScroll:true});
  }
 });
 function point(state,event){const rect=state.dial.getBoundingClientRect(),style=getComputedStyle(state.dial),diameter=parseFloat(style.getPropertyValue('--_time-dial-size')),center=parseFloat(style.getPropertyValue('--_time-dial-center-size'))*rect.width/diameter,value=dialValue(event.clientX-rect.left,event.clientY-rect.top,rect.width,state.active,state.is24,center);if(value===null)return;if(state.active==='hour')state.hour=state.is24?value:value+(state.hour>=12?12:0);else state.minute=value;render(state,false);}
 document.addEventListener('pointerdown',event=>{
  const dial=event.target.closest?.('[data-time-dial]');if(!dial||event.button!==0||event.isPrimary===false)return;const state=stateFor(dial);if(!state)return;
  state.choice=undefined;state.gesture={id:event.pointerId,part:state.active,hour:state.hour,minute:state.minute};dial.dataset.dragging='';dial.setPointerCapture(event.pointerId);dial.focus({preventScroll:true});event.preventDefault();point(state,event);
 });
 document.addEventListener('pointermove',event=>{
  const state=stateFor(event.target);if(state?.gesture?.id===event.pointerId){event.preventDefault();point(state,event);}
 });
 function finishPointer(event,cancel){const state=stateFor(event.target),gesture=state?.gesture;if(!gesture||gesture.id!==event.pointerId)return;cancelGesture(state);if(cancel){state.hour=gesture.hour;state.minute=gesture.minute;}else if(gesture.part==='hour')state.active='minute';render(state,false);}
 document.addEventListener('pointerup',event=>finishPointer(event,false));document.addEventListener('pointercancel',event=>finishPointer(event,true));document.addEventListener('lostpointercapture',event=>finishPointer(event,true));
 function fitOpenPanels(){for(const panel of panels(document)){const state=stateFor(panel);if(state)adapt(state);}}
 window.addEventListener('resize',fitOpenPanels);
 window.visualViewport?.addEventListener('resize',fitOpenPanels);
 window.addEventListener('blur',()=>{for(const panel of document.querySelectorAll('.md-time-panel')){const state=stateFor(panel);if(state?.gesture){const g=state.gesture;cancelGesture(state);state.hour=g.hour;state.minute=g.minute;render(state,false);}}});
 function dispose(root){
  for(const panel of panels(root)){
   const state=stateFor(panel);if(!state)continue;
   state.binding.abort();stopModeMotion(state);cancelGesture(state);states.delete(state.root);
  }
 }
 globalThis.MaterialTimePickers=Object.freeze({initialize,dispose});
 document.addEventListener('htmx:beforeCleanupElement',event=>dispose(event.detail?.elt||event.target));
 for(const name of ['htmx:afterSwap','DOMContentLoaded','tokens:assets-ready'])document.addEventListener(name,()=>initialize());
 initialize();
})();
