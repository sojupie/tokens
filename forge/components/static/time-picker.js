(() => {
 const wrap=(value,length)=>(value%length+length)%length,pad=value=>String(value).padStart(2,'0');
 function parseFields(hour,minute,is24,pm){
  if(!/^\d{1,2}$/.test(hour)||!/^\d{1,2}$/.test(minute))return null;
  const h=Number(hour),m=Number(minute);if(m>59||h>(is24?23:12)||h<(is24?0:1))return null;
  return{hour:is24?h:h%12+(pm?12:0),minute:m};
 }
 function nearestAngle(previous,target){return previous+wrap(target-previous+180,360)-180;}
 function dialValue(x,y,size,part,is24,centerSize=0){
  const dx=x-size/2,dy=y-size/2,distance=Math.hypot(dx,dy);if(distance<=centerSize/2)return null;
  const angle=wrap(Math.atan2(dx,-dy)*180/Math.PI,360);
  if(part==='minute')return Math.round(angle/6)%60;
  return Math.round(angle/30)%12+(is24&&distance<size*74/256?12:0);
 }
 if(typeof module!=='undefined')module.exports={parseFields,nearestAngle,dialValue,wrap};
 if(typeof document==='undefined')return;
 const states=new WeakMap(),reduced=matchMedia('(prefers-reduced-motion: reduce)');
 const stateFor=target=>states.get(target.closest?.('.md-time-demo'));
 const hourDisplay=state=>state.is24?state.hour:state.hour%12||12;
 const timeText=state=>pad(state.hour)+':'+pad(state.minute);
 const pressed=(button,value)=>{const text=String(value);if(button.getAttribute('aria-pressed')!==text)button.setAttribute('aria-pressed',text);};
 function face(state){
  const faceKey=state.active+':'+state.is24;if(state.faceKey===faceKey)return;state.faceKey=faceKey;
  const nodes=[];
  for(let ring=0;ring<(state.is24&&state.active==='hour'?2:1);ring++)for(let i=0;i<12;i++){
   const tick=document.createElement('span'),radius=(ring?69:101)/256,theta=i*Math.PI/6;
   tick.className='md-time-tick';tick.style.setProperty('--_time-label-x',(50+Math.sin(theta)*radius*100)+'%');tick.style.setProperty('--_time-label-y',(50-Math.cos(theta)*radius*100)+'%');
   tick.textContent=state.active==='minute'?pad(i*5):ring?String(i+12):i?String(i):state.is24?'00':'12';nodes.push(tick);
  }
  state.dial.querySelector('.md-time-labels').replaceChildren(...nodes);
  state.dial.querySelector('.md-time-selected-labels').replaceChildren(...nodes.map(n=>n.cloneNode(true)));
 }
 function render(state,syncInputs=true){
  const {panel,dial}=state;panel.dataset.active=state.active;
  panel.querySelector('[data-time-title]').textContent=panel.dataset.mode==='input'?'Enter time':'Select time';
  panel.querySelector('[data-time-mode]').setAttribute('aria-label',panel.dataset.mode==='input'?'Use clock':'Use keyboard input');
  for(const part of ['hour','minute']){
   const value=part==='hour'?hourDisplay(state):state.minute,button=panel.querySelector(`[data-time-value="${part}"]`);
   button.textContent=pad(value);pressed(button,state.active===part);
   if(syncInputs){const input=panel.querySelector(`[data-time-input="${part}"]`);input.value=pad(value);input.setAttribute('aria-invalid','false');}
  }
  if(syncInputs)panel.querySelector('[data-time-error]').textContent='';
  const pm=state.hour>=12;for(const button of panel.querySelectorAll('[data-time-period]')){const selected=(button.dataset.timePeriod==='pm')===pm;pressed(button,selected);button.tabIndex=selected?0:-1;}
  panel.querySelector('.md-time-period-group').hidden=state.is24;
  const value=state.active==='hour'?(state.is24?state.hour:hourDisplay(state)):state.minute;
  dial.setAttribute('aria-label',state.active==='hour'?'Hours':'Minutes');dial.setAttribute('aria-valuemin',String(state.active==='hour'&&!state.is24?1:0));dial.setAttribute('aria-valuemax',String(state.active==='minute'?59:state.is24?23:12));dial.setAttribute('aria-valuenow',String(value));dial.setAttribute('aria-valuetext',state.active==='minute'?value+' minutes':state.is24?pad(value)+' hours':value+(pm?' PM':' AM'));
  const target=state.active==='hour'?state.hour%12*30:state.minute*6;
  const current=state.angle===undefined?target:state.gesture?state.angle:parseFloat(getComputedStyle(dial).getPropertyValue('--_time-angle'));
  state.angle=nearestAngle(current,target);dial.style.setProperty('--_time-angle',state.angle+'deg');
  dial.dataset.ring=state.active==='hour'&&state.is24&&pm?'inner':'outer';dial.dataset.dot=String(state.active==='minute'&&state.minute%5!==0);
  if(panel.dataset.mode==='clock')face(state);
 }
 function validate(state,show=true){
  if(state.panel.dataset.mode!=='input')return true;
  const hour=state.panel.querySelector('[data-time-input="hour"]'),minute=state.panel.querySelector('[data-time-input="minute"]'),value=parseFields(hour.value,minute.value,state.is24,state.hour>=12);
  if(value){Object.assign(state,value);hour.setAttribute('aria-invalid','false');minute.setAttribute('aria-invalid','false');state.panel.querySelector('[data-time-error]').textContent='';return true;}
  if(show){const validHour=/^\d{1,2}$/.test(hour.value)&&Number(hour.value)>=(state.is24?0:1)&&Number(hour.value)<=(state.is24?23:12),input=validHour?minute:hour;
   hour.setAttribute('aria-invalid',String(!validHour));minute.setAttribute('aria-invalid',String(!/^\d{1,2}$/.test(minute.value)||Number(minute.value)>59));
   state.panel.querySelector('[data-time-error]').textContent=validHour?'Enter minutes from 0 to 59.':state.is24?'Enter an hour from 0 to 23.':'Enter an hour from 1 to 12.';input.focus({preventScroll:true});input.select();}
  return false;
 }
 function timing(state){return MaterialMotion.timing(state.panel,'fast-effects');}
 function cancelGesture(state){const g=state.gesture;if(!g)return;state.gesture=null;delete state.dial.dataset.dragging;if(state.dial.hasPointerCapture(g.id))state.dial.releasePointerCapture(g.id);}
 function open(state){
  if(state.panel.open&&!state.closing)return;
  const from=state.panel.open?getComputedStyle(state.panel).opacity:0;state.closing=false;state.panel.inert=false;
  state.revision=(state.revision||0)+1;state.animation?.cancel();render(state);if(!state.panel.open)state.panel.showModal();state.opener.setAttribute('aria-expanded','true');
  if(state.panel.dataset.mode==='input'){const input=state.panel.querySelector('[data-time-input="hour"]');input.focus({preventScroll:true});input.select();}
  else state.panel.querySelector(`[data-time-value="${state.active}"]`).focus({preventScroll:true});
  if(!reduced.matches)state.animation=state.panel.animate([{opacity:from},{opacity:1}],timing(state));
 }
 async function close(state,save){
  if(save&&!validate(state))return;
  cancelGesture(state);
  if(save){state.committed={hour:state.hour,minute:state.minute};state.root.querySelector('[data-time-output]').textContent=timeText(state);}else Object.assign(state,state.committed);
  render(state);if(state.panel.hasAttribute('data-time-inline'))return;
  const from=getComputedStyle(state.panel).opacity;state.closing=true;state.panel.inert=true;const revision=state.revision=(state.revision||0)+1;state.animation?.cancel();
  if(!reduced.matches){const animation=state.panel.animate([{opacity:from},{opacity:0}],timing(state));state.animation=animation;try{await animation.finished;}catch{return;}if(state.revision!==revision)return;animation.cancel();}
  state.panel.close();state.panel.inert=false;state.closing=false;state.opener.setAttribute('aria-expanded','false');state.opener.focus({preventScroll:true});
 }
 function init(){for(const root of document.querySelectorAll('.md-time-demo')){
  if(states.has(root))continue;const panel=root.querySelector('.md-time-panel'),state={root,panel,opener:root.querySelector('[data-time-open]'),dial:panel.querySelector('[data-time-dial]'),is24:panel.dataset.format==='24h',hour:10,minute:10,active:panel.dataset.active};state.committed={hour:10,minute:10};states.set(root,state);render(state);
  panel.addEventListener('cancel',event=>{event.preventDefault();close(state,false);});
 }}
 document.addEventListener('click',event=>{
  const state=stateFor(event.target);if(!state)return;
  if(event.target.closest('[data-time-open]')){open(state);return;}
  const value=event.target.closest('[data-time-value]');if(value){state.active=value.dataset.timeValue;render(state,false);return;}
  const period=event.target.closest('[data-time-period]');if(period){state.hour=state.hour%12+(period.dataset.timePeriod==='pm'?12:0);render(state,false);return;}
  if(event.target.closest('[data-time-mode]')){if(!validate(state))return;cancelGesture(state);state.panel.dataset.mode=state.panel.dataset.mode==='clock'?'input':'clock';render(state);if(state.panel.dataset.mode==='input'){const input=state.panel.querySelector(`[data-time-input="${state.active}"]`);input.focus({preventScroll:true});input.select();}else state.panel.querySelector(`[data-time-value="${state.active}"]`).focus({preventScroll:true});return;}
  if(event.target.closest('[data-time-save]'))close(state,true);if(event.target.closest('[data-time-cancel]'))close(state,false);
 });
 document.addEventListener('input',event=>{
  const input=event.target;if(!input.matches?.('[data-time-input]'))return;const state=stateFor(input);if(!state)return;
  const h=state.panel.querySelector('[data-time-input="hour"]'),m=state.panel.querySelector('[data-time-input="minute"]'),parsed=parseFields(h.value,m.value,state.is24,state.hour>=12);
  if(parsed){Object.assign(state,parsed);h.setAttribute('aria-invalid','false');m.setAttribute('aria-invalid','false');state.panel.querySelector('[data-time-error]').textContent='';}
  render(state,false);
  if(parsed&&input.dataset.timeInput==='hour'&&input.value.length===2){m.focus({preventScroll:true});m.select();}
 });
 document.addEventListener('focusin',event=>{const input=event.target;if(!input.matches?.('[data-time-input]'))return;const state=stateFor(input);if(state){state.active=input.dataset.timeInput;render(state,false);}});
 document.addEventListener('keydown',event=>{
  const state=stateFor(event.target);if(!state)return;
  if(event.key==='Escape'&&state.panel.open){event.preventDefault();close(state,false);return;}
  const period=event.target.closest?.('[data-time-period]');if(period){const next=['ArrowRight','ArrowDown','End'].includes(event.key)?'pm':['ArrowLeft','ArrowUp','Home'].includes(event.key)?'am':null;if(next){event.preventDefault();state.hour=state.hour%12+(next==='pm'?12:0);render(state,false);state.panel.querySelector(`[data-time-period="${next}"]`).focus({preventScroll:true});}return;}
  const input=event.target.closest?.('[data-time-input]'),dial=event.target.closest?.('[data-time-dial]');if(!input&&!dial)return;
  if(input&&event.key==='Enter'){event.preventDefault();close(state,true);return;}
  const part=input?input.dataset.timeInput:state.active,min=part==='hour'&&!state.is24?1:0,max=part==='minute'?59:state.is24?23:12;
  let current=part==='hour'?hourDisplay(state):state.minute,value;
  if(['ArrowRight','ArrowUp'].includes(event.key))value=current+1;if(['ArrowLeft','ArrowDown'].includes(event.key))value=current-1;
  if(event.key==='PageUp')value=current+5;if(event.key==='PageDown')value=current-5;if(event.key==='Home')value=min;if(event.key==='End')value=max;
  if(value!==undefined){event.preventDefault();value=wrap(value-min,max-min+1)+min;if(part==='hour')state.hour=state.is24?value:value%12+(state.hour>=12?12:0);else state.minute=value;render(state);if(input)input.select();}
  if(dial&&(event.key==='Enter'||event.key===' ')){event.preventDefault();if(state.active==='hour'){state.active='minute';render(state,false);}}
 });
 function point(state,event){const rect=state.dial.getBoundingClientRect(),style=getComputedStyle(state.dial),diameter=parseFloat(style.getPropertyValue('--_time-dial-size')),center=parseFloat(style.getPropertyValue('--_time-dial-center-size'))*rect.width/diameter,value=dialValue(event.clientX-rect.left,event.clientY-rect.top,rect.width,state.active,state.is24,center);if(value===null)return;if(state.active==='hour')state.hour=state.is24?value:value+(state.hour>=12?12:0);else state.minute=value;render(state,false);}
 document.addEventListener('pointerdown',event=>{
  const dial=event.target.closest?.('[data-time-dial]');if(!dial||event.button!==0||event.isPrimary===false)return;const state=stateFor(dial);if(!state)return;
  state.gesture={id:event.pointerId,part:state.active,hour:state.hour,minute:state.minute};dial.dataset.dragging='';dial.setPointerCapture(event.pointerId);dial.focus({preventScroll:true});event.preventDefault();point(state,event);
 });
 document.addEventListener('pointermove',event=>{const state=stateFor(event.target);if(state?.gesture?.id===event.pointerId){event.preventDefault();point(state,event);}});
 function finishPointer(event,cancel){const state=stateFor(event.target),gesture=state?.gesture;if(!gesture||gesture.id!==event.pointerId)return;cancelGesture(state);if(cancel){state.hour=gesture.hour;state.minute=gesture.minute;}else if(gesture.part==='hour')state.active='minute';render(state,false);}
 document.addEventListener('pointerup',event=>finishPointer(event,false));document.addEventListener('pointercancel',event=>finishPointer(event,true));document.addEventListener('lostpointercapture',event=>finishPointer(event,true));
 window.addEventListener('blur',()=>{for(const panel of document.querySelectorAll('.md-time-panel')){const state=stateFor(panel);if(state?.gesture){const g=state.gesture;cancelGesture(state);state.hour=g.hour;state.minute=g.minute;render(state,false);}}});
 document.addEventListener('htmx:beforeCleanupElement',event=>{const node=event.detail?.elt;if(!node)return;for(const root of [...(node.matches?.('.md-time-demo')?[node]:[]),...node.querySelectorAll('.md-time-demo')]){const state=states.get(root);if(!state)continue;state.revision++;cancelGesture(state);state.animation?.cancel();if(state.panel.open)state.panel.close();states.delete(root);}});
 reduced.addEventListener('change',()=>{if(!reduced.matches)return;for(const root of document.querySelectorAll('.md-time-demo')){const state=states.get(root);if(!state)continue;state.animation?.finish();state.monthAnimation?.finish();}});
 document.addEventListener('htmx:afterSwap',init);if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',init);else init();
})();
