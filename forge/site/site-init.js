// Establish the responsive layout before content paints; site.js binds updates.
{
 const header=document.currentScript.parentElement;
 const wide=matchMedia(`(min-width:${header.dataset.drawerBreakpoint}px)`).matches;
 header.dataset.docked=String(wide);
 document.documentElement.dataset.siteDrawer=wide?'expanded':'modal';
 if(wide){
  const drawer=header.querySelector('#tokens-site-menu');
  drawer.dataset.variant='standard';
  drawer.setAttribute('role','navigation');
  drawer.setAttribute('open','');
 }
}
