let handle;
document.querySelector('#layout').onchange=event=>{drop();list.dataset.variant=event.target.value;};
function drop(){handle?.dispatchEvent(new PointerEvent('pointerup',{bubbles:true,pointerId:1}));}
document.querySelector('#drop').onclick=drop;
document.querySelector('#run').onclick=()=>{
 drop();const row=list.firstElementChild;handle=row.querySelector('[data-list-grip]');
 // Synthetic pointer events do not create a browser pointer-capture session.
 handle.setPointerCapture=()=>{};handle.hasPointerCapture=()=>false;
 const rect=row.getBoundingClientRect(),height=list.getBoundingClientRect().height;
 const send=(type,y)=>handle.dispatchEvent(new PointerEvent(type,{bubbles:true,pointerId:1,isPrimary:true,button:0,clientX:rect.left+20,clientY:y}));
 send('pointerdown',rect.top+20);send('pointermove',rect.bottom+40);
 const marker=list.querySelector('.md-list-drop-zone'),checks={
  'List tokens set the visible row height':rect.height>=48,
  'Indicator is absolutely positioned':getComputedStyle(marker).position==='absolute',
  'Indicator matches the destination row':Math.abs(marker.getBoundingClientRect().top-(list.getBoundingClientRect().top+row.offsetTop))<1,
  'Indicator does not add a list row':Math.abs(list.getBoundingClientRect().height-height)<1,
  'Item changed position':[...list.querySelectorAll('[data-list-row]')].indexOf(row)===1
 };
 document.querySelector('#results').textContent=Object.entries(checks).map(([label,pass])=>`${pass?'PASS':'FAIL'}: ${label}`).join('\n');
};
