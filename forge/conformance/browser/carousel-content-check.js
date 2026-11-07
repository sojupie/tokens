const run=document.querySelector('#run'),out=document.querySelector('#result'),samples=document.querySelector('#samples');
run.onclick=async()=>{run.disabled=true;const checks=[],check=(ok,name)=>{if(!ok)throw Error(name);checks.push(name)},wait=ms=>new Promise(r=>setTimeout(r,ms));
try{
for(const variant of ['multi-browse','hero','center-aligned-hero','uncontained','uncontained-multi-aspect-ratio','full-screen'])for(const content of ['overlay']){
 samples.replaceChildren();const frame=document.createElement('iframe');samples.append(frame);
 await new Promise(resolve=>{frame.onload=resolve;frame.src=`/components/carousel?preview.compare=interactive&preview.variant=${variant}&preview.content=${content}&preview.viewport=360`});
 const d=frame.contentDocument,w=frame.contentWindow;await w.TokenAssets.ready;await wait(80);
 const root=d.querySelector('#component-preview .md-carousel'),items=[...root.querySelectorAll('.md-carousel-item')];
 const tag=variant+'/'+content;
 if(variant==='full-screen')check(parseFloat(w.getComputedStyle(items[0]).borderTopLeftRadius)===0,tag+': square corners');
 if(content==='overlay'){
  const caption=items[0].querySelector('.md-carousel-caption'),style=w.getComputedStyle(caption);
  check(style.backgroundColor==='rgba(0, 0, 0, 0)'&&parseFloat(style.paddingTop)===0&&parseFloat(style.borderRadius)===0,tag+': plain text without a container');
  check(caption.textContent.trim().length<=14,tag+': brief label');
  check(parseFloat(style.fontSize)===16&&parseFloat(style.lineHeight)===24,tag+': title-medium typography');
  check(parseFloat(style.insetInlineStart)===16&&parseFloat(style.bottom)===16,tag+': 16px text inset');
  const text=items.filter(item=>item.getBoundingClientRect().width>100).map(item=>parseFloat(w.getComputedStyle(item.querySelector('.md-carousel-caption')).opacity));
  check(text.some(opacity=>opacity===1),tag+': readable item has opaque text');
  const narrow=items.find(item=>item.getBoundingClientRect().width<=56);
  if(narrow)check(parseFloat(w.getComputedStyle(narrow.querySelector('.md-carousel-caption')).opacity)===0,tag+': small mask hides text');
 }
}
out.textContent=JSON.stringify({passed:checks.length,checks},null,2);
}catch(error){out.textContent=JSON.stringify({passed:checks.length,error:error.stack,checks},null,2)}finally{run.disabled=false}
};
