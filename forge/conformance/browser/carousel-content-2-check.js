// Exercise text fitting independently of carousel size roles, including clipping
// at both viewport edges and wrapped labels in mixed-aspect-ratio items.
const fitRun=document.createElement('button');fitRun.textContent='Run text-fit regression checks';document.body.append(fitRun);
fitRun.onclick=async()=>{fitRun.disabled=true;const checks=[],check=(ok,name)=>{if(!ok)throw Error(name);checks.push(name)},wait=ms=>new Promise(r=>setTimeout(r,ms));
try{
for(const variant of ['multi-browse','hero','center-aligned-hero','uncontained','uncontained-multi-aspect-ratio']){
 samples.replaceChildren();const f=document.createElement('iframe');f.style.width='1000px';samples.append(f);
 await new Promise(resolve=>{f.onload=resolve;f.src=`/components/carousel?preview.compare=interactive&preview.variant=${variant}&preview.content=overlay&preview.viewport=600`});
 const d=f.contentDocument,w=f.contentWindow;await w.TokenAssets.ready;await d.fonts.ready;await wait(100);
 const root=d.querySelector('#component-preview .md-carousel'),view=root.querySelector('.md-carousel-viewport');view.style.scrollSnapType='none';
 const captions=[...root.querySelectorAll('.md-carousel-caption')],heights=captions.map(c=>c.getBoundingClientRect().height),widths=captions.map(c=>c.getBoundingClientRect().width);
 const max=view.scrollWidth-view.clientWidth;
 for(const scroll of [0,60,160,248,496,744,Math.max(0,max-80),max]){
  view.scrollLeft=scroll;await wait(80);
  const v=view.getBoundingClientRect();
  for(const [index,caption] of captions.entries()){
   const item=caption.closest('.md-carousel-item'),b=item.getBoundingClientRect(),c=caption.getBoundingClientRect();
   if(b.right<=v.left||b.left>=v.right||w.getComputedStyle(item).visibility==='hidden')continue;
   const range=d.createRange();range.selectNodeContents(caption);const text=range.getBoundingClientRect();
   const clearance=Math.min(text.left-Math.max(v.left,b.left,c.left),Math.min(v.right,b.right,c.right)-text.right,text.top-Math.max(v.top,b.top,c.top),Math.min(v.bottom,b.bottom,c.bottom)-text.bottom);
   check(Math.abs(c.height-heights[index])<.1&&Math.abs(c.width-widths[index])<.1,`${variant}/${scroll}/${caption.textContent.trim()}: text layout does not rewrap with the mask`);
   const opacity=parseFloat(w.getComputedStyle(caption).opacity),name=`${variant}/${scroll}/${caption.textContent.trim()}`;
   if(caption.textContent.trim()==='Green River'&&b.width<=56.1)check(opacity===0,name+': no fragments in small preview');
   if(clearance>=0)check(opacity===1,name+': fitting text remains opaque');
   else if(clearance<-.1)check(opacity<1,name+': clipped text fades');
  }
 }
}
out.textContent=JSON.stringify({passed:checks.length,checks},null,2);
}catch(error){out.textContent=JSON.stringify({passed:checks.length,error:error.stack,checks},null,2)}finally{fitRun.disabled=false}
};
