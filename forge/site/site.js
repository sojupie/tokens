(() => {
 const compact=matchMedia('(max-width: 960px)');
 const menu=document.getElementById('tokens-site-menu'),toggle=document.getElementById('tokens-nav-toggle');
 if(!menu||!toggle)return;
 function position(){const rect=toggle.getBoundingClientRect();menu.style.top=(rect.bottom+8)+'px';menu.style.maxHeight=`calc(100dvh - ${rect.bottom+24}px)`;}
 function layout(){
  if(menu.matches(':popover-open'))menu.hidePopover();
  toggle.hidden=!compact.matches;menu.classList.toggle('md-menu',compact.matches);
  if(compact.matches){menu.setAttribute('popover','auto');position();}
  else{menu.removeAttribute('popover');menu.style.top='';menu.style.maxHeight='';}
 }
 menu.addEventListener('toggle',()=>{const open=menu.matches(':popover-open');toggle.setAttribute('aria-expanded',String(open));if(open){position();menu.querySelector('a')?.focus({preventScroll:true});}});
 compact.addEventListener('change',layout);
 window.addEventListener('resize',()=>{if(compact.matches)position();});
 layout();
})();
// Keep a viewport-height tool workspace below the measured navigation height.
{
 const header=document.getElementById('tokens-site-header');
 if(header)new ResizeObserver(([entry])=>document.documentElement.style.setProperty('--site-header-height',`${entry.borderBoxSize[0].blockSize}px`)).observe(header);
}
