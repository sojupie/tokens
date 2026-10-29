// The same drawer is docked on wide windows and modal when content needs the space.
{
 const header=document.getElementById('tokens-site-header');
 if(header){
  const drawer=header.querySelector('#tokens-site-menu');
  const opener=header.querySelector('[data-navigation-open]');
  const wide=matchMedia(`(min-width:${header.dataset.drawerBreakpoint}px)`);
  function sync(){
   header.dataset.docked=String(wide.matches);
   document.documentElement.dataset.siteDrawer=wide.matches?'expanded':'modal';
   opener.setAttribute('aria-expanded',String(drawer.open));
   opener.setAttribute('aria-haspopup',wide.matches?'false':'dialog');
   document.documentElement.style.setProperty('--site-header-height',wide.matches?'0px':`${header.getBoundingClientRect().height}px`);
  }
  function arrange(){
   if(drawer.open)drawer.close();
   drawer.dataset.variant=wide.matches?'standard':'modal';
   if(wide.matches){
    drawer.setAttribute('role','navigation');
    {
     const previous=document.activeElement;
     drawer.show();
     if(previous===document.body)document.activeElement?.blur();
     else if(previous?.getClientRects().length)previous.focus({preventScroll:true});
    }
   } else drawer.removeAttribute('role');
   sync();
  }
  drawer.addEventListener('close',sync);
  wide.addEventListener('change',arrange);
  new ResizeObserver(sync).observe(header);
  arrange();
 }
}
