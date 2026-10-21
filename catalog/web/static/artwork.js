(() => {
  // Same unit-mass, underdamped spring model as the Material button renderer.
  // The captured slow.spatial tokens determine both the curve and settling time.
  function artworkSpring(damping,stiffness){
    if(!(damping>0&&damping<1&&stiffness>0))return null;
    const omega=Math.sqrt(stiffness),frequency=omega*Math.sqrt(1-damping*damping),samples=[0];
    for(let i=1;i<=600;i++){
      const time=i/60;
      samples.push(1-Math.exp(-damping*omega*time)*(Math.cos(frequency*time)+damping*omega/frequency*Math.sin(frequency*time)));
      if(i>2&&Math.abs(1-samples[i])<.0001&&Math.abs(samples[i]-samples[i-1])<.0001)break;
    }
    samples[samples.length-1]=1;
    return {samples,duration:(samples.length-1)*1000/60};
  }
  function artworkFrames(states,property,spring,loop=8000,unit='fr'){
    const frames=[];
    states.forEach((from,scene)=>{
      const to=states[(scene+1)%states.length];
      spring.samples.forEach((progress,i)=>frames.push({
        offset:(scene*loop/states.length+i*spring.duration/(spring.samples.length-1))/loop,
        [property]:from.map((value,j)=>(value+(to[j]-value)*progress).toFixed(6)+unit).join(' '),
      }));
    });
    frames.push({offset:1,[property]:states[0].map(value=>value+unit).join(' ')});
    return frames;
  }
  function artworkDirectionalCorners(from,to,[small,,base,soft]){
    // Outward-moving edges soften; edges being pushed inward flatten.
    const edges=[from.left-to.left,to.right-from.right,from.top-to.top,to.bottom-from.bottom];
    const distance=Math.max(...edges.map(Math.abs));
    if(distance<.00001)return [base,base,base,base];
    return [[0,2],[1,2],[1,3],[0,3]].map(([a,b])=>{
      const direction=Math.max(-1,Math.min(1,(edges[a]+edges[b])/distance));
      return base+direction*(direction>0?soft-base:base-small);
    });
  }
  function artworkCornerStates(tracks,col,row,corners){
    const rect=scene=>{
      const widths=tracks[0][scene],heights=tracks[col+1][scene];
      const left=widths.slice(0,col).reduce((sum,n)=>sum+n,0),top=heights.slice(0,row).reduce((sum,n)=>sum+n,0);
      return {left,right:left+widths[col],top,bottom:top+heights[row]};
    };
    return tracks[0].map((_,scene)=>artworkDirectionalCorners(rect((scene+3)%4),rect(scene),corners));
  }
  if(typeof module!=='undefined')module.exports={artworkSpring,artworkFrames,artworkDirectionalCorners,artworkCornerStates};
  if(typeof document==='undefined')return;
  const reduced=matchMedia('(prefers-reduced-motion: reduce)');
  let current;
  function initialize(){
    const artwork=document.querySelector('[data-artwork]');
    if(current?.artwork===artwork)return;
    current?.dispose();current=null;
    if(!artwork)return;
    const style=getComputedStyle(artwork),spring=artworkSpring(parseFloat(style.getPropertyValue('--_art-spring-damping')),parseFloat(style.getPropertyValue('--_art-spring-stiffness')));
    const stagger=parseFloat(style.getPropertyValue('--_art-stagger')),animations=[];
    const shapeSpring=artworkSpring(parseFloat(style.getPropertyValue('--_art-shape-damping')),parseFloat(style.getPropertyValue('--_art-shape-stiffness')));
    const tracks=[
        [[1.2,.8,1],[.65,1.45,.9],[1,.65,1.35],[1.5,.85,.65]],
        [[1.3,.7,1],[.6,1,1.4],[1,1.4,.6],[1.6,.8,.6]],
        [[.7,1.5,.8],[1.5,.7,.8],[.6,1,1.4],[1,.6,1.4]],
        [[1,.7,1.3],[.8,1.5,.7],[1.4,.6,1],[.7,1.4,.9]],
      ];
    if(spring&&Number.isFinite(stagger)){
      [artwork,...artwork.querySelectorAll('.tokens-art-column')].forEach((node,i)=>{
        animations.push(node.animate(artworkFrames(tracks[i],i?'gridTemplateRows':'gridTemplateColumns',spring),{duration:8000,delay:i*stagger,iterations:Infinity,fill:'both',easing:'linear'}));
      });
      // Turn selected stripes on the same beats as their column resize and
      // corner morph. Held scenes keep neighboring tiles from turning together.
      const stripes=[...artwork.querySelectorAll('.tokens-art-sections')];
      stripes.forEach((node,index)=>{
        const col=Math.floor(index/3),phase=(col+index%3)%4;
        const states=Array.from({length:4},(_,scene)=>[Math.floor(((scene+phase)%4)/2)*90]);
        animations.push(node.animate(artworkFrames(states,'--_art-angle',shapeSpring||spring,8000,'deg'),{duration:8000,delay:(col+1)*stagger,iterations:Infinity,fill:'both',easing:'linear'}));
      });
    }
    const corners=['small','large','base','soft'].map(name=>parseFloat(style.getPropertyValue('--_art-corner-'+name)));
    if(shapeSpring&&corners.every(Number.isFinite)&&Number.isFinite(stagger)){
      // Share the owning column's resize clock; corners never pulse alone.
      const tiles=[...artwork.querySelectorAll('.tokens-art-tile')],properties=['--_art-radius-tl','--_art-radius-tr','--_art-radius-br','--_art-radius-bl'];
      [0,4,8,2,6,1,7,3,5].forEach(index=>{
        const states=artworkCornerStates(tracks,Math.floor(index/3),index%3,corners);
        const frames=artworkFrames(states,'borderRadius',shapeSpring,8000,'px').map(frame=>Object.fromEntries([['offset',frame.offset],...frame.borderRadius.split(' ').map((radius,i)=>[properties[i],radius])]));
        animations.push(tiles[index].animate(frames,{duration:8000,delay:(Math.floor(index/3)+1)*stagger,iterations:Infinity,fill:'both',easing:'linear'}));
      });
    }
    let visible=true;
    const toggle=artwork.parentElement.querySelector('.tokens-art-pause input');
    function sync(){
      const paused=!visible||document.hidden||reduced.matches||toggle.checked;
      artwork.toggleAttribute('data-paused',paused);
      animations.forEach(animation=>paused?animation.pause():animation.play());
    }
    const observer=new IntersectionObserver(entries=>{visible=entries[0].isIntersecting;sync();});observer.observe(artwork);
    toggle.addEventListener('change',sync);reduced.addEventListener('change',sync);document.addEventListener('visibilitychange',sync);
    current={artwork,dispose(){animations.forEach(animation=>animation.cancel());observer.disconnect();toggle.removeEventListener('change',sync);reduced.removeEventListener('change',sync);document.removeEventListener('visibilitychange',sync);}};
    sync();
  }
  document.addEventListener('htmx:afterSwap',initialize);document.addEventListener('htmx:historyRestore',initialize);
  if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',initialize);else initialize();
})();
