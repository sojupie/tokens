(async()=>{
 await document.fonts.ready;
 const checks=[];
 const near=(name,actual,expected)=>checks.push({name,actual,expected,pass:Math.abs(actual-expected)<.1});
 for(const [id,start,end] of [['plain',16,16],['dropdown',16,8],['leading',8,16],['unchecked',16,16],['checked',8,16],['removable',16,8]]){
  const chip=document.getElementById(id),primary=chip.querySelector('.md-chip-primary')||chip,cs=getComputedStyle(chip),ps=getComputedStyle(primary);
  near(id+' height',chip.getBoundingClientRect().height,32);
  near(id+' radius',parseFloat(cs.borderTopLeftRadius),8);
  near(id+' leading padding',parseFloat(ps.paddingInlineStart),start);
  const endElement=chip.querySelector('.md-chip-remove')||primary;
  near(id+' trailing padding',parseFloat(getComputedStyle(endElement).paddingInlineEnd),end);
  near(id+' gap',parseFloat(ps.columnGap),8);
  for(const icon of chip.querySelectorAll('[data-slot]')){
   if(getComputedStyle(icon).display==='none')continue;
   near(id+' icon size',icon.getBoundingClientRect().width,18);
   near(id+' svg size',icon.querySelector('svg').getBoundingClientRect().width,18);
  }
 }
 const plain=document.querySelector('#plain .md-chip-label'),menu=document.querySelector('#menu .md-menu-item span');
 near('plain label inset',plain.getBoundingClientRect().left-document.getElementById('plain').getBoundingClientRect().left,16);
 near('menu label inset',menu.getBoundingClientRect().left-document.getElementById('menu').getBoundingClientRect().left,16);
 document.getElementById('results').textContent=JSON.stringify({passed:checks.filter(c=>c.pass).length,total:checks.length,failures:checks.filter(c=>!c.pass)},null,2);
 document.body.dataset.result=checks.every(c=>c.pass)?'pass':'fail';
})();
