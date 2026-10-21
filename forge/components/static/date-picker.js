(() => {
 const key=date=>date.toISOString().slice(0,10),make=(y,m,d=1)=>new Date(Date.UTC(y,m,d));
 function parse(value){if(!/^\d{4}-\d{2}-\d{2}$/.test(value))return null;const [y,m,d]=value.split('-').map(Number);if(y<1900||y>2100)return null;const date=make(y,m-1,d);return key(date)===value?date:null;}
 function cells(month){const first=parse(month+'-01');if(!first)return[];const days=make(first.getUTCFullYear(),first.getUTCMonth()+1,0).getUTCDate(),n=Math.ceil((first.getUTCDay()+days)/7)*7;return Array.from({length:n},(_,i)=>key(make(first.getUTCFullYear(),first.getUTCMonth(),1-first.getUTCDay()+i)));}
 function addMonths(iso,delta){const d=parse(iso);if(!d)return null;const first=make(d.getUTCFullYear(),d.getUTCMonth()+delta),day=Math.min(d.getUTCDate(),make(first.getUTCFullYear(),first.getUTCMonth()+1,0).getUTCDate()),date=make(first.getUTCFullYear(),first.getUTCMonth(),day);return date.getUTCFullYear()<1900||date.getUTCFullYear()>2100?iso:key(date);}
 function selectRange(start,end,date){if(!start||end||date<start)return{start:date,end:null};return{start,end:date};}
 // Accept ISO year-first and en-US month-first input; normalize only after entry.
 function normalizeInput(value){
  const parts=value.trim().split(/[-/.\s]+/);if(parts.length!==3)return null;
  const [a,b,c]=parts;let iso;
  if(/^\d{4}$/.test(a))iso=`${a}-${b.padStart(2,'0')}-${c.padStart(2,'0')}`;
  else if(/^\d{4}$/.test(c))iso=`${c}-${a.padStart(2,'0')}-${b.padStart(2,'0')}`;
  return iso&&parse(iso)?iso:null;
 }
 function monthWindow(month){
  const result=[];for(let n=-3;n<=12;n++){const m=addMonths(month+'-01',n).slice(0,7);if(!result.includes(m))result.push(m);}return result;
 }
 if(typeof module!=='undefined')module.exports={parse,cells,addMonths,selectRange,readEntries,normalizeInput,monthWindow,setMode};
 if(typeof document==='undefined')return;
 const states=new WeakMap(),reduced=matchMedia('(prefers-reduced-motion: reduce)'),months=['January','February','March','April','May','June','July','August','September','October','November','December'],weekdays=['Sunday','Monday','Tuesday','Wednesday','Thursday','Friday','Saturday'];
 const stateFor=target=>states.get(target.closest?.('.md-date-demo'));
 const element=(tag,className,text)=>{const e=document.createElement(tag);if(className)e.className=className;if(text!==undefined)e.textContent=text;return e;};
 const monthLabel=month=>{const d=parse(month+'-01');return months[d.getUTCMonth()]+' '+d.getUTCFullYear();};
 const display=iso=>iso?new Intl.DateTimeFormat('en-US',{month:'short',day:'numeric',timeZone:'UTC'}).format(parse(iso)):'—';
 function syncEntries(state){for(const input of state.fields){input.value=state[input.dataset.dateValue]||'';input.setAttribute('aria-invalid','false');input.parentElement.setAttribute('aria-invalid','false');state.root.querySelector(`[data-date-value-help="${input.dataset.dateValue}"]`).textContent='YYYY-MM-DD';}}
 function readEntries(state){
  if(!state.fields.length)return true;
  const values={};let valid=true;
  for(const input of state.fields)values[input.dataset.dateValue]=input.value.trim()?(normalizeInput(input.value)||input.value.trim()):null;
  for(const input of state.fields){
   const part=input.dataset.dateValue,value=values[part],reversed=part==='end'&&value&&values.start&&value<values.start;
   const message=part==='start'&&!value&&values.end?'Enter a start date.':value&&!parse(value)?'Enter a date from 1900 to 2100 as YYYY-MM-DD.':reversed?'End date must follow the start date.':'';
   input.setAttribute('aria-invalid',String(!!message));input.parentElement.setAttribute('aria-invalid',String(!!message));state.root.querySelector(`[data-date-value-help="${part}"]`).textContent=message||'YYYY-MM-DD';if(message)valid=false;
  }
  if(!valid)return false;
  state.start=values.start;state.end=state.range?values.end:null;if(state.start)state.month=state.start.slice(0,7);
  state.committed={start:state.start,end:state.end};syncEntries(state);return true;
 }
 // Keep the calendar presentation while switching to the shared modal-input surface.
 function setMode(state,mode){
  const {panel}=state;if(panel.dataset.mode==='calendar'&&mode==='input')state.calendarScroll={month:state.month,top:panel.querySelector('[data-date-calendars]').scrollTop};panel.dataset.mode=mode;
  if(mode==='calendar'&&state.calendarScroll&&state.calendarScroll.month!==state.month)state.months=null;
  if(state.variant!=='full-screen'||!state.inline)return;
  if(mode==='input'&&panel.hasAttribute('data-date-inline')){
   panel.close();panel.removeAttribute('data-date-inline');panel.showModal();
  }else if(mode==='calendar'&&!panel.hasAttribute('data-date-inline')){
   panel.close();panel.setAttribute('data-date-inline','');panel.open=true;
  }
 }
 function announce(state,text){state.panel.querySelector('.md-date-status').textContent=text;}
 function dayButton(state,iso,month){
  const button=element('button','md-date-day',Number(iso.slice(8))),date=parse(iso),selected=iso===state.start||(state.range&&iso===state.end),inRange=state.range&&state.start&&state.end&&iso>=state.start&&iso<=state.end;
  button.type='button';button.setAttribute('role','gridcell');button.setAttribute('aria-label',new Intl.DateTimeFormat('en-US',{weekday:'long',year:'numeric',month:'long',day:'numeric',timeZone:'UTC'}).format(date));button.setAttribute('aria-selected',String(!!(selected||inRange)));button.tabIndex=-1;
  Object.assign(button.dataset,{date:iso,dateSelected:String(selected),today:String(iso==='2026-10-18'),outsideMonth:String(iso.slice(0,7)!==month),inRange:String(!!inRange)});
  if(iso==='2026-10-18')button.setAttribute('aria-current','date');
  const representative={'selected':'2026-10-21','unselected':'2026-10-22','today':'2026-10-18','in-range':'2026-10-23'}[state.panel.dataset.selection];
  if(state.root.dataset.control==='day'&&state.panel.dataset.state!=='interactive'&&iso===representative)button.dataset.componentState=state.panel.dataset.state;
  return button;
 }
 function calendar(state,month){
  const section=element('section','md-date-month-calendar');section.dataset.calendarMonth=month;
  if(state.range&&state.variant!=='docked')section.append(element('h4','',monthLabel(month)));
  const grid=element('div','md-date-grid');grid.setAttribute('role','grid');grid.setAttribute('aria-label',monthLabel(month));
  const head=element('div','md-date-week');head.setAttribute('role','row');for(const name of weekdays){const cell=element('span','md-date-weekday',name[0]);cell.setAttribute('role','columnheader');cell.setAttribute('aria-label',name);cell.title=name;head.append(cell);}grid.append(head);
  const dates=cells(month);grid.setAttribute('aria-colcount','7');grid.setAttribute('aria-rowcount',String(dates.length/7+1));let row;
  dates.forEach((iso,index)=>{
   if(index%7===0){row=element('div','md-date-row');row.setAttribute('role','row');grid.append(row);}
   const cell=element('span','md-date-cell');cell.setAttribute('role','presentation');
   if(state.range)Object.assign(cell.dataset,{inRange:String(!!(state.start&&state.end&&iso>=state.start&&iso<=state.end)),rangeStart:String(iso===state.start),rangeEnd:String(iso===state.end)});
   if(parse(iso)&&(iso.slice(0,7)===month||state.variant==='docked')){const b=dayButton(state,iso,month);b.setAttribute('aria-colindex',String(index%7+1));b.setAttribute('aria-rowindex',String(Math.floor(index/7)+2));cell.append(b);}row.append(cell);
  });
  const buttons=[...grid.querySelectorAll('.md-date-day')],chosen=buttons.find(b=>b.dataset.date===state.start)||buttons.find(b=>b.dataset.date===month+'-01');if(chosen)chosen.tabIndex=0;
  section.append(grid);return section;
 }
 function render(state,focusISO){
  state.panel.dataset.month=state.month;
  const {panel}=state,calendars=panel.querySelector('[data-date-calendars]');
  const scroll=calendars.scrollTop;
  if(state.range&&state.variant!=='docked'){
   const reset=!state.months||!state.months.includes(state.month);if(reset)state.months=monthWindow(state.month);
   calendars.replaceChildren(...state.months.map(month=>calendar(state,month)));
   calendars.scrollTop=reset?calendars.querySelector(`[data-calendar-month="${state.month}"]`).offsetTop-calendars.offsetTop:scroll;
  }else calendars.replaceChildren(calendar(state,state.month));
  for(const b of calendars.querySelectorAll('.md-date-day'))b.tabIndex=-1;
  const tab=calendars.querySelector(`[data-date="${focusISO||state.start}"]`)||calendars.querySelector(`[data-date="${state.month}-01"]`);if(tab)tab.tabIndex=0;
  const years=panel.querySelector('[data-date-years]');if(panel.dataset.pane==='year'&&!years.children.length){for(let y=1900;y<=2100;y++){
   const b=element('button',state.variant==='docked'?'md-date-month-item':'md-date-year');
   if(state.variant==='docked'){b.append(panel.querySelector('[data-date-month] svg').cloneNode(true),element('span','',y));}else b.textContent=String(y);
   b.type='button';b.dataset.dateYear=String(y);years.append(b);
  }}
  panel.querySelector('[data-date-years]').hidden=panel.dataset.pane!=='year';panel.querySelector('[data-date-months]').hidden=panel.dataset.pane!=='month';
  panel.querySelector('[data-date-month-label]').textContent=state.variant==='docked'?months[Number(state.month.slice(5))-1].slice(0,3):monthLabel(state.month);
  for(const label of panel.querySelectorAll('[data-date-year-label]'))label.textContent=state.month.slice(0,4);
  panel.style.setProperty('--_date-rows',String(cells(state.month).length/7));
  panel.querySelector('[data-date-title]').textContent=state.range?'Select dates':'Select date';
  panel.querySelector('[data-date-headline]').textContent=panel.dataset.mode==='input'?(state.range?'Enter dates':'Enter date'):state.range?display(state.start)+' – '+display(state.end):state.start?new Intl.DateTimeFormat('en-US',{weekday:'short',month:'short',day:'numeric',timeZone:'UTC'}).format(parse(state.start)):'Select date';
  for(const menu of panel.querySelectorAll('[data-date-menu]')){
   menu.setAttribute('aria-expanded',String(panel.dataset.pane===menu.dataset.dateMenu));
   menu.disabled=menu.dataset.componentState==='disabled'||(state.variant==='docked'&&['year','month'].includes(panel.dataset.pane)&&panel.dataset.pane!==menu.dataset.dateMenu);
  }
  for(const step of panel.querySelectorAll('[data-date-step]'))step.disabled=addMonths(state.month+'-01',Number(step.dataset.dateStep)).slice(0,7)===state.month;
  panel.querySelector('[data-date-mode]').setAttribute('aria-label',panel.dataset.mode==='input'?'Choose from calendar':'Enter date');
  for(const input of panel.querySelectorAll('[data-date-input]')){input.value=state[input.dataset.dateInput]||'';input.setCustomValidity('');input.setAttribute('aria-invalid','false');input.closest('.md-outlined-text-field').setAttribute('aria-invalid','false');}
  panel.querySelector('[data-date-error]').textContent='';
  for(const year of panel.querySelectorAll('[data-date-year]')){const selected=year.dataset.dateYear===state.month.slice(0,4);year.setAttribute('aria-pressed',String(selected));year.tabIndex=selected?0:-1;}
  for(const month of panel.querySelectorAll('[data-date-month]')){const selected=Number(month.dataset.dateMonth)===Number(state.month.slice(5));month.setAttribute('aria-pressed',String(selected));month.tabIndex=selected?0:-1;}
  for(const save of panel.querySelectorAll('[data-date-save]'))save.disabled=panel.dataset.mode==='calendar'&&(!state.start||(state.range&&!state.end));
  if(panel.dataset.mode==='calendar'&&state.calendarScroll){if(state.calendarScroll.month===state.month)calendars.scrollTop=state.calendarScroll.top;delete state.calendarScroll;}
  if(state.variant==='docked'&&panel.open&&!panel.hasAttribute('data-date-inline'))positionDocked(state);
  if(focusISO)panel.querySelector(`[data-date="${focusISO}"]`)?.focus({preventScroll:true});
 }
 function timing(state){const s=getComputedStyle(state.panel);return{duration:parseFloat(s.getPropertyValue('--md-sys-motion-duration-short3')),easing:s.getPropertyValue('--md-sys-motion-easing-standard').trim()};}
 function positionDocked(state){
  const r=state.anchor.getBoundingClientRect(),box=state.panel.getBoundingClientRect(),s=getComputedStyle(state.panel),gap=parseFloat(s.getPropertyValue('--md-sys-measurement-space50')),gutter=parseFloat(s.getPropertyValue('--md-sys-measurement-space200'));
  const below=r.bottom+gap,above=r.top-box.height-gap,maxTop=Math.max(gutter,innerHeight-box.height-gutter),top=below+box.height+gutter<=innerHeight?below:above>=gutter?above:Math.max(gutter,Math.min(below,maxTop));
  Object.assign(state.panel.style,{left:Math.max(gutter,Math.min(r.left,innerWidth-box.width-gutter))+'px',top:top+'px'});
 }
 async function close(state,save=false,focus=true,preserveEntry=false){
  if(save){render(state);state.committed={start:state.start,end:state.end};}
  else {Object.assign(state,state.committed);render(state);}
  if(!preserveEntry)syncEntries(state);
  if(state.inline&&state.variant==='full-screen'&&!state.panel.hasAttribute('data-date-inline')){setMode(state,'calendar');render(state);initialFocus(state);return;}
  if(state.panel.hasAttribute('data-date-inline'))return;
  state.closing=true;const revision=state.revision=(state.revision||0)+1;state.animation?.cancel();
  if(!reduced.matches){const a=state.panel.animate([{opacity:1},{opacity:0}],timing(state));state.animation=a;try{await a.finished;}catch{return;}if(revision!==state.revision)return;a.cancel();}
  state.closing=false;if(state.panel.matches(':popover-open'))state.panel.hidePopover();state.panel.close();state.opener.setAttribute('aria-expanded','false');for(const input of state.fields)if(input.hasAttribute('aria-expanded'))input.setAttribute('aria-expanded','false');if(focus)(state.returnFocus||state.opener).focus({preventScroll:true});
 }
 function open(state,focusCalendar=true){
  if(state.panel.open&&!state.closing){close(state,false);return;}
  readEntries(state);state.months=null;state.returnFocus=focusCalendar?state.opener:state.fields[0];
  const reopening=state.panel.open,from=reopening?getComputedStyle(state.panel).opacity:0;state.closing=false;
  state.revision=(state.revision||0)+1;state.animation?.cancel();state.panel.style.left='';state.panel.style.top='';state.panel.style.position='';state.panel.style.margin='';
  if(!reopening&&state.variant==='docked'){
   Object.assign(state.panel.style,{position:'fixed',inset:'auto',margin:'0'});state.panel.setAttribute('popover','manual');state.panel.open=true;state.panel.showPopover();
  }else if(!reopening)state.panel.showModal();
  state.opener.setAttribute('aria-expanded','true');for(const input of state.fields)if(input.hasAttribute('aria-expanded'))input.setAttribute('aria-expanded','true');render(state);if(focusCalendar)initialFocus(state);else state.fields[0].focus({preventScroll:true});
  if(!reduced.matches)state.animation=state.panel.animate([{opacity:from},{opacity:1}],timing(state));
 }
 function centerChoice(state){
 const year=state.panel.dataset.pane==='year',scroll=state.panel.querySelector(year?'[data-date-years]':'[data-date-months]'),item=scroll.querySelector('[aria-pressed="true"]');
 if(item){let top=item.offsetTop-scroll.offsetTop-scroll.clientHeight/2+item.offsetHeight/2;if(state.variant==='docked')top=Math.round(top/item.offsetHeight)*item.offsetHeight;scroll.scrollTop=top;}return item;
 }
 function initialFocus(state){
  if(state.panel.dataset.mode==='input')state.panel.querySelector('[data-date-input]').focus({preventScroll:true});
  else if(state.panel.dataset.pane==='year'){centerChoice(state)?.focus({preventScroll:true});}
  else if(state.panel.dataset.pane==='month')centerChoice(state)?.focus({preventScroll:true});
  else (state.panel.querySelector(`[data-date="${state.start}"]`)||state.panel.querySelector(`[data-date="${state.month}-01"]`)||state.panel.querySelector('.md-date-day')).focus({preventScroll:true});
 }
 function validate(state){
  if(state.panel.dataset.mode!=='input')return!!(state.start&&(!state.range||state.end));
  const inputs=[...state.panel.querySelectorAll('[data-date-input]')];let bad;
  for(const input of inputs){const normalized=normalizeInput(input.value),valid=!!normalized;if(valid)input.value=normalized;input.setAttribute('aria-invalid',String(!valid));input.closest('.md-outlined-text-field').setAttribute('aria-invalid',String(!valid));if(!valid&&!bad)bad=input;}
  if(!bad&&state.range&&inputs[1].value<inputs[0].value)bad=inputs[1];
  if(bad){const message=bad.value&&parse(bad.value)?'End date must follow the start date.':'Enter a date from 1900 to 2100 as YYYY-MM-DD.';bad.setAttribute('aria-invalid','true');bad.closest('.md-outlined-text-field').setAttribute('aria-invalid','true');state.panel.querySelector('[data-date-error]').textContent=message;bad.focus({preventScroll:true});return false;}
  for(const input of inputs)state[input.dataset.dateInput]=input.value;
  state.month=state.start.slice(0,7);return true;
 }
 function changeMonth(state,delta){
  state.month=addMonths(state.month+'-01',delta).slice(0,7);if(state.range)state.months=null;render(state);announce(state,monthLabel(state.month));
  if(!state.range&&state.variant!=='docked'&&!reduced.matches){const view=state.panel.querySelector('[data-date-calendars]');state.monthAnimation?.cancel();state.monthAnimation=view.animate([{translate:`${delta>0?10:-10}% 0`,opacity:.4},{translate:'0 0',opacity:1}],timing(state));}
 }
 function extendMonths(state,view){
  if(!state.range||state.variant==='docked'||view.hidden||!view.clientHeight)return;
  const before=view.scrollTop<80,after=view.scrollTop+view.clientHeight>view.scrollHeight-80;if(!before&&!after)return;
  const edge=state.months[before?0:state.months.length-1],added=[];
  for(let n=1;n<=6;n++){const month=addMonths(edge+'-01',before?-n:n).slice(0,7);if(!state.months.includes(month)&&!added.includes(month))added.push(month);}
  if(!added.length)return;const top=view.scrollTop,height=view.scrollHeight;
  if(before){added.reverse();state.months.unshift(...added);view.prepend(...added.map(m=>calendar(state,m)));view.scrollTop=top+view.scrollHeight-height;}
  else {state.months.push(...added);view.append(...added.map(m=>calendar(state,m)));}
  for(const section of [...view.children].filter(e=>added.includes(e.dataset.calendarMonth)))for(const b of section.querySelectorAll('.md-date-day'))b.tabIndex=-1;
 }

 function init(){for(const root of document.querySelectorAll('.md-date-demo')){
  if(states.has(root))continue;const panel=root.querySelector('.md-date-panel'),state={root,panel,opener:root.querySelector('[data-date-open]'),anchor:root.querySelector('[data-date-anchor]'),fields:[...root.querySelectorAll('[data-date-value]')],variant:panel.dataset.variant,inline:panel.hasAttribute('data-date-inline'),range:panel.dataset.range==='true',month:panel.dataset.month,start:'2026-10-21',end:panel.dataset.range==='true'?'2026-10-26':null};state.committed={start:state.start,end:state.end};states.set(root,state);render(state);if(panel.open&&['year','month'].includes(panel.dataset.pane))centerChoice(state);
  const view=panel.querySelector('[data-date-calendars]');view.addEventListener('scroll',()=>extendMonths(state,view),{passive:true});
  let gesture;
  view.addEventListener('pointerdown',event=>{if(state.range||state.variant==='docked'||event.button!==0)return;gesture={id:event.pointerId,x:event.clientX,y:event.clientY};});
  view.addEventListener('pointermove',event=>{if(!gesture||gesture.id!==event.pointerId)return;const dx=event.clientX-gesture.x,dy=event.clientY-gesture.y;if(!gesture.dragging&&Math.abs(dx)>10&&Math.abs(dx)>Math.abs(dy)*1.5){gesture.dragging=true;view.setPointerCapture(event.pointerId);}if(gesture.dragging){event.preventDefault();view.style.translate=`${dx/3}px 0`;}});
  const finishSwipe=event=>{if(!gesture||gesture.id!==event.pointerId)return;const g=gesture;gesture=null;view.style.translate='';if(!g.dragging)return;state.swiped=true;setTimeout(()=>state.swiped=false,0);const dx=event.clientX-g.x;if(event.type!=='pointercancel'&&Math.abs(dx)>=48){const rtl=getComputedStyle(view).direction==='rtl';changeMonth(state,(dx<0?1:-1)*(rtl?-1:1));}else if(!reduced.matches)view.animate([{translate:`${dx/3}px 0`},{translate:'0 0'}],timing(state));};
  view.addEventListener('pointerup',finishSwipe);view.addEventListener('pointercancel',finishSwipe);
  panel.addEventListener('cancel',event=>{event.preventDefault();close(state,false);});
 }}
 document.addEventListener('click',event=>{
  const state=stateFor(event.target);if(!state)return;
  if(event.target.closest('[data-date-open]')){open(state);return;}
  const date=event.target.closest('[data-date]');if(date){if(state.swiped){event.preventDefault();return;}const iso=date.dataset.date;if(state.range)Object.assign(state,selectRange(state.start,state.end,iso));else state.start=iso;
   if(!state.range)state.month=iso.slice(0,7);render(state,iso);announce(state,state.range?`${display(state.start)} to ${display(state.end)}`:date.getAttribute('aria-label'));return;}
  const step=event.target.closest('[data-date-step]');if(step){changeMonth(state,Number(step.dataset.dateStep));return;}
  const menu=event.target.closest('[data-date-menu]');if(menu){state.panel.dataset.pane=state.panel.dataset.pane===menu.dataset.dateMenu?'day':menu.dataset.dateMenu;render(state);initialFocus(state);return;}
  if(event.target.closest('[data-date-mode]')){if(state.panel.dataset.mode==='input'&&!validate(state))return;setMode(state,state.panel.dataset.mode==='input'?'calendar':'input');state.panel.dataset.pane='day';render(state);initialFocus(state);return;}
  const year=event.target.closest('[data-date-year]');if(year){state.month=year.dataset.dateYear+state.month.slice(4);state.months=null;state.panel.dataset.pane='day';render(state);initialFocus(state);return;}
  const month=event.target.closest('[data-date-month]');if(month){state.month=state.month.slice(0,4)+'-'+String(month.dataset.dateMonth).padStart(2,'0');state.months=null;state.panel.dataset.pane='day';render(state);initialFocus(state);return;}
  if(event.target.closest('[data-date-cancel]')){close(state,false);return;}
  if(event.target.closest('[data-date-save]')&&validate(state))close(state,true);
 });
 document.addEventListener('change',event=>{if(event.target.matches?.('[data-date-input]')){const iso=normalizeInput(event.target.value);if(iso)event.target.value=iso;return;}if(!event.target.matches?.('[data-date-value]'))return;const state=stateFor(event.target);if(state&&readEntries(state))render(state);});
 document.addEventListener('keydown',event=>{
  let state=stateFor(event.target);if(!state&&event.key==='Escape'){const panel=document.querySelector('.md-date-panel[open][data-variant="docked"]');if(panel)state=stateFor(panel);}if(!state)return;
  if(event.key==='Escape'&&state.panel.open){event.preventDefault();if(state.variant==='docked'&&['year','month'].includes(state.panel.dataset.pane)){const pane=state.panel.dataset.pane;state.panel.dataset.pane='day';render(state);state.panel.querySelector(`[data-date-menu="${pane}"]`).focus();}else close(state,false);return;}
  if(event.target.matches?.('[data-date-value]')){if(event.altKey&&event.key==='ArrowDown'){event.preventDefault();if(!state.panel.open)open(state);else initialFocus(state);}if(event.key==='Enter'){event.preventDefault();if(readEntries(state)){if(state.panel.open)close(state,true);else render(state);}}return;}
  if(event.target.matches?.('[data-date-input]')&&event.key==='Enter'){event.preventDefault();if(validate(state))close(state,true);return;}
  if(event.shiftKey&&['M','Y'].includes(event.key.toUpperCase())&&state.panel.dataset.mode==='calendar'){
   const pane=event.key.toUpperCase()==='M'?'month':'year';const menu=state.panel.querySelector(`[data-date-menu="${pane}"]`);if(menu){event.preventDefault();state.panel.dataset.pane=pane;render(state);initialFocus(state);}return;
  }
  const button=event.target.closest?.('[data-date]');if(button){
   if(event.key==='Enter'){event.preventDefault();const iso=button.dataset.date;if(state.range)Object.assign(state,selectRange(state.start,state.end,iso));else state.start=iso;render(state);if(validate(state))close(state,true);return;}const iso=button.dataset.date,d=parse(iso),rtl=getComputedStyle(button).direction==='rtl';let target;
   const delta={ArrowLeft:rtl?1:-1,ArrowRight:rtl?-1:1,ArrowUp:-7,ArrowDown:7,Home:1-d.getUTCDate(),End:make(d.getUTCFullYear(),d.getUTCMonth()+1,0).getUTCDate()-d.getUTCDate()}[event.key];
   if(delta!==undefined)target=key(make(d.getUTCFullYear(),d.getUTCMonth(),d.getUTCDate()+delta));
   if(event.key==='PageUp'||event.key==='PageDown')target=addMonths(iso,(event.key==='PageUp'?-1:1)*(event.shiftKey?12:1));
   if(!target||!parse(target))return;event.preventDefault();
   if(!state.panel.querySelector(`[data-date="${target}"]`)){state.month=target.slice(0,7);state.months=null;}
   render(state,target);return;
  }
  const year=event.target.closest?.('[data-date-year]'),month=event.target.closest?.('[data-date-month]');if(year||month){const buttons=[...state.panel.querySelectorAll(year?'[data-date-year]':'[data-date-month]')],index=buttons.indexOf(event.target),columns=year&&state.variant!=='docked'?3:1,delta=event.key==='ArrowDown'?columns:event.key==='ArrowUp'?-columns:event.key==='ArrowLeft'?-1:event.key==='ArrowRight'?1:0;if(delta){event.preventDefault();for(const b of buttons)b.tabIndex=-1;const target=buttons[Math.max(0,Math.min(buttons.length-1,index+delta))];target.tabIndex=0;target.focus();}}
 });
 document.addEventListener('pointerdown',event=>{for(const panel of document.querySelectorAll('.md-date-panel[open]:not([data-date-inline])')){const state=stateFor(panel);if(!state)continue;if(state.variant==='docked'){if(!panel.contains(event.target)&&!state.anchor.contains(event.target)){readEntries(state);close(state,false,false,true);}}else if(event.target===panel){const r=panel.getBoundingClientRect();if(event.clientX<r.left||event.clientY<r.top||event.clientX>r.right||event.clientY>r.bottom)close(state,false);}}});
 document.addEventListener('htmx:beforeCleanupElement',event=>{const node=event.detail?.elt;if(!node)return;for(const root of [...(node.matches?.('.md-date-demo')?[node]:[]),...node.querySelectorAll('.md-date-demo')]){const state=states.get(root);if(!state)continue;state.revision++;state.animation?.cancel();state.monthAnimation?.cancel();if(state.panel.matches(':popover-open'))state.panel.hidePopover();if(state.panel.open)state.panel.close();states.delete(root);}});
 window.addEventListener('resize',()=>{for(const panel of document.querySelectorAll('.md-date-panel[open][data-variant="docked"]:not([data-date-inline])')){const state=stateFor(panel);if(state)positionDocked(state);}});
 document.addEventListener('htmx:afterSwap',init);if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',init);else init();
})();
