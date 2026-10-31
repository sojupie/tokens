// Application state belongs to the host. Component geometry, gestures and menus stay in Forge.
(() => {
 const root=document.querySelector('.application-screen');
 function showDialog(dialog,opener) {
  if(!dialog||dialog.open)return;
  opener?.focus({preventScroll:true});
  if(globalThis.MaterialDialogs)MaterialDialogs.open(dialog,opener);else dialog.showModal();
 }
 function reviewFrame(entry,route){
  const frame=entry.querySelector('iframe'),url=new URL(route,location.href),width=entry.querySelector('[data-app-width]')?.value;
  if(width){url.searchParams.set('width',width);frame.width=width;frame.parentElement.style.setProperty('--frame-width',width);}
  entry.querySelector('[data-app-dimensions]').textContent=frame.width+' × '+frame.height;
  frame.src=url.href;entry.querySelector('[data-app-native]').href=url.href;
 }
 const completedTasks=new Set([0,1]);
 function refreshTasks() {
  const layout=document.querySelector('.app-task-layout');if(!layout)return;
  for(const row of layout.querySelectorAll('[data-app-task]')){
   const index=Number(row.dataset.appTask),panel=row.closest('[role="tabpanel"]'),filter=Number(panel.id.split('-').at(-1));
   row.closest('li').hidden=filter===1?![0,5].includes(index):filter===2?!completedTasks.has(index):false;
   row.querySelector('input').checked=completedTasks.has(index);
  }
  const tab=layout.querySelector('[role="tab"][aria-selected="true"]')?.id;
  const hidden=tab==='project-tabs-tab-1'||(tab==='project-tabs-tab-2'&&!completedTasks.has(2));
  layout.querySelector('.app-task-detail').hidden=hidden;
  layout.querySelector('.app-task-empty').hidden=!hidden;
  layout.querySelector('[data-app-task-status]').textContent=completedTasks.has(2)?'Completed':'In progress';
  const subtitle=layout.querySelector('.md-appbar-subtitle');if(subtitle)subtitle.textContent='Exhibition 2026 · '+(9-completedTasks.size)+' open';
 }
 document.addEventListener('click',event=>{
  const target=event.target;
  const state=target.closest('[data-app-view]');
  if(state){
   const entry=state.closest('[data-app-review]');
   for(const button of entry.querySelectorAll('[data-app-view]'))button.setAttribute('aria-pressed',String(button===state));
   reviewFrame(entry,state.dataset.appView);
  }
  const open=target.closest('[data-app-dialog-open]');
  if(open)showDialog(document.getElementById(open.dataset.appDialogOpen),open);
  if(target.closest('.app-task-layout [role="tab"]'))queueMicrotask(refreshTasks);
  const lesson=target.closest('[data-app-lesson]');
  if(lesson){
   for(const item of document.querySelectorAll('[data-app-lesson]'))item.setAttribute('aria-current',item===lesson?'page':'false');
   document.querySelector('[data-app-lesson-title]').textContent=lesson.dataset.appLesson;
   document.querySelector('.app-main > .md-app-bar .md-appbar-subtitle').textContent=lesson.dataset.appLesson;
   document.querySelector('[data-app-lesson-time]').textContent='00:00 / '+lesson.dataset.appDuration;
   document.querySelector('[data-app-lesson-summary]').textContent='Follow the lesson and complete the '+lesson.dataset.appLesson.toLocaleLowerCase()+' exercise in your sketchbook.';
   document.querySelector('[data-app-lesson-notes]').hidden=lesson.dataset.appLesson!=='Framing the subject';
   const progress=document.querySelector('[aria-label="Lesson playback"]');progress.setAttribute('aria-valuenow','0');
  }
  const action=target.closest('[data-app-document-action]');
  if(action?.dataset.appDocumentAction==='4')showDialog(document.getElementById('document-delete-panel'),document.querySelector('[data-app-dialog-open]'));
 });
 document.addEventListener('keydown',event=>{
  if(event.target.closest('.app-task-layout [role="tab"]'))queueMicrotask(refreshTasks);
 });
 document.addEventListener('input',event=>{
  const input=event.target;
  if(input.matches('.app-adjustments input[type="range"]')){
   const output=input.closest('section').querySelector('output');if(output)output.value=input.value;
  }
  if(input.matches('#artwork-search-view input[type="search"],input[aria-label="Search artworks"]')){
   const query=input.value.trim().toLocaleLowerCase();
   for(const card of document.querySelectorAll('.app-art-feed > *'))card.hidden=!card.textContent.toLocaleLowerCase().includes(query);
  }
 });
 document.addEventListener('change',event=>{
  if(event.target.matches('[data-app-task] input')){
   const index=Number(event.target.closest('[data-app-task]').dataset.appTask);
   if(event.target.checked)completedTasks.add(index);else completedTasks.delete(index);refreshTasks();
  }
  if(event.target.closest('.app-calendar [role="radiogroup"]')){
   const calendar=event.target.closest('.app-calendar'),mode=event.target.getAttribute('aria-label').toLowerCase();
   calendar.dataset.calendarView=mode;
   calendar.closest('.app-pane-content').querySelector('.md-appbar-subtitle').textContent={day:'Wednesday, 21 October 2026',week:'19–23 October 2026',month:'October 2026'}[mode];calendar.querySelector('.app-week-calendar').hidden=mode==='month';calendar.querySelector('.app-month-calendar').hidden=mode!=='month';
  }
  if(event.target.matches('[data-app-width]')){
   const entry=event.target.closest('[data-app-review]');reviewFrame(entry,entry.querySelector('[data-app-view][aria-pressed="true"]').dataset.appView);
  }
  if(event.target.closest('[data-app-product-option]')){
   const size=document.querySelector('[data-app-product-option="size"] input:checked').value,frame=document.querySelector('[data-app-product-option="frame"] input:checked').value;
   const price=48+(size==='2'?18:0)+(frame!=='1'?20:0);
   document.querySelector('[data-app-product-price]').textContent='€'+price;
   const buy=document.querySelector('[data-app-product-buy]');buy.textContent='Add to bag · €'+price;
   const url=new URL(buy.closest('a').href);url.searchParams.set('size',size);url.searchParams.set('frame',frame);buy.closest('a').href=url.href;
  }
  if(event.target.matches('.app-delivery-methods input[type="radio"]')){
   const controls=[...document.querySelectorAll('.app-delivery-methods input')],fee=[4.9,9.9,0][controls.indexOf(event.target)];
   const cells=document.querySelectorAll('.app-order-totals dd');
   const subtotal=Number(cells[0].textContent.replace(/[^\d.]/g,''));
   cells[1].textContent=fee?'€'+fee.toFixed(2):'Free';cells[3].textContent='€'+(subtotal+fee).toFixed(2);
  }
 });
 async function initialize(){
  await globalThis.TokenAssets?.ready;
  refreshTasks();
  for(const host of document.querySelectorAll('[data-app-initial-open]')){
   if(host.matches('dialog'))showDialog(host,document.querySelector('[data-app-dialog-open]'));
   else host.querySelector('[data-date-open],[data-time-open],[data-sheet-open]')?.click();
  }
  document.querySelector('[data-app-initial-menu] [data-menu-open]')?.click();
  for(const panel of document.querySelectorAll('.md-date-panel,.md-time-panel'))panel.addEventListener('close',()=>{
   if(panel.matches('.md-date-panel')){
    const selected=panel.querySelector('[data-date][aria-selected="true"]');
    if(selected)document.getElementById('event-date').value=new Date(selected.dataset.date+'T12:00:00').toLocaleDateString('en-GB',{weekday:'long',day:'numeric',month:'long',year:'numeric'});
   }else{
    const value=panel.closest('.md-time-demo,[data-time-picker]').querySelector('[data-time-output]').textContent;
    if(value)document.getElementById('event-time').value=value;
   }
  });
  for(const dialog of document.querySelectorAll('[data-app-dialog]'))dialog.addEventListener('close',()=>{
   if(dialog.returnValue==='delete'){
    const paper=document.querySelector('.scene-document-paper');paper.hidden=true;
    const notice=document.createElement('p');notice.className='md-typescale-body-large';notice.setAttribute('role','status');notice.textContent='Installation plan moved to Trash.';paper.after(notice);
   }
  });
 }
 if(root)initialize().catch(error=>{console.error(error);const notice=document.createElement('p');notice.setAttribute('role','alert');notice.textContent='Scene controls could not load. Reload this scene to retry.';root.prepend(notice);});
})();
