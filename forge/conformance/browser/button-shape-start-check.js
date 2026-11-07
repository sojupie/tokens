const frame=document.querySelector('iframe'),out=document.querySelector('#result');
frame.onload=async()=>{
 const w=frame.contentWindow,d=frame.contentDocument;await w.TokenAssets.ready;
 const button=d.querySelector('#component-preview [data-playground-specimens] .md-button'),checks=[];
 button.addEventListener('pointerdown',()=>{
  const animation=button.getAnimations().find(a=>a.effect?.getKeyframes().some(f=>f.borderRadius));
  const frames=animation?.effect.getKeyframes(),rect=button.getBoundingClientRect();
  const first=frames?.[0].borderRadius,last=frames?.at(-1).borderRadius;
  const expected=Math.min(rect.width,rect.height)/2;
  const passed=button.matches(':active')&&Math.abs(parseFloat(first)-expected)<.1&&parseFloat(last)<expected;
  checks.push({passed,nativeActive:button.matches(':active'),first,last,expected,duration:animation?.effect.getTiming().duration});
  out.textContent=JSON.stringify(checks,null,2);
 });
 out.textContent='Ready';
};
