import { readFileSync } from "node:fs";

const contract = JSON.parse(
  readFileSync(new URL("./component-contract.json", import.meta.url)),
);
const extension = "dev.tokencatalog.material";
const selection =
  ':where(:checked,[aria-pressed="true"],[aria-selected="true"],[aria-current="page"],[data-selected="true"],:has(> input:checked))';
const enabled =
  ':not(:disabled):not([aria-disabled="true"]):not(:has(> input:disabled, > select:disabled))';
const stateNames = {
  hover: "hovered",
  hovered: "hovered",
  focus: "focused",
  focused: "focused",
  pressed: "pressed",
  disabled: "disabled",
  dragged: "dragged",
};
const states = {
  hovered: `:where(:hover,[data-component-state="hovered"])${enabled}:not(:where([data-touch]))`,
  focused: `:where(:focus-visible,:has(> input:focus-visible),[data-component-state="focused"])${enabled}`,
  pressed: `:where(:active,[data-pressed],[data-component-state="pressed"])${enabled}:not(:where([data-button-cancelled]))`,
  dragged: `:where([data-dragged="true"],[data-component-state="dragged"])${enabled}`,
  disabled:
    ':is(:disabled,[aria-disabled="true"],:has(> input:disabled, > select:disabled))',
};
const order = ["", "hovered", "focused", "pressed", "dragged", "disabled"];
const roleProperties = {
  "container.color": "container",
  "container.opacity": "container-opacity",
  "container.height": "height",
  "container.elevation": "elevation",
  "container.shadow-color": "shadow",
  "container.shape": "shape",
  "outline.color": "outline",
  "outline.opacity": "outline-opacity",
  "outline.width": "outline-width",
  "state-layer.color": "state-color",
  "state-layer.opacity": "state-opacity",
  "icon.color": "icon",
  "icon.opacity": "icon-opacity",
  "icon.size": "icon-size",
  "label-text.color": "label",
  "label-text.opacity": "label-opacity",
  "label-text.font": "font-family",
  "label-text.size": "font-size",
  "label-text.line-height": "line-height",
  "label-text.tracking": "tracking",
  "label-text.weight": "font-weight",
  "focus.indicator.color": "focus-color",
  "focus.indicator.outline.offset": "focus-offset",
  "focus.indicator.thickness": "focus-width",
  "focus.indicator.outline.color": "focus-color",
  "focus.indicator.outline.thickness": "focus-width",
};
const textProperty = {
  font: "font-family",
  size: "font-size",
  "line-height": "line-height",
  tracking: "tracking",
  weight: "font-weight",
  color: "color",
  opacity: "opacity",
};

