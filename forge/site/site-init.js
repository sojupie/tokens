// Establish the responsive layout before content paints; site.js binds updates.
{
 const header=document.currentScript.parentElement;
 const wide=matchMedia(`(min-width:${header.dataset.drawerBreakpoint}px)`).matches;
 const rail=matchMedia('(min-width:600px)').matches;
 header.dataset.docked=String(rail);
 header.dataset.siteLayout=wide?'wide':rail?'medium':'mobile';
 document.documentElement.dataset.siteLayout=header.dataset.siteLayout;
 const section=header.querySelector('[data-site-navigation]')?.dataset.activeSection??'';
 document.documentElement.dataset.siteSection=section;
 header.dataset.drawerPersistent=String(wide&&!!section);
 header.dataset.mobileLevel=section?'section':'main';
 document.documentElement.dataset.siteDrawer=rail?'expanded':'modal';
 if(rail){
  const drawer=header.querySelector('#tokens-site-menu');
  drawer.dataset.variant='standard';
  drawer.setAttribute('open','');
  drawer.setAttribute('role','navigation');
 }
}
