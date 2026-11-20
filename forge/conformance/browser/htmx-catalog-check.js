const frame=document.querySelector('iframe'),results=document.querySelector('#results');
const wait=async(test,label)=>{for(let i=0;i<240;i++){if(test())return;await new Promise(r=>setTimeout(r,50));}throw Error('Timed out: '+label);};
const assert=(ok,label)=>{if(!ok)throw Error(label);results.insertAdjacentHTML('beforeend','<li>PASS: '+label+'</li>');};
const load=path=>new Promise(resolve=>{frame.onload=resolve;frame.src=path;});
document.querySelector('#run').onclick=async()=>{
 results.textContent='';document.querySelector('#run').disabled=true;
 try{
 const selectPreviewOption=async(field,value)=>{
   const doc=frame.contentDocument, previous=doc.querySelector('#component-preview');
   let swapped=false;
   const afterSwap=event=>{if(event.detail.ctx.target===previous)swapped=true;};
   doc.addEventListener('htmx:after:swap',afterSwap);
   previous.querySelector('#preview-'+field).click();
   previous.querySelector('#preview-'+field+'-options [data-field-value="'+value+'"]').click();
   await wait(()=>swapped&&!previous.isConnected,field+' preview swap');
   doc.removeEventListener('htmx:after:swap',afterSwap);
   assert(new URLSearchParams(frame.contentWindow.location.search).get('preview.'+field)===value,field+' filter updates the preview URL');
 };
 frame.style.width='390px';
 await load('/components/time-pickers?preview.compare=variants');
 await frame.contentWindow.TokenAssets.ready;
 const dialPreview=frame.contentDocument.querySelector('[data-time-preview][data-mode="clock"]');
 const expressivePanels=frame.contentDocument.querySelectorAll('[data-time-preview][data-time-style="expressive"]');
 assert(expressivePanels.length===3&&Array.from(expressivePanels).every(panel=>frame.contentWindow.getComputedStyle(panel).boxShadow==='none'),'Expressive modal variants use the same flat surface');
 assert(!dialPreview.closest('.md-card')&&!frame.contentDocument.querySelector('#component-preview [data-preview-dimension]'),
   'modal Time Picker examples render directly without sizing controls');
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
 await load('/components/time-pickers?preview.variant=scroll&preview.compare=interactive&preview.width=12h');
 await frame.contentWindow.TokenAssets.ready;
 const pickerDocument=frame.contentDocument,pickerWindow=frame.contentWindow;
 pickerDocument.querySelector('[data-time-open]').click();
 const scrollModal=pickerDocument.querySelector('[data-time-panel]');
 await wait(()=>scrollModal.matches(':modal'),'scroll time picker opens');
 const modeButton=scrollModal.querySelector('[data-time-mode]');
 const periodGroup=scrollModal.querySelector('.md-time-period-group');
 const pm=periodGroup.querySelector('[data-time-period="pm"]');
 pm.click();
 const body=scrollModal.querySelector('.md-time-body');
 const before=periodGroup.getBoundingClientRect();
 for(const mode of ['input','scroll']){
   modeButton.click();
   const after=periodGroup.getBoundingClientRect();
   assert(scrollModal.dataset.mode===mode&&periodGroup===scrollModal.querySelector('.md-time-period-group')&&
     pm.getAttribute('aria-checked')==='true'&&after.x===before.x&&after.y===before.y&&
     pickerWindow.getComputedStyle(body).opacity==='1',
     'scroll modal keeps AM/PM in place and opaque when switching to '+mode);
 }
 scrollModal.querySelector('[data-time-cancel]').click();
 for(const mode of ['interactive','states']){
   await load('/components/time-pickers?preview.variant=scroll&preview.compare='+mode);
   await frame.contentWindow.TokenAssets.ready;
   await selectPreviewOption('placement','inline');
   const inline=frame.contentDocument.querySelector('#component-preview [data-time-inline][data-mode="scroll"]');
   assert(inline?.getAttribute('role')==='group'&&!inline.hasAttribute('data-time-inline-container'),'Time Picker renders inline without a container in '+mode);
   if(mode==='interactive'){
     const timeWindow=frame.contentWindow;
     await timeWindow.TokenAssets.ready;
     const scene=inline.closest('.tokens-preview-width-scene');
     const hug=inline.getBoundingClientRect().width;
     const available=timeWindow.document.querySelector('#preview-dimension-width-end');
     assert(available.value===available.min&&Number(available.min)===hug,'inline Time Picker starts at its intrinsic minimum width');
     available.value=640;
     available.dispatchEvent(new timeWindow.Event('input',{bubbles:true}));
     assert(inline.dataset.timeWidth==='hug'&&scene.getBoundingClientRect().width>hug,'inline Time Picker hugs its two fields');
     await selectPreviewOption('inline-width','fill');
     await timeWindow.TokenAssets.ready;
     const fill=frame.contentDocument.querySelector('#component-preview [data-time-inline]');
     const filledScene=fill.closest('.tokens-preview-width-scene');
     const tracks=[...fill.querySelectorAll('[data-time-scroll]')];
     assert(fill.dataset.timeWidth==='fill'&&fill.offsetWidth===filledScene.offsetWidth&&fill.offsetWidth>hug&&
       tracks.length===2&&tracks.every(track=>track.dataset.scrollFill==='true'&&track.offsetWidth===track.parentElement.offsetWidth),
       'the inline Time Picker fills the scene and each scroll field fills its column');
   }
   await selectPreviewOption('container','filled');
   const contained=frame.contentDocument.querySelector('#component-preview [data-time-inline-container][data-mode="scroll"]');
   assert(contained?.getAttribute('role')==='group'&&frame.contentWindow.getComputedStyle(contained).backgroundColor!=='rgba(0, 0, 0, 0)','the container option fills the inline surface in '+mode);
 }
 frame.style.width='1440px';
 await load('/components/button-groups?preview.compare=interactive');
 const groupWindow=frame.contentWindow;
 await groupWindow.TokenAssets.ready;
 const groupDocument=groupWindow.document;
 const groupPreview=()=>groupDocument.querySelector('.tokens-group-preview');
 const groupSurface=()=>groupDocument.querySelector('#component-preview .md-button-group-demo .md-button-group');
 const groupDimension=name=>groupDocument.querySelector('#preview-dimension-'+name+'-end');
 const setGroupDimension=(name,value)=>{
   const input=groupDimension(name);
   input.value=value;
   input.dispatchEvent(new groupWindow.Event('input',{bubbles:true}));
 };
 const sizeGuide=()=>groupPreview().offsetWidth;
 assert(groupSurface().dataset.control==='connected'&&groupSurface().dataset.width==='fill','Button Group opens with the connected Fill example');
 const groupWidth=groupSurface().offsetWidth;
 assert(groupWidth===sizeGuide()&&['width','max-width'].every(name=>{
   const input=groupDimension(name);
   return Number(input.min)===groupWidth&&input.value===input.min;
 }),'both Button Group sliders start at the intrinsic minimum width');
 setGroupDimension('max-width',480);
 setGroupDimension('width',400);
 assert(groupSurface().offsetWidth===400&&sizeGuide()===400,'a filling Button Group grows with the available width');
 setGroupDimension('width',640);
 assert(groupSurface().offsetWidth===480&&sizeGuide()===640,'the group stops growing at its maximum width');
 setGroupDimension('max-width',800);
 assert(groupSurface().offsetWidth===640&&groupDocument.querySelector('[data-preview-dimension="max-width"] output').value==='No limit',
   'the maximum width control can remove the cap');
 setGroupDimension('max-width',480);
 await selectPreviewOption('width','hug');
 await groupWindow.TokenAssets.ready;
 assert(groupSurface().dataset.width==='hug'&&Math.abs(groupSurface().offsetWidth-groupWidth)<=1&&
   !groupDimension('width')&&!groupDimension('max-width')&&!groupPreview(),
   'Hug hides both width controls and the width guide');
 let fillingGroupAtSwap;
 const observeGroup=event=>{
   if(event.target.id==='component-preview') fillingGroupAtSwap=[groupDimension('width').value,groupDimension('max-width').value,sizeGuide(),groupSurface().offsetWidth];
 };
 groupDocument.addEventListener('htmx:before:settle',observeGroup);
 await selectPreviewOption('width','fill');
 groupDocument.removeEventListener('htmx:before:settle',observeGroup);
 assert(fillingGroupAtSwap?.join() === '640,480,640,480','returning to Fill restores both width values and the scene during the swap');
 await load('/components/button-groups?preview.compare=sizes');
 await frame.contentWindow.TokenAssets.ready;
 const sizeInputs=[...frame.contentDocument.querySelectorAll('[data-preview-dimension] input')];
 assert(sizeInputs.length===2&&sizeInputs.every(input=>input.value===input.min&&Number(input.min)<Number(input.max)),
   'the sizes comparison starts at its widest group minimum with room to adjust both sliders');
 await load('/components/button-groups');
 const variantsWindow=frame.contentWindow,variantsDocument=frame.contentDocument;
 await variantsWindow.TokenAssets.ready;
 const variantPreview=()=>variantsDocument.getElementById('component-preview');
 await wait(()=>!variantPreview().hasAttribute('data-preview-sizing-pending'),'initial Button Group sizing');
 const originalPreview=variantPreview();
 originalPreview.setAttribute('data-preview-sizing-pending','');
 const sizingContent=[...originalPreview.querySelectorAll('.tokens-preview-dimensions,[data-playground-specimens]')];
 const paintsBeforeSizing=sizingContent.every(element=>variantsWindow.getComputedStyle(element).visibility==='visible');
 originalPreview.removeAttribute('data-preview-sizing-pending');
 assert(paintsBeforeSizing,'the sizing controls and specimens render before minimum-width measurement');
 const variantInput=name=>variantPreview().querySelector('#preview-dimension-'+name+'-end');
 const originalMinimum=variantInput('width').min;
 const navLink=variantsDocument.querySelector('#tokens-site-destinations a[href$="/components/button-groups"]');
 let drawerSwapFinished=false;
 const afterDrawerSwap=event=>{
   if(event.detail.ctx.request.headers['HX-Boosted']==='true')drawerSwapFinished=true;
 };
 variantsDocument.addEventListener('htmx:after:swap',afterDrawerSwap);
 navLink.click();
 await wait(()=>drawerSwapFinished&&!originalPreview.isConnected,'Button Group drawer swap');
 variantsDocument.removeEventListener('htmx:after:swap',afterDrawerSwap);
 await variantsWindow.TokenAssets.ready;
 await wait(()=>!variantPreview().hasAttribute('data-preview-sizing-pending'),'Button Group sizing after drawer swap');
 const maxInput=variantInput('max-width');
 assert(Number(originalMinimum)>240&&['width','max-width'].every(name=>{
   const input=variantInput(name);
   return input.min===originalMinimum&&input.value===input.min&&
     input.closest('.md-slider').style.getPropertyValue('--_slider-end')==='0'&&
     variantPreview().style.getPropertyValue('--_preview-dimension-'+name)===originalMinimum+'px';
 }),'the settled drawer swap preserves Button Group minima, thumbs and scene widths');
 maxInput.value=Number(originalMinimum)+80;
 maxInput.dispatchEvent(new variantsWindow.Event('input',{bubbles:true}));
 assert([...variantPreview().querySelectorAll('.tokens-group-preview')].every(scene=>{
   const group=scene.querySelector('.md-button-group'), figure=scene.closest('figure');
   return scene.offsetWidth===Number(originalMinimum)&&group.offsetWidth===scene.offsetWidth&&figure.scrollWidth===figure.clientWidth;
 }),'raising the cap alone keeps the settled groups inside their available width');
 await load('/components/scroll-fields?preview.compare=sizes&preview.scene=standalone');
 const dimensionsWindow=frame.contentWindow;
 await dimensionsWindow.TokenAssets.ready;
 const dimensionsDocument=dimensionsWindow.document;
 const widthControl=dimensionsDocument.querySelector('[data-preview-dimension="width"]');
 const widthInput=widthControl.querySelector('input');
 const setWidth=value=>{widthInput.value=value;widthInput.dispatchEvent(new dimensionsWindow.Event('input',{bubbles:true}));};
 const scrollScenes=[...dimensionsDocument.querySelectorAll('.tokens-scroll-preview')];
 const sceneRows=()=>new Set(scrollScenes.map(scene=>scene.offsetTop)).size;
 const controls=dimensionsDocument.querySelector('.tokens-playground-controls');
 const headings=[...controls.querySelectorAll(':scope > .tokens-playground-control-section > h3')];
 const [showSection,apiCard,environmentCard]=headings.map(heading=>heading.parentElement);
 assert(controls.matches('div:not(.md-card)')&&headings.map(heading=>heading.textContent).join('|')==='Show|Component API|Environment'&&
   headings.every(heading=>heading.classList.contains('md-typescale-title-small')&&heading.parentElement.getAttribute('aria-labelledby')===heading.id)&&
   showSection.matches('section:not(.md-card)')&&[apiCard,environmentCard].every(card=>card.matches('article.md-outlined-card')),
   'Show is unboxed and API and Environment use outlined cards');
 const [heightControl,gapControl]=apiCard.querySelectorAll('[data-preview-dimension]');
 const sameRow=(a,b)=>a.getBoundingClientRect().top===b.getBoundingClientRect().top;
 const tracksFill=(a,b)=>Math.abs(a.offsetWidth-b.offsetWidth)<=1;
 const fillsTrack=element=>Math.abs(element.getBoundingClientRect().width-element.parentElement.clientWidth)<1;
 assert(!sameRow(apiCard,environmentCard)&&[apiCard,environmentCard].every(card=>tracksFill(card,showSection))&&
   sameRow(heightControl,gapControl)&&tracksFill(heightControl,gapControl)&&widthControl.offsetWidth<widthControl.parentElement.offsetWidth,
   'the cards stay full width while wide slider grids use equal columns');
 const dimensionInputs=[...dimensionsDocument.querySelectorAll('[data-preview-dimension] input')];
 assert(dimensionInputs.every(input=>input.value===input.min)&&scrollScenes.every(scene=>scene.offsetWidth===Number(widthInput.min)),
   'the Scroll Field sizing controls and scene start at their minimums');
 setWidth(280);
 assert(sceneRows()===1&&scrollScenes.every(scene=>scene.offsetWidth===280),'narrow Scroll Field size examples fit in one row');
 setWidth(800);
 assert(sceneRows()===scrollScenes.length&&scrollScenes.every(scene=>scene.offsetWidth===800),'wider Scroll Field size examples wrap at the requested width');
 const dimensionSliders=[...dimensionsDocument.querySelectorAll('[data-preview-dimension] .md-slider')];
 assert(dimensionSliders.length===3&&dimensionSliders.every(slider=>slider.dataset.valueDisplay==='hidden'&&
   slider.dataset.size==='small'&&!slider.querySelector('[data-slider-inset], .md-slider-tick'))&&
   widthControl.querySelector('output').value==='800 px',
   'the dimensions show values beside small sliders');
 const comparison=dimensionsDocument.querySelector('#component-preview [data-preview-compare]');
 const comparisonAligns=group=>{
   const bounds=group.parentElement.getBoundingClientRect();
   const heading=group.closest('section').querySelector('h3').getBoundingClientRect();
   return Math.abs(group.firstElementChild.getBoundingClientRect().left-heading.left)<1&&
     Math.abs(group.lastElementChild.getBoundingClientRect().right-bounds.right)<1;
 };
 assert(comparison.dataset.width==='fill'&&comparisonAligns(comparison),
   'the playground comparison buttons fill the controls surface edge to edge');
 frame.style.width='390px';
 await wait(()=>dimensionsWindow.innerWidth===390,'compact controls viewport');
 await new Promise(resolve=>dimensionsWindow.requestAnimationFrame(()=>dimensionsWindow.requestAnimationFrame(resolve)));
 assert(!sameRow(apiCard,environmentCard)&&[apiCard,environmentCard].every(card=>tracksFill(card,showSection))&&
   !sameRow(heightControl,gapControl)&&fillsTrack(heightControl)&&fillsTrack(gapControl)&&fillsTrack(widthControl),
   'sliders wrap to full-width rows in the full-width cards on compact screens');
 frame.style.removeProperty('width');
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
 const wheelDocument=wheelWindow.document;
 const gapInput=()=>wheelDocument.querySelector('#preview-dimension-gap-end');
 gapInput().value=12;
 gapInput().dispatchEvent(new wheelWindow.Event('input',{bubbles:true}));
 let gapAtSwap;
 const observeGap=event=>{
   if(event.target.id!=='component-preview')return;
   const input=gapInput();
   gapAtSwap=[input.value,input.closest('[data-preview-dimension]').querySelector('output').value,
     event.target.style.getPropertyValue('--_preview-dimension-gap'),
     Number(wheelWindow.getComputedStyle(input.closest('.md-slider')).getPropertyValue('--_slider-end'))];
 };
 wheelDocument.addEventListener('htmx:before:settle',observeGap);
 await selectPreviewOption('entry','input');
 wheelDocument.removeEventListener('htmx:before:settle',observeGap);
 assert(gapAtSwap?.join() === '12,12 px,12px,0.375','the replacement slider has its saved thumb, readout and preview spacing during the swap');

 await load('/components/scroll-fields?preview.compare=interactive&preview.content=text&preview.direct-entry=on&preview.entry=input');
 const autocompleteWindow=frame.contentWindow;
 await autocompleteWindow.TokenAssets.ready;
 const textEntry=autocompleteWindow.document.querySelector('#component-preview .md-scroll-entry[aria-label="Climate mode"]');
 textEntry.focus();
 textEntry.value='';
 const editEntry=(text,inputType)=>{
   textEntry.setRangeText(text,textEntry.selectionStart,textEntry.selectionEnd,'end');
   textEntry.dispatchEvent(new autocompleteWindow.InputEvent('input',{bubbles:true,inputType,data:text}));
 };
 const suffixes=[];
 for(const key of 'Hea'){
   editEntry(key,'insertText');
   suffixes.push(textEntry.value==='Heat'&&textEntry.value.slice(textEntry.selectionStart,textEntry.selectionEnd));
 }
 assert(suffixes.join(',')==='eat,at,t','Scroll Field keeps the remaining suggestion selected as the prefix grows');
 editEntry('','deleteContentBackward');
 assert(textEntry.value==='Hea'&&textEntry.selectionStart===textEntry.selectionEnd,'deleting the suggestion leaves the typed prefix intact');

 frame.style.width='390px';
 await load('/components/buttons?preview.compare=interactive');
 await frame.contentWindow.TokenAssets.ready;
 const compactShow=frame.contentDocument.querySelector('#component-preview [data-preview-compare]');
 const showRight=compactShow.getBoundingClientRect().right;
 const configurationHeading=frame.contentDocument.querySelector('#preview-configuration-heading');
 assert(configurationHeading.textContent==='Configuration'&&
   configurationHeading.parentElement.matches('.md-outlined-card[aria-labelledby="preview-configuration-heading"]')&&
   configurationHeading.closest('.tokens-playground-controls').matches('div:not(.md-card)'),
   'a playground with only Configuration uses one outlined card');
 assert(comparisonAligns(compactShow)&&compactShow.querySelectorAll('button').length===4&&[...compactShow.querySelectorAll('button > span:last-child')].every(label=>
   label.getBoundingClientRect().right<=showRight),'the four compact Show buttons stay edge aligned and their labels fit');

 frame.style.width='1440px';
 await load('/components/search?preview.compare=interactive');
 await frame.contentWindow.TokenAssets.ready;
 const searchPreviewWindow=frame.contentWindow,searchPreviewDocument=frame.contentDocument;
 const searchPreview=()=>searchPreviewDocument.getElementById('component-preview');
 const compositionWidth=()=>searchPreview().querySelector('#preview-dimension-composition-width-end');
 const compositionSlider=compositionWidth();
 const compositionCard=compositionSlider.closest('.md-outlined-card');
 const apiCardForSearch=searchPreview().querySelector('[aria-labelledby="preview-api-heading"]');
 assert(compositionCard.matches('[aria-labelledby="preview-environment-heading"]')&&
   compositionCard.closest('form')&&!searchPreview().querySelector('[data-playground-specimens] [data-preview-dimension]')&&
   !!apiCardForSearch.querySelector('#preview-control')&&!apiCardForSearch.querySelector('#preview-shadow')&&
   !!compositionCard.querySelector('#preview-disclosure')&&!!compositionCard.querySelector('#preview-content'),
   'Search puts component options and scene controls in their respective cards above the specimens');
 compositionSlider.value=560;
 compositionSlider.dispatchEvent(new searchPreviewWindow.Event('input',{bubbles:true}));
 const compositionSliderCell=compositionSlider.closest('.tokens-preview-dimension');
 assert(compositionSliderCell.offsetWidth<compositionSliderCell.parentElement.offsetWidth&&
   compositionSlider.closest('.md-slider').dataset.size==='small'&&compositionCard.querySelector('output').value==='560 px'&&
   searchPreview().querySelector('[data-search-viewport]').offsetWidth===560,
   'the half-width Search slider sizes the adaptive composition');
 let rangeRequests=0;
 const countRangeRequest=()=>rangeRequests++;
 searchPreviewDocument.addEventListener('htmx:before:request',countRangeRequest);
 compositionSlider.dispatchEvent(new searchPreviewWindow.Event('change',{bubbles:true}));
 searchPreviewDocument.removeEventListener('htmx:before:request',countRangeRequest);
 assert(rangeRequests===0,'changing Search composition width stays local');
 await selectPreviewOption('content','text');
 await searchPreviewWindow.TokenAssets.ready;
 assert(compositionWidth().value==='560'&&searchPreview().querySelector('[data-search-viewport]').offsetWidth===560,
   'changing a Search environment option preserves the selected composition width');
 await selectPreviewOption('variant','docked');
 assert(!compositionWidth(),'Search hides composition width when no adaptive example is shown');
 await selectPreviewOption('variant','adaptive');
 assert(compositionWidth().value==='560','returning to adaptive Search restores the composition width');

 await load('/components/buttons?q=icon&preview.compare=interactive');
 await frame.contentWindow.TokenAssets.ready;
 const filterDocument=frame.contentDocument;
 const filteredSearch=filterDocument.querySelector('#tokens-token-search');
 assert(filteredSearch.getBoundingClientRect().width>0,'the token toolbar leaves room for its search input');
 filteredSearch.closest('.md-search-bar').click();
 assert(filterDocument.activeElement===filteredSearch,'the inline Search Bar surface focuses the input');
 const inlineSearchBar=filteredSearch.closest('.md-search-bar');
 const inlineSearchStyle=frame.contentWindow.getComputedStyle(inlineSearchBar);
 assert(inlineSearchStyle.outlineStyle==='solid'&&Number.parseFloat(inlineSearchStyle.outlineWidth)>0&&
   Number.parseFloat(inlineSearchStyle.outlineOffset)===-Number.parseFloat(inlineSearchStyle.outlineWidth),
   'the inline Search Bar draws its focus indicator along the surface edge');
 const filterPreview=filterDocument.querySelector('#component-preview');
 filterDocument.querySelector('#preview-compare-states').click();
 await wait(()=>!filterPreview.isConnected&&new URL(frame.contentWindow.location.href).searchParams.get('preview.compare')==='states','playground configuration before clearing the token search');
 filterDocument.querySelector('.tokens-token-search [data-search-clear]').click();
 await wait(()=>!new URL(frame.contentWindow.location.href).searchParams.get('q'),'clearing token search');
 assert(filteredSearch.isConnected&&filterDocument.activeElement===filteredSearch,'clearing token search keeps the same input focused');
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
 const tokenBrowser=d.querySelector('#tokens-token-browser');
 const search=d.querySelector('.tokens-token-search input');
 const rows=Array.from(tokenBrowser.querySelectorAll('[data-token-search-fields]'));
 const filterTokens=value=>{search.value=value;search.dispatchEvent(new w.Event('input',{bubbles:true}));};
 filterTokens('icon');
 const matches=rows.filter(row=>JSON.parse(row.dataset.tokenSearchFields).some(field=>field.includes('icon')));
 assert(matches.length>0&&matches.length<rows.length&&rows.every(row=>row.hidden===!matches.includes(row)),
   'token search shows matching rows from the current set as you type');
 d.querySelector('.tokens-token-filters').requestSubmit();
 assert(tokenBrowser.isConnected&&search.isConnected&&requests.length===0,'searching and submitting the token query stays local');
 assert(d.querySelector('#component-preview')===preview&&d.querySelector('#tokens-site-header')===header,'table filtering preserves playground and shell identity');
 assert(w.location.search.includes('preview.compare=states'),'table submission preserves newly changed playground settings');
 filterTokens('no-token-matches-this-query');
 assert(rows.every(row=>row.hidden)&&!tokenBrowser.querySelector('[data-token-search-empty]').hidden,
   'token search shows the empty state when all rows are filtered out');
 filterTokens('icon');
 w.history.back();
 await wait(()=>search.value===''&&rows.every(row=>!row.hidden),'restore unfiltered token table');
 w.history.forward();
 await wait(()=>search.value==='icon'&&rows.some(row=>row.hidden),'restore token search');
 assert(tokenBrowser.isConnected&&requests.length===0,'token search history restores results without a table request');
 const headStyles=()=>Array.from(d.head.querySelectorAll('link[data-component-stylesheet]'));
 const originalStyles=headStyles();
 assert(originalStyles.length>0&&originalStyles.every(link=>link.sheet),'component stylesheets are loaded in the head');
 for(let i=0;i<3;i++){
   const previousForm=d.querySelector('.tokens-token-filters');
   previousForm.querySelector('[data-token-set-link]:not([aria-current="page"])').click();
   await wait(()=>!previousForm.isConnected,'repeated token set change');
   assert(requests.at(-1).options.headers['HX-Target'].includes('tokens-token-browser')&&
     new URL(w.location.href).searchParams.get('q')==='icon','changing token set requests its fragment with the current query');
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
 const searchBounds=d.querySelector('.tokens-search-form .md-search-bar').getBoundingClientRect();
 const resultsBounds=d.querySelector('#subject-list').getBoundingClientRect();
 assert(searchBounds.left===resultsBounds.left&&searchBounds.right===resultsBounds.right,'the index Search Bar spans the results width');
 const styleSearch=d.querySelector('#subject-search');
 const stylesList=d.querySelector('#subject-list');
 const styleCards=[...stylesList.querySelectorAll('[data-subject-slug]')];
 const motionCard=styleCards.find(card=>card.dataset.subjectName==='Motion');
 const motionImage=motionCard.querySelector('img');
 let styleSearchRequests=0;
 const trackStyleSearch=()=>styleSearchRequests++;
 d.addEventListener('htmx:before:request',trackStyleSearch);
 styleSearch.value='motion';
 styleSearch.dispatchEvent(new w.Event('input',{bubbles:true}));
 assert(w.location.search.includes('motion')&&styleCards.every(card=>card.hidden===(card!==motionCard))&&
   motionCard.querySelector('img')===motionImage&&styleSearchRequests===0,
   'index search filters mounted cards and thumbnails without an HTMX request');
 w.history.back();
 await wait(()=>!w.location.search&&styleCards.every(card=>!card.hidden),'search history');
 w.history.forward();
 await wait(()=>w.location.search.includes('motion')&&styleCards.some(card=>card.hidden),'search history forward');
 w.history.back();
 await wait(()=>!w.location.search&&styleCards.every(card=>!card.hidden),'return to the unfiltered index');
 d.removeEventListener('htmx:before:request',trackStyleSearch);
 assert(d.querySelector('#tokens-site-header')===header&&d.querySelector('#subject-list')===stylesList&&styleSearchRequests===0,
   'index search history restores the filter without replacing the shell or gallery');
 // A simulated deployment changes the core asset signature in the response.
 w.fetch=async(url,options)=>{const response=await originalFetch(url,options);const html=await response.text();const changed=html.replace(/(src="\/static\/site-ui.js\?v=)[^"]*/, '$1changed-deployment');const result=new w.Response(changed,{status:200,headers:response.headers});Object.defineProperty(result,'url',{value:String(url)});return result;};
 d.querySelector('#tokens-site-destinations a[href="/components"]').click();
 await wait(()=>frame.contentDocument!==d&&frame.contentWindow.location.pathname==='/components','deployed asset reload');
 assert(frame.contentDocument!==d,'changed core assets force a document reload');

 await load('/components/buttons');
 await frame.contentWindow.TokenAssets.ready;
 const navWindow=frame.contentWindow,navDocument=navWindow.document;
 const navDestinations=navDocument.querySelector('#tokens-site-destinations');
 const componentGroups=navDestinations.querySelectorAll('[data-site-panel="/components"] [data-site-group]');
 assert(Array.from(componentGroups,group=>group.dataset.siteGroup).join('|')==='Buttons|Date & time pickers|Loading & progress|Navigation|Sheets',
   'the drawer organizes related components into collapsible groups');
 const [buttonGroup,pickerGroup]=componentGroups;
 assert(buttonGroup.open,'the group containing the current component starts expanded');
 const pickerLink=pickerGroup.querySelector('a[href="/components/time-pickers"]');
 pickerGroup.querySelector('summary').click();
 assert(pickerGroup.open,'the group summary reveals its links');
 const controlClick=new navWindow.MouseEvent('click',{bubbles:true,cancelable:true,ctrlKey:true});
 pickerLink.dispatchEvent(controlClick);
 assert(!controlClick.defaultPrevented&&pickerLink.getAttribute('aria-current')!=='page',
   'Ctrl+click preserves native link behavior and the original drawer selection');
 await wait(()=>[buttonGroup,pickerGroup].every(group=>{
   const summary=group.querySelector('summary');
   return group.offsetHeight===summary.offsetHeight+group.querySelector('.tokens-site-group-items').offsetHeight&&
     !summary.getAnimations({subtree:true}).length;
 }),'expanded group layout');
 const navHeader=navDestinations.closest('#tokens-site-header');
 const scroller=buttonGroup.closest('[data-site-scroll]');
 scroller.scrollTop=240;
 const initialScroll=scroller.scrollTop;
 let headerMoved=false;
 const watchHeader=records=>{headerMoved ||= records.some(record=>Array.from(record.removedNodes).includes(navHeader));};
 const shellObserver=new navWindow.MutationObserver(watchHeader);
 shellObserver.observe(navDocument.body,{childList:true});
 const drawerGeometry=()=>[buttonGroup,pickerGroup].map(group=>{
   const child=group.querySelector('a').getBoundingClientRect();
   return [group.open,group.offsetHeight,child.left,child.top];
 });
 const initialGeometry=JSON.stringify(drawerGeometry());
 let settlingGeometry,settlingScroll;
 const observeDrawer=event=>{
   if(event.target!==navDestinations)return;
   settlingGeometry=JSON.stringify(drawerGeometry());
   settlingScroll=scroller.scrollTop;
 };
 navDocument.addEventListener('htmx:before:settle',observeDrawer);
 pickerLink.click();
 await wait(()=>navWindow.location.pathname==='/components/time-pickers'&&settlingScroll!==undefined,'grouped component navigation');
 navDocument.removeEventListener('htmx:before:settle',observeDrawer);
 watchHeader(shellObserver.takeRecords());
 shellObserver.disconnect();
 assert(!headerMoved&&initialScroll>0&&settlingScroll===initialScroll&&scroller.scrollTop===initialScroll,
   'boosted navigation keeps the drawer mounted at the same scroll position');
 assert(settlingGeometry===initialGeometry,'open drawer groups keep their geometry during the HTMX swap');
 assert(frame.contentDocument===navDocument&&navDocument.querySelector('#tokens-site-destinations')===navDestinations&&
   navDestinations.querySelector('[data-site-group="Buttons"]')===buttonGroup&&buttonGroup.open&&pickerGroup.open&&
   pickerLink.getAttribute('aria-current')==='page','navigation retains open groups and selects the new page');

 }catch(error){const li=document.createElement('li');li.className='fail';li.textContent='FAIL: '+error.message;results.append(li);}
 frame.style.removeProperty('width');
 document.querySelector('#run').disabled=false;
};