function parse(suffix) {
  let parts = suffix.split("."),
    state = "",
    selected = "",
    variant = "";
  // Focus indicators describe the focus ring, not a transient focus-state value.
  const indicator = suffix.includes("focus.indicator.");
  parts = parts.filter((p, i) => {
    if (
      Object.hasOwn(stateNames, p) &&
      !(indicator && p === "focus" && parts[i + 1] === "indicator")
    ) {
      state = stateNames[p];
      return false;
    }
    if (p === "selected" || p === "unselected") {
      selected = p;
      return false;
    }
    if (p === "flat" || p === "elevated") {
      variant = p;
      return false;
    }
    return true;
  });
  return { sourceProperty: parts.join("."), state, selected, variant };
}
function condition({ state, selected }) {
  return (
    (selected === "selected"
      ? selection
      : selected === "unselected"
        ? `:not(${selection})`
        : "") + (state ? states[state] : "")
  );
}
function textRole(property, roles) {
  for (const role of roles) {
    const tail = property.slice(role.length + 1);
    if (property.startsWith(role + ".") && Object.hasOwn(textProperty, tail))
      return `${role}-${textProperty[tail]}`;
  }
}
function binding(set, suffix) {
  let parsed = parse(suffix),
    property = parsed.sourceProperty,
    selector = "",
    family = "",
    target = "";
  if(set==='md.comp.time-picker'||set==='md.comp.time-input') {
    if(suffix.endsWith('.type'))return{excluded:'Composite typography duplicates the individual font, size, weight, line-height and tracking fields.'};
    if(suffix==='surface-tint-layer.color')return{excluded:'Deprecated surface tint is replaced by the captured tonal container color.'};
    if(suffix.endsWith('.ignore'))return{excluded:'Deprecated source name; use the canonical '+suffix.slice(0,-7)+' field.'};
    const input=set==='md.comp.time-input',root='.md-time-panel:where([data-mode="'+(input?'input':'clock')+'"])';
    let key=suffix,rank=0;selector=root;family='time';
    if(key.startsWith('clock-dial.')){key=key.slice(11);family='time-dial';}
    else if(key.startsWith('period-selector.')){key=key.slice(16);family='time-period';
      if(key.startsWith('horizontal.')){selector+=':where([data-orientation="landscape"])';key=key.slice(11);rank++;}
      else if(key.startsWith('vertical.')){selector+=':where([data-orientation="portrait"])';key=key.slice(9);rank++;}
    }
    else if(key.startsWith('time-input-field.')){key=key.slice(17);family='time-field';}
    else if(key.startsWith('time-selector.')){key=key.slice(14);family='time-value';if(key.startsWith('24h-vertical.')){selector+=':where([data-format="24h"][data-orientation="portrait"])';key=key.slice(13);rank++;}}
    const parsed=parse(key),property=parsed.sourceProperty;
    if(family==='time-field'&&property.startsWith('supporting-text.'))return{...parsed,family:'time-support',selector:root,property:textRole('text.'+property.slice(16),['text']).replace(/^text-/,''),state:''};
    if(property.startsWith('separator.')){family='time-separator';key=property.slice(10);return{...parsed,family,selector:root,property:textRole('text.'+key,['text']).replace(/^text-/,''),state:''};}
    const roles={...roleProperties,'container.size':'size','container.width':'width','color':'container','shape':'shape','selector.center.container.color':'center-color','selector.center.container.shape':'center-shape','selector.center.container.size':'center-size','selector.handle.container.color':'handle-color','selector.handle.container.shape':'handle-shape','selector.handle.container.size':'handle-size','selector.track.container.color':'track-color','selector.track.container.width':'track-width'};
    let role=roles[property]||textRole(property,['headline','supporting-text']);
    if(family==='time-dial'&&parsed.selected){role=property==='label-text.color'?(parsed.selected==='selected'?'selected-label':'label'):role;parsed.selected='';}
    else if(['time-value','time-period','time-field'].includes(family)&&(parsed.state||parsed.selected||/(?:color|opacity)$/.test(property)&&!property.startsWith('outline.')))selector+=' .md-'+family;
    return{...parsed,family,selector,property:role,rank};
  } else if(set.startsWith('md.comp.date-picker.')||set==='md.comp.date-input.modal') {
    if(suffix.endsWith('.type'))return{excluded:'Composite typography duplicates individual font, size, weight, line-height and tracking fields.'};
    if(suffix==='container.surface-tint-layer.color')return{excluded:'Deprecated surface tint is replaced by the captured tonal container color.'};
    if(suffix.endsWith('.opcaity'))return{excluded:'Misspelled source duplicate; the canonical opacity field is used after checking equality.'};
    let key=suffix,rank=0;family='date';selector='.md-date-panel:where('+(set==='md.comp.date-input.modal'?'[data-mode="input"]':set.endsWith('.docked')?'[data-variant="docked"]':'[data-calendar="modal"]:not([data-mode="input"])')+')';
    if(key.startsWith('range-selection.')){key=key.slice(16);selector+=':where([data-range="true"])';rank=2;}
    if(key.startsWith('year-selection.year.')){key=key.slice(20);family='date-year';selector+=' .md-date-year';}
    else if(key.startsWith('menu.list-item.')){key=key.slice(15);family='date-month';selector+=' .md-date-month-item';}
    else if(key.startsWith('menu-button.')){key=key.slice(12);family='date-menu';selector+=' .md-date-menu-button';}
    else if(key.startsWith('weekdays.')){key=key.slice(9);family='date-weekday';}
    else if(key.startsWith('date.')){key=key.slice(5);family='date-day';selector+=' .md-date-day';
      if(key.startsWith('today.')){key=key.slice(6);selector+=':where([data-today="true"]):not(:where([data-date-selected="true"]))';rank++;}
      else if(key.startsWith('selected.')){key=key.slice(9);selector+=':where([data-date-selected="true"])';rank+=2;}
      else if(key.startsWith('unselected.')){key=key.slice(11);selector+=':not(:where([data-date-selected="true"]))';
        if(key.startsWith('outside-month.')){key=key.slice(14);selector+=':where([data-outside-month="true"])';rank++;}
      }
      else if(key.startsWith('in-range.')){key=key.slice(9);selector+=':where([data-in-range="true"]):not(:where([data-date-selected="true"]))';rank++;}
    }
    key=key.replace(/^header\./,'').replace('container.outline.','outline.');
    const parsed=parse(key),roles={...roleProperties,'container.width':'width','header.height':'navigation-height','height':'navigation-height','container.shadow-color':'shadow','state-layer.height':'state-height','state-layer.width':'state-width','state-layer.shape':'state-shape','label-text.type':'font','leading-icon.color':'icon','leading-icon.size':'icon-size','active-indicator.container.color':'range-color','active-indicator.container.height':'range-height','active-indicator.container.shape':'range-shape','month.subhead.color':'month-color','month.subhead.font':'month-font-family','month.subhead.line-height':'month-line-height','month.subhead.size':'month-font-size','month.subhead.tracking':'month-tracking','month.subhead.weight':'month-font-weight'};
    let property=roles[parsed.sourceProperty]||textRole(parsed.sourceProperty,['headline','supporting-text']);
    if(family==='date-day'&&!parsed.state&&/^(?:container\.(?:width|height|shape)|state-layer\.(?:width|height|shape)|label-text\.(?:font|size|weight|tracking|line-height))$/.test(parsed.sourceProperty)&&selector.endsWith(' .md-date-day'))selector=selector.slice(0,-' .md-date-day'.length);
    if(family==='date'&&suffix.includes('header.container.'))property=suffix.endsWith('width')?'header-width':'header-height';
    return{...parsed,family,selector,property,rank};
  } else if(set==='md.comp.search-bar'||set==='md.comp.search-view') {
    if(suffix.endsWith('.type'))return{excluded:'Composite typography duplicates the individual font, size, weight, line-height and tracking fields.'};
    if(suffix==='container.surface-tint-layer.color')return{excluded:'Deprecated surface tint is replaced by the captured tonal container color.'};
    const view=set.endsWith('view');family=view?'search-view':'search';selector=view?'.md-search-view':'.md-search-bar';
    let key=suffix,rank=0;
    if(key.startsWith('contained.')){selector+=':where([data-control="contained"])';key=key.slice(10);rank=1;}
    for(const variant of ['docked','full-screen'])if(key.startsWith(variant+'.')){selector+=`:where([data-variant="${variant}"])`;key=key.slice(variant.length+1);rank++;}
    const parsed=parse(key);key=parsed.sourceProperty;
    if(view)key=key.replace(/^header\./,'');
    const roles={...roleProperties,'container.shadow-color':'shadow','avatar.shape':'avatar-shape','avatar.size':'avatar-size','avatar.target-size':'avatar-target','leading-margin':'margin-leading','trailing-margin':'margin-trailing','icon-label.gap':'gap','leading-space':'leading','trailing-space':'trailing','no-actions.leading-space':'no-action-leading','no-actions.trailing-space':'no-action-trailing','trailing-actions.gap':'action-gap','trailing-actions.leading-space':'action-leading','trailing-actions.trailing-space':'action-trailing','background.color':'background','bar-results.gap':'results-gap','bar.shape':'header-shape','results.shape':'results-shape','bar.container.height':'height','divider.color':'divider','leading-icon.color':'leading-icon','trailing-icon.color':'trailing-icon','leading-icon.leading-icon-label-space':'gap','trailing-icon.label-trailing-icon-space':'trailing-gap'};
    return{...parsed,family,selector,property:roles[key]||textRole(key,['input-text','supporting-text']),rank};
  } else if(set==='md.comp.carousel-item') {
    if(suffix==='container.surface-tint-layer.color')return{excluded:'Deprecated surface tint is replaced by the captured tonal container color.'};
    const outlined=suffix.startsWith('with-outline.');
    const parsed=parse(suffix.replace(/^with-outline\./,''));
    const roles={...roleProperties,'container.shadow-color':'shadow'};
    return{...parsed,family:'carousel-item',selector:'.md-carousel'+(outlined?':where([data-control="outlined"])':'')+' .md-carousel-item',property:roles[parsed.sourceProperty]};
  } else if(set==='md.comp.list'||set.startsWith('md.comp.list.')) {
    let key=(set==='md.comp.list'?'':set.slice('md.comp.list.'.length)+'.')+suffix;
    if(key.startsWith('md.comp.list.'))return{excluded:'Duplicate source names repeat the one-, two- and three-line minimum heights.'};
    if(key.endsWith('.type'))return{excluded:'Composite typography duplicates its individual font, size, weight, line-height and tracking fields.'};
    if(key==='container.shape'||key==='segmented.gap'||key.startsWith('divider.')||key.startsWith('focus.indicator.')) {
      const roles={'container.shape':'shape','segmented.gap':'gap','divider.color':'divider','divider.height':'divider-width','divider.leading-space':'divider-leading','divider.trailing-space':'divider-trailing','divider.top-space':'divider-top','divider.bottom-space':'divider-bottom',...roleProperties};
      return{family:'list',selector:'.md-list',property:roles[key],state:''};
    }
    family='list-item';selector='.md-list .md-list-item';let rank=0,state='',selected='';
    if(key.startsWith('expand.')){
      selector='.md-list:where([data-control="expand"]) .md-list-item';key=key.slice(7);rank=2;
      if(key==='container.shape')return{family:'list',selector:'.md-list:where([data-control="expand"])',property:'shape',state:'',rank};
      if(key.startsWith('collapsed.')||key.startsWith('expanded.')){const expanded=key.startsWith('expanded.');selector+=`:where([aria-expanded="${expanded}"])`;key=key.slice(expanded?9:10);}
    } else if(key.startsWith('reorder.')){
      key=key.slice(8);rank=5;
      if(key==='list-item.drop-zone.color')return{family:'list',selector:'.md-list:where([data-control="reorder"])',property:'drop-color',state:''};
      selector='.md-list:where([data-control="reorder"]) .md-list-item';state='dragged';
    } else if(key.startsWith('reveal.')){
      key=key.slice(7);rank=2;selector='.md-list:where([data-control="reveal"]) .md-list-row:where([data-revealed="true"]) > .md-list-item';
      const action=key.includes('action.icon-button.')||key.includes('icon-button.action.');
      if(key.includes('icon-button.')){
        key=key.replace(/^list-item\./,'').replace('action.icon-button.','').replace('icon-button.action.','').replace('icon-button.','');
        const parsed=parse(key);return{...parsed,family:'list-reveal',selector:'.md-list '+(action?'.md-list-reveal-action':'.md-list-reveal-toggle'),property:roleProperties[parsed.sourceProperty],rank};
      }
    }
    key=key.replace(/^list-item\./,'');
    const parsed=parse(key);state=state||parsed.state;selected=parsed.selected;
    if(selected)rank+=3;
    let property=parsed.sourceProperty.replace('.icon.color','.color');
    if(property.includes('.expressive.')){selector='.md-list:where([data-shape="expressive"])'+selector.slice('.md-list'.length);property=property.replace('.expressive.','.');rank++;}
    if(property.startsWith('segmented.')){selector=selector.replace('.md-list','.md-list:where([data-variant="segmented"])');property=property.slice(10);rank++;}
    for(const size of ['one-line','two-line','three-line'])if(property.startsWith(size+'.')){selector=selector.replace('.md-list',`.md-list:where([data-size="${size}"])`);property=property.slice(size.length+1);rank++;}
    if(property.startsWith('small.leading-video.')||property.startsWith('large.leading-video.')){
      const large=property.startsWith('large.');selector=selector.replace('.md-list',`.md-list:where(${large?'[data-media="large"]':'[data-media="small"]'})`);property=property.slice(large?6:6);rank++;
    }
    const roles={...roleProperties,'shape':'shape','shadow-color':'shadow','between-space':'gap','leading-space':'leading','trailing-space':'trailing','top-space':'top','bottom-space':'bottom','leading-icon.color':'leading-icon','leading-icon.opacity':'leading-icon-opacity','leading-icon.size':'leading-icon-size','trailing-icon.color':'trailing-icon','trailing-icon.opacity':'trailing-icon-opacity','trailing-icon.size':'trailing-icon-size','leading-avatar.color':'avatar','leading-avatar.size':'avatar-size','leading-avatar.shape':'avatar-shape','leading-image.width':'image-width','leading-image.height':'image-height','leading-image.shape':'image-shape','leading-video.width':'video-width','leading-video.height':'video-height','leading-video.shape':'video-shape','trailing-icon.container.color':'expand-icon-container','trailing-icon.shape':'expand-icon-shape'};
    let role=roles[property]||textRole(property,['overline','supporting-text','trailing-supporting-text','leading-avatar-label']);
    if(property==='trailing-icon.icon.color')role='trailing-icon';
    if(!role)throw new Error(`Unmapped list token: ${set}.${suffix} (${property})`);
    if(state==='pressed'&&role==='shape'){role='pressed-shape';state='';}
    return{family,selector,property:role,state,selected,rank};
  } else if (set==='md.comp.fab-menu' ||set.startsWith('md.comp.fab-menu.')) {
    let key=suffix;
    const common=set==='md.comp.fab-menu';
    const close=common?key.startsWith('close-button.'):set.endsWith('.close-button');
    family=close?'fab-menu-close':'fab-menu-item';
    selector='.md-fab-menu';
    if(!common)selector+=`:where([data-variant="${set.slice('md.comp.fab-menu.'.length).split('.')[0].replace('-container','')}"])`;
    if(common&&key.endsWith('between-space'))return{family:'fab-menu',selector,property:close?'close-gap':'item-gap',state:''};
    selector+=close?' .md-fab-menu-close':' .md-fab-menu-item';
    key=key.replace(/^(?:close-button|menu-item|list-item)\./,'');
    const parsed=parse(key);
    const roles={...roleProperties,'container.shadow-color':'shadow','container.width':'width','leading-space':'leading','trailing-space':'trailing','icon-label-space':'gap','label-text':'font'};
    const role=roles[parsed.sourceProperty];if(!role)throw new Error(`Unmapped FAB menu token: ${set}.${suffix}`);
    return{...parsed,family,selector,property:role};
  } else if (set.startsWith("md.comp.button-group.")) {
    const [control,size]=set.slice("md.comp.button-group.".length).split('.');
    family="group";selector=`.md-button-group:where([data-control="${control}"][data-size="${size}"])`;
    const roles={"between-space":"gap","container.height":"height","container.shape":"outer","inner-corner.corner-size":"inner","pressed.inner-corner.corner-size":"pressed-inner","selected.inner-corner.corner-size":"selected-inner","pressed.item.width.multiplier":"width-growth","pressed.item.width.motion.spring.dampening":"width-damping","pressed.item.width.motion.spring.stiffness":"width-stiffness"};
    const role=roles[suffix];if(!role)throw new Error(`Unmapped button-group token: ${suffix}`);
    return{family,selector,property:role,state:""};
  } else if (set.startsWith("md.comp.split-button.")) {
    const size=set.slice("md.comp.split-button.".length);
    family="split"; selector=`.md-split-button:where([data-size="${size}"])`;
    const roles={"between-space":"gap","container.height":"height","container.shape":"outer","outer-corner.corner-size":"outer","trailing-button.inner-corner.selected.corner-size":"selected-inner","inner-corner.corner-size":"inner","inner-corner.hovered.corner-size":"inner","inner-corner.pressed.corner-size":"pressed-inner","leading-button.leading-space":"leading","leading-button.trailing-space":"trailing","trailing-button.icon.size":"icon-size","trailing-button.leading-space":"leading","trailing-button.trailing-space":"trailing"};
    const role=roles[suffix];if(!role)throw new Error(`Unmapped split-button token: ${suffix}`);
    if(suffix.startsWith("leading-button.")){family="split-item";selector+=' > .md-split-leading';}
    if(suffix.startsWith("trailing-button.")){family="split-item";selector+=' > .md-split-trailing';}
    if(suffix.includes("hovered")){family="split-item";selector+=' > .md-button';return{family,selector,property:role,state:"hovered"};}
    // The pressed target must remain available while the shared spring releases.
    return{family,selector,property:role,state:""};
  } else if (set === "md.comp.sheet.bottom" || set === "md.comp.sheet.side") {
    const bottom=set.endsWith("bottom");family="sheet";selector=bottom?'.md-bottom-sheet':'.md-side-sheet';
    let key=suffix.replace(/^docked\./,""),rank=0;
    if(key.includes("surface-tint-layer.color"))return {excluded:"Surface container colors supply their tonal treatment without an additional tint overlay."};
    if(key.endsWith(".type"))return {excluded:"Composite headline typography duplicates the individual font, size, weight, line-height and tracking tokens."};
    if(key.startsWith("action.")) {
      const parsed=parse(key.slice(7));family="sheet-action";selector+=' .md-sheet-action';
      const role=roleProperties[parsed.sourceProperty];
      if(!role)throw new Error(`Unmapped sheet action: ${suffix}`);
      return {...parsed,family,selector,property:role};
    }
    if(key.startsWith("modal.")||key.startsWith("standard.")){const type=key.split('.')[0];selector+=`:where([data-control="${type}"])`;key=key.slice(type.length+1);rank=1;}
    if(key.startsWith("minimized.")){selector+=':where([data-variant="minimized"])';key=key.slice(10);rank=2;}
    if(key.startsWith("detached.")){selector+=':where([data-variant="detached"])';key=key.slice(9);rank=2;}
    const roles={"container.width":"width","drag-handle.color":"handle","drag-handle.opacity":"handle-opacity","drag-handle.width":"handle-width","drag-handle.height":"handle-height","divider.color":"divider",...roleProperties};
    const role=roles[key]||textRole(key,["headline"]);
    if(!role)throw new Error(`Unmapped sheet token: ${set}.${suffix}`);
    return {family,selector,property:role,state:"",rank};
  } else if (/^md\.comp\.(app-bar|top-app-bar|bottom-app-bar)(\.|$)/.test(set)) {
    family="appbar";
    const modern=set.startsWith("md.comp.app-bar"),bottom=set==="md.comp.bottom-app-bar";
    const part=set.replace(/^md\.comp\.(?:top-app-bar|app-bar)\.?/,"");
    const variant=part==="small.centered"?"baseline-centered":modern?part:`baseline-${part}`;
    selector=bottom?'.md-app-bar:where([data-variant="bottom"])':modern&&part===""?'.md-app-bar:where([data-modern])':modern&&part==="small"?'.md-app-bar:where([data-variant="small"],[data-variant="centered"])':`.md-app-bar:where([data-variant="${variant}"])`;
    let key=suffix;const onScroll=key.includes("on-scroll.");let detail=false;
    key=key.replace("on-scroll.","");
    if(onScroll)selector+=':where([data-scrolled],[data-component-state="scrolled"])';
    if(key.startsWith("with-subtitle.")){selector+=':where([data-control="subtitle"])';key=key.slice(14);detail=true;}
    if(key.startsWith("with-fab.")){selector+=':where([data-control="fab"])';key=key.slice(9);detail=true;}
    if(key.endsWith("surface-tint-layer.color"))return {excluded:"Surface container colors supply their tonal treatment without an additional tint overlay."};
    if(key.endsWith(".type"))return {excluded:"Composite headline typography duplicates the individual font, size, weight, line-height and tracking tokens."};
    const roles={"container.color":"container","container.height":"height","container.shape":"shape","container.elevation":"elevation","container.shadow-color":"shadow","avatar.size":"avatar-size","avatar.shape":"avatar-shape","icon.size":"icon-size","icon-button-space":"button-gap","leading-space":"leading","trailing-space":"trailing","leading-icon.color":"leading-icon","leading-icon.size":"leading-icon-size","trailing-icon.color":"trailing-icon","trailing-icon.size":"trailing-icon-size","title.color":"title-color","title.font":"title-font","subtitle.color":"subtitle-color","subtitle.font":"subtitle-font","search.container.color":"search-container","search.container.height":"search-height","search.container.shape":"search-shape","search.label.color":"search-label","search.label-text.font":"search-font","search.leading-space":"search-leading","search.trailing-space":"search-trailing"};
    const role=roles[key]||textRole(key,["headline"])?.replace(/^headline-/,"title-");
    if(!role)throw new Error(`Unmapped app bar token: ${set}.${suffix}`);
    return {family,selector,property:role,state:"",rank:(modern&&part===""?0:1)+(detail?1:0)+(onScroll?3:0)};
  } else if (set.startsWith("md.comp.toolbar.")) {
    const root='.md-toolbar'; family="toolbar"; let key=suffix,rank=0;
    const subtype=set.slice("md.comp.toolbar.".length);
    if(subtype==="standard"||subtype==="vibrant") {
      selector=root+`:where([data-control="${subtype}"])`;
      if(key==="container.color"||key==="container.shape")return {family,selector,property:roleProperties[key],state:"",rank};
      family="toolbar-item";selector+=' .md-toolbar-item';
      const parsed=parse(key);key=parsed.sourceProperty;
      const role=key==="button.container.color"?"container":roleProperties[key];
      if(!role)throw new Error(`Unmapped toolbar item: ${set}.${suffix}`);
      return {...parsed,family,selector,property:role,rank};
    }
    if(subtype==="floating.fab") {
      family="toolbar-fab";selector=root+':where([data-variant="floating"]) + .md-toolbar-fab';rank=2;
      if(key.startsWith("medium.")){selector+=':where([data-size="medium"])';key=key.slice(7);rank++;}
      if(key.startsWith("standard.")||key.startsWith("vibrant.")){const color=key.split('.')[0];selector+=`:where([data-color="${color}"])`;key=key.slice(color.length+1);}
      const roles={"between-space":"gap","container.width":"width",...roleProperties};
      if(!roles[key])throw new Error(`Unmapped toolbar FAB: ${suffix}`);
      return {family,selector,property:roles[key],state:"",rank};
    }
    if(subtype==="floating" && key==="container.external-padding")return {family:"toolbar-wrapper",selector:'.md-toolbar-demo:has(> .md-toolbar[data-variant^="floating"])',property:"padding",state:"",rank:2};
    selector=root+(subtype==="docked"?':where([data-variant="docked"])':':where([data-variant^="floating"])');rank=2;
    if(key.startsWith("standard.")||key.startsWith("vibrant.")) {
      const color=key.split('.')[0];selector+=`:where([data-control="${color}"])`;key=key.slice(color.length+1);
      if(key.startsWith("button.")) {
        const parsed=parse(key.slice(7));key=parsed.sourceProperty;family="toolbar-item";selector+=' .md-toolbar-item';
        const roles={"container.color":"container","icon.color":"icon","text.color":"label"};
        if(!roles[key])throw new Error(`Unmapped floating toolbar item: ${suffix}`);
        return {...parsed,family,selector,property:roles[key],rank};
      }
    }
    if(subtype==="docked"&&key==="container.color")selector+=':where([data-control="standard"])';
    if(key.startsWith("horizontal.")||key.startsWith("vertical.")){const axis=key.split('.')[0];selector+=`:where([data-orientation="${axis}"])`;key=key.slice(axis.length+1);rank++;}
    const roles={"container.width":"width","container.leading-space":"leading","container.trailing-space":"trailing","container.between-space":"gap","container.max-spacing":"max-gap","container.min-spacing":"min-gap","container.external-padding":"external-padding","container.external-space":"external-space",...roleProperties};
    if(!roles[key])throw new Error(`Unmapped toolbar token: ${set}.${suffix}`);
    return {family,selector,property:roles[key],state:"",rank};
  } else if (set === "md.comp.loading-indicator") {
    if(suffix === "container.color") return {excluded:"AndroidX uses a transparent uncontained indicator and contained.container.color for contained indicators. The separate base container color remains available in token exports."};
    family="loading"; selector=".md-loading";
    let key=suffix;
    if(key.startsWith("contained.")){selector+=':where([data-variant="contained"])';key=key.slice(10);}
    const roles={"active-indicator.color":"color","active-indicator.size":"size","container.color":"container","container.height":"height","container.width":"width","container.shape":"shape"};
    if(!roles[key])throw new Error(`Unmapped loading token: ${set}.${suffix}`);
    return {family,selector,property:roles[key],state:"",selected:"",variant:""};
  } else if (/^md\.comp\.(progress-indicator|linear-progress-indicator|circular-progress-indicator|loading-indicator|app-bar|top-app-bar|bottom-app-bar|toolbar|sheet)(\.|$)/.test(set)) {
    family = "progress";
    const legacy = !set.startsWith("md.comp.progress-indicator");
    const kind = set.includes("circular") ? "circular" : set.includes("linear") ? "linear" : "";
    selector = legacy ? `.md-progress:where([data-variant="legacy-${kind}"])` : kind ? `.md-progress:where([data-variant="${kind}"])` : '.md-progress:where([data-variant="linear"],[data-variant="circular"])';
    // These suffixes describe structural variants, not interaction states.
    let key = suffix;
    if(key.startsWith("thick.")) { selector += ':where([data-size="thick"])'; key = key.slice(6); }
    if(key.startsWith("indeterminate.")) { selector += ':where([data-control="indeterminate"])'; key=key.slice(14); }
    if(key.startsWith("with-wave.")) { selector += ':where([data-shape="wave"])'; key=key.slice(10); }
    const roles = {"active-indicator.color":"color", "active-indicator.shape":"shape", "active-indicator.thickness":"thickness", "active-indicator.width":"thickness", "active-indicator.height":"thickness", "active-indicator.wave.amplitude":"amplitude", "active-indicator.wave.wavelength":"wavelength", "active-indicator-track-space":"gap", "track-active-indicator-space":"gap", "track.color":"track", "track.shape":"track-shape", "track.thickness":"track-thickness", "track.height":"track-thickness", "stop-indicator.color":"stop-color", "stop-indicator.shape":"stop-shape", "stop-indicator.size":"stop-size", "stop-indicator.trailing-space":"stop-trailing", "size":"size", "height":"height"};
    const color = key.match(/^four-color\.active-indicator\.(one|two|three|four)\.color$/);
    const role=roles[key] || (color && `color-${color[1]}`);
    if(!role) throw new Error(`Unmapped progress token: ${set}.${suffix}`);
    return {family,selector,property:role,state:"",selected:"",variant:""};
  } else if (set === "md.comp.slider" || set.startsWith("md.comp.slider.")) {
    family = "slider"; selector = ".md-slider";
    const size = set.slice("md.comp.slider.".length);
    if(set !== "md.comp.slider") selector += `:where([data-size="${size}"])`;
    if (property.endsWith(".type")) return {excluded:"Individual source font metrics supply slider labels."};
    if (property === "label.label-text.font-size") return {excluded:"Duplicate size field; label.label-text.size supplies the label font size."};
    if (property.startsWith("active.handle.")) property = property.slice(7);
    if (/^handle\.(color|width|opacity|elevation|shadow-color)$/.test(property) || /^state-layer\.(color|opacity)$/.test(property)) {
      family="slider-control"; selector+=" .md-slider-input";
    }
    const extra={
      "active.container.opacity":"active-opacity", "inactive.container.opacity":"inactive-opacity",
      "active.track.color":"active-color", "active.track.opacity":"active-opacity", "active.track.height":"active-height", "active.track.shape":"active-shape", "active.track.shape.leading":"active-leading", "active.track.outer-corner.corner-size":"active-outer",
      "inactive.track.color":"inactive-color", "inactive.track.opacity":"inactive-opacity", "inactive.track.height":"inactive-height", "inactive.track.shape":"inactive-shape", "inactive.track.shape.trailing":"inactive-trailing",
      "handle.color":"color", "handle.height":"handle-height", "handle.width":"width", "handle.shape":"handle-shape", "handle.opacity":"opacity", "handle.elevation":"elevation", "handle.shadow-color":"shadow", "handle.padding":"padding", "handle.leading-space":"leading", "handle.trailing-space":"trailing",
      "state-layer.size":"state-size", "state-layer.color":"state-color", "state-layer.opacity":"state-opacity",
      "with-overlap.handle.outline.color":"overlap-color", "with-overlap.handle.outline.width":"overlap-width",
      "active.stop-indicator.container.color":"active-stop-color", "active.stop-indicator.container.opacity":"active-stop-opacity", "inactive.stop-indicator.container.color":"stop-color", "inactive.stop-indicator.container.opacity":"inactive-stop-opacity",
      "stop-indicator.container.color":"stop-color", "stop-indicator.container.opacity":"stop-opacity", "stop-indicator.color":"stop-color", "stop-indicator.color-selected":"active-stop-color", "stop-indicator.size":"stop-size", "stop-indicator.shape":"stop-shape", "stop-indicator.trailing-space":"stop-trailing", "stop.color":"stop-color",
      "with-tick-marks.active.container.color":"active-tick-color", "with-tick-marks.active.container.opacity":"active-tick-opacity", "with-tick-marks.inactive.container.color":"tick-color", "with-tick-marks.inactive.container.opacity":"tick-opacity", "with-tick-marks.container.color":"disabled-tick-color", "with-tick-marks.container.opacity":"disabled-tick-opacity", "with-tick-marks.container.shape":"tick-shape", "with-tick-marks.container.size":"tick-size",
      "track.elevation":"elevation", "icon.size":"icon-size", "icon.padding":"icon-padding",
      "label.container.color":"label-container", "label.container.elevation":"label-elevation", "label.container.height":"label-height", "label.label-text.font-size":"label-text-font-size",
      "value-indicator.active.bottom-space":"value-gap", "value-indicator.container.color":"value-container"
    };
    const mapped=extra[property]||textRole(property,["label.label-text","value-indicator.label.label-text"]);
    if(!mapped) throw new Error(`Unmapped slider token: ${set}.${suffix} (${property})`);
    // Keep readable role names for both independently published value labels.
    return {family,selector,property:mapped.replaceAll(".","-").replace(/^label-label-text-/,"label-text-").replace(/^value-indicator-label-label-text-/,"value-text-"),...parsed};
  } else if (set === "md.comp.outlined-segmented-button" || /^md\.comp\.(primary|secondary)-navigation-tab$/.test(set)) {
    const segmented = set === "md.comp.outlined-segmented-button";
    family = segmented ? "segment" : "tab";
    selector = segmented ? ".md-segmented-button" : `.md-tabs:where([data-variant="${set.includes("primary") ? "primary" : "secondary"}"]) .md-tab`;
    parsed = parse(suffix.replace(/(^|\.)active(?=\.)/g,"$1selected").replace(/(^|\.)inactive(?=\.)/g,"$1unselected"));
    property = parsed.sourceProperty;
    if (property.endsWith(".type")) return {excluded:"Individual source font metrics supply typography."};
    property = property.replace(/^with-icon\./, "").replace(/^with-label-text\./, "");
    if (property.startsWith("with-icon-and-label-text.")) {
      property = property.slice("with-icon-and-label-text.".length);
      selector = selector.replace(" .md-tab", ':where([data-control="icon-label"]) .md-tab');
    }
    const extra = {"shape":"shape", "active-indicator.color":"indicator", "active-indicator.height":"indicator-height", "active-indicator.shape":"indicator-shape", "divider.color":"divider", "divider.height":"divider-height"};
    const mapped = extra[property] || roleProperties[property];
    if (!mapped) throw new Error(`Unmapped ${family} token: ${set}.${suffix}`);
    return {family,selector,property:mapped,...parsed};
  } else if (/^md\.comp\.navigation-(bar|rail)$/.test(set)) {
    return {excluded:"Older navigation set; the renderer uses the current separate nav-bar and nav-rail sets."};
  } else if (/^md\.comp\.nav-(bar|rail)(\.|$)/.test(set) || set === "md.comp.navigation-drawer") {
    parsed = parse(suffix.replace(/(^|\.)active(?=\.)/g,"$1selected").replace(/(^|\.)inactive(?=\.)/g,"$1unselected"));
    property = parsed.sourceProperty;
    family = "nav";
    const bar = set.startsWith("md.comp.nav-bar"), drawer = set === "md.comp.navigation-drawer";
    const root = bar ? ".md-navigation-bar" : drawer ? ".md-navigation-drawer" : ".md-navigation-rail";
    selector = root;
    const extra = {"container.width":"width", "container.width.minimum":"min-width", "container.width.maximum":"max-width", "active-indicator.color":"indicator", "active-indicator.height":"indicator-height", "active-indicator.width":"indicator-width", "active-indicator.shape":"indicator-shape", "active-indicator.leading-space":"leading", "active-indicator.trailing-space":"trailing", "active-indicator.icon-label-space":"indicator-gap", "icon-label-space":"indicator-gap", "container.between-space":"label-gap", "container.vertical-space":"label-gap", "full-width.leading-space":"outer-leading", "full-width.trailing-space":"outer-trailing", "leading-space":"outer-leading", "trailing-space":"outer-trailing", "indicator.color":"indicator", "divider.color":"divider", "scrim.color":"scrim", "scrim.opacity":"scrim-opacity", "top-space":"top", "between-item-space":"gap", "item.vertical-space":"gap", "item.between-space":"gap", "vertical.trailing-space":"end-padding"};
    if (!drawer) {
      const part = set.slice((bar ? "md.comp.nav-bar" : "md.comp.nav-rail").length).replace(/^\./, "");
      if ((part === "collapsed" || part === "expanded") && property === "item.vertical-space") {
        return {family,selector:root+`:where([data-variant="${part}"])`,property:"gap",...parsed};
      }
      if (part.startsWith("item") || property.startsWith("item.")) {
        if (property.startsWith("item.")) property = property.slice(5);
        if (property === "between-space") return {family:"nav",selector:root,property:"gap",...parsed};
        if (property === "header-space-minimum") return {family:"nav",selector:root,property:"header-gap",...parsed};
        family = "nav-item";
        const orientation = part.split(".")[1];
        selector = root + (orientation ? `:where([data-orientation="${orientation}"])` : "") + " .md-nav-item";
        if (property.startsWith("short.")) { selector = root + ':where([data-size="short"]) .md-nav-item'; property = property.slice(6); }
        if (property === "label-text.font") return {family,selector,property:"font",...parsed};
      } else {
        if (part === "collapsed" || part === "expanded") selector += part === "collapsed" ? ':where([data-variant="collapsed"])' : ':where([data-variant="expanded"])';
        if (property.startsWith("modal.")) { selector += ':where([data-layout="modal"])'; property = property.slice(6); }
        if (property.startsWith("narrow.")) { selector += ':where([data-width="narrow"])'; property = property.slice(7); }
      }
    } else {
      if(property.startsWith("bottom."))return {excluded:"Bottom placement is not offered by the drawer preview; retained in source exports."};
      if (property.startsWith("modal.") || property.startsWith("standard.")) {
        const variant = property.split(".")[0]; selector += `:where([data-variant="${variant}"])`; property = property.slice(variant.length+1);
      } else if (parsed.state || parsed.selected || property.startsWith("icon.") || property.startsWith("label-text.") || property.startsWith("large-badge-label.") || property.startsWith("focus.indicator.")) {
        family = "nav-item"; selector += " .md-nav-item";
      }
    }
    if (property === "container.surface-tint-layer.color") return {excluded:"The source surface color supplies its tonal treatment."};
    if (property.endsWith(".type")) return {excluded:"Navigation typography uses the individual source font metrics."};
    const mapped = !bar&&!drawer&&property==="container.vertical-space" ? "vertical-padding" : extra[property] || roleProperties[property] || textRole(property,["headline","large-badge-label"]);
    if (!mapped) throw new Error(`Unmapped navigation token: ${set}.${suffix} (${property})`);
    return {family,selector,property:mapped,...parsed};
  } else if (set === "md.comp.badge") {
    selector = '.md-badge:where(:not([data-variant="labeled"]))';
    family = "badge";
    property = suffix;
    if (property.startsWith("large.")) {
      selector = '.md-badge:where([data-variant="labeled"])';
      property = property.slice(6);
    }
    if (property.endsWith(".type")) return {excluded:"Badge typography uses the individual source font metrics."};
    const mapped = {color:"color",size:"size",shape:"shape"}[property] || textRole(property,["label-text"]);
    if (!mapped) throw new Error(`Unmapped badge token: ${suffix}`);
    return {family,selector,property:mapped};
  } else if (["md.comp.snackbar", "md.comp.plain-tooltip", "md.comp.rich-tooltip"].includes(set)) {
    family = set === "md.comp.snackbar" ? "snackbar" : "tooltip";
    selector = set === "md.comp.snackbar" ? ".md-snackbar" : set.endsWith("plain-tooltip") ? '.md-tooltip:where([data-variant="plain"])' : '.md-tooltip:where([data-variant="rich"])';
    if (property.startsWith("action.")) {
      family = "support-action";
      selector += " .md-feedback-action";
      property = property.slice(7);
    } else if (property.startsWith("icon.") && parsed.state) {
      family = "support-icon";
      selector += " .md-snackbar-close";
      property = property.slice(5);
    }
    if (property.startsWith("with-single-line.") || property.startsWith("with-two-lines.")) {
      selector += property.startsWith("with-single-line.") ? ':where([data-variant="single-line"])' : ':where([data-variant="two-lines"])';
      property = property.slice(property.indexOf(".")+1);
    }
    parsed.sourceProperty = property;
  } else if (/^md\.comp\.(fab|extended-fab)(\.|$)/.test(set)) {
    family = "fab";
    const extended = set.startsWith("md.comp.extended-fab");
    const root = extended ? "md.comp.extended-fab" : "md.comp.fab";
    const part = set.slice(root.length).replace(/^\./, "");
    selector = extended ? ".md-extended-fab" : ".md-fab";
    if (["", "small", "medium", "large"].includes(part)) {
      selector += `:where([data-size="${part || "default"}"])`;
      const geometry = {"container.height":"height", "container.width":"width", "container.shape":"shape", "icon.size":"icon-size", "icon-label-space":"gap", "leading-space":"leading", "trailing-space":"trailing", "label-text":"font"};
      if (!geometry[suffix]) throw new Error(`Unmapped FAB geometry: ${set}.${suffix}`);
      return {family,selector,property:geometry[suffix], ...parsed};
    }
    if (["primary", "secondary", "tertiary", "primary-container", "secondary-container", "tertiary-container"].includes(part)) {
      const current = ["container.color", "container.elevation", "container.shadow-color", "icon.color", "label-text.color"];
      const currentState = /^(focused|hovered|pressed)\.(container.elevation|icon.color|label-text.color|state-layer.color|state-layer.opacity)$/.test(suffix);
      if (current.includes(suffix) || currentState) {
        selector += `:where([data-variant="${part}"])`;
        return {family,selector,property:roleProperties[property], ...parsed};
      }
    }
    return {excluded:"Older combined FAB set; the renderer uses the separate current size and color sets."};
  } else if (set === "md.comp.checkbox" || set === "md.comp.switch") {
    family = set.slice("md.comp.".length);
    selector = `.md-${family}`;
    const error = property.startsWith("error.");
    if (error) {
      selector += ':where([aria-invalid="true"])';
      property = property.slice(6);
    }
    const withIcon = property.startsWith("with-icon.");
    if (withIcon) {
      selector += ':where([data-with-icon="true"])';
      property = property.slice(10);
    }
    const extra = {
      "container.size": "size", "container.width": "width", "container.outline.width": "outline-width",
      "state-layer.shape": "state-shape", "handle.shape": "handle-shape", "track.shape": "track-shape",
      "track.width": "track-width", "track.height": "track-height",
      "track.color": "track", "track.opacity": "track-opacity",
      "track.outline.width": "track-outline-width", "track.outline.color": "track-outline",
      "handle.width": "handle-width", "handle.height": "handle-height",
      "handle.color": "handle", "handle.opacity": "handle-opacity",
      "handle.elevation": "handle-elevation", "handle.shadow-color": "handle-shadow",
      "focus.indicator.offset": "focus-offset"
    };
    const mapped = extra[property] || roleProperties[property] || (property === "state-layer.size" ? "state-size" : "");
    if (!mapped) throw new Error(`Unmapped selection control token: ${set}.${suffix}`);
    return { ...parsed, family, selector, error, withIcon, property: mapped };
  } else if (/^md.comp.(elevated|filled|outlined)-card$/.test(set)) {
    family = "card";
    selector = ".md-" + set.split(".")[2];
    if (parsed.state) selector += ".md-interactive-card";
  } else if (set.startsWith("md.comp.icon-button.")) {
    family = "icon";
    const name = set.slice("md.comp.icon-button.".length);
    selector = ["xsmall", "small", "medium", "large", "xlarge"].includes(name)
      ? `.md-icon-button[data-size="${name}"]${name === "small" ? ", .md-icon-button:not([data-size])" : ""}`
      : `.md-icon-button.md-${name}-icon-button`;
    const special = {
      "container.shape.round": "round-shape",
      "container.shape.square": "square-shape",
      "container.shape": "shape",
      "outlined.outline.width": "size-outline-width",
      "pressed.container.corner-size.motion.spring.damping": "spring-damping",
      "pressed.container.corner-size.motion.spring.stiffness":
        "spring-stiffness",
    };
    // Size-set pressed geometry and spring parameters are available at rest.
    if (suffix in special) {
      parsed = { state: "", selected: "" };
      return { family, selector, property: special[suffix], ...parsed };
    }
    if (
      property === "container.shape.round" ||
      property === "container.shape.square"
    )
      return {
        family,
        selector,
        property:
          "selected-" +
          (property.endsWith("round") ? "round" : "square") +
          "-shape",
        ...parsed,
      };
    if (property === "container.shape" && parsed.state === "pressed")
      return {
        family,
        selector,
        property: "pressed-shape",
        state: "",
        selected: "",
      };
    const spacing = property.match(
      /^(default|narrow|wide)\.(leading|trailing)-space$/,
    );
    if (spacing) {
      const width =
        spacing[1] === "default"
          ? ':is(:not([data-width]),[data-width="default"])'
          : `[data-width="${spacing[1]}"]`;
      selector = selector
        .split(", ")
        .map((s) => s + width)
        .join(", ");
      return {
        family,
        selector,
        property: spacing[2],
        state: "",
        selected: "",
      };
    }
  } else if (/^md\.comp\.(?:filter|assist|suggestion|input)-chip$/.test(set)) {
    family = "chip";
    selector = ".md-" + set.slice(8);
    if (parsed.variant)
      selector += parsed.variant === "flat" ? ':where(:not([data-elevated="true"]))' : ':where([data-elevated="true"])';
    if(property.startsWith('with-avatar.')) {
      property=property.slice(12);
      if(['avatar.size','avatar.opacity','avatar.shape'].includes(property))return {...parsed,family,selector,property:property.replace('.','-')};
    }
    for(const slot of ['with-leading-icon','with-trailing-icon','with-icon'])
      if(property.startsWith(slot+'.')) {
        property=property.slice(slot.length+1);
        if(slot==='with-icon')break;
        const role=slot.slice(5),metric=property.slice(role.length+1);
        if(['color','opacity','size'].includes(metric))return {...parsed,family,selector,property:role+(metric==='color'?'':'-'+metric)};
      }
  } else if (set==='md.comp.full-screen-dialog') {
    if(suffix.endsWith('.type'))return{excluded:'Composite typography duplicates its individual font metrics.'};
    if(suffix.endsWith('surface-tint-layer.color'))return{excluded:'The captured header color supplies its tonal treatment.'};
    selector='.md-full-screen-dialog';family='full';let key=suffix;
    if(key.startsWith('header.action.')){family='full-action';selector+=' .md-full-header .md-text-button';key=key.slice(14);}
    else if(key.startsWith('header.')){family='full-header';key=key.slice(7);if(key.startsWith('on-scroll.')){selector+=':where([data-top-scrolled="true"])';key=key.slice(10);}selector+=' .md-full-header';}
    else if(key.startsWith('action-bar.')){
      key=key.slice(11);
      if(key.startsWith('on-scroll.')){selector+=':where([data-bottom-scrolled="true"])';key=key.slice(10);}
      if(key.startsWith('container.')){family='full-bar';selector+=' .md-full-action-bar';}
      else {family='full-action';selector+=' .md-full-action-bar .md-text-button';}
    }
    const parsed=parse(key),property=parsed.sourceProperty,role=roleProperties[property]||textRole(property,['headline'])||({'with-divider.divider.color':'divider-color','with-divider.divider.height':'divider-height'}[property]);
    if(!role)throw new Error(`Unmapped fullscreen dialog field: ${suffix}`);
    if(role==='height'&&['full-header','full-bar'].includes(family))selector=selector.replace(/ \.md-full-(?:header|action-bar)$/,'');
    return {...parsed,selector,family,property:role,rank:selector.includes('data-top-scrolled')||selector.includes('data-bottom-scrolled')?1:0};
  } else if (set === "md.comp.dialog") {
    family = "dialog";
    selector = ".md-dialog";
    if (property.startsWith("action.")) {
      selector += " .md-text-button";
      property = property.slice(7);
      family = "dialog-action";
    }
    const extra = {
      "with-divider.divider.color": "divider",
      "with-divider.divider.height": "divider-height",
      "with-icon.icon.color": "icon",
      "with-icon.icon.size": "icon-size",
    };
    if (extra[property])
      return { family, selector, property: extra[property], ...parsed };
  } else if (set === "md.comp.radio-button") {
    family = "radio";
    selector = ".md-radio";
  } else if (/^md\.comp\.(?:filled|outlined)-(?:text-field|select|autocomplete)$/.test(set)) {
    family='field';selector='.md-'+set.slice(8);
    if(suffix.startsWith('menu.')) {
      if(suffix.includes('cascading-menu-indicator'))return {excluded:'Cascading menu indicators do not belong to a single-choice listbox popup.'};
      if(suffix.endsWith('.type'))return{excluded:'Composite typography duplicates its individual font metrics.'};
      if(suffix.endsWith('surface-tint-layer.color'))return{excluded:'The captured popup container color supplies its tonal treatment.'};
      const key=suffix.slice(5),roles={'container.color':'container','container.elevation':'elevation','container.shadow-color':'shadow','container.shape':'shape','divider.color':'divider-color','divider.height':'divider-height','list-item.container.height':'item-height','list-item.label-text.color':'label-color','list-item.selected.container.color':'selected-container','list-item.with-leading-icon.leading-icon.color':'leading-icon-color','list-item.with-leading-icon.leading-icon.size':'leading-icon-size','list-item.with-trailing-icon.trailing-icon.color':'trailing-icon-color','list-item.with-trailing-icon.trailing-icon.size':'trailing-icon-size'};
      return{family:'field-menu',selector,property:roles[key]||textRole(key,['list-item.label-text'])?.replace('list-item.label-text-',''),state:''};
    }
    const raw=suffix.replace(/^text-field\./,''),error=raw.startsWith('error.'),key=error?raw.slice(6):raw;
    parsed=parse(key);property=parsed.sourceProperty;
    if(error)selector+=':where([aria-invalid="true"],:has([aria-invalid="true"]))';
    const extras={'leading-icon.size':'leading-icon-size','trailing-icon.size':'trailing-icon-size','caret.color':'caret','input-text.placeholder.color':'placeholder','input-text.prefix.color':'prefix','input-text.suffix.color':'suffix','label-text.populated.size':'populated-label-size','label-text.populated.line-height':'populated-label-line-height','active-indicator.color':'indicator-color','active-indicator.height':'indicator-height','active-indicator.opacity':'indicator-opacity','active-indicator.thickness':'focus-width'};
    if(extras[property])return {...parsed,family,selector,property:extras[property],error};
    if(property.endsWith('.type'))return{excluded:'Composite typography duplicates its individual font metrics.'};
    if(property==='container.surface-tint-layer.color')return{excluded:'The captured container color supplies its tonal treatment.'};
    const role=roleProperties[property]||textRole(property,['input-text','supporting-text','leading-icon','trailing-icon'])||(/^(?:leading|trailing)-icon\.(?:color|opacity)$/.test(property)?property.replace('.','-'):undefined);
    if(!role)throw new Error(`Unmapped field token: ${set}.${suffix}`);
    return {...parsed,family,selector,property:role,error};
  } else if(set==='md.comp.menu') {
    if(suffix.endsWith('.type'))return{excluded:'Composite typography duplicates the individual font metrics.'};
    if(suffix==='container.surface-tint-layer.color')return{excluded:'The captured container color supplies its tonal treatment.'};
    const root='.md-menu[data-variant="baseline"]';
    let selector=root,key=suffix,family='menu';
    if(key.startsWith('list-item.')){family='menu-item';selector+=' .md-menu-item';key=key.slice(10);}
    else if(key.startsWith('focus.indicator.')){family='menu-item';selector+=' .md-menu-item';}
    else if(key.startsWith('cascading-menu-indicator.')){return{family:'menu-item',selector:root+' .md-menu-item:where([aria-haspopup="menu"])',property:key.endsWith('.size')?'trailing-icon-size':'trailing-icon-color',state:'',rank:1,baseline:true};}
    const parsed=parse(key);key=parsed.sourceProperty;
    if(key.includes('with-leading-icon.'))key=key.replace('with-leading-icon.','').replace(/^icon\./,'leading-icon.');
    if(key.includes('with-trailing-icon.'))key=key.replace('with-trailing-icon.','').replace(/^icon\./,'trailing-icon.');
    const roles={'container.height':'height','leading-icon.size':'leading-icon-size','trailing-icon.size':'trailing-icon-size','divider.color':'divider-color','divider.height':'divider-height'};
    return {...parsed,family,selector,property:roles[key]||roleProperties[key]||textRole(key,['leading-icon','trailing-icon']),baseline:true};
  } else if (set.startsWith("md.comp.menus")) {
    family = "menu";
    selector = '.md-menu:where(:not([data-variant="baseline"]))';
    if (set !== "md.comp.menus")
      selector += set.endsWith(".standard")
        ? ':not([data-variant="vibrant"])'
        : '[data-variant="vibrant"]';
    if(suffix==='disabled.icon-button.icon.color'||suffix.startsWith('icon-button.')) {
      family='menu-icon';selector+=' .md-menu-icon-button';
      let key=suffix==='disabled.icon-button.icon.color'?'disabled.icon.color':suffix.slice(12);
      const parsed=parse(key);
      // The Standard source places icon-button.disabled under Disabled, Selected.
      if(set==='md.comp.menus.standard'&&suffix==='icon-button.disabled.icon.color')parsed.selected='selected';
      const roles={'container.color':'container','icon.color':'icon','state-layer.color':'state-color'};
      return {...parsed,family,selector,property:roles[parsed.sourceProperty]};
    }
    if(suffix.startsWith('horizontal.')) {
      selector+=':where([data-orientation="horizontal"])';
      let key=suffix.slice(11),rank=2;
      if(key.startsWith('icon-only.')){selector+=':where([data-content="icon-only"])';key=key.slice(10);rank=3;}
      if(key.startsWith('menu-item.')){selector+=' .md-menu-item';family='menu-item';key=key.slice(10);}
      const parsed=parse(key),roles={'container.shape':'shape','container.top-space':'top','container.bottom-space':'bottom','gap':'gap','shape':'shape','between-space':'gap','leading-space':'leading','trailing-space':'trailing','top-space':'top','bottom-space':'bottom'};
      return {...parsed,family,selector,property:roles[parsed.sourceProperty],rank};
    }
    if(suffix==='active.container.shape'||suffix==='inactive.container.shape')return{family,selector:selector+':where(:not([data-orientation="horizontal"])[data-menu-activity="'+suffix.split('.')[0]+'"])',property:'shape',state:'',rank:1};
    if(suffix.startsWith('menu-item.active.'))return{family:'menu-item',selector:selector+' .md-menu-item:where([aria-expanded="true"])'+enabled,property:suffix.endsWith('.color')?'active-state-color':'active-state-opacity',state:''};
    if (property.startsWith("menu-item.")) {
      property = property.slice(10);
      selector += " .md-menu-item";
      family = "menu-item";
    }
    if (
      property.startsWith("first-child.") ||
      property.startsWith("last-child.")
    ) {
      selector=selector.replace(' .md-menu-item',':where(:not([data-orientation="horizontal"])) .md-menu-item');
      const part = property.split(".")[0];
      const inner = property.endsWith("inner-corner.corner-size");
      const position = `:${part}${inner ? ":not(:only-child)" : ""}`;
      selector = selector
        .split(", ")
        .flatMap((s) => [
          s + `:where(${position}):not(:where(li > *))`,
          s.replace(" .md-menu-item", ` :where(li${position}) > .md-menu-item`),
        ])
        .join(", ");
      property = property.slice(part.length + 1);
      return {
        family,
        selector,
        ...parsed,
        property:
          (part === "first-child") !== inner ? "top-shape" : "bottom-shape",
      };
    }
    const extra = {
      gap: "gap",
      "group.padding": "group-padding",
      "group.shape": "group-shape",
      color: "container",
      shape: "shape",
      height: "height",
      "between-space": "gap",
      "leading-space": "leading",
      "trailing-space": "trailing",
      "top-space": "top",
      "bottom-space": "bottom",
      "leading-icon.size": "leading-icon-size",
      "trailing-icon.size": "trailing-icon-size",
    };
    if (extra[property])
      return { family, selector, property: extra[property], ...parsed };
  } else if (set === "md.comp.divider") {
    family = "divider";
    selector = ".md-divider";
    return {
      family,
      selector,
      property: { color: "color", thickness: "thickness" }[property],
      ...parsed,
    };
  } else return null;
  if (property === "container.surface-tint-layer.color")
    return {
      excluded:
        "Surface container colors include their tonal treatment; no second tint layer is applied.",
    };
  if (property.endsWith(".type"))
    return {
      excluded:
        "Composite typography is represented by the individual font, size, weight, line-height and tracking tokens.",
    };
  let prop =
    roleProperties[property] ||
    textRole(property, [
      "headline",
      "subhead",
      "supporting-text",
      "input-text",
      "leading-icon",
      "trailing-icon",
      "trailing-supporting-text",
      "section-label-text",
    ]);
  if (
    !prop &&
    [
      "leading-icon.color",
      "trailing-icon.color",
      "leading-icon.opacity",
      "trailing-icon.opacity",
    ].includes(property)
  )
    prop = property.replaceAll(".", "-");
  if (!prop && property === "state-layer.size") prop = "state-size";
  if (!prop)
    throw new Error(`Unmapped component token: ${set}.${suffix} (${property})`);
  return { family, selector, target, ...parsed, property: prop };
}

