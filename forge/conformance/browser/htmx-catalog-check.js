const frame=document.querySelector('iframe'),results=document.querySelector('#results');
const wait=async(test,label)=>{for(let i=0;i<240;i++){if(test())return;await new Promise(r=>setTimeout(r,50));}throw Error('Timed out: '+label);};
const assert=(ok,label)=>{if(!ok)throw Error(label);results.insertAdjacentHTML('beforeend','<li>PASS: '+label+'</li>');};
const load=path=>new Promise(resolve=>{frame.onload=resolve;frame.src=path;});
document.querySelector('#run').onclick=async()=>{
 results.textContent='';document.querySelector('#run').disabled=true;
 try{
 const selectPreviewOption=async(field,value)=>{
   const doc=frame.contentDocument, previous=doc.querySelector('#component-preview');
   previous.querySelector('#preview-'+field).click();
   previous.querySelector('#preview-'+field+'-options [data-field-value="'+value+'"]').click();
   await wait(()=>!previous.isConnected,field+' preview swap');
   assert(new URLSearchParams(frame.contentWindow.location.search).get('preview.'+field)===value,field+' filter updates the preview URL');
 };
 frame.style.width='390px';
 await load('/components/time-pickers?preview.compare=variants');
 await frame.contentWindow.TokenAssets.ready;
 const dialPreview=frame.contentDocument.querySelector('[data-time-preview][data-mode="clock"]');
 const expressivePanels=frame.contentDocument.querySelectorAll('[data-time-preview][data-time-style="expressive"]');
 assert(expressivePanels.length===3&&Array.from(expressivePanels).every(panel=>frame.contentWindow.getComputedStyle(panel).boxShadow==='none'),'Expressive modal variants use the same flat surface');
 const nextTimeCase=dialPreview.closest('figure').nextElementSibling;
 const dialField=dialPreview.querySelector('[data-time-value="hour"]').getBoundingClientRect();
 const nextTimeCaseTop=nextTimeCase.getBoundingClientRect().top+frame.contentWindow.scrollY;
 dialPreview.querySelector('[data-time-mode]').click();
 if(!frame.contentWindow.matchMedia('(prefers-reduced-motion: reduce)').matches){
   const hasMotion=(element,property)=>element.getAnimations().some(animation=>property in animation.effect.getKeyframes()[0]);
   assert(hasMotion(dialPreview,'height')&&hasMotion(dialPreview.querySelector('.md-time-dial'),'scale'),'the Variants preview animates the modal size and clock face');
 }
 await wait(()=>!dialPreview.hasAttribute('data-time-resizing'),'dial-to-input transition');
 const inputField=dialPreview.querySelector('[data-time-input="hour"]').getBoundingClientRect();
 assert(dialPreview.dataset.mode==='input'&&inputField.width===dialField.width&&inputField.height===dialField.height,'switching from dial to input preserves field dimensions');
 assert(nextTimeCase.getBoundingClientRect().top+frame.contentWindow.scrollY===nextTimeCaseTop,'switching the dial preview preserves the following case position');
 frame.style.removeProperty('width');
 for(const mode of ['interactive','states']){
   await load('/components/time-pickers?preview.variant=scroll&preview.compare='+mode);
   await frame.contentWindow.TokenAssets.ready;
   await selectPreviewOption('placement','inline');
   const inline=frame.contentDocument.querySelector('#component-preview [data-time-inline][data-mode="scroll"]');
   assert(inline?.getAttribute('role')==='group'&&!inline.hasAttribute('data-time-inline-container'),'Time Picker renders inline without a container in '+mode);
   await selectPreviewOption('container','filled');
   const contained=frame.contentDocument.querySelector('#component-preview [data-time-inline-container][data-mode="scroll"]');
   assert(contained?.getAttribute('role')==='group'&&frame.contentWindow.getComputedStyle(contained).backgroundColor!=='rgba(0, 0, 0, 0)','the container option fills the inline surface in '+mode);
 }
 await load('/components/scroll-fields?preview.compare=states');
 await frame.contentWindow.TokenAssets.ready;
 await selectPreviewOption('content','four-digits');
 assert(!!frame.contentDocument.querySelector('#component-preview [role="spinbutton"][aria-label="Light level"]'),'Scroll Field renders the selected values');
 await load('/components/scroll-fields?preview.compare=interactive&preview.scene=room');
 await frame.contentWindow.TokenAssets.ready;
 assert(!frame.contentDocument.querySelector('#component-preview [data-scroll-toggle]'),'Scroll Field starts without the optional mode button');
 const wheelWindow=frame.contentWindow;
 const wheelField=wheelWindow.document.querySelector('#component-preview [data-scroll-field]');
 const wheelTrack=wheelField.querySelector('.md-scroll-track');
 const selectedRow=wheelTrack.querySelector('[data-scroll-current]');
 const center=element=>{const rect=element.getBoundingClientRect();return rect.top+rect.height/2;};
 const initialIndex=Number(wheelField.dataset.scrollIndex);
 const initialCenter=center(selectedRow);
 for(let n=0;n<3;n++)wheelTrack.dispatchEvent(new wheelWindow.WheelEvent('wheel',{deltaY:100,bubbles:true,cancelable:true}));
 if(!wheelWindow.matchMedia('(prefers-reduced-motion: reduce)').matches){
   assert(Number(wheelField.dataset.scrollIndex)===initialIndex&&center(selectedRow)===initialCenter,'rapid wheel input starts from the painted row');
 }
 await wait(()=>Number(wheelField.dataset.scrollIndex)===initialIndex+3&&
   Math.abs(center(selectedRow)-center(wheelTrack))<.1,'wheel settles in the center');
 assert(wheelTrack.querySelector('[role=spinbutton]').getAttribute('aria-valuetext')===selectedRow.textContent,'wheel keeps the emphasized row and accessible value in sync');
 await selectPreviewOption('direct-entry','on');
 await frame.contentWindow.TokenAssets.ready;
 const scrollScene=frame.contentDocument.querySelector('#component-preview .tokens-scroll-preview');
 const modeButtons=scrollScene.querySelectorAll('[data-scroll-toggle]');
 const rotors=scrollScene.querySelectorAll('[role="spinbutton"]');
 assert(modeButtons.length===1&&!modeButtons[0].closest('[data-scroll-field]')&&rotors.length===2,'the pair uses one host-placed mode button');
 rotors[1].focus();
 modeButtons[0].click();
 const inputs=Array.from(scrollScene.querySelectorAll('.md-scroll-entry'));
 assert(inputs.every(input=>!input.hidden)&&frame.contentDocument.activeElement===inputs[1],'shared text entry focuses the last used field and exposes both inputs');

 frame.style.width='1440px';
 await load('/components/buttons?q=icon&preview.compare=interactive');
 await frame.contentWindow.TokenAssets.ready;
 const filterDocument=frame.contentDocument;
 const filteredSearch=filterDocument.querySelector('#tokens-token-search');
 assert(filteredSearch.getBoundingClientRect().width>0,'the token toolbar leaves room for its search input');
 const filterPreview=filterDocument.querySelector('#component-preview');
 filterDocument.querySelector('#preview-compare-states').click();
 await wait(()=>!filterPreview.isConnected&&new URL(frame.contentWindow.location.href).searchParams.get('preview.compare')==='states','playground configuration before clearing the token search');
 filterDocument.querySelector('.tokens-token-search [data-search-clear]').click();
 await wait(()=>!filteredSearch.isConnected&&!new URL(frame.contentWindow.location.href).searchParams.get('q'),'clearing token search');
 assert(filterDocument.activeElement===filterDocument.querySelector('#tokens-token-search'),'clearing token search keeps input focus through the table swap');
 assert(new URL(frame.contentWindow.location.href).searchParams.get('preview.compare')==='states','clearing token search preserves the latest playground settings');
 frame.style.removeProperty('width');

 await load('/components/buttons?preview.compare=interactive');
 let w=frame.contentWindow,d=w.document;
 await wait(()=>d.querySelector('.tokens-token-filters')&&d.querySelector('[data-export-open]')&&!d.querySelector('[data-app-theme]').disabled,'page initialization');
 const header=d.querySelector('#tokens-site-header'),drawer=d.querySelector('#tokens-site-menu');let preview=d.querySelector('#component-preview');
 assert(!!preview,'playground exists before filtering');
 d.querySelector('#preview-compare-states').click();await wait(()=>d.querySelector('#component-preview')!==preview,'playground configuration');preview=d.querySelector('#component-preview');
 const originalFetch=w.fetch.bind(w),requests=[];
 w.fetch=async(url,options)=>{requests.push({url:String(url),options});return originalFetch(url,options);};
 const search=d.querySelector('.tokens-token-search input');search.value='icon';d.querySelector('.tokens-token-filters').requestSubmit();
 await wait(()=>d.querySelector('.tokens-token-search input')!==search,'table swap');
 assert(d.querySelector('#component-preview')===preview&&d.querySelector('#tokens-site-header')===header,'table filtering preserves playground and shell identity');
 assert(requests.at(-1).options.headers['HX-Target'].includes('tokens-token-browser'),'table request targets its fragment');
 assert(w.location.search.includes('preview.compare=states'),'table submission preserves newly changed playground settings');
 const headStyles=()=>Array.from(d.head.querySelectorAll('link[data-component-stylesheet]'));
 const originalStyles=headStyles();
 assert(originalStyles.length>0&&originalStyles.every(link=>link.sheet),'component stylesheets are loaded in the head');
 for(let i=0;i<3;i++){
   const previousForm=d.querySelector('.tokens-token-filters');
   previousForm.requestSubmit();
   await wait(()=>!previousForm.isConnected,'repeated table filter');
   const styles=headStyles();
   assert(styles.length===originalStyles.length&&styles.every((link,index)=>link===originalStyles[index]&&link.sheet)&&
     new Set(styles.map(link=>link.href)).size===styles.length&&!d.body.querySelector('link[data-component-stylesheet]'),
     'table swap '+(i+1)+' keeps the loaded head stylesheets without duplicates');
 }

 d.querySelector('[data-action="context"]').click();
 let dialog=d.querySelector('#tokens-context-picker');
 dialog.querySelector('.tokens-context-advanced').open=true;
 let contextInsertedOpen;
 const observeContextInsertion=event=>{
   if(event.detail.task.target?.id==='tokens-context-fields') contextInsertedOpen=dialog.querySelector('.tokens-context-advanced')?.open;
 };
 d.addEventListener('htmx:before:settle',observeContextInsertion);
 const platform=dialog.querySelector('input[value="Android"]');platform.click();
 await wait(()=>dialog.getAttribute('aria-busy')!=='true'&&dialog.querySelector('input[value="Android"]')!==platform,'context validation');
 assert(!dialog.querySelector('[type="submit"]').disabled&&dialog.querySelector('input[value="Android"]').checked,'context validation replaces choices and enables Save');
 assert(contextInsertedOpen===true,'table context expansion is preserved before settling');
 d.removeEventListener('htmx:before:settle',observeContextInsertion);
 dialog.querySelector('[data-action="cancel"]').click();
 await wait(()=>dialog.querySelector('input[value="Web"]').checked,'cancel restoration');
 assert(dialog.querySelector('input[value="Web"]').checked,'cancel restores unapplied context');
 const start=requests.length;d.querySelector('[data-export-open]').click();dialog=d.querySelector('#tokens-site-download');
 const button=dialog.querySelector('[data-export-submit]');
 const busy=()=>button.getAttribute('aria-busy')==='true';
 const disabled=()=>button.matches(':disabled');
 const ready=()=>!busy()&&!disabled();
 const summary=()=>dialog.querySelector('#tokens-export-summary');
 const changeExportOption=async(selector,label)=>{
   const previous=summary();
   dialog.querySelector(selector).click();
   await wait(()=>!previous.isConnected&&!busy(),label);
 };
 await wait(ready,'download selection');
 assert(requests.length===start+1&&requests.at(-1).url.includes('/downloads/export/options'),'opening download uses one HTML request');
 assert(dialog.querySelector('[data-export-submit]')===button,'loading action remains mounted');
 dialog.querySelector('.tokens-export-context').open=true;
 const insertedDisclosures=[];
 const observeInsertion=event=>{
   if(event.detail.task.target?.id==='tokens-export-fields') insertedDisclosures.push(dialog.querySelector('.tokens-export-context').open);
 };
 d.addEventListener('htmx:before:settle',observeInsertion);
 await changeExportOption('input[value="css"]','format selection');
 assert(ready()&&dialog.querySelector('input[value="css"]').checked&&dialog.querySelector('input[value="dtcg"]').checked&&summary().textContent.includes('ZIP'),'multiple formats are reflected in the server summary');
 assert(insertedDisclosures.length===0&&dialog.querySelector('.tokens-export-context').open,'format changes update the summary without replacing or collapsing the options');
 d.removeEventListener('htmx:before:settle',observeInsertion);
 // Hold real responses after they arrive to exercise htmx cancellation deterministically.
 const held=[];
 w.fetch=async(url,options)=>{requests.push({url:String(url),options});const response=await originalFetch(url,options);if(String(url).includes('/downloads/export/summary'))await new Promise(resolve=>held.push({resolve,options}));return response;};
 dialog.querySelector('input[value="source"]').click();await wait(()=>held.length===1,'first held request');
 dialog.querySelector('input[value="css"]').click();await wait(()=>held.length===2,'replacement held request');
 held[0].resolve();await new Promise(r=>setTimeout(r,100));
 assert(held[0].options.signal.aborted&&busy()&&disabled(),'aborted request cannot clear the newer loading state');
 w.fetch=async(url,options)=>{requests.push({url:String(url),options});return originalFetch(url,options);};
 held[1].resolve();await wait(ready,'replacement completion');
 assert(dialog.querySelector('input[value="source"]').checked&&!dialog.querySelector('input[value="css"]').checked,'latest request retains the latest selections');

 await changeExportOption('input[value="implementation"]','implementation selection');
 // Match the reported case: an incomplete selection followed by another chip change.
 insertedDisclosures.length=0;d.addEventListener('htmx:before:settle',observeInsertion);
 await changeExportOption('input[name="context.Theme"][value="Light"]','empty theme');
 await changeExportOption('input[name="context.Contrast"][value="High contrast"]','contrast with empty theme');
 assert(insertedDisclosures.length===2&&insertedDisclosures.every(Boolean),'incomplete selections keep the context expanded throughout replacement');
 assert(disabled()&&!dialog.querySelector('input[name="context.Theme"]:checked')&&dialog.querySelector('input[name="context.Contrast"][value="High contrast"]').checked,'validation preserves an intentionally empty selection and disables download');
 d.removeEventListener('htmx:before:settle',observeInsertion);
 await changeExportOption('input[name="context.Theme"][value="Light"]','restore theme');
 assert(ready(),'restoring the theme enables download');

 await changeExportOption('input[value="tokens"]','return to tokens');
 assert(ready()&&dialog.querySelector('input[value="dtcg"]').checked&&dialog.querySelector('input[value="source"]').checked&&
   !dialog.querySelector('input[value="css"]').checked,'switching content preserves token formats');
 assert(dialog.querySelector('.tokens-export-context').open,'expansion survives field replacement');
 const beforeNative=requests.length;let native=false;
 dialog.querySelector('form').addEventListener('submit',event=>{native=!event.defaultPrevented;event.preventDefault();},{once:true});
 button.click();assert(native&&requests.length===beforeNative,'download submission remains native');
 dialog.querySelector('[data-export-cancel]').click();
 const destinations=d.querySelector('#tokens-site-destinations');
 d.querySelector('#tokens-site-destinations a[href="/styles"]').click();
 await wait(()=>w.location.pathname==='/styles'&&d.querySelector('h1')?.textContent.includes('Styles'),'boosted navigation');
 assert(w.document===d&&d.querySelector('#tokens-site-header')===header&&d.querySelector('#tokens-site-menu')===drawer,'boosted navigation preserves header and drawer instances');
 assert(d.querySelector('#tokens-site-destinations')===destinations&&
   destinations.querySelector('[data-site-section="/styles"][aria-current="page"]')&&
   destinations.querySelector('a[href="/styles"][aria-current="page"]'),'partial navigation updates the active drawer section and destination');
 assert(d.querySelector('meta[name="description"]').content.includes('styles')&&d.querySelectorAll('meta[name="description"]').length===1,'hx-head updates metadata without duplication');
 d.querySelector('#subject-search').value='motion';d.querySelector('#subject-search').dispatchEvent(new w.Event('input',{bubbles:true}));
 await wait(()=>w.location.search.includes('motion'),'index search');
 assert(d.querySelector('#subject-search').getAttribute('hx-sync')==='this:replace'&&d.querySelector('#subject-list').textContent.includes('Motion'),'index search uses replacement synchronization');
 w.history.back();await wait(()=>!w.location.search,'search history');
 assert(d.querySelector('#tokens-site-header')===header,'history navigation preserves shell');
 // A simulated deployment changes the core asset signature in the response.
 w.fetch=async(url,options)=>{const response=await originalFetch(url,options);const html=await response.text();const changed=html.replace(/(src="\/static\/site-ui.js\?v=)[^"]*/, '$1changed-deployment');const result=new w.Response(changed,{status:200,headers:response.headers});Object.defineProperty(result,'url',{value:String(url)});return result;};
 d.querySelector('#tokens-site-destinations a[href="/components"]').click();
 await wait(()=>frame.contentDocument!==d&&frame.contentWindow.location.pathname==='/components','deployed asset reload');
 assert(frame.contentDocument!==d,'changed core assets force a document reload');

 }catch(error){const li=document.createElement('li');li.className='fail';li.textContent='FAIL: '+error.message;results.append(li);}
 frame.style.removeProperty('width');
 document.querySelector('#run').disabled=false;
};
