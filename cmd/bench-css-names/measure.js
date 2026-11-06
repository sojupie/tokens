(async()=>{
  await document.fonts.ready;
  await new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r)));
  const nav=performance.getEntriesByType('navigation')[0],resources=performance.getEntriesByType('resource');
  const result={userAgent:navigator.userAgent,documentMode:document.compatMode,elements:document.querySelectorAll('*').length,
    fcpMs:performance.getEntriesByName('first-contentful-paint')[0]?.startTime||0,
    domContentLoadedMs:nav.domContentLoadedEventEnd,loadMs:nav.loadEventEnd,
    cssTransferBytes:resources.filter(r=>r.initiatorType==='link'&&r.name.includes('.css')).reduce((n,r)=>n+r.transferSize,0),
    totalTransferBytes:nav.transferSize+resources.reduce((n,r)=>n+r.transferSize,0)};
  // Capture page-load counters before the isolated interaction tests.
  result.recalculateMs=[];
  const theme=document.documentElement.getAttribute('data-theme');
  for(let i=0;i<25;i++){
    const start=performance.now();
    document.documentElement.setAttribute('data-theme',i%2?'light':'dark');
    document.body.getBoundingClientRect();
    result.recalculateMs.push(performance.now()-start);
  }
  if(theme===null)document.documentElement.removeAttribute('data-theme');else document.documentElement.setAttribute('data-theme',theme);
  document.body.getBoundingClientRect();
  const root=document.documentElement,contrast=root.getAttribute('data-contrast');
  const frames=()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r)));
  const set=state=>{const [theme,contrast]=state.split('/');root.dataset.theme=theme;root.dataset.contrast=contrast;};
  const transitions=[];
  for(const c of ['default','medium','high']){
    transitions.push({kind:'theme',from:'light/'+c,to:'dark/'+c},{kind:'theme',from:'dark/'+c,to:'light/'+c});
  }
  for(const t of ['light','dark'])for(const from of ['default','medium','high'])for(const to of ['default','medium','high']){
    if(from!==to)transitions.push({kind:'contrast',from:t+'/'+from,to:t+'/'+to});
  }
  result.appearance=[];
  // Warm every transition first. Reverse the sequence in alternate blocks so
  // individual theme/contrast directions do not always occupy the same position.
  for(let block=0;block<3;block++)for(const step of block%2?[...transitions].reverse():transitions){
    set(step.from);document.body.getBoundingClientRect();await frames();
    const start=performance.now();set(step.to);document.body.getBoundingClientRect();
    const flushMs=performance.now()-start;await frames();
    if(block)result.appearance.push({...step,flushMs,frameMs:performance.now()-start});
  }
  if(theme===null)delete root.dataset.theme;else root.dataset.theme=theme;
  if(contrast===null)delete root.dataset.contrast;else root.dataset.contrast=contrast;
  document.body.getBoundingClientRect();
  // Allocate repeated stylesheet copies only after appearance measurements.
  // Fetch source text: CSSOM serialization can normalize whitespace or syntax.
  const sources=await Promise.all([...document.querySelectorAll('link[rel=stylesheet]')].map(async l=>await(await fetch(l.href)).text()));
  result.parseMs=[];
  for(let i=0;i<25;i++){
    const start=performance.now();
    for(const css of sources)new CSSStyleSheet().replaceSync(css);
    result.parseMs.push(performance.now()-start);
  }
  return result;
})()