export function generateComponents(document, values) {
  const source = new Map();
  function walk(node) {
    const meta = node.$extensions?.[extension];
    if (meta?.sourceName) source.set(meta.sourceName, meta);
    for (const [key, child] of Object.entries(node))
      if (
        (!key.startsWith("$") || key === "$root") &&
        child &&
        typeof child === "object"
      )
        walk(child);
  }
  walk(document);
  const mapped = [],
    excluded = [],
    rules = [],
    requiredNames = new Set();
  for (const [set, suffixes] of Object.entries(contract)) {
    for (const suffix of suffixes)
      if (!source.has(set + "." + suffix))
        throw new Error(`Missing required component token: ${set}.${suffix}`);
  }
  const records = [];
  for (const [name, meta] of source) {
    const set = meta.presentation?.setName;
    if (
      !set &&
      /^md\.comp\.(?:elevated-card|filled-card|outlined-card|filter-chip|assist-chip|suggestion-chip|input-chip|dialog|full-screen-dialog|radio-button|outlined-text-field|filled-text-field|outlined-select|filled-select|outlined-autocomplete|filled-autocomplete|icon-button|menu|menus|checkbox|switch|fab|extended-fab|badge|snackbar|plain-tooltip|rich-tooltip|nav-bar|nav-rail|navigation-drawer|primary-navigation-tab|secondary-navigation-tab|outlined-segmented-button|slider|progress-indicator|linear-progress-indicator|circular-progress-indicator|loading-indicator|app-bar|top-app-bar|bottom-app-bar|toolbar|sheet)\./.test(
        name,
      )
    )
      throw new Error(`Missing component source set: ${name}`);
    if (!(set in contract)) {
      if (
        set === "md.comp.icon-button" ||
        /^md.comp.(filled|filled-tonal|outlined)-icon-button$/.test(
          set || "",
        )
      ) {
        excluded.push({
          token: name,
          reason:
            "Superseded baseline or deprecated set; current icon-button sets are used.",
        });
        continue;
      }
      if (
        set?.startsWith("md.comp.icon-button.") ||
        set?.startsWith("md.comp.menus.") || set?.startsWith("md.comp.menu.") ||
        /^md\.comp\.nav-(?:bar|rail)(?:\.|$)/.test(set || "") ||
        /^md\.comp\.(?:full-screen-dialog|(?:outlined|filled)-(?:text-field|select|autocomplete)|assist-chip|suggestion-chip|input-chip|primary-navigation-tab|secondary-navigation-tab|outlined-segmented-button|slider|progress-indicator|linear-progress-indicator|circular-progress-indicator|loading-indicator|app-bar|top-app-bar|bottom-app-bar|toolbar|sheet|split-button|button-group|fab-menu|list|carousel-item|search-bar|search-view|date-picker|date-input|time-picker|time-input)(?:\.|$)/.test(set || "")
      )
        throw new Error(`Unknown component set: ${set}`);
      continue;
    }
    const suffix = name.slice(set.length + 1);
    if (!contract[set].includes(suffix))
      throw new Error(`Unknown component token: ${name}`);
    const b = binding(set, suffix);
    if (!b) throw new Error(`Missing component mapping: ${name}`);
    if (b.excluded) {
      if(name.endsWith('.opcaity')){const canonical=name.replace('.opcaity','.opacity');if(values[name.replaceAll('.','-')]!==values[canonical.replaceAll('.','-')])throw new Error(`Conflicting duplicate date opacity: ${name}`);}
      if(name.startsWith('md.comp.list.md.comp.list.')) {
        const canonical=name.replace('md.comp.list.md.comp.list.','md.comp.list.');
        if(values[name.replaceAll('.','-')]!==values[canonical.replaceAll('.','-')])throw new Error(`Conflicting duplicate list height: ${name}`);
      }
      excluded.push({ token: name, reason: b.excluded });
      continue;
    }
    const variable = name.replaceAll(".", "-");
    if (
      (name.includes(".corner-size.motion.spring.") || name.includes(".item.width.motion.spring.")) &&
      !(Number(values[variable]) > 0)
    )
      throw new Error(`Invalid component spring: ${name}`);
    if (!(variable in values))
      throw new Error(`Missing converted component value: ${name}`);
    if (!b.property) throw new Error(`Missing component property: ${name}`);
    // Indicators retain their normal property selector; interaction states are emitted in precedence order.
    records.push({ ...b, variable, token: name, dimensionFromElevation: b.family === "carousel-item" && b.property === "outline-width" && meta.sourceType === "elevation" });
  }
  records.sort(
    (a, b) =>
      // A selected menu item retains its selected colors when a state only
      // supplies opacity or a state layer, rather than a replacement color.
      // Baseline disabled colors apply to both selected and unselected items.
      Number(!!a.baseline && a.state==='disabled')-Number(!!b.baseline && b.state==='disabled') ||
      Number(a.family === "loading" && a.selector.includes("[data-variant")) -
        Number(b.family === "loading" && b.selector.includes("[data-variant")) ||
      Number(a.family === "progress" && a.selector.includes("[data-size")) -
        Number(b.family === "progress" && b.selector.includes("[data-size")) ||
      Number(a.family === "menu-item" && !!a.selected) -
        Number(b.family === "menu-item" && !!b.selected) ||
      order.indexOf(a.state) - order.indexOf(b.state) ||
      (a.rank||0) - (b.rank||0) ||
      Number(!!a.withIcon) - Number(!!b.withIcon) ||
      Number(!!a.selected) - Number(!!b.selected) ||
      Number(!!a.error) - Number(!!b.error) ||
      Number(a.family === "nav" && a.selector.includes("[data-variant")) -
        Number(b.family === "nav" && b.selector.includes("[data-variant")) ||
      Number(/:first-child|:last-child/.test(a.selector)) -
        Number(/:first-child|:last-child/.test(b.selector)),
  );
  for (const b of records) {
    const sliderState = {
      hovered: ':where([data-slider-hover],:has([data-slider-hover]),[data-component-state="hovered"]):not(:has(input:disabled))',
      focused: ':where(:has(input:focus-visible),[data-component-state="focused"]):not(:has(input:disabled))',
      pressed: ':where(:has([data-pressed]),[data-pressed],[data-component-state="pressed"]):not(:has(input:disabled))',
      disabled: ':where(:has(input:disabled))'
    };
    const fieldCondition =
      ["slider","slider-control"].includes(b.family) ? (b.state ? sliderState[b.state] : "") :
      b.family === "list-item" && b.state === "focused"
        ? `:where(${condition(b).replace(':has(> input:focus-visible)', ':has(.md-list-selection-control > input:focus-visible)')})` :
      b.family === "chip" && b.state === "focused"
        ? condition(b).replace(":focus-visible,:has(> input:focus-visible)", ":focus-visible,:has(> input:focus-visible),:has(> button:focus-visible)")
        : b.family === "field" && b.state === "focused"
        ? condition({ ...b, state: "" }) +
          `:where(:focus,:focus-within,[data-component-state="focused"])${enabled}`
        : ["menu-item", "checkbox", "switch", "nav-item", "segment", "tab", "toolbar-item", "sheet-action", "split-item", "fab-menu-item", "fab-menu-close", "list-item", "list-reveal", "carousel-item", "date-day", "date-year", "date-month", "date-menu", "time-value", "time-period", "time-field"].includes(b.family) && condition(b)
          ? `:where(${condition(b)})`
          : condition(b);
    const selectors = b.selector
      .split(", ")
      .map((s) => s + fieldCondition + (b.target || ""))
      .join(", ");
    const property =
      b.family === "dialog-action" &&
      ["label", "state-color", "state-opacity"].includes(b.property)
        ? `--_button-${b.property}`
        : `--_${b.family === "split-item" && b.property === "inner" ? "split" : b.family}-${b.property}`;
    rules.push(`${selectors} { ${property}: ${b.dimensionFromElevation ? `calc(var(--${b.variable}) * 1px)` : `var(--${b.variable})`}; }`);
    if((b.family === "slider" && ["active-shape","inactive-shape"].includes(b.property)) || (b.family === "sheet" && b.property === "shape")) {
      for(const [corner,name]of [["topLeft","tl"],["topRight","tr"],["bottomRight","br"],["bottomLeft","bl"]]) {
        const child=`${b.variable}-${corner}`, variable=child in values?child:b.variable;
        rules.push(`${selectors} { ${property}-${name}: var(--${variable}); }`);requiredNames.add(variable);
      }
    }

    if (
      b.family === "menu-item" &&
      b.property === "shape" &&
      (b.selected === "selected" || b.selector.includes('data-orientation="horizontal"'))
    )
      rules.push(
        `${selectors} { --_menu-item-top-shape: var(--${b.variable}); --_menu-item-bottom-shape: var(--${b.variable}); }`,
      );
    if (
      !["pressed", "disabled"].includes(b.state) &&
      /state-(color|opacity)$/.test(b.property)
    )
      rules.push(
        `${selectors} { ${property.replace("state-", "background-state-")}: var(--${b.variable}); }`,
      );
    if (b.state === "pressed" && /state-(color|opacity)$/.test(b.property)) {
      const persistent = b.selector
        .split(", ")
        .map((s) => s + condition({ ...b, state: "" }) + (b.target || ""))
        .join(", ");
      rules.push(
        `${persistent} { ${property.replace("state-", "pressed-state-")}: var(--${b.variable}); }`,
      );
    }
    if ((["fab", "nav-item", "fab-menu-item"].includes(b.family) && b.property === "font") || (b.family === "appbar" && b.property.endsWith("-font"))) {
      const tracking = `${b.variable}-letter-spacing`;
      if (!(tracking in values)) throw new Error(`Missing composite typography tracking: ${tracking}`);
      rules.push(`${selectors} { --_${b.family}-${b.family==="appbar"?b.property.replace(/font$/,"tracking"):"tracking"}: var(--${tracking}); }`);
      requiredNames.add(tracking);
    }
    mapped.push({ token: b.token, selector: selectors, property });
    requiredNames.add(b.variable);
  }
  const configurations = [];
  for(const variant of ['clock','input'])for(const orientation of variant==='clock'?['portrait','landscape']:['portrait'])
    for(const width of ['12h','24h'])for(const control of variant==='clock'?['value','dial',...(width==='12h'?['period']:[])]:['field',...(width==='12h'?['period']:[])])
      for(const shape of ['value','field','dial'].includes(control)?['hour','minute']:['hour'])
        for(const selection of ['value','period'].includes(control)?['selected','unselected']:[''])
          for(const state of control==='dial'?['enabled']:control==='field'?['enabled','hovered','focused']:['enabled','hovered','focused','pressed'])
            configurations.push({family:'time',variant,orientation,width,control,shape,selection,state});
  for(const variant of ['docked','modal','input','full-screen'])
    for(const dateRange of [false,true])
      for(const control of variant==='docked'?['day','month','year','menu-button']:variant==='input'?['day']:['day','year'])
        for(const selection of variant==='input'?['']:control==='day'?['selected','unselected','today',...(dateRange?['in-range']:[])]:['text','menu-button'].includes(control)?['']:['selected','unselected'])
          for(const state of variant==='input'?['enabled']:['enabled','hovered','focused','pressed',...(control==='menu-button'?['disabled']:[])])
            configurations.push({family:'date',variant,dateRange,control,selection,state});
  for(const variant of ['adaptive','docked','full-screen'])for(const control of ['baseline','contained'])
    for(const shape of ['icon','menu','none'])for(const width of ['clear','avatar','none','settings','clear-settings','settings-avatar'])for(const expanded of [false,true])
      for(const state of expanded?['enabled']:['enabled','hovered','focused','pressed'])
        configurations.push({family:'search',variant,control,shape,width,expanded,state});
  for(const variant of ['multi-browse','uncontained','hero','full-screen'])
    for(const control of ['standard','outlined'])for(const size of ['small','medium','large'])
      for(const state of ['enabled','hovered','focused','pressed','disabled'])
        configurations.push({family:'carousel',variant,control,size,state});
  for(const control of ['select','expand','reorder','reveal'])for(const variant of ['standard','segmented'])
    for(const shape of ['expressive','baseline'])for(const size of ['one-line','two-line','three-line'])
      for(const width of ['icon','none','avatar','image','video'])
        for(const selection of control==='select'?['multiple','single']:[''])
          for(const selected of control==='select'?[false,true]:[false])
            for(const expanded of control==='expand'?[false,true]:[false])for(const state of control==='reorder'?['enabled','hovered','focused','pressed','dragged','disabled']:['enabled','hovered','focused','pressed','disabled'])
              configurations.push({family:'list',control,variant,shape,size,width,selection,selected,expanded,state});
  for(const variant of ['primary','secondary','tertiary'])
    for(const control of ['items','close'])for(const state of ['enabled','hovered','focused','pressed'])
      configurations.push({family:'fab-menu',variant,control,state});
  for(const control of ["standard","connected"])
    for(const size of ["xsmall","small","medium","large","xlarge"])
      for(const variant of ["filled","tonal","outlined","elevated"])
        for(const shape of control==="connected"?["single","multiple"]:[""])
          for(const selected of control==="connected"?[false,true]:[false])
            for(const state of ["enabled","hovered","focused","pressed","disabled"])
              configurations.push({family:"button-group",control,size,variant,shape,selected,state});
  for(const size of ["xsmall","small","medium","large","xlarge"])
    for(const variant of ["filled","tonal","outlined","elevated"])
      for(const state of ["enabled","hovered","focused","pressed","disabled"])
        configurations.push({family:"split",size,variant,state});
  for(const control of ["standard","modal"]) {
    for(const variant of ["expanded","minimized"])for(const state of ["enabled","focused"])configurations.push({family:"bottom-sheet",variant,control,state});
    for(const variant of ["docked","detached"])for(const orientation of ["end","start"])for(const state of ["enabled","hovered","focused","pressed","disabled"])configurations.push({family:"side-sheet",variant,control,orientation,state});
  }
  for(const variant of ["small","centered","medium","large","medium-flexible","large-flexible","baseline-small","baseline-centered","baseline-medium","baseline-large"]) {
    const modern=!variant.startsWith("baseline");
    for(const control of modern?(variant==="small"?["title","subtitle","search","avatar"]:["title","subtitle","avatar"]):variant==="baseline-centered"?["title","avatar"]:["title"])
      for(const state of modern||["baseline-small","baseline-centered"].includes(variant)?["enabled","scrolled"]:["enabled"])
        configurations.push({family:"appbar",variant,control,state});
  }
  for(const control of ["actions","fab"])configurations.push({family:"appbar",variant:"bottom",control,state:"enabled"});
  for(const variant of ["docked","floating"])
    for(const control of ["standard","vibrant"])
      for(const orientation of variant==="docked"?["horizontal"]:["horizontal","vertical"])
        for(const size of variant==="floating"?["none","small","medium"]:["none"])
          for(const shape of orientation==="vertical"?["icon"]:["icon","label","icon-label"])
            for(const selected of [false,true])for(const state of ["enabled","hovered","focused","pressed","disabled"])
              configurations.push({family:"toolbar",variant,control,orientation,size,shape,selected,state});
  for(const control of ["indeterminate","determinate"])
    for(const variant of ["uncontained","contained"])
      configurations.push({family:"loading",control,variant,state:"enabled"});
  for(const variant of ["linear","circular","legacy-linear","legacy-circular"])
    for(const control of ["determinate","indeterminate"])
      for(const size of variant.startsWith("legacy") ? ["normal"] : ["normal","thick"])
        for(const shape of variant.startsWith("legacy") ? (control==="indeterminate" ? ["flat","four-color"] : ["flat"]) : ["flat","wave"])
          configurations.push({family:"progress",variant,control,size,shape,state:"enabled"});
  for(const variant of ["continuous","discrete"]) for(const control of ["single","range"])
    for(const size of ["xsmall","small","medium","large","xlarge"])
      for(const shape of ["label","value-indicator"]) for(const orientation of ["horizontal","vertical"])
        for(const state of ["enabled","hovered","focused","pressed","disabled"]) configurations.push({family:"slider",variant,control,size,shape,orientation,state});
  for (const selected of [false,true]) for (const state of ["enabled","hovered","focused","pressed"]) {
    for (const variant of ["vertical","horizontal"]) configurations.push({family:"nav-bar",variant,selected,state});
    for (const variant of ["standard","modal"]) configurations.push({family:"nav-drawer",variant,selected,state});
    for (const variant of ["collapsed","expanded"])
      for (const orientation of variant === "collapsed" ? ["vertical"] : ["horizontal"])
        for (const width of variant === "collapsed" ? ["default","narrow"] : ["default"])
          for (const size of ["default","short"]) configurations.push({family:"nav-rail",variant,orientation,width,size,selected,state});
  }
  for (const selected of [false,true]) for (const state of ["enabled","hovered","focused","pressed"]) {
    for (const variant of ["primary","secondary"]) for (const control of ["label","icon","icon-label"]) configurations.push({family:"tabs",variant,control,selected,state});
  }
  for (const selected of [false,true]) for (const state of ["enabled","hovered","focused","pressed","disabled"])
    for (const variant of ["single","multiple"]) for (const control of ["label","icon","icon-label"]) configurations.push({family:"segmented",variant,control,selected,state});
  configurations.push({family:"badge", variant:"dot", state:"enabled"},{family:"badge", variant:"labeled", state:"enabled"});
  for (const family of ["snackbar", "tooltip"])
    for (const variant of family === "snackbar" ? ["single-line", "two-lines"] : ["plain", "rich"])
      for (const state of family === "tooltip" && variant === "plain" ? ["enabled"] : ["enabled","hovered","focused","pressed"])
        for (const control of family === "snackbar" ? ["action", "dismiss"] : [""])
          configurations.push({family,variant,state,control});
  for (const family of ["fab", "extended-fab"])
    for (const variant of ["primary", "secondary", "tertiary", "primary-container", "secondary-container", "tertiary-container"])
      for (const size of ["default", "small", "medium", "large"])
        for (const state of ["enabled", "hovered", "focused", "pressed"])
          configurations.push({family, variant, size, state});
  for (const variant of ["elevated", "filled", "outlined"])
    for (const state of [
      "enabled",
      "hovered",
      "focused",
      "pressed",
      "dragged",
      "disabled",
    ])
      configurations.push({ family: "card", variant, state });
  for (const variant of ["standard", "filled", "tonal", "outlined"])
    for (const size of ["xsmall", "small", "medium", "large", "xlarge"])
      for (const shape of ["round", "square"])
        for (const width of ["default", "narrow", "wide"])
          for (const selected of [false, true])
            for (const state of [
              "enabled",
              "hovered",
              "focused",
              "pressed",
              "disabled",
            ])
              configurations.push({
                family: "icon",
                variant,
                size,
                shape,
                width,
                selected,
                state,
              });
  for(const variant of ['filter','assist','suggestion','input'])
    for(const elevated of variant==='input'?[false]:[false,true])
      for(const selected of ['filter','input'].includes(variant)?[false,true]:[false])
        for(const shape of variant==='input'?['icon','none','avatar']:['icon','none'])
          for(const width of variant==='filter'?['plain','removable']:['input'].includes(variant)?['removable']:['plain'])
            for(const state of ['enabled','hovered','focused','pressed','dragged','disabled'])
              configurations.push({family:'chip',variant,elevated,selected,shape,width,state});
  for (const selected of [false, true])
    for (const state of [
      "enabled",
      "hovered",
      "focused",
      "pressed",
      "disabled",
    ])
      configurations.push({ family: "radio", selected, state });
  for (const family of ["checkbox", "switch"])
    for (const selected of [false, true])
      for (const invalid of family === "checkbox" ? [false, true] : [false])
        for (const variant of family === "switch" ? ["plain", "icon"] : selected ? ["checked", "indeterminate"] : ["unchecked"])
          for (const state of ["enabled", "hovered", "focused", "pressed", "disabled"])
            configurations.push({ family, selected, invalid, variant, state });
  for(const family of ['field','select','autocomplete'])for(const variant of ['outlined','filled'])
    for(const shape of family==='field'?['plain','icons','affixes']:['plain','icons'])for(const width of ['empty','populated'])
      for(const invalid of [false,true])for(const state of ['enabled','hovered','focused','disabled'])
        configurations.push({family,variant,shape,width,invalid,state});
  for(const variant of ['standard','vibrant'])for(const orientation of ['vertical','horizontal'])
    for(const shape of orientation==='vertical'?['plain','supporting','cascade']:['plain','icon-only'])
    for(const selected of shape==='cascade'?[false]:[false,true])for(const expanded of shape==='cascade'?[false,true]:[false])
    for(const state of ['enabled','hovered','focused','pressed','disabled'])configurations.push({family:'menu',variant,orientation,shape,selected,expanded,state});
  for(const shape of ['plain','divided','cascade'])for(const selected of shape==='cascade'?[false]:[false,true])for(const expanded of shape==='cascade'?[false,true]:[false])for(const state of ['enabled','hovered','focused','pressed','disabled'])configurations.push({family:'menu',variant:'baseline',orientation:'vertical',shape,selected,expanded,state});
  for(const dialogHeadline of ['shown','hidden'])for(const shape of ['plain','icon','divided','icon-divided'])for(const expanded of [false,true])for(const state of ['enabled','hovered','focused','pressed'])configurations.push({family:'dialog',variant:'basic',dialogHeadline,control:'action',shape,expanded,state});
  for(const dialogHeadline of ['shown','hidden'])for(const orientation of ['header','bottom'])for(const control of ['action','close'])for(const shape of ['plain','divided'])for(const width of ['top','middle','bottom'])for(const state of ['enabled','hovered','focused','pressed'])configurations.push({family:'dialog',variant:'full-screen',dialogHeadline,orientation,control,shape,width,state});
  for(const variant of ['full-width','inset'])for(const shape of variant==='inset'?['start','both']:[''])for(const control of ['plain','supporting'])for(const width of ['none','right','bottom','right-bottom'])configurations.push({family:'divider',variant,shape,control,width,state:'enabled'});
  return {
    css:
      "/* Generated from the reviewed Material component contract. */\n" +
      rules.join("\n") +
      "\n",
    coverage: { mapped, excluded, configurations },
    requiredNames,
  };
}
