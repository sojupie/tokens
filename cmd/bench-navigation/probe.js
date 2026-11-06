(() => {
  let active=false, start=0, marks={}, target='', kind='', expected='', done=false;
  const mark=name=>{if(!active || name==='ready' && marks.ready!==undefined)return;marks[name]=performance.now()-start;performance.mark('tokens:'+name);};
  const tasks=[];
  new PerformanceObserver(list=>tasks.push(...list.getEntries())).observe({type:'longtask',buffered:true});
  function check(){
    if(!active)return;
    const matches=kind==='download'
      ? document.querySelector('#tokens-export-fields')?.textContent.includes('ZIP archive') &&
        document.querySelector('input[name=content][value=implementation]')?.checked &&
        !document.querySelector('#tokens-export-fields fieldset:disabled') &&
        document.querySelector('[data-export-confirmation]')?.disabled===false
      : location.pathname===target && document.querySelector('#component-preview '+expected);
    if(!matches)return;
    if(marks.dom===undefined){mark('dom');requestAnimationFrame(()=>requestAnimationFrame(()=>{mark('frame-opportunity');if(marks.ready!==undefined)done=true;}));}
    if(kind==='download' && marks.ready===undefined)mark('ready');
    if(marks.ready!==undefined && marks['frame-opportunity']!==undefined)done=true;
  }
  new MutationObserver(check).observe(document,{subtree:true,childList:true,attributes:true});
  for(const name of ['htmx:before:request','htmx:before:swap','htmx:after:swap','htmx:before:settle','htmx:after:settle'])document.addEventListener(name,()=>{mark(name);check();});
  document.addEventListener('tokens:assets-ready',()=>{check();if(active&&marks.dom!==undefined){mark('ready');check();}});
  globalThis.TokenBenchmark={mark,
    reset(nextKind,nextTarget,nextExpected){performance.clearMarks();tasks.length=0;marks={};kind=nextKind;target=nextTarget;expected=nextExpected;done=false;start=performance.now();active=true;mark('activation');},
    get done(){return done;},
    result(){active=false;return {userAgent:navigator.userAgent,viewport:{width:innerWidth,height:innerHeight},marks,longTasks:tasks.filter(t=>t.startTime>=start).map(t=>({start:t.startTime-start,duration:t.duration})),resources:performance.getEntriesByType('resource').filter(r=>r.startTime>=start-10000).map(r=>({url:r.name,start:r.startTime-start,duration:r.duration,protocol:r.nextHopProtocol,decodedBytes:r.decodedBodySize,transferBytes:r.transferSize,encodedBytes:r.encodedBodySize,serverTiming:r.serverTiming.map(t=>({name:t.name,duration:t.duration,description:t.description}))}))};}
  };
})();
