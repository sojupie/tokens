// Read-only checks of actual rendered properties against captured source variables.
export function checkComponentRendering() {
  const failures = [];
  let checked = 0,
    cases = 0;
  const color = (value) => {
    if (value === "transparent") return [0, 0, 0, 0];
    if (value.startsWith("#"))
      return [1, 3, 5]
        .map((i) => Number.parseInt(value.slice(i, i + 2), 16) / 255)
        .concat(
          value.length === 9 ? Number.parseInt(value.slice(7, 9), 16) / 255 : 1,
        );
    const n = (value.match(/-?\d*\.?\d+(?:e[+-]?\d+)?/g) || []).map(Number);
    return value.startsWith("rgb")
      ? n
          .slice(0, 3)
          .map((x) => x / 255)
          .concat(n[3] ?? 1)
      : n.slice(0, 3).concat(n[3] ?? 1);
  };
  for (const el of document.querySelectorAll("[data-component-case]")) {
    if (!el.getClientRects().length) continue;
    cases++;
    const cs = getComputedStyle(el),
      f = el.dataset.family,
      state = el.dataset.componentState || (["field","select","autocomplete"].includes(f) ? el.querySelector("input")?.disabled ? "disabled" : el.matches(":focus-within") ? "focused" : el.matches(":hover") ? "hovered" : "enabled" : undefined);
    const short = { hovered: "hover", focused: "focus" }[state] || state;
    const selected =
      el.checked || el.dataset.selected === "true" ||
      el.getAttribute("aria-pressed") === "true" ||
      el.getAttribute("aria-current") === "page" ||
      el.querySelector("input")?.checked;
    const sel = selected ? "selected" : "unselected";
    const get = (name) =>
      cs.getPropertyValue("--" + name.replaceAll(".", "-")).trim();
    const first = (names, fallback) => names.map(get).find(Boolean) ?? fallback;
    function compare(name, actual, expected, colors = false, opacity = 1) {
      checked++;
      let okay = actual === expected;
      if (name === "letterSpacing" && actual === "normal" && Number.parseFloat(expected) === 0) okay = true;
      if (colors) {
        const a = color(actual),
          b = color(expected || "");
        b[3] *= opacity;
        okay =
          a.length === 4 &&
          b.length === 4 &&
          a.every((x, i) => Math.abs(x - b[i]) < 0.001);
      } else if (expected?.includes("infinity"))
        okay = Number.parseFloat(actual) > 1000000;
      else if (
        Number.isFinite(Number.parseFloat(actual)) &&
        Number.isFinite(Number.parseFloat(expected))
      )
        okay =
          Math.abs(Number.parseFloat(actual) - Number.parseFloat(expected)) <
          0.01;
      if (!okay)
        failures.push({
          id: el.dataset.componentCase,
          property: name,
          actual,
          expected,
          opacity,
        });
    }
    function c(name, expected, opacity = 1) {
      compare(name, cs[name], expected, true, Number(opacity));
    }
    function value(name, expected) {
      compare(name, cs[name], expected);
    }
    function typography(element, prefix) {
      const style = getComputedStyle(element);
      for (const [property, suffix] of [["fontSize","size"],["fontWeight","weight"],["lineHeight","line-height"],["letterSpacing","tracking"]])
        compare(property,style[property],get(prefix + suffix));
      compare("fontFamily",style.fontFamily.replaceAll('"', '').replaceAll(" ", ""),get(prefix + "font").replaceAll('"', '').replaceAll(" ", ""));
    }
    function layer(prefix, colorName, opacityName) {
      const st = getComputedStyle(el, "::before");
      compare("state layer color", st.backgroundColor, colorName, true);
      compare("state layer opacity", st.opacity, opacityName);
    }
    function shadow(element, level, pseudo = null) {
      const clamp = (x, max = 1) => Math.max(0, Math.min(x, max));
      const kY = clamp(level) + clamp(level - 3) + 2 * clamp(level - 4);
      const kBlur = 2 * clamp(level) + clamp(level - 2) + clamp(level - 4);
      const aY = clamp(level) + clamp(level - 1) + 2 * clamp(level - 2, 3);
      const aBlur = 3 * clamp(level, 2) + 2 * clamp(level - 2, 3);
      const spread = clamp(level, 4) + 2 * clamp(level - 4);
      const offsets =
        getComputedStyle(element, pseudo)
          .boxShadow.replace(/rgba?\([^)]*\)|color\([^)]*\)/g, "color")
          .match(/-?\d*\.?\d+px/g) || [];
      const expected = [0, kY, kBlur, 0, 0, aY, aBlur, spread];
      checked++;
      if (
        offsets.length !== 8 ||
        offsets.some(
          (x, i) => Math.abs(Number.parseFloat(x) - expected[i]) > 0.01,
        )
      )
        failures.push({
          id: el.dataset.componentCase,
          property: "shadow geometry",
          actual: offsets,
          expected,
        });
    }
    if (f === "badge") {
      const prefix = "md.comp.badge." + (el.dataset.variant === "labeled" ? "large." : "");
      value("height", get(prefix + "size"));
      value("minWidth", get(prefix + "size"));
      c("backgroundColor", get(prefix + "color"));
      value("borderTopLeftRadius", get(prefix + "shape"));
      if (el.dataset.variant === "labeled") { c("color",get(prefix + "label-text.color")); typography(el,prefix + "label-text."); }
    } else if (f === "snackbar" || f === "tooltip") {
      const prefix = f === "snackbar" ? "md.comp.snackbar." : `md.comp.${el.dataset.variant}-tooltip.`;
      c("backgroundColor", get(prefix + "container.color"));
      value("borderTopLeftRadius", get(prefix + "container.shape"));
      if (f === "snackbar") value("minHeight",get(prefix + (el.dataset.variant === "two-lines" ? "with-two-lines" : "with-single-line") + ".container.height"));
      const text = el.querySelector(f === "snackbar" ? ".md-feedback-text" : "p");
      compare("supporting text",getComputedStyle(text).color,get(prefix + "supporting-text.color"),true);
      typography(text,prefix + "supporting-text.");
      if (f === "snackbar" || el.dataset.variant === "rich") shadow(el,Number.parseFloat(get(prefix + "container.elevation")));
      if (el.dataset.variant === "rich") typography(el.querySelector("h3"),prefix + "subhead.");
      const action = el.querySelector(".md-feedback-action");
      if (action) {
        const actionState = action.dataset.componentState;
        const actionShort = {hovered:"hover", focused:"focus"}[actionState] || actionState;
        compare("action label",getComputedStyle(action).color,first([prefix + "action." + actionShort + ".label-text.color",prefix + "action.label-text.color"]),true);
        compare("action state opacity",getComputedStyle(action,"::before").opacity,first([prefix + "action." + actionShort + ".state-layer.opacity"],"0"));
        typography(action,prefix + "action.label-text.");
      }
      const close = el.querySelector(".md-snackbar-close");
      if (close) {
        const closeState = close.dataset.componentState;
        const closeShort = {hovered:"hover", focused:"focus"}[closeState] || closeState;
        compare("close state opacity",getComputedStyle(close,"::before").opacity,first([prefix + "icon." + closeShort + ".state-layer.opacity"],"0"));
      }
    } else if (f === "fab" || f === "extended-fab") {
      const root = `md.comp.${f}.`, variant = root + el.dataset.variant + ".";
      const geometry = root + (el.dataset.size === "default" ? "" : el.dataset.size + ".");
      const source = property => first([variant + state + "." + property, variant + property]);
      value("height", get(geometry + "container.height"));
      value("borderTopLeftRadius", get(geometry + "container.shape"));
      if (f === "fab") value("width", get(geometry + "container.width"));
      c("backgroundColor", source("container.color"));
      const icon = getComputedStyle(el.querySelector("svg"));
      compare("icon color", icon.color, source("icon.color"), true);
      compare("icon width", icon.width, get(geometry + "icon.size"));
      shadow(el, Number.parseFloat(source("container.elevation")));
      if (f === "extended-fab") {
        c("color", source("label-text.color"));
        value("paddingLeft", get(geometry + "leading-space"));
        value("paddingRight", get(geometry + "trailing-space"));
        value("columnGap", get(geometry + "icon-label-space"));
        value("letterSpacing", get(geometry + "label-text-letter-spacing"));
      }
    } else if (f === "checkbox" || f === "switch") {
      const prefix = `md.comp.${f}.`;
      const cssValue = name => cs.getPropertyValue(`--_${f}-${name}`).trim();
      const painted = (element, pseudo, name, token, opacity = "1") => {
        compare(name, getComputedStyle(element, pseudo)[name], token, true, Number(opacity));
      };
      if (f === "checkbox") {
        const box = el.querySelector(".md-checkbox-box");
        compare("box width", getComputedStyle(box).width, get(prefix + "container.width"));
        compare("box height", getComputedStyle(box).height, get(prefix + "container.height"));
        painted(box, null, "backgroundColor", cssValue("container"), cssValue("container-opacity"));
        const mark = box.querySelector(el.querySelector("input").indeterminate ? ".md-checkbox-dash" : ".md-checkbox-check");
        if (selected) painted(mark, null, "color", cssValue("icon"), cssValue("icon-opacity"));
      } else {
        const track = el.querySelector(".md-switch-track"), handle = el.querySelector(".md-switch-handle");
        compare("track width", getComputedStyle(track).width, get(prefix + "track.width"));
        compare("track height", getComputedStyle(track).height, get(prefix + "track.height"));
        compare("handle width", getComputedStyle(handle).width, cssValue("handle-width"));
        compare("handle height", getComputedStyle(handle).height, cssValue("handle-height"));
        painted(track, null, "backgroundColor", cssValue("track"), cssValue("track-opacity"));
        painted(handle, "::after", "backgroundColor", cssValue("handle"), cssValue("handle-opacity"));
      }
      compare("native disabled", String(el.querySelector("input").disabled), String(state === "disabled"));
    } else if (f === "card") {
      value("paddingInlineStart", get("md.sys.measurement.space200"));
      value("paddingInlineEnd", get("md.sys.measurement.space200"));
      value("textAlign", "start");
      const grid=el.closest(".tokens-preview-grid");
      if(grid){
        compare("card grid gap",getComputedStyle(grid).columnGap,get("md.sys.measurement.space100"));
        compare("card specimen padding",getComputedStyle(el.parentElement).paddingInlineStart,"0px");
      }
      c("color", get("md.sys.color.on-surface"));
      const prefix =
        "md.comp." +
        ["elevated", "filled", "outlined"].find((v) =>
          el.classList.contains("md-" + v + "-card"),
        ) +
        "-card.";
      const source = (key) => first([prefix + short + "." + key, prefix + key]);
      c(
        "backgroundColor",
        source("container.color"),
        source("container.opacity") || 1,
      );
      value("borderTopWidth", get(prefix + "outline.width") || "0px");
      if (cs.borderTopWidth !== "0px")
        c(
          "borderTopColor",
          source("outline.color"),
          source("outline.opacity") || 1,
        );
      value("borderTopLeftRadius", get(prefix + "container.shape"));
      compare(
        "elevation",
        cs.getPropertyValue("--_elevation-level").trim(),
        source("container.elevation"),
      );
      compare(
        "shadow color",
        cs.getPropertyValue("--_shadow-color").trim(),
        get(prefix + "container.shadow-color"),
        true,
      );
      shadow(el, Number(source("container.elevation")));
      const icon = el.querySelector("[data-slot=icon]");
      if (icon) {
        compare(
          "card icon width",
          getComputedStyle(icon).width,
          get(prefix + "icon.size"),
        );
        compare(
          "card icon color",
          getComputedStyle(icon).color,
          get(prefix + "icon.color"),
          true,
        );
      }
      if (["hovered", "focused", "pressed", "dragged"].includes(state))
        layer(
          prefix,
          get(prefix + short + ".state-layer.color"),
          get(prefix + short + ".state-layer.opacity"),
        );
    } else if (f === "icon") {
      const variant = ["standard", "filled", "tonal", "outlined"].find((v) =>
        el.classList.contains("md-" + v + "-icon-button"),
      );
      const p = "md.comp.icon-button." + variant + ".",
        size = "md.comp.icon-button." + el.dataset.size + ".";
      const source = (key) =>
        state === "disabled"
          ? first([
              p + sel + ".disabled." + key,
              p + "disabled." + key,
              p + sel + "." + key,
              p + key,
            ])
          : first([
              p + sel + "." + state + "." + key,
              p + state + "." + key,
              p + sel + "." + key,
              p + key,
            ]);
      c(
        "backgroundColor",
        source("container.color") || "transparent",
        source("container.opacity") || 1,
      );
      c("color", source("icon.color"), source("icon.opacity") || 1);
      value("height", get(size + "container.height"));
      value("paddingLeft", get(size + el.dataset.width + ".leading-space"));
      value("paddingRight", get(size + el.dataset.width + ".trailing-space"));
      const icon = el.querySelector("svg");
      if (icon)
        compare(
          "icon width",
          getComputedStyle(icon).width,
          get(size + "icon.size"),
        );
      value(
        "borderTopWidth",
        variant === "outlined" && !selected
          ? get(size + "outlined.outline.width")
          : "0px",
      );
      const shape =
        state === "pressed"
          ? "pressed.container.shape"
          : (selected ? "selected." : "") +
            "container.shape." +
            el.dataset.shape;
      value("borderTopLeftRadius", get(size + shape));
      if (["hovered", "focused", "pressed"].includes(state))
        layer(p, source("state-layer.color"), source("state-layer.opacity"));
    } else if (f === "chip") {
      const p = "md.comp."+(el.dataset.variant||"filter")+"-chip.",
        elevated = el.dataset.elevated === "true",
        variant = elevated ? "elevated" : "flat";
      const source = (key) =>
        first([
          p +
            (state === "disabled" ? "disabled." : "") +
            sel +
            "." +
            (state === "disabled" ? "" : short + ".") +
            key,
          p + sel + "." + short + "." + key,
          p + sel + "." + key,
          p + (state === "disabled" ? "disabled." : "") + key,
          p + short + "." + key,
          p + key,
        ]);
      const bg=first([
        ...(state==='disabled'?[p+variant+'.disabled.'+sel+'.container.color',p+variant+'.disabled.container.color',p+'disabled.'+sel+'.container.color',p+'disabled.container.color']:[]),
        p+variant+'.'+sel+'.container.color',p+variant+'.container.color',p+sel+'.container.color',p+'container.color'
      ],'transparent');
      const opacity=state==='disabled'?first([p+variant+'.disabled.'+sel+'.container.opacity',p+variant+'.disabled.container.opacity',p+'disabled.'+sel+'.container.opacity',p+'disabled.container.opacity'],'1'):'1';
      c("backgroundColor", bg, opacity);
      c(
        "color",
        state === "disabled"
          ? get(p + "disabled.label-text.color")
          : source("label-text.color"),
        state === "disabled" ? get(p + "disabled.label-text.opacity") : 1,
      );
      const elevation =
        state === "dragged"
          ? get(p + "dragged.container.elevation")
          : first([
              p +
                variant +
                "." +
                (state === "disabled" ? "disabled." : "") +
                sel +
                "." +
                (state === "disabled" ? "" : short + ".") +
                "container.elevation",
              p + variant + "." + short + ".container.elevation",
              p + variant + "." + sel + "." + short + ".container.elevation",
              p + variant + ".container.elevation",
              p + "container.elevation",
            ]);
      compare(
        "chip elevation",
        cs.getPropertyValue("--_elevation-level").trim(),
        elevation,
      );
      shadow(el, Number(elevation));
      value("minHeight", get(p + "container.height"));
      value("fontSize", get(p + "label-text.size"));
      value("fontWeight", get(p + "label-text.weight"));
      value("letterSpacing", get(p + "label-text.tracking"));
      if (["hovered", "focused", "pressed", "dragged"].includes(state))
        layer(p, source("state-layer.color"), source("state-layer.opacity"));
      for(const icon of el.querySelectorAll('.md-chip-icon')) {
        if(getComputedStyle(icon).display==='none')continue;
        const slot=icon.dataset.slot, scope='with-'+slot;
        const expected=first([
          p+scope+'.'+sel+'.'+short+'.'+slot+'.color',p+scope+'.'+short+'.'+slot+'.color',
          p+scope+'.'+sel+'.'+slot+'.color',p+scope+'.'+slot+'.'+sel+'.color',p+scope+'.'+slot+'.color',
          p+'with-icon.'+sel+'.'+short+'.icon.color',p+'with-icon.'+short+'.icon.color',p+'with-icon.'+sel+'.icon.color',p+'with-icon.icon.color'
        ]);
        const opacity=state==='disabled'?first([p+scope+'.disabled.'+slot+'.opacity',p+'with-icon.disabled.icon.opacity'],'1'):'1';
        compare(slot+' color',getComputedStyle(icon).color,expected,true,Number(opacity));
        compare(slot+' size',getComputedStyle(icon).width,first([p+scope+'.'+slot+'.size',p+'with-icon.icon.size']));
      }
      const avatar=el.querySelector('.md-chip-avatar');if(avatar){const st=getComputedStyle(avatar);compare('avatar size',st.width,get(p+'with-avatar.avatar.size'));compare('avatar shape',st.borderTopLeftRadius,get(p+'with-avatar.avatar.shape'));compare('avatar opacity',st.opacity,state==='disabled'?get(p+'with-avatar.disabled.avatar.opacity'):'1');}
    } else if (f === "radio") {
      const p='md.comp.radio-button.',input=el.querySelector('input'),icon=el.querySelector('.md-radio-icon');
      const state=el.dataset.componentState||(input.disabled?'disabled':el.hasAttribute('data-pressed')?'pressed':input.matches(':focus-visible')?'focused':el.matches(':hover')?'hovered':'enabled'),short={hovered:'hover',focused:'focus'}[state]||state;
      const prefix=p+(state==='disabled'?'disabled.':'')+sel+'.'+(['hovered','focused','pressed'].includes(state)?short+'.':'');
      c('color',get(prefix+'icon.color'),state==='disabled'?get(prefix+'icon.opacity'):1);
      value('width',get(p+'state-layer.size'));value('height',get(p+'state-layer.size'));
      compare('radio icon width',getComputedStyle(icon).width,get(p+'icon.size'));
      compare('radio dot opacity',getComputedStyle(icon,'::after').opacity,input.checked?'1':'0');
      if(state==='focused'){value('outlineWidth',get('md.sys.state.focus-indicator.thickness'));value('outlineOffset',get('md.sys.state.focus-indicator.outer-offset'));c('outlineColor',get('md.sys.color.secondary'));}
      if(['hovered','focused','pressed'].includes(state)){
        const style=getComputedStyle(el,'::before');compare('radio state color',style.backgroundColor,get(prefix+'state-layer.color'),true);
        if(state!=='pressed'||el.dataset.componentState==='pressed')compare('radio state opacity',style.opacity,get(prefix+'state-layer.opacity'));
        const ripple=el.querySelector('.md-button-ripple > span');if(ripple){compare('radio ink color',getComputedStyle(el).getPropertyValue('--_button-pressed-state-color').trim(),get(p+sel+'.pressed.state-layer.color'),true);compare('radio ink opacity',getComputedStyle(ripple).opacity,get(p+sel+'.pressed.state-layer.opacity'));}
      } else compare('radio rest opacity',getComputedStyle(el,'::before').opacity,'0');
    } else if (['field','select','autocomplete'].includes(f)) {
      const variant=el.dataset.variant||'outlined',p='md.comp.'+variant+'-'+(f==='field'?'text-field':f)+'.'+(f==='field'?'':'text-field.'),invalid=el.getAttribute('aria-invalid')==='true';
      const prefixes=state==='disabled'?['disabled.','']:invalid?['error.'+(short==='enabled'?'':short+'.'),'error.',short+'.','']:[short+'.',''];
      const source=key=>first(prefixes.map(prefix=>p+prefix+key));
      c('color',source('input-text.color'),source('input-text.opacity')||1);
      value('minHeight',get(p+'container.height'));value('height',get(p+'container.height'));
      value('fontSize',get(p+'input-text.size'));value('fontWeight',get(p+'input-text.weight'));value('lineHeight',get(p+'input-text.line-height'));
      const input=el.querySelector('input'),label=el.querySelector('[data-slot="label"]'),floating=state==='focused'||!!input?.value||el.matches(':focus-within');
      if(label){const st=getComputedStyle(label);compare('label color',st.color,source('label-text.color'),true,Number(source('label-text.opacity')||1));compare('label size',st.fontSize,get(p+'label-text.'+(floating?'populated.':'')+'size'));}
      for(const slot of ['leading-icon','trailing-icon','supporting-text','prefix','suffix']) {
        const element=el.querySelector('[data-slot="'+slot+'"]');if(!element)continue;
        const st=getComputedStyle(element),key=['prefix','suffix'].includes(slot)?state==='disabled'?'input-text':'input-text.'+slot:slot;
        compare(slot+' color',st.color,source(key+'.color'),true,Number(source(key+'.opacity')||1));
        if(slot.endsWith('icon'))compare(slot+' size',st.width,get(p+slot+'.size'));
      }
      if(variant==='filled'&&label&&input) {
        const r=input.getBoundingClientRect(),ls=label.getBoundingClientRect(),is=getComputedStyle(input),rtl=cs.direction==='rtl',labelStyle=getComputedStyle(label);compare('filled label background',labelStyle.backgroundColor,'transparent',true);compare('filled label padding',labelStyle.paddingInlineStart,'0px');
        if(!el.querySelector('[data-slot="prefix"]'))compare('filled label alignment',String(rtl?ls.right:ls.left),String(rtl?r.right:r.left));
        if(floating)for(const affix of el.querySelectorAll('[data-slot="prefix"],[data-slot="suffix"]')){const a=affix.getBoundingClientRect();compare('affix top',String(a.top),String(r.top+Number.parseFloat(is.paddingTop)));compare('affix bottom',String(a.bottom),String(r.bottom-Number.parseFloat(is.paddingBottom)));}
      }
      if(variant==='filled') {
        c('backgroundColor',source('container.color'),source('container.opacity')||1);
        if(state==='hovered'){const layer=cs.backgroundImage.match(/(?:rgba?\([^)]*\)|color\(srgb[^)]*\))/)?.[0];compare('hover layer',layer||'',source('state-layer.color'),true,Number(source('state-layer.opacity')));}
        const line=getComputedStyle(el,'::after');compare('active indicator height',line.height,source('active-indicator.height'));compare('active indicator color',line.backgroundColor,source('active-indicator.color'),true,Number(source('active-indicator.opacity')||1));
      } else {
        const outline=el.querySelector('.md-field-outline'),st=outline?getComputedStyle(outline):getComputedStyle(el,'::after');
        compare('outline color',st.borderTopColor,source('outline.color'),true,Number(source('outline.opacity')||1));compare('outline width',st.borderTopWidth,source('outline.width'));
        const ring=el.querySelector('.md-field-focus-outline');if(ring&&state==='focused'){const rs=getComputedStyle(ring);compare('focus ring width',rs.borderTopWidth,get(p+'focus.indicator.outline.thickness')||get('md.sys.state.focus-indicator.thickness'));compare('focus ring color',rs.borderTopColor,get(p+(invalid?'error.':'')+'focus.indicator.outline.color')||get('md.sys.color.secondary'),true);compare('focus label background',getComputedStyle(label).backgroundColor,'transparent',true);compare('focus ring visibility',rs.opacity,'1');const rr=ring.getBoundingClientRect(),er=el.getBoundingClientRect(),offset=Number.parseFloat(rs.borderTopWidth)+Number.parseFloat(get('md.sys.state.focus-indicator.outer-offset'));compare('focus ring left',String(rr.left),String(er.left-offset));compare('focus ring bottom',String(rr.bottom),String(er.bottom+offset));}
      }
      const menu=el.querySelector('.md-field-menu');if(menu?.matches(':popover-open')) {
        const prefix='md.comp.'+variant+'-'+f+'.menu.',st=getComputedStyle(menu);
        compare('popup color',st.backgroundColor,get(prefix+'container.color'),true);shadow(menu,Number(get(prefix+'container.elevation')));
        for(const option of menu.querySelectorAll('[data-field-option]:not([hidden])')) {
          const os=getComputedStyle(option);compare('option height',os.minHeight,get(prefix+'list-item.container.height'));compare('option label',os.color,get(prefix+'list-item.label-text.color'),true);compare('option font size',os.fontSize,get(prefix+'list-item.label-text.size'));
          if(option.getAttribute('aria-selected')==='true')compare('option selected color',os.backgroundColor,get(prefix+'list-item.selected.container.color'),true);
          for(const slot of ['leading','trailing']) {const icon=option.querySelector('.md-field-menu-'+slot);if(icon){compare('option '+slot+' icon',getComputedStyle(icon).color,get(prefix+'list-item.with-'+slot+'-icon.'+slot+'-icon.color'),true);compare('option '+slot+' size',getComputedStyle(icon).width,get(prefix+'list-item.with-'+slot+'-icon.'+slot+'-icon.size'));}}
        }
        const divider=getComputedStyle(menu.querySelector('hr'));compare('popup divider color',divider.borderTopColor,get(prefix+'divider.color'),true);compare('popup divider height',divider.borderTopWidth,get(prefix+'divider.height'));
      }
    } else if(f==='menu'&&el.closest('.md-menu').dataset.variant==='baseline') {
      if(!el.getClientRects().length)continue;
      const menu=el.closest('.md-menu'),p='md.comp.menu.',item=p+'list-item.',state=el.dataset.componentState||(el.disabled?'disabled':el.matches(':focus-visible')?'focused':el.matches(':hover')?'hovered':'enabled'),short={hovered:'hover',focused:'focus'}[state]||state;
      shadow(menu,Number(get(p+'container.elevation')));compare('baseline container color',getComputedStyle(menu).backgroundColor,get(p+'container.color'),true);compare('baseline container shape',getComputedStyle(menu).borderTopLeftRadius,get(p+'container.shape'));
      c('color',get(item+(state==='disabled'?'disabled.':selected?'selected.':['hovered','focused','pressed'].includes(state)?short+'.':'')+'label-text.color'),state==='disabled'?Number(get(item+'disabled.label-text.opacity')):1);
      c('backgroundColor',selected?get(item+'selected.container.color'):'transparent');
      value('minHeight',get(item+'container.height'));typography(el,item+'label-text.');
      for(const slot of ['leading','trailing']){
        const node=el.querySelector('[data-slot="'+slot+'-icon"]');if(!node)continue;
        const prefix=item+'with-'+slot+'-icon.',color=state==='disabled'?get(prefix+'disabled.'+slot+'-icon.color'):selected?get(item+'selected.with-leading-icon.'+slot+'-icon.color'):first([prefix+short+'.icon.color',prefix+slot+'-icon.color']);
        compare(slot+' icon color',getComputedStyle(node).color,color,true,state==='disabled'?Number(get(prefix+'disabled.'+slot+'-icon.opacity')):1);
        compare(slot+' icon size',getComputedStyle(node).width,get(prefix+slot+'-icon.size'));
        const glyph=[...node.querySelectorAll('svg')].find(x=>x.getClientRects().length);compare(slot+' visible glyph',glyph?'1':'0','1');if(glyph)compare(slot+' glyph size',getComputedStyle(glyph).width,get(prefix+slot+'-icon.size'));
      }
      if(['hovered','focused','pressed'].includes(state))layer(p,get(item+short+'.state-layer.color'),get(item+short+'.state-layer.opacity'));
      if(state==='disabled')compare('disabled state opacity',getComputedStyle(el,'::before').opacity,'0');
      if(state==='focused'){value('outlineWidth',get(p+'focus.indicator.thickness'));value('outlineOffset',get(p+'focus.indicator.outline.offset'));c('outlineColor',get(p+'focus.indicator.color'));}
      for(const line of menu.querySelectorAll('.md-divider')){const st=getComputedStyle(line);compare('baseline divider color',st.borderTopColor,get(p+'divider.color'),true);compare('baseline divider height',st.borderTopWidth,get(p+'divider.height'));}
    } else if (f === "menu") {
      if(!el.getClientRects().length)continue;
      const menu=el.closest('.md-menu'),variant=menu.dataset.variant||'standard',icon=el.classList.contains('md-menu-icon-button'),p='md.comp.menus.'+variant+'.'+(icon?'icon-button.':'menu-item.');
      const state=el.dataset.componentState||(el.disabled?'disabled':el.matches(':focus-visible')?'focused':el.matches(':hover')?'hovered':'enabled');
      const prefixes=state==='disabled'?[sel+'.disabled.',sel+'.','disabled.','']:[sel+'.'+state+'.',sel+'.',state+'.',''];
      const source=key=>first(prefixes.map(s=>p+s+key)),itemPrefix='md.comp.menus.'+variant+'.menu-item.';
      const opacity=key=>first(prefixes.map(s=>itemPrefix+s+key),'1');
      shadow(menu,Number(get('md.comp.menus.container.elevation')));
      if(!(icon&&state==='disabled'))c('color',source(icon?'icon.color':'label-text.color'),opacity('label-text.opacity'));
      c('backgroundColor',source('container.color')||first([p+'color'],'transparent'),opacity('container.opacity'));
      compare('menu container color',getComputedStyle(menu).backgroundColor,get('md.comp.menus.'+variant+'.container.color'),true);
      const heading=menu.querySelector('h3');
      if(heading){const style=getComputedStyle(heading);compare('section label minimum height',style.minBlockSize,'32px');compare('section label horizontal inset',style.paddingInlineStart,get('md.sys.measurement.space150'));compare('section label vertical padding',style.paddingTop,'0px');}
      for(const slot of ['leading-icon','trailing-icon','supporting-text','trailing-supporting-text']){
        const child=el.querySelector('[data-slot="'+slot+'"]');if(!child)continue;
        const expected=icon?state==='disabled'?(variant==='standard'?selected?get(p+'disabled.icon.color'):get('md.comp.menus.standard.disabled.icon-button.icon.color'):source('icon.color')):source('icon.color'):source(slot+'.color');
        compare(slot+' color',getComputedStyle(child).color,expected,true,Number(opacity(slot+'.opacity')));
        if(slot.endsWith('icon')){compare(slot+' size',getComputedStyle(child).width,get('md.comp.menus.menu-item.'+slot+'.size'));const glyph=[...child.querySelectorAll('svg')].find(x=>x.getClientRects().length);compare(slot+' visible glyph',glyph?'1':'0','1');if(glyph)compare(slot+' glyph size',getComputedStyle(glyph).width,get('md.comp.menus.menu-item.'+slot+'.size'));}
        else typography(child,'md.comp.menus.menu-item.'+slot+'.');
      }
      if(icon&&state==='disabled')c('color',variant==='standard'?(selected?get(p+'disabled.icon.color'):get('md.comp.menus.standard.disabled.icon-button.icon.color')):source('icon.color'),opacity('label-text.opacity'));
      value('minHeight',get('md.comp.menus.menu-item.height'));typography(el,'md.comp.menus.menu-item.label-text.');
      if(['hovered','focused','pressed'].includes(state))layer(p,source('state-layer.color'),first(prefixes.map(s=>itemPrefix+s+'state-layer.opacity')));
      else if(state!=='disabled'&&el.getAttribute('aria-expanded')==='true')layer(p,get(itemPrefix+'active.state-layer.color'),get(itemPrefix+'active.state-layer.opacity'));
      if(state==='disabled')compare('disabled state opacity',getComputedStyle(el,'::before').opacity,'0');
      const horizontal=menu.dataset.orientation==='horizontal',shape=horizontal?'md.comp.menus.horizontal.container.shape':menu.dataset.menuActivity?'md.comp.menus.'+menu.dataset.menuActivity+'.container.shape':'md.comp.menus.container.shape';
      compare('menu container shape',getComputedStyle(menu).borderTopLeftRadius,get(shape));
      if(horizontal){const rootStyle=getComputedStyle(menu.querySelector('section'));
        compare('horizontal top padding',rootStyle.paddingTop,get('md.comp.menus.horizontal.container.top-space'));
        compare('horizontal bottom padding',rootStyle.paddingBottom,get('md.comp.menus.horizontal.container.bottom-space'));
        const prefix='md.comp.menus.horizontal.'+(icon?'icon-only.':'')+'menu-item.';
        for(const [property,key] of [['paddingTop','top-space'],['paddingBottom','bottom-space'],['paddingLeft','leading-space'],['paddingRight','trailing-space']])compare(property,cs[property],get(prefix+key));
      }
    } else if (f === "dialog") {
      if (!el.open) continue;
      const full=el.dataset.variant==='full-screen',p=full?'md.comp.full-screen-dialog.':'md.comp.dialog.';
      shadow(el,Number(get(p+'container.elevation')));
      c('backgroundColor',get(p+'container.color'));
      value('borderTopLeftRadius',get(p+'container.shape'));
      const heading=el.querySelector('h2'),headline=p+(full?'header.headline.':'headline.');
      if(heading){compare('headline color',getComputedStyle(heading).color,get(headline+'color'),true);typography(heading,headline);}
      compare('dialog role',el.getAttribute('role'),full?'dialog':'alertdialog');
      compare('dialog accessible name',heading?el.getAttribute('aria-labelledby'):el.getAttribute('aria-label'),heading?heading.id:(full?'New item':'Edit item'));
      if(full){
        for(const [selector,key,scrolled] of [['.md-full-header','header.',el.dataset.topScrolled],['.md-full-action-bar','action-bar.',el.dataset.bottomScrolled]]){
          const bar=el.querySelector(selector);if(!bar)continue;const style=getComputedStyle(bar);
          compare(key+'height',style.height,get(p+key+'container.height'));
          compare(key+'color',style.backgroundColor,get(p+key+(key==='header.'&&scrolled==='true'?'on-scroll.':'')+'container.color'),true);
          shadow(bar,Number(get(p+key+(scrolled==='true'?'on-scroll.':'')+'container.elevation')));
        }
        const icon=el.querySelector('.md-dialog-close svg');
        compare('close icon size',getComputedStyle(icon).width,get(p+'header.icon.size'));
        compare('close icon color',getComputedStyle(icon).color,get(p+'header.icon.color'),true);
      } else {
        const icon=el.querySelector('.md-basic-header > [data-slot="icon"]');
        if(icon){compare('icon size',getComputedStyle(icon).width,get(p+'with-icon.icon.size'));compare('icon color',getComputedStyle(icon).color,get(p+'with-icon.icon.color'),true);}
      }
      for(const node of el.querySelectorAll('h3[data-slot="subhead"],p[data-slot="supporting-text"]')){
        const slot=node.dataset.slot,style=getComputedStyle(node),prefix=full?'md.sys.typescale.'+(slot==='subhead'?'title-medium.':'body-large.'):p+slot+'.';
        typography(node,prefix);
        if(!full)compare(slot+' color',style.color,get(prefix+'color'),true);
      }
      compare('entry duration',cs.getPropertyValue('--_dialog-enter-duration').trim(),get('md.sys.motion.duration.'+(full?'medium2':'short3')));
      compare('exit duration',cs.getPropertyValue('--_dialog-exit-duration').trim(),get('md.sys.motion.duration.'+(full?'short4':'short3')));
      for(const button of el.querySelectorAll('.md-text-button')){
        const q=p+(full?(button.closest('.md-full-header')?'header.action.':'action-bar.'):'action.'),style=getComputedStyle(button);
        typography(button,q+'label-text.');
        const state=button.dataset.componentState||'enabled',suffix={hovered:'hover.',focused:'focus.',pressed:'pressed.'}[state]||'';
        compare('action label',style.color,get(q+suffix+'label-text.color'),true);
        if(suffix){const before=getComputedStyle(button,'::before');compare('action state color',before.backgroundColor,get(q+suffix+'state-layer.color'),true);compare('action state opacity',before.opacity,get(q+suffix+'state-layer.opacity'));}
      }
      const body=el.querySelector('[data-dialog-content]');
      for(const side of ['Left','Right'])compare('body '+side+' padding',getComputedStyle(body)['padding'+side],full?get('md.sys.measurement.space300'):get('md.sys.measurement.space100'));
      compare('body horizontal overflow',String(body.scrollWidth>body.clientWidth),'false');
      if(full)compare('header element gap',getComputedStyle(el.querySelector('.md-full-header')).gap,get('md.sys.measurement.space100'));
      else {
        for(const side of ['Top','Right','Bottom','Left'])compare('outer '+side+' padding',cs['padding'+side],get('md.sys.measurement.space300'));
        compare('title to body gap',cs.rowGap,get('md.sys.measurement.space200'));
        compare('body top padding',getComputedStyle(body).paddingTop,'0px');
        compare('body to actions gap',String(Number.parseFloat(getComputedStyle(body).paddingBottom)+Number.parseFloat(cs.rowGap)),get('md.sys.measurement.space300'));
        const icon=el.querySelector('.md-basic-header > [data-slot="icon"]');
        if(icon&&heading)compare('icon to headline',String(heading.getBoundingClientRect().top-icon.getBoundingClientRect().bottom),get('md.sys.measurement.space200'));
      }
      for(const button of el.querySelectorAll('button')){
        const target=getComputedStyle(button,'::after');checked++;
        if(Number.Number.parseFloat(target.width)<48 || Number.Number.parseFloat(target.height)<48)failures.push({id:el.dataset.componentCase,property:'dialog target below 48px'});
      }
      for(const divider of el.querySelectorAll('.md-divider')){
        const style=getComputedStyle(divider),prefix=p+'with-divider.divider.';
        compare('divider color',style.borderTopColor,get(prefix+'color'),true);compare('divider height',style.borderTopWidth,get(prefix+'height'));
      }
    } else if (f === "time") {
      const panel = el.querySelector(".md-time-panel"); if (!panel.open) continue;
      const input = panel.dataset.mode === "input", p = "md.comp." + (input ? "time-input." : "time-picker."), style = getComputedStyle(panel);
      compare("time container color", style.backgroundColor, get(p + "container.color"), true);
      compare("time container shape", style.borderTopLeftRadius, get(p + "container.shape"));
      shadow(panel, Number(get(p + "container.elevation")));
      const title = panel.querySelector(".md-time-title");
      compare("time headline", getComputedStyle(title).color, get(p + "headline.color"), true);
      typography(title, p + "headline.");
      for (const [selector,key] of [[".md-time-value","time-selector."],[".md-time-period","period-selector."],[".md-time-field","time-input-field."]]) for (const node of panel.querySelectorAll(selector)) {
        if (!node.getClientRects().length) continue;
        const style = getComputedStyle(node), selected = node.getAttribute("aria-pressed") === "true", state = node.dataset.componentState || (node.hasAttribute("data-pressed") ? "pressed" : node.matches(":focus-visible") ? "focused" : node.matches(":hover") ? "hovered" : "enabled"), short = {hovered:"hover",focused:"focus"}[state] || state,
          source = property => first([p + key + (selected ? "selected." : "unselected.") + short + "." + property, p + key + short + "." + property, p + key + (selected ? "selected." : "unselected.") + property, p + key + property]);
        compare(key + "label", style.color, source("label-text.color"), true);
        let background = source("container.color") || "transparent";
        if (input && key === "time-input-field." && state === "hovered") {
          const base = color(background), layer = color(source("state-layer.color")), opacity = Number(source("state-layer.opacity"));
          background = "color(srgb " + base.slice(0,3).map((value,i) => value * (1-opacity) + layer[i] * opacity).join(" ") + ")";
        }
        compare(key + "container", style.backgroundColor, background, true);
        typography(node, p + key + "label-text.");
        if (key !== "period-selector.") {
          for (const [property,dimension] of [["width","width"],["height","height"]]) compare(key + property, style[property], first([...(panel.dataset.format === "24h" && panel.dataset.orientation === "portrait" && key === "time-selector." ? [p + key + "24h-vertical.container." + dimension] : []), p + key + "container." + dimension]));
          compare(key + "shape", style.borderTopLeftRadius, get(p + key + "container.shape"));
        }
        if (!input && ["hovered","focused","pressed"].includes(state) || input && key === "period-selector." && ["hovered","focused","pressed"].includes(state)) {
          const layer = getComputedStyle(node,"::before");
          compare(key + "state color", layer.backgroundColor, source("state-layer.color"), true);
          compare(key + "state opacity", layer.opacity, source("state-layer.opacity"));
        }
        if (input && key === "time-input-field." && state === "focused") {
          compare("input focus border", style.borderTopWidth, get(p + key + "focus.outline.width"));
          compare("input focus border color", style.borderTopColor, get(p + key + "focus.outline.color"), true);
        }
      }
      const period = panel.querySelector(".md-time-period-group");
      if (period.getClientRects().length) {
        const style = getComputedStyle(period), q = p + "period-selector.", orientation = panel.dataset.orientation === "landscape" ? "horizontal." : "vertical.";
        for (const dimension of ["width","height"]) compare("period " + dimension, style[dimension], first([p + (input ? "period-selector." : "period-selector." + orientation) + "container." + dimension]));
        compare("period outline", style.borderTopWidth, get(q + "outline.width"));
        compare("period outline color", style.borderTopColor, get(q + "outline.color"), true);
        compare("period shape", style.borderTopLeftRadius, get(q + "container.shape"));
      }
      if (!input) {
        const q = p + "clock-dial.", dial = panel.querySelector(".md-time-dial"), style = getComputedStyle(dial);
        compare("dial size", style.width, get(q + "container.size"));
        compare("dial color", style.backgroundColor, get(q + "color"), true);
        compare("dial shape", style.borderTopLeftRadius, get(q + "shape"));
        for (const [selector,key] of [[".md-time-hand","track"],[".md-time-handle","handle"],[".md-time-center","center"]]) {
          const part = getComputedStyle(panel.querySelector(selector)), prefix = q + "selector." + key + ".container.";
          compare(key + " color", part.backgroundColor, get(prefix + "color"), true);
          compare(key + " width", part.width, get(prefix + (key === "track" ? "width" : "size")));
          if (key !== "track") compare(key + " shape", part.borderTopLeftRadius, get(prefix + "shape"));
        }
        const label = panel.querySelector(".md-time-labels .md-time-tick");
        typography(label, q + "label-text.");
        compare("dial label color", getComputedStyle(label).color, get(q + "unselected.label-text.color"), true);
        compare("dial selected label", getComputedStyle(panel.querySelector(".md-time-selected-labels")).color, get(q + "selected.label-text.color"), true);
      }
    } else if (f === "date") {
      const panel = el.querySelector(".md-date-panel"); if (!panel.open) continue;
      const input = panel.dataset.mode === "input", docked = panel.dataset.variant === "docked", range = panel.dataset.range === "true",
        p = input ? "md.comp.date-input.modal." : "md.comp.date-picker." + (docked ? "docked." : "modal."), style = getComputedStyle(panel),
        source = key => first([...(range ? [p + "range-selection." + key] : []), p + key]);
      compare("date container color", style.backgroundColor, source("container.color"), true);
      compare("date container shape", style.borderTopLeftRadius, panel.dataset.variant === "full-screen" && !input ? get("md.comp.full-screen-dialog.container.shape") : source("container.shape"));
      shadow(panel, Number(panel.dataset.variant === "full-screen" && !input ? get("md.comp.full-screen-dialog.container.elevation") : source("container.elevation")));
      if(input) {
        const header=panel.querySelector('.md-date-header'),fields=panel.querySelector('.md-date-inputs'),actions=panel.querySelector('.md-date-actions');
        const expected=header.getBoundingClientRect().height+fields.getBoundingClientRect().height+actions.getBoundingClientRect().height;
        compare('input content height',style.height,String(Math.max(expected,Number.parseFloat(source('container.height')))));
        const field=fields.querySelector('.md-outlined-text-field'),label=field.querySelector('[data-slot="label"]'),hint=header.querySelector('p'),headline=header.querySelector('h3');
        const rect=node=>node.getBoundingClientRect();
        compare('divider to floating label',String(rect(label).top-rect(header).bottom),'10px');
        compare('header top padding',String(rect(hint).top-rect(header).top),get('md.sys.measurement.space200'));
        compare('headline to divider',String(rect(header).bottom-Number.parseFloat(getComputedStyle(header).borderBottomWidth)-rect(headline).bottom),'10px');
        compare('supporting text to headline',String(rect(headline).top-rect(hint).bottom),get('md.sys.measurement.space450'));
        compare('input left padding',String(rect(field).left-rect(panel).left),get('md.sys.measurement.space300'));
        compare('input bottom padding',getComputedStyle(fields).paddingBottom,get('md.sys.measurement.space200'));
        compare('actions top padding',getComputedStyle(actions).paddingTop,get('md.sys.measurement.space100'));
        compare('actions bottom padding',getComputedStyle(actions).paddingBottom,get('md.sys.measurement.space150'));
        const lastField=[...fields.querySelectorAll('.md-outlined-text-field')].at(-1),firstAction=actions.querySelector('button'),lastAction=actions.querySelector('button:last-child');
        if(!panel.querySelector('[data-date-error]').textContent)compare('field to action',String(rect(firstAction).top-rect(lastField).bottom),'24px');
        compare('actions to panel bottom',String(rect(panel).bottom-rect(lastAction).bottom),get('md.sys.measurement.space150'));
        compare('input uses dialog footer',getComputedStyle(actions).display,'flex');
        const fullActions=panel.querySelector('.md-date-fullscreen-actions');if(fullActions)compare('input hides full-screen actions',getComputedStyle(fullActions).display,'none');
      }
      if(docked) {
        const nav=panel.querySelector('.md-date-docked-navigation'),menuPane=['month','year'].includes(panel.dataset.pane),rows=panel.querySelectorAll('.md-date-calendars .md-date-row').length,
          expectedHeight=Number.parseFloat(source('container.height'))-(menuPane?0:(6-rows)*Number.parseFloat(source('date.container.height'))),
          maxHeight=document.documentElement.clientHeight-2*Number.parseFloat(get('md.sys.measurement.space200'));
        compare('docked height',style.height,String(Math.min(expectedHeight,maxHeight)));
        compare('docked header height',getComputedStyle(nav).height,source('header.height'));
        compare('docked action visibility',getComputedStyle(panel.querySelector('.md-date-actions')).display,menuPane?'none':'flex');
        for(const menu of nav.querySelectorAll('[data-date-menu]')) {
          compare('menu expanded',menu.getAttribute('aria-expanded'),String(panel.dataset.pane===menu.dataset.dateMenu));
          if(menuPane&&panel.dataset.pane!==menu.dataset.dateMenu)compare('inactive menu disabled',String(menu.disabled),'true');
        }
        if(menuPane) {
          const list=panel.querySelector(panel.dataset.pane==='year'?'[data-date-years]':'[data-date-months]');
          compare('menu starts below header',String(list.getBoundingClientRect().top),String(nav.getBoundingClientRect().bottom));
          compare('menu scrolls independently',getComputedStyle(list).overflowY,'auto');
        } else {
          checked++;
          if(panel.querySelector('.md-date-actions').getBoundingClientRect().bottom>panel.getBoundingClientRect().bottom+.1)failures.push({id:el.dataset.componentCase,property:'clipped date actions'});
        }
      }
      if (!docked) for (const [selector,key] of [[".md-date-header h3","headline"],[".md-date-header p","supporting-text"]]) {
        const node = panel.querySelector(selector); if (!node) continue;
        compare(key + " color", getComputedStyle(node).color, source("header." + key + ".color"), true);
        typography(node, p + (range && get(p + "range-selection.header." + key + ".size") ? "range-selection." : "") + "header." + key + ".");
      }
      for (const button of panel.querySelectorAll(".md-date-day")) {
        if (!button.getClientRects().length) continue;
        const style = getComputedStyle(button), state = button.dataset.componentState || (button.hasAttribute("data-pressed") ? "pressed" : (button.matches(":focus-visible") || button.querySelector("input")?.matches(":focus-visible")) ? "focused" : button.matches(":hover") ? "hovered" : "enabled"), short = {hovered:"hover",focused:"focus"}[state] || state,
          selected = button.dataset.dateSelected === "true", today = button.dataset.today === "true", inRange = button.dataset.inRange === "true", outside = button.dataset.outsideMonth === "true",
          dateSource = key => first([
            ...((range && inRange && !selected) ? [p + "range-selection.date.in-range." + short + "." + key, p + "range-selection.date.in-range." + key] : []),
            p + "date." + (selected ? "selected." : today ? "today." : "unselected.") + short + "." + key,
            ...(outside && !selected ? [p + "date.unselected.outside-month." + key] : []),
            p + "date." + (selected ? "selected." : today ? "today." : "unselected.") + key,
            p + "date.unselected." + key, p + "date." + short + "." + key, p + "date." + key]);
        const target=getComputedStyle(button,'::after'),rowHeight=button.closest('.md-date-cell').getBoundingClientRect().height;
        checked++;
        if(Number.Number.parseFloat(target.width)<48 || Number.Number.parseFloat(target.height)<48 || rowHeight<48)failures.push({id:el.dataset.componentCase,property:'date target below 48px or overlapping rows'});
        compare("day width", style.width, get(p + "date.container.width"));
        compare("day height", style.height, get(p + "date.container.height"));
        compare("day color", style.color, dateSource("label-text.color"), true, Number(dateSource("label-text.opacity") || 1));
        compare("day background", style.backgroundColor, dateSource("container.color") || "transparent", true);
        typography(button, p + "date.label-text.");
        if (["hovered","focused","pressed"].includes(state)) {
          const layer = getComputedStyle(button, "::before");
          compare("day state color", layer.backgroundColor, dateSource("state-layer.color"), true);
          compare("day state opacity", layer.opacity, dateSource("state-layer.opacity"));
          compare("day state width", layer.width, get(p + "date.state-layer.width"));
          compare("day state height", layer.height, get(p + "date.state-layer.height"));
        }
      }
      for (const [selector,part] of [[".md-date-year","year-selection.year."],[".md-date-month-item","menu.list-item."]]) for (const button of panel.querySelectorAll(selector)) {
        if (!button.getClientRects().length) continue;
        const style = getComputedStyle(button), selected = button.getAttribute("aria-pressed") === "true", state = button.dataset.componentState || (button.disabled ? "disabled" : button.hasAttribute("data-pressed") ? "pressed" : (button.matches(":focus-visible") || button.querySelector("input")?.matches(":focus-visible")) ? "focused" : button.matches(":hover") ? "hovered" : "enabled"), short = {hovered:"hover",focused:"focus"}[state] || state,
          value = key => first([p + part + (selected ? "selected." : "unselected.") + short + "." + key, p + part + short + "." + key, p + part + (selected ? "selected." : "unselected.") + key, p + part + key]);
        compare(part + "label", style.color, value("label-text.color"), true, Number(value("label-text.opacity") || 1));
        compare(part + "background", style.backgroundColor, value("container.color") || "transparent", true);
        typography(button, p + part + "label-text.");
        if (["hovered","focused","pressed"].includes(state)) {
          const layer = getComputedStyle(button, "::before");
          compare(part + "state color", layer.backgroundColor, value("state-layer.color"), true);
          compare(part + "state opacity", layer.opacity, value("state-layer.opacity"));
        }
      }
    } else if (f === "search") {
      const panel=el.querySelector('.md-search-view'),bar=panel.querySelector('.md-search-bar'),input=panel.querySelector('.md-search-input'),results=panel.querySelector('.md-search-results'),open=panel.dataset.open==='true',contained=panel.dataset.control==='contained',full=panel.dataset.variant==='full-screen',p='md.comp.search-bar.',q='md.comp.search-view.',ps=getComputedStyle(panel),bs=getComputedStyle(bar),rs=getComputedStyle(results),scene=el.querySelector('.md-search-scene');
      if(!scene.getClientRects().length)continue;
      if(panel.dataset.layout==='adaptive'&&scene.getClientRects().length){
        compare('adaptive search belongs to inbox',String(scene.parentElement.classList.contains('md-search-app')),'true');
        compare('adaptive search follows inbox width',String(Math.abs(panel.getBoundingClientRect().width-(scene.clientWidth-2*Number.parseFloat(ps.left)))<1),'true');
      }
      compare('one persistent input',String(el.querySelectorAll('.md-search-input').length),'1');
      compare('one persistent header',String(el.querySelectorAll('.md-search-bar').length),'1');
      compare('no popup dialog',String(el.querySelectorAll('dialog.md-search-view').length),'0');
      compare('focus disclosure',input.getAttribute('aria-expanded'),String(open));
      compare('search header height',bs.height,open?first([...(contained?[q+'contained.'+panel.dataset.variant+'.bar.container.height']:[]),q+panel.dataset.variant+'.header.container.height']):get(p+'container.height'));
      compare('header fill',bs.backgroundColor,get(open?q+'container.color':p+'container.color'),true);
      const elevated=panel.dataset.elevation==='token';
      if(!open){compare('closed wrapper has no duplicate shadow',ps.boxShadow,'none');if(elevated)shadow(bar,Number(get(p+'container.elevation')));else compare('flat search bar',bs.boxShadow,'none');}
      else if(!elevated||full){for(const [name,node] of [['view',panel],['header',bar],['results',results]])compare('flat '+name,getComputedStyle(node).boxShadow,'none');}
      else if(contained){compare('contained wrapper has no shadow',ps.boxShadow,'none');for(const [name,node] of [['bar',bar],['results',results]])compare(name+' casts no shadow over sibling',getComputedStyle(node).boxShadow,'none');for(const pseudo of ['::before','::after']){shadow(panel,Number(get(q+'container.elevation')),pseudo);compare('surface shadow behind content',getComputedStyle(panel,pseudo).zIndex,'-1');}}
      else{shadow(panel,Number(get(q+'container.elevation')));compare('no duplicate header shadow',bs.boxShadow,'none');compare('no duplicate results shadow',rs.boxShadow,'none');}
      if(bar.dataset.componentState==='focused'||input.matches(':focus-visible')){compare('unclipped focus indicator',ps.overflowX,'visible');compare('focus thickness',bs.outlineWidth,get(p+'focus.indicator.thickness'));compare('focus offset',bs.outlineOffset,get(p+'focus.indicator.outline.offset'));}
      for(const avatar of bar.querySelectorAll('.md-search-avatar')) {
        const image=avatar.querySelector(':scope > span:not(.md-button-ripple)'),a=avatar.getBoundingClientRect(),i=image.getBoundingClientRect();
        compare('avatar image size',String(i.width),get(p+'avatar.size'));
        compare('avatar centered horizontally',String(i.left+i.width/2),String(a.left+a.width/2));
        compare('avatar centered vertically',String(i.top+i.height/2),String(a.top+a.height/2));
        const ripple=avatar.querySelector('.md-button-ripple');
        if(ripple){const r=ripple.getBoundingClientRect();compare('avatar ripple width',String(r.width),String(a.width));compare('avatar ripple height',String(r.height),String(a.height));compare('avatar ripple left',String(r.left),String(a.left));compare('avatar ripple top',String(r.top),String(a.top));}
      }
      if(!open&&bar.dataset.leading==='none'){
        const noActions=contained&&![...bar.querySelectorAll('.md-search-action-target:not([hidden]) button,.md-search-avatar')].some(b=>b.getClientRects().length);
        compare('no-leading padding',bs.paddingInlineStart,get(p+(noActions?'contained.no-actions.':contained?'contained.':'')+'leading-space'));
        if(noActions)compare('no-actions trailing padding',bs.paddingInlineEnd,get(p+'contained.no-actions.trailing-space'));
      }
      typography(input,(open?q+'header.':p)+'input-text.');
      compare('input color',getComputedStyle(input).color,get((open?q+'header.':p)+'input-text.color'),true);
      if(open){
        const app=el.querySelector('.md-search-app');
        if(app){const scrim=el.querySelector('.md-search-scrim'),a=app.getBoundingClientRect(),r=scrim.getBoundingClientRect();for(const dimension of ['left','top','width','height'])compare('app scrim '+dimension,String(r[dimension]),String(a[dimension]));for(const region of el.querySelectorAll('.md-search-mail-navigation,.md-search-mail-detail'))compare('background is inert',String(region.hasAttribute('inert')),'true');}
        compare('focused header has no shared hover layer',getComputedStyle(bar,'::before').content,'none');
        compare('canvas contains search',String(panel.getBoundingClientRect().left>=scene.getBoundingClientRect().left&&panel.getBoundingClientRect().right<=scene.getBoundingClientRect().right),'true');
        if(full)compare('full-screen height',String(panel.getBoundingClientRect().height),String(scene.clientHeight));
        else compare('docked height limit',String(panel.getBoundingClientRect().height<=scene.clientHeight*2/3+1),'true');
        if(contained&&!full){
          compare('separate contained surfaces',ps.backgroundColor,'transparent',true);
          compare('results fill',rs.backgroundColor,get(q+'container.color'),true);
          compare('bar/results gap',ps.rowGap,get(q+'contained.docked.bar-results.gap'));
          compare('results corners',rs.borderTopLeftRadius,get(q+'contained.docked.results.shape'));
        }else{
          compare('focused surface fill',ps.backgroundColor,get(q+(contained?'contained.background.color':'container.color')),true);
          if(!contained)compare('divided rule',rs.borderTopColor,get(q+'divider.color'),true);
        }
        for(const button of results.querySelectorAll('[data-search-result] button'))compare('result Tab stop',String(button.tabIndex),'0');
      }
    } else if (f === "list") {
      const p = "md.comp.list.", item = p + "list-item.", expressive = el.dataset.shape === "expressive", segmented = el.dataset.variant === "segmented";
      value("borderTopLeftRadius", get(p + (el.dataset.control === "expand" ? "expand." : "") + "container.shape"));
      value("rowGap", segmented ? get(p + "segmented.gap") : "0px");
      for(const group of el.querySelectorAll(':scope > li[data-list-expanded="true"]')) {
        const panel=group.querySelector('.md-list-expand-content'),parent=group.querySelector('.md-list-row > .md-list-item'),inset=el.dataset.expandInset==='inset'?Number.parseFloat(get(item+'leading-space')):0;
        if(panel&&!panel.hidden){const ps=getComputedStyle(panel);compare('expanded content inset',ps.paddingInlineStart,String(inset));compare('expanded group background',getComputedStyle(group).backgroundColor,el.dataset.expandSurface==='none'?'transparent':el.dataset.expandSurface==='tonal'?get('md.sys.color.surface-container'):get(p+'expand.expanded.list-item.'+(segmented?'segmented.':'')+'container.color'),true);
          for(const child of panel.querySelectorAll('.md-list-item')){compare('expanded child width',String(child.getBoundingClientRect().width),String(parent.getBoundingClientRect().width-inset));compare('expanded label alignment',String(child.querySelector('.md-list-label').getBoundingClientRect().left),String(parent.querySelector('.md-list-label').getBoundingClientRect().left+inset));}}
      }
      const firstRow=el.querySelector(':scope > li');
      if(firstRow){const divider=getComputedStyle(firstRow,'::after'),visible=!segmented&&el.dataset.divider!=='none';
        compare('divider visibility',divider.content!=='none'?'visible':'hidden',visible?'visible':'hidden');
        if(visible){compare('divider thickness',divider.borderBlockStartWidth,get(p+'divider.height'));compare('divider color',divider.borderBlockStartColor,get(p+'divider.color'),true);for(const [property,key] of [['marginInlineStart','leading-space'],['marginInlineEnd','trailing-space']])compare('divider '+key,divider[property],el.dataset.divider==='full-width'?'0px':get(p+'divider.'+key));}}
      for (const button of el.querySelectorAll(".md-list-item")) {
        if (!button.getClientRects().length) continue;
        const style = getComputedStyle(button), selected = button.getAttribute("aria-pressed") === "true" || button.dataset.selected === "true", disabled = button.disabled || button.getAttribute('aria-disabled') === 'true',
          state = button.dataset.componentState || (disabled ? "disabled" : button.hasAttribute("data-dragged") ? "dragged" : button.hasAttribute("data-pressed") ? "pressed" : (button.matches(":focus-visible") || button.querySelector("input")?.matches(":focus-visible")) ? "focused" : button.matches(":hover") ? "hovered" : "enabled"),
          short = {hovered: "hover", focused: "focus"}[state] || state,
          revealed = button.closest("[data-list-row]").dataset.revealed === "true",
          expanded = button.getAttribute("aria-expanded") === "true",
          source = key => first([
            ...(el.dataset.control === "reorder" && state === "dragged" ? [p + "reorder.list-item." + key] : []),
            ...(selected ? [item + "selected." + short + "." + key, item + "selected." + state + "." + key, item + "selected." + key] : []),
            ...(revealed ? [p + "reveal.list-item." + (segmented ? "segmented." : "") + key, p + "reveal.list-item." + key] : []),
            ...(expanded ? [p + "expand.expanded.list-item." + (segmented ? "segmented." : "") + key, p + "expand.expanded.list-item." + key] : []),
            item + short + "." + key, item + state + "." + key,
            ...(segmented ? [item + "segmented." + key] : []), item + "unselected." + key, item + key]);
        const itemBackground = expanded && el.dataset.expandSurface==='none' ? get(item+(segmented?'segmented.':'')+'container.color') : expanded && el.dataset.expandSurface==='tonal' ? 'transparent' : button.closest('.md-list-expand-content') && el.dataset.expandSurface!=='none' ? 'transparent' : source('container.color');
        compare("item background", style.backgroundColor, itemBackground, true, disabled ? Number(source("container.opacity") || 1) : 1);
        const disclosure=button.querySelector('.md-list-expand-icon');
        if(disclosure)compare('disclosure icon background',getComputedStyle(disclosure).backgroundColor,el.dataset.expandIcon==='plain'?'transparent':get(p+'expand.'+(expanded?'expanded':'collapsed')+'.list-item.trailing-icon.container.color'),true);
        compare("item minimum height", style.minHeight, get(item + el.dataset.size + ".container.height"));
        if(!expressive && button.getBoundingClientRect().height>=88)compare('tall baseline alignment',style.alignItems,'flex-start');
        const video=button.querySelector('.md-list-video');
        if(video){const vs=getComputedStyle(video),prefix=item+(el.dataset.media||'small')+'.leading-video.';compare('video width',vs.width,get(prefix+'width'));compare('video height',vs.height,get(prefix+'height'));}
        for (const [property,key] of [["paddingInlineStart","leading-space"],["paddingInlineEnd","trailing-space"],["paddingTop","top-space"],["paddingBottom","bottom-space"],["columnGap","between-space"]]) compare(property, style[property], property === "paddingInlineEnd" && ["reorder","reveal"].includes(el.dataset.control) ? String(Number.parseFloat(get(item + key)) + Number.parseFloat(get("md.comp.icon-button.small.container.height")) + Number.parseFloat(get(item + "between-space"))) : get(item + key));
        const shape = el.dataset.control === "reorder" && state === "dragged" ? get(p + "reorder.list-item.shape") : source("container." + (expressive ? "expressive." : "") + "shape");
        compare("item corner", style.borderTopLeftRadius, shape);
        shadow(button, Number(source("container.elevation")));
        for (const [selector,key] of [[".md-list-label","label-text"],[".md-list-supporting","supporting-text"],[".md-list-overline","overline"],[".md-list-meta","trailing-supporting-text"]]) {
          const text = button.querySelector(selector); if (!text) continue;
          compare(key + " color", getComputedStyle(text).color, source(key + ".color"), true, disabled ? Number(source(key + ".opacity") || (key === "trailing-supporting-text" ? source("supporting-text.opacity") : undefined) || 1) : 1);
          typography(text, item + key + ".");
        }
        for (const [selector,key] of [[".md-list-icon svg","leading-icon"],[".md-list-trailing-icon svg","trailing-icon"]]) {
          const icon = button.querySelector(selector); if (!icon) continue; const iconStyle = getComputedStyle(icon);
          compare(key + " color", iconStyle.color, first([...(selected ? [item + "selected." + short + "." + key + ".color"] : []), ...(state === "dragged" && el.dataset.control === "reorder" ? [p + "reorder.list-item." + key + ".color"] : []), item + short + "." + key + ".icon.color"], source(key + ".color")), true, disabled ? Number(source(key + ".opacity")) : 1);
          compare(key + " size", iconStyle.width, first([...(expressive ? [item + key + ".expressive.size"] : []), item + key + ".size"]));
        }
        if (["hovered","focused","pressed","dragged"].includes(state)) {
          const layer = getComputedStyle(button, "::before");
          compare("item state color", layer.backgroundColor, source("state-layer.color"), true);
          compare("item state opacity", layer.opacity, source("state-layer.opacity"));
        }
        if (state === "focused") {
          compare("item focus width", style.outlineWidth, get(p + "focus.indicator.thickness"));
          compare("item focus offset", style.outlineOffset, get(p + "focus.indicator.outline.offset"));
          compare("item focus color", style.outlineColor, get(p + "focus.indicator.color"), true);
        }
      }
    } else if (f === "loading") {
      if (!el.getClientRects().length) continue;
      const p = "md.comp.loading-indicator.", contained = el.dataset.variant === "contained", svg = el.querySelector("svg"), style = getComputedStyle(svg);
      value("width", get(p + "container.width"));
      value("height", get(p + "container.height"));
      value("borderTopLeftRadius", get(p + "container.shape"));
      c("backgroundColor", contained ? get(p + "contained.container.color") : "transparent");
      c("color", get(p + (contained ? "contained." : "") + "active-indicator.color"));
      compare("shape width", style.width, get(p + "active-indicator.size"));
      compare("shape height", style.height, get(p + "active-indicator.size"));
      checked++;
      const path = el.querySelector("path").getAttribute("d") || "";
      if (!path.startsWith("M") || !path.endsWith("Z") || /NaN|Infinity/.test(path))
        failures.push({id: el.dataset.componentCase, property: "closed finite loading shape", actual: path});
    } else if (f === "carousel") {
      const p = "md.comp.carousel-item.";
      const viewport=el.querySelector('.md-carousel-viewport'),track=el.querySelector('.md-carousel-track'),vr=viewport.getBoundingClientRect(),tr=track.getBoundingClientRect(),vs=getComputedStyle(viewport),rtl=cs.direction==='rtl';
      const offset=rtl?vr.right-Number.parseFloat(vs.paddingInlineStart)-tr.right:tr.left-vr.left-Number.parseFloat(vs.paddingInlineStart);
      checked++;
      if(Math.abs(offset)>.04)failures.push({id:el.dataset.componentCase,property:'stationary item layer',actual:offset,expected:0});
      const visible=[...el.querySelectorAll('.md-carousel-item')].filter(b=>getComputedStyle(b).visibility!=='hidden').map(b=>b.getBoundingClientRect()).sort((a,b)=>a.left-b.left);
      for(let i=1;i<visible.length;i++) {
        // Layout quantization can contribute less than one CSS subpixel per edge.
        checked++;
        const actual=visible[i].left-visible[i-1].right,expected=Number.parseFloat(cs.getPropertyValue('--_carousel-gap'));
        if(Math.abs(actual-expected)>.04)failures.push({id:el.dataset.componentCase,property:'item gap',actual,expected});
      }
      for (const button of el.querySelectorAll(".md-carousel-item")) {
        if (!button.getClientRects().length || getComputedStyle(button).visibility === "hidden") continue;
        const style = getComputedStyle(button), disabled = button.disabled,
          state = button.dataset.componentState || (disabled ? "disabled" : button.hasAttribute("data-pressed") ? "pressed" : (button.matches(":focus-visible") || button.querySelector("input")?.matches(":focus-visible")) ? "focused" : button.matches(":hover") ? "hovered" : "enabled"),
          suffix = {hovered: "hover.", focused: "focus.", pressed: "pressed.", disabled: "disabled."}[state] || "",
          source = key => first([p + suffix + key, p + key]);
        compare("item background", style.backgroundColor, el.dataset.expandSurface==='none'&&button.closest('li[data-list-expanded="true"]')?'transparent':source("container.color"), true, disabled ? Number(source("container.opacity")) : 1);
        compare("item shape", style.borderTopLeftRadius, get(p + "container.shape"));
        shadow(button, Number(source("container.elevation")));
        if (disabled) compare("disabled media opacity", getComputedStyle(button.querySelector(".md-carousel-art")).opacity, source("container.opacity"));
        if (el.dataset.control === "outlined") {
          compare("item outline width", style.borderTopWidth, get(p + "with-outline.outline.width"));
          compare("item outline color", style.borderTopColor, first([p + "with-outline." + suffix + "outline.color", p + "with-outline.outline.color"]), true,
            disabled ? Number(get(p + "with-outline.disabled.outline.opacity")) : 1);
        }
        if (["hovered", "focused", "pressed"].includes(state)) {
          const layer = getComputedStyle(button, "::before");
          compare("item state color", layer.backgroundColor, source("state-layer.color"), true);
          compare("item state opacity", layer.opacity, source("state-layer.opacity"));
        }
        if (state === "focused") {
          compare("item focus width", style.outlineWidth, get(p + "focus.indicator.thickness"));
          compare("item focus offset", style.outlineOffset, get(p + "focus.indicator.outline.offset"));
          compare("item focus color", style.outlineColor, get(p + "focus.indicator.color"), true);
        }
      }
    } else if (f === "progress") {
      if (!el.getClientRects().length) continue;
      const variant = el.dataset.variant, circular = variant.endsWith("circular"), legacy = variant.startsWith("legacy"),
        base = legacy ? "md.comp." + (circular ? "circular" : "linear") + "-progress-indicator." : "md.comp.progress-indicator.",
        kind = legacy ? base : base + variant + ".", thick = el.dataset.size === "thick", wave = el.dataset.shape === "wave",
        indeterminate = el.dataset.control === "indeterminate",
        source = key => first([...(thick ? [kind + "thick." + key] : []), kind + key, base + key]);
      const active = el.querySelector(".md-progress-active"), track = el.querySelector(".md-progress-track"), activeStyle = getComputedStyle(active), trackStyle = getComputedStyle(track),
        thickness = legacy ? get(base + "active-indicator." + (circular ? "width" : "height")) : source("active-indicator.thickness");
      compare("active thickness", activeStyle.strokeWidth, thickness);
      compare("active caps", activeStyle.strokeLinecap, get(base + "active-indicator.shape").includes("infinity") || Number.parseFloat(get(base + "active-indicator.shape")) > 0 ? "round" : "butt");
      if (el.dataset.shape === "four-color") {
        checked++;
        const actual = color(activeStyle.stroke), colors = ["one", "two", "three", "four"].map(name => color(get(base + "four-color.active-indicator." + name + ".color")));
        if (!colors.some(expected => expected.length === actual.length && actual.every((x, i) => Math.abs(x - expected[i]) < .001)))
          failures.push({id: el.dataset.componentCase, property: "four-color stroke", actual: activeStyle.stroke});
      } else compare("active color", activeStyle.stroke, get(base + "active-indicator.color"), true);
      if (circular) {
        const size = first([...(thick ? [kind + "thick.size"] : []), ...(wave ? [kind + "with-wave.size"] : []), kind + "size"]);
        value("width", size); value("height", size);
      } else value("height", (wave && first([...(thick ? [kind + "thick.with-wave.height"] : []), kind + "with-wave.height"])) || source("height") || thickness);
      if (variant !== "legacy-circular") {
        compare("track color", trackStyle.stroke, get(base + "track.color"), true);
        compare("track thickness", trackStyle.strokeWidth, legacy ? get(base + "track.height") : source("track.thickness"));
        const shape = get(base + "track.shape");
        compare("track caps", trackStyle.strokeLinecap, shape.includes("infinity") || Number.parseFloat(shape) > 0 ? "round" : "butt");
      } else compare("legacy track", track.getAttribute("d"), "");
      if (!legacy) {
        compare("track gap", cs.getPropertyValue("--_progress-gap").trim(), source("track-active-indicator-space") || get(base + "active-indicator-track-space"));
        if (wave) {
          const wavelength = first([...(indeterminate ? [kind + "indeterminate.active-indicator.wave.wavelength"] : []), kind + "active-indicator.wave.wavelength"]);
          compare("wavelength", cs.getPropertyValue("--_progress-wavelength").trim(), wavelength);
          compare("amplitude", cs.getPropertyValue("--_progress-amplitude").trim(), get(kind + "active-indicator.wave.amplitude"));
        }
        if (!circular && !indeterminate) {
          const stop = el.querySelector(".md-progress-stop"), stopStyle = getComputedStyle(stop), size = source("stop-indicator.size"), trailing = source("stop-indicator.trailing-space");
          compare("stop color", stopStyle.fill, get(base + "stop-indicator.color"), true);
          compare("stop width", stop.getAttribute("width"), size);
          compare("stop height", stop.getAttribute("height"), size);
          compare("stop trailing space", String(el.clientWidth - Number(stop.getAttribute("x")) - Number(stop.getAttribute("width"))), trailing);
        }
      }
      for (const path of el.querySelectorAll("path")) {
        checked++;
        if (/NaN|Infinity/.test(path.getAttribute("d") || "")) failures.push({id: el.dataset.componentCase, property: "finite path", actual: path.getAttribute("d")});
      }
    } else if (f === "fab-menu") {
      const panel = el.querySelector(".md-fab-menu-panel");
      if (!panel.getClientRects().length) continue;
      const base = "md.comp.fab-menu.", panelStyle = getComputedStyle(panel), list = panel.querySelector("ul");
      compare("close spacing", panelStyle.rowGap, get(base + "close-button.between-space"));
      compare("item spacing", getComputedStyle(list).rowGap, get(base + "menu-item.between-space"));
      const toggle=el.querySelector(".md-fab-menu-opener");
      if(toggle){
        compare("single toggle",String(el.querySelectorAll(".md-fab-menu-opener,.md-fab-menu-close").length),"1");
        const toggleBox=toggle.getBoundingClientRect();
        compare("bounded toggle radius",String(Number.parseFloat(getComputedStyle(toggle).borderTopLeftRadius)<=Math.min(toggleBox.width,toggleBox.height)/2),"true");
        compare("popover close button",String(panel.querySelectorAll(".md-fab-menu-close").length),"0");
        compare("toggle spacing",String(toggle.getBoundingClientRect().top-panel.getBoundingClientRect().bottom),get(base+"close-button.between-space"));
      }
      for (const button of el.querySelectorAll(".md-fab-menu-item,.md-fab-menu-close")) {
        const close = button.classList.contains("md-fab-menu-close"), kind = close ? "close-button." : "menu-item.",
          profile = base + el.dataset.variant + (close ? ".close-button." : "-container.list-item."), style = getComputedStyle(button),
          state = button.dataset.componentState || (button.hasAttribute("data-pressed") ? "pressed" : (button.matches(":focus-visible") || button.querySelector("input")?.matches(":focus-visible")) ? "focused" : button.matches(":hover") ? "hovered" : "enabled"),
          source = key => first([profile + state + "." + key, profile + key, base + kind + key]);
        compare("button height", style.height, get(base + kind + "container.height"));
        compare("button color", style.backgroundColor, source("container.color"), true);
        shadow(button, Number(source("container.elevation")));
        const rect = button.getBoundingClientRect();
        const shape = get(base + kind + "container.shape"), bound = Math.min(rect.height, rect.width) / 2;
        compare("button corner", String(Math.min(Number.parseFloat(style.borderTopLeftRadius), bound)), String(Math.min(shape.includes("infinity") ? Infinity : Number.parseFloat(shape), bound)));
        const iconStyle = getComputedStyle(button.querySelector("svg"));
        compare("icon size", iconStyle.width, get(base + kind + "icon.size"));
        compare("icon color", iconStyle.color, source("icon.color"), true);
        if (close) compare("close width", style.width, get(base + kind + "container.width"));
        else {
          const content = getComputedStyle(button.querySelector(".md-fab-menu-item-content"));
          compare("label color", style.color, source("label-text.color"), true);
          compare("leading padding", content.paddingInlineStart, get(base + kind + "leading-space"));
          compare("trailing padding", content.paddingInlineEnd, get(base + kind + "trailing-space"));
          compare("icon label gap", content.columnGap, get(base + kind + "icon-label-space"));
          const font = get(base + kind + "label-text"), metrics = font.match(/^(\d+) ([\d.]+)px\/([\d.]+)/);
          compare("label weight", style.fontWeight, metrics?.[1]);
          compare("label size", style.fontSize, metrics?.[2]);
          compare("label line height", style.lineHeight, String(Number(metrics?.[2]) * Number(metrics?.[3])));
        }
        if (["hovered", "focused", "pressed"].includes(state)) {
          const layer = getComputedStyle(button, "::before");
          compare("state color", layer.backgroundColor, source("state-layer.color"), true);
          compare("state opacity", layer.opacity, source("state-layer.opacity"));
        }
      }
    } else if (f === "bottom-sheet" || f === "side-sheet") {
      if (!el.getClientRects().length) continue;
      const bottom = f === "bottom-sheet", p = "md.comp.sheet." + (bottom ? "bottom." : "side."),
        docked = p + "docked.", modal = el.dataset.control === "modal";
      c("backgroundColor", first([docked + (modal ? "modal." : "standard.") + "container.color", docked + "container.color"]));
      shadow(el, Number(get(docked + (modal ? "modal." : "standard.") + "container.elevation")));
      const shape = bottom ? docked + (el.dataset.variant === "minimized" ? "minimized." : "") + "container.shape"
        : el.dataset.variant === "detached" ? p + "detached.container.shape"
        : first([docked + (modal ? "modal." : "standard.") + "container.shape"]) ? docked + (modal ? "modal." : "standard.") + "container.shape" : docked + "container.shape";
      const sizes = get(shape).split(/\s+/), corners = [sizes[0], sizes[1] || sizes[0], sizes[2] || sizes[0], sizes[3] || sizes[1] || sizes[0]];
      if(!bottom){
        const width=el.tagName==='DIALOG'?document.documentElement.clientWidth:el.closest('.md-sheet-frame').clientWidth,
          inset=el.dataset.variant==='detached'?Number.parseFloat(get('md.sys.measurement.space200')):0,
          left=(el.dataset.orientation==='start')!==(cs.direction==='rtl');
        compare('side-sheet logical edge',cs.left,String(left?inset:width-Number.parseFloat(cs.width)-inset));
      }
      // Side-sheet geometry is expressed at the trailing edge in the source.
      if (!bottom && ((el.dataset.orientation === "start") !== (cs.direction === "rtl")))
        [corners[0], corners[1], corners[2], corners[3]] = [corners[1], corners[0], corners[3], corners[2]];
      ["TopLeft", "TopRight", "BottomRight", "BottomLeft"].forEach((corner, index) => value("border" + corner + "Radius", corners[index]));
      const focus = bottom ? p + "focus.indicator." : docked + "focus.indicator.";
      function ring(node) {
        const style = getComputedStyle(node);
        compare("focus thickness", style.outlineWidth, get(focus + "thickness"));
        compare("focus offset", style.outlineOffset, get(focus + "outline.offset"));
        compare("focus color", style.outlineColor, get(focus + "color"), true);
      }
      if (bottom) {
        const mark = el.querySelector(".md-sheet-handle-mark"), style = getComputedStyle(mark);
        compare("handle width", style.width, get(docked + "drag-handle.width"));
        compare("handle height", style.height, get(docked + "drag-handle.height"));
        compare("handle color", style.backgroundColor, get(docked + "drag-handle.color"), true, Number(get(docked + "drag-handle.opacity")));
        if (state === "focused" || mark.parentElement.matches(":focus-visible")) ring(mark);
      } else {
        value("width", get(docked + "container.width"));
        const heading = el.querySelector(".md-sheet-title");
        compare("headline color", getComputedStyle(heading).color, get(docked + "headline.color"), true);
        typography(heading, docked + "headline.");
        if (!modal && el.dataset.variant === "docked") {
          value(el.dataset.orientation === "start" ? "borderInlineEndWidth" : "borderInlineStartWidth", get("md.comp.divider.thickness"));
          c(el.dataset.orientation === "start" ? "borderInlineEndColor" : "borderInlineStartColor", get(docked + "divider.color"));
        }
        for (const button of el.querySelectorAll(".md-sheet-action")) {
          const style = getComputedStyle(button), actionState = button.dataset.componentState || (button.disabled ? "disabled" : button.hasAttribute("data-pressed") ? "pressed" : (button.matches(":focus-visible") || button.querySelector("input")?.matches(":focus-visible")) ? "focused" : button.matches(":hover") ? "hovered" : "enabled"),
            suffix = { hovered: "hover.", focused: "focus.", pressed: "pressed." }[actionState], disabled = actionState === "disabled";
          compare("action label", style.color, suffix ? get(docked + "action." + suffix + "label-text.color") : get("md.comp.button.text." + (disabled ? "disabled." : "") + "label-text.color"), true,
            disabled ? Number(get("md.comp.button.text.disabled.label-text.opacity")) : 1);
          if (suffix) {
            const layer = getComputedStyle(button, "::before");
            compare("action state color", layer.backgroundColor, get(docked + "action." + suffix + "state-layer.color"), true);
            compare("action state opacity", layer.opacity, get(docked + "action." + suffix + "state-layer.opacity"));
          }
          if (actionState === "focused") ring(button);
        }
      }
    } else if (f === "divider") {
      value("borderTopWidth", get("md.comp.divider.thickness"));
      c("borderTopColor", get("md.comp.divider.color"));
      const inset=el.dataset.inset,space=get('md.sys.measurement.space200');
      value('marginInlineStart',inset==='true'||inset==='middle'?space:'0px');
      value('marginInlineEnd',inset==='middle'?space:'0px');
      const layout=el.closest('.md-divider-layout');
      if(layout){const ls=getComputedStyle(layout),spacing=layout.dataset.spacing;
        compare('divider available width',cs.width,String(Number.parseFloat(ls.width)-Number.parseFloat(cs.marginInlineStart)-Number.parseFloat(cs.marginInlineEnd)));
        compare('divider outer right margin',ls.marginInlineEnd,['right','right-bottom'].includes(spacing)?get('md.sys.measurement.space100'):'0px');
        compare('divider outer bottom margin',ls.marginBlockEnd,['bottom','right-bottom'].includes(spacing)?get('md.sys.measurement.space100'):'0px');
        const supporting=layout.querySelector('p');if(supporting)compare('divider supporting text gap',getComputedStyle(supporting).marginBlockStart,get('md.sys.measurement.space50'));
      }
    }
  }
  for (const item of document.querySelectorAll(".md-menu .md-menu-item")) {
    if(!item.getClientRects().length)continue;
    const node =
      item.parentElement.tagName === "LI" ? item.parentElement : item;
    const style = getComputedStyle(item);
    const token = (suffix) =>
      style.getPropertyValue("--md-comp-menus-menu-item-" + suffix).trim();
    const selected = item.matches(
      ':checked,[aria-pressed="true"],[aria-selected="true"],[aria-current="page"],[data-selected="true"],:has(> input:checked)',
    );
    const first = node.matches(":first-child"),
      last = node.matches(":last-child");
    const baseline=item.closest('.md-menu').dataset.variant==='baseline';
    const horizontal=item.closest('.md-menu').dataset.orientation==='horizontal',icon=item.classList.contains('md-menu-icon-button'),state=item.dataset.componentState||(item.matches(':focus-visible')?'focused':item.matches(':hover')?'hovered':'enabled');
    const horizontalShape=()=>{
      const prefix='--md-comp-menus-horizontal-'+(icon?'icon-only-':'')+'menu-item-';
      if(selected&&icon)return style.getPropertyValue(prefix+'selected-shape').trim();
      if(['hovered','focused','pressed'].includes(state))return style.getPropertyValue('--md-comp-menus-horizontal-menu-item-'+(selected?'selected-':'')+state+'-shape').trim();
      return token(selected?'selected-shape':'shape');
    };
    const top = baseline?'0px':horizontal?horizontalShape():selected
      ? token("selected-shape")
      : first
        ? token("first-child-shape")
        : last
          ? token("last-child-inner-corner-corner-size")
          : token("shape");
    const bottom = baseline?'0px':horizontal?horizontalShape():selected
      ? token("selected-shape")
      : last
        ? token("last-child-shape")
        : first
          ? token("first-child-inner-corner-corner-size")
          : token("shape");
    for (const [corner, expected] of [
      ["TopLeft", top],
      ["TopRight", top],
      ["BottomLeft", bottom],
      ["BottomRight", bottom],
    ]) {
      checked++;
      if (expected.includes('infinity') ? Number.parseFloat(style['border'+corner+'Radius'])<1000000 : style["border" + corner + "Radius"] !== expected)
        failures.push({
          id: item.dataset.componentCase || item.textContent.trim(),
          property: "border" + corner + "Radius",
          actual: style["border" + corner + "Radius"],
          expected,
        });
    }
  }
  return {
    cases,
    checked,
    failureCount: failures.length,
    failures: failures.slice(0, 80),
  };
}

// Navigation has separate container and item source sets. Check their computed
// output together so correct variables on the wrong element cannot pass.
export function checkNavigationRendering() {
 const failures=[]; let checked=0;
 const number=value=>Number.parseFloat(value);
 const rgba=value=>{
  const v=value.trim();
  if(v==='transparent') return '0,0,0,0';
  if(v.startsWith('#')) return [1,3,5].map(i=>Number.parseInt(v.slice(i,i+2),16)).concat(v.length===9?Number.parseInt(v.slice(7,9),16)/255:1).join(',');
  const n=(v.match(/[\d.]+/g)||[]).map(Number);return [...n.slice(0,3),n[3]??1].join(',');
 };
 for(const root of document.querySelectorAll('#component-preview .md-navigation')) {
  if(!root.getClientRects().length)continue;
  const rs=getComputedStyle(root), family=root.dataset.family, variant=root.dataset.variant, orientation=root.dataset.orientation;
  const token=(name,style=rs)=>style.getPropertyValue('--'+name.replaceAll('.','-')).trim();
  const check=(label,actual,expected,kind='number')=>{checked++;if(!expected || (kind==='color'?rgba(actual)!==rgba(expected):(expected.includes('infinity') ? number(actual)<1000000 : Math.abs(number(actual)-number(expected))>.1 || !Number.isFinite(number(expected))) || !Number.isFinite(number(actual)))) failures.push({family,variant,orientation,label,actual,expected});};
  const drawer=family==='nav-drawer', bar=family==='nav-bar';
  const base=drawer?'md.comp.navigation-drawer':bar?'md.comp.nav-bar':'md.comp.nav-rail';
  const container=drawer?(variant==='modal'?base+'.modal':base):bar?base:base+(variant==='collapsed'?'.collapsed':root.dataset.layout==='modal'?'.expanded.modal':'.expanded');
  check('container',rs.backgroundColor,token(container+'.container.color'),'color');
  if(bar) check('bar height',rs.height,token(base+'.container.height'));
  if(drawer) check('drawer width',rs.width,token(base+'.container.width'));
  if(!drawer&&!bar&&variant==='collapsed') check('rail width',rs.width,token(base+'.collapsed.'+(root.dataset.width==='narrow'?'narrow.':'')+'container.width'));
  const shape=(token(container+'.container.shape')||token(base+'.container.shape')).split(/\s+/)[0];
  check('container corner',rs.borderTopLeftRadius,shape);
  check('elevation',rs.getPropertyValue('--_elevation-level'),token(container+'.container.elevation')||token(base+'.container.elevation')||'0');
  if(!drawer&&!bar){
   const menu=root.querySelector('.md-rail-menu');
   if(menu){const glyphs=[...menu.querySelectorAll('[data-navigation-menu-icon],[data-navigation-collapse-icon]')];
    check('one menu glyph',String(glyphs.filter(e=>getComputedStyle(e).opacity==='1'&&getComputedStyle(e).display!=='none').length),'1');
    for(const glyph of glyphs)check('menu glyph centered',String(glyph.getBoundingClientRect().x+glyph.getBoundingClientRect().width/2-menu.getBoundingClientRect().x),String(menu.getBoundingClientRect().width/2));
   }
  }
  for(const item of root.querySelectorAll('.md-nav-item')) {
   if(!item.getClientRects().length)continue;
   const cs=getComputedStyle(item), indicator=item.querySelector('.md-nav-indicator'), is=getComputedStyle(indicator), selected=item.getAttribute('aria-current')==='page';
   if(!drawer&&!bar){
    check('full-width destination target',String(item.getBoundingClientRect().width),rs.width);
    if(variant==='expanded'&&root.dataset.indicator==='fill')check('filled indicator width',is.width,String(item.getBoundingClientRect().width-number(cs.paddingLeft)-number(cs.paddingRight)));
    if(variant==='expanded'&&orientation==='vertical')check('leading indicator inset',String(indicator.getBoundingClientRect().x-item.getBoundingClientRect().x),token(base+'.item.horizontal.full-width.leading-space'));
   }
   const state=item.dataset.componentState, short={hovered:'hover',focused:'focus'}[state]||state;
   const prefix=base+(drawer?'':'.item');
   const selectedPrefix=prefix+(selected?'.active':'.inactive');
   const statePrefix=drawer && ['hovered','focused','pressed'].includes(state)?selectedPrefix+'.'+short:selectedPrefix;
   check('label',cs.color,token(statePrefix+'.label-text.color'),'color');
   const icon=[...indicator.querySelectorAll('svg')].find(e=>e.getClientRects().length);
   check('icon',getComputedStyle(icon).color,token(statePrefix+'.icon.color'),'color');
   check('icon size',String(icon.getBoundingClientRect().width),token(base+(drawer?'':'.item')+'.icon.size'));
   check('indicator color',is.backgroundColor,selected?token(drawer?base+'.active-indicator.color':base+'.item.active.indicator.color'):'transparent','color');
   check('indicator height',is.height,token(drawer?base+'.active-indicator.height':base+'.item.'+orientation+'.active-indicator.height'));
   check('indicator corner',is.borderTopLeftRadius,token(base+(drawer?'':'.item')+'.active-indicator.shape'));
   if(drawer) check('indicator width',is.width,token(base+'.active-indicator.width'));
   else if(orientation==='vertical') check('indicator width',is.width,token(base+'.item.vertical.active-indicator.width'));
   if(!drawer&&!bar) check('item minimum height',cs.minHeight,token(base+'.item.'+(root.dataset.size==='short'?'short.':'')+'container.height'));
   if(drawer) {
    for(const [prop,suffix] of [['fontSize','size'],['lineHeight','line-height'],['fontWeight','weight'],['letterSpacing','tracking']]) check('label '+prop,cs[prop],token(base+(selected&&suffix==='weight'?'.active':'')+'.label-text.'+suffix));
   } else {
    const font=token(base+'.item.'+orientation+'.label-text.font'), m=font.match(/([\d.]+)\s+([\d.]+)px\s*\/\s*([\d.]+)(px)?/);
    if(m) {check('font weight',cs.fontWeight,m[1]);check('font size',cs.fontSize,m[2]);check('font line height',cs.lineHeight,m[4]?m[3]:String(Number(m[2])*Number(m[3])));}
    else failures.push({family,label:'composite font',actual:font});
    check('tracking',cs.letterSpacing,token(base+'.item.'+orientation+'.label-text.font-letter-spacing'));
   }
   const layer=getComputedStyle(indicator,'::before');
   if(['hovered','focused','pressed'].includes(state)) {
    check('state color',layer.backgroundColor,token(selectedPrefix+'.'+(drawer?short:state)+'.state-layer.color'),'color');
    check('state opacity',layer.opacity,token(drawer?base+'.'+short+'.state-layer.opacity':selectedPrefix+'.'+state+'.state-layer.opacity')||token('md.sys.state.'+short+'.state-layer-opacity'));
   } else check('no state layer',layer.opacity,'0');
  }
 }
 return {checked,failures};
}

export function checkChoiceRendering() {
 const failures=[];let checked=0;
 const rgba=value=>{const v=value.trim();if(v==='transparent')return[0,0,0,0];if(v.startsWith('#'))return[1,3,5].map(i=>Number.parseInt(v.slice(i,i+2),16)/255).concat(v.length===9?Number.parseInt(v.slice(7,9),16)/255:1);const n=(v.match(/-?\d*\.?\d+(?:e[+-]?\d+)?/g)||[]).map(Number);return(v.startsWith('rgb')?n.slice(0,3).map(x=>x/255):n.slice(0,3)).concat(n[3]??1);};
 for(const item of document.querySelectorAll('#component-preview .md-tab,#component-preview .md-segmented-button')) {
  const segment=item.classList.contains('md-segmented-button'),group=item.closest(segment?'.md-segmented-buttons':'.md-tabs'),variant=group.dataset.variant,cs=getComputedStyle(item),input=item.querySelector('input');
  const selected=segment?input.checked:item.getAttribute('aria-selected')==='true', state=input?.disabled?'disabled':item.dataset.componentState, short={hovered:'hover',focused:'focus'}[state]||state;
  const base=segment?'md.comp.outlined-segmented-button':`md.comp.${variant}-navigation-tab`, selection=segment?(selected?'selected':'unselected'):(selected?'active':'inactive');
  const token=suffix=>cs.getPropertyValue('--'+(base+'.'+suffix).replaceAll('.','-')).trim();
  const check=(label,actual,expected,color=false,opacity=1)=>{checked++;let bad=!expected;if(color){const a=rgba(actual),b=rgba(expected);b[3]*=opacity;bad||=a.some((v,i)=>Math.abs(v-b[i])>.005);}else bad||=!Number.isFinite(Number.parseFloat(actual))||(expected.includes('infinity')?Number.parseFloat(actual)<1000000:!Number.isFinite(Number.parseFloat(expected))||Math.abs(Number.parseFloat(actual)-Number.parseFloat(expected))>.1);if(bad)failures.push({variant,segment,selected,state,label,actual,expected});};
  if(segment){const face=item.querySelector('.md-segment-face'),fs=getComputedStyle(face),first=group.firstElementChild,rect=item.getBoundingClientRect();check('touch target height',String(input.getBoundingClientRect().height),'48');if(input.getBoundingClientRect().width<48)failures.push({variant,label:'minimum touch target width',actual:input.getBoundingClientRect().width,expected:48});check('content gap',fs.columnGap,group.dataset.control==='icon'?'0':'8');check('leading padding',fs.paddingInlineStart,'12');check('trailing padding',fs.paddingInlineEnd,'12');check('centered content',fs.justifyContent==='center'?'1':'0','1');if(group.dataset.control==='label'&&!input.checked){const label=face.lastElementChild.getBoundingClientRect();check('unselected label centered',String(Math.abs(label.left+label.width/2-rect.left-rect.width/2)),'0');}if(Math.abs(rect.width-first.getBoundingClientRect().width)>.1)failures.push({variant,label:'equal segment widths',actual:rect.width,expected:first.getBoundingClientRect().width});}
  const activeState=['hovered','focused','pressed','disabled'].includes(state);
  const prefix=segment?'':variant==='primary'?'with-label-text.':'';
  const label=token(prefix+selection+'.'+short+'.label-text.color')||token(prefix+short+'.label-text.color')||token(prefix+selection+'.label-text.color');
  check('label',cs.color,label,true,segment&&state==='disabled'?Number(token('disabled.label-text.opacity')):1);
  const typography=prefix+'label-text.';
  for(const[prop,suffix]of [['fontSize','size'],['fontWeight','weight'],['lineHeight','line-height'],['letterSpacing','tracking']])check(prop,cs[prop],token(typography+suffix));
  const height=!segment&&variant==='primary'&&group.dataset.control==='icon-label'?'with-icon-and-label-text.container.height':'container.height';
  check('height',cs.height,segment?String(Number.parseFloat(token(height))+Number(group.dataset.density||0)*Number.parseFloat(cs.getPropertyValue('--md-sys-measurement-space50'))):token(height));
  check('container',cs.backgroundColor,segment?(selected?token('selected.container.color'):'transparent'):token('container.color'),true);
  const iconPrefix=segment?'':'with-icon.';
  for(const icon of item.querySelectorAll('svg')) {
   const is=getComputedStyle(icon),iconColor=token(iconPrefix+selection+'.'+short+'.icon.color')||token(iconPrefix+short+'.icon.color')||token(iconPrefix+selection+'.icon.color')||token(selection+'.with-icon.icon.color');
   check('icon color',is.color,iconColor,true,segment&&state==='disabled'?Number(token('disabled.icon.opacity')):1);
   check('icon size',is.width,token('with-icon.icon.size'));
  }
  const layer=getComputedStyle(item,'::before');
  if(activeState&&state!=='disabled') {
   check('state color',layer.backgroundColor,token(selection+'.'+short+'.state-layer.color')||token(short+'.state-layer.color'),true);
   check('state opacity',layer.opacity,token(selection+'.'+short+'.state-layer.opacity')||token(short+'.state-layer.opacity'));
  }else check('no state layer',layer.opacity,'0');
  if(state==='focused') {
   check('focus color',cs.outlineColor,token('focus.indicator.color'),true);
   check('focus width',cs.outlineWidth,token('focus.indicator.thickness'));
   check('focus offset',cs.outlineOffset,token('focus.indicator.outline.offset'));
  }
  const rawShape=token(segment?'shape':'container.shape');
  const shape=rawShape.includes('infinity')?[rawShape]:rawShape.split(/\s+/);
  for(const [i,corner]of ['TopLeft','TopRight','BottomRight','BottomLeft'].entries()) {
   const rounded=!segment || (i===0||i===3?item===group.firstElementChild:item===group.lastElementChild);
   check('corner '+corner,cs['border'+corner+'Radius'],rounded?(shape[i]||shape[i%2]||shape[0]):'0');
  }
  if(segment) {
   check('outline width',cs.borderTopWidth,token('outline.width'));
   check('outline color',cs.borderTopColor,token(state==='disabled'?'disabled.outline.color':'outline.color'),true,state==='disabled'?Number(token('disabled.outline.opacity')):1);
  }else {
   const indicator=item.querySelector('.md-tab-active-indicator'),is=getComputedStyle(indicator);
   check('indicator color',is.backgroundColor,token('active-indicator.color'),true);
   check('indicator height',is.height,token('active-indicator.height'));
   check('indicator width',is.width,String(variant==='primary'?item.querySelector('.md-tab-content').getBoundingClientRect().width:item.getBoundingClientRect().width));
   const indicatorShape=token('active-indicator.shape')||'0';
   const corners=indicatorShape.split(/\s+/);
   for(const[i,corner]of ['TopLeft','TopRight','BottomRight','BottomLeft'].entries()) check('indicator corner '+corner,is['border'+corner+'Radius'],corners[i]||corners[i%2]||corners[0]);
   check('indicator visibility',is.opacity,selected?'1':'0');
   check('divider',cs.borderBottomColor,token('divider.color'),true);check('divider height',cs.borderBottomWidth,token('divider.height'));
  }
 }
 return{checked,failures};
}

export function checkSliderRendering() {
 const failures=[];let checked=0;
 const rgba=value=>{const v=value.trim();if(v==='transparent')return[0,0,0,0];if(v.startsWith('#'))return[1,3,5].map(i=>Number.parseInt(v.slice(i,i+2),16)/255).concat(v.length===9?Number.parseInt(v.slice(7,9),16)/255:1);const n=(v.match(/-?\d*\.?\d+(?:e[+-]?\d+)?/g)||[]).map(Number);return(v.startsWith('rgb')?n.slice(0,3).map(x=>x/255):n.slice(0,3)).concat(n[3]??1);};
 for(const root of document.querySelectorAll('#component-preview .md-slider')) {
  checked++;if(!root.hasAttribute('data-slider-ready')) failures.push({label:'slider not initialized'});
  const cs=getComputedStyle(root),size=root.dataset.size,vertical=root.dataset.orientation==='vertical',state=root.dataset.componentState,short={hovered:'hover',focused:'focus'}[state]||state;
  const token=name=>cs.getPropertyValue('--md-comp-slider-'+name.replaceAll('.','-')).trim(),num=Number.parseFloat;
  const check=(label,actual,expected,color=false)=>{checked++;const bad=!expected||(color?rgba(actual).some((v,i)=>Math.abs(v-rgba(expected)[i])>.005):!Number.isFinite(num(actual))||!Number.isFinite(num(expected))||Math.abs(num(actual)-num(expected))>.1);if(bad)failures.push({size,vertical,state,label,actual,expected});};
  const part=vertical?'width':'height',axis=vertical?'height':'width',bounds=root.getBoundingClientRect(),stateSize=num(token('state-layer.size')),length=bounds[axis]-stateSize;
  const start=num(cs.getPropertyValue('--_slider-start')),end=num(cs.getPropertyValue('--_slider-end'));
  const active=getComputedStyle(root.querySelector('.md-slider-active')),inactive=getComputedStyle(root.querySelector('.md-slider-after'));
  for(const [name,style]of [['active',active],['inactive',inactive]]) {
   check(name+' thickness',style[part],token(size+'.'+name+'.track.height'));
   check(name+' color',style.backgroundColor,token(short+'.'+name+'.track.color')||token(name+'.track.color'),true);
   check(name+' opacity',style.opacity,state==='disabled'?token('disabled.'+name+'.track.opacity'):token(name+'.container.opacity'));
  }
  check('active length',active[axis],String(Math.max(0,(end-start)*length-num(cs.getPropertyValue('--_slider-start-gap'))-num(cs.getPropertyValue('--_slider-end-gap')))));
  const activeCorner=vertical?'borderBottomLeftRadius':'borderTopLeftRadius',inactiveCorner=vertical?'borderTopLeftRadius':'borderTopRightRadius';
  check('active outer corner',active[activeCorner],String(Math.min(num(token(size+'.active.track.shape.leading')),num(token(size+'.active.track.height'))/2)));
  check('inactive outer corner',inactive[inactiveCorner],String(Math.min(num(token(size+'.inactive.track.shape.trailing')),num(token(size+'.inactive.track.height'))/2)));
  for(const control of root.querySelectorAll('.md-slider-input')) {
   const input=control.querySelector('input'),controlState=input.disabled?'disabled':control.dataset.componentState,st={hovered:'hover',focused:'focus'}[controlState]||controlState;
   const nub=control.querySelector('.md-slider-nub'),ns=getComputedStyle(nub),thumb=control.querySelector('.md-slider-thumb'),ts=getComputedStyle(thumb),value=control.querySelector('.md-slider-value'),vs=getComputedStyle(value);
   check('handle thickness',ns[vertical?'height':'width'],token(st+'.handle.width')||token('handle.width'));
   check('handle length',ns[vertical?'width':'height'],token(size+'.active.handle.height'));
   check('handle color',ns.backgroundColor,token(st+'.handle.color')||token('handle.color'),true);
   check('handle opacity',ns.opacity,controlState==='disabled'?token('disabled.handle.opacity'):'1');
   check('handle elevation',ns.getPropertyValue('--_elevation-level'),token(st+'.handle.elevation')||token('handle.elevation'));
   check('state size',ts.width,token('state-layer.size'));check('state height',ts.height,token('state-layer.size'));
   const layer=getComputedStyle(control.querySelector('.md-slider-state-layer'),'::before');
   if(['hovered','focused','pressed'].includes(controlState)) {check('state color',layer.backgroundColor,token(st+'.state-layer.color'),true);check('state opacity',layer.opacity,token(st+'.state-layer.opacity'));}else check('no state layer',layer.opacity,'0');
   const fraction=(input.valueAsNumber-Number(input.min))/(Number(input.max)-Number(input.min)),tb=thumb.getBoundingClientRect();
   check('handle position',String(vertical?bounds.bottom-(tb.top+tb.height/2):tb.left+tb.width/2-bounds.left),String(stateSize/2+length*fraction));
   const label=root.dataset.shape==='label'?'label':'value-indicator.label';
   for(const[prop,suffix]of [['fontSize','size'],['lineHeight','line-height'],['fontWeight','weight'],['letterSpacing','tracking']])check('value '+prop,vs[prop],token(label+'.label-text.'+suffix));
   check('value label color',vs.color,token(label+'.label-text.color'),true);check('value background',vs.backgroundColor,token(root.dataset.shape==='label'?'label.container.color':'value-indicator.container.color'),true);
   const shown=['hovered','focused','pressed'].includes(controlState);
   check('value visibility',vs.opacity,shown?'1':'0');
   if(shown)check('value gap',String(nub.getBoundingClientRect().top-value.getBoundingClientRect().bottom),token('value-indicator.active.bottom-space'));
  }
  for(const tick of root.querySelectorAll('.md-slider-tick')) {
   const s=getComputedStyle(tick),prefix=state==='disabled'?'disabled':tick.dataset.active==='true'?'active':'inactive';
   check('tick size',s.width,token('with-tick-marks.container.size'));check('tick color',s.backgroundColor,token('with-tick-marks.'+prefix+'.container.color'),true);check('tick opacity',s.opacity,token('with-tick-marks.'+prefix+'.container.opacity'));
  }
  for(const icon of root.querySelectorAll('.md-slider-icon svg'))check('icon size',getComputedStyle(icon).width,token(size+'.icon.size'));
 }
 return{checked,failures};
}

export function checkBarRendering() {
 const failures=[];let checked=0;
 const rgba=value=>{
  if(value==='transparent')return[0,0,0,0];
  if(value.startsWith('#'))return[1,3,5].map(i=>Number.parseInt(value.slice(i,i+2),16)/255).concat(value.length===9?Number.parseInt(value.slice(7,9),16)/255:1);
  const values=(value.match(/-?\d*\.?\d+(?:e[+-]?\d+)?/g)||[]).map(Number);
  return values.slice(0,3).map(v=>value.startsWith('rgb')?v/255:v).concat(values[3]??1);
 };
 const variable=(style,name)=>style.getPropertyValue('--'+name).trim();
 const check=(element,label,actual,expected,color=false,alpha=1)=>{
  checked++;let valid=!!expected;
  if(color){const a=rgba(actual),b=rgba(expected);b[3]*=alpha;valid&&=a.every((v,i)=>Math.abs(v-b[i])<.01);}
  else if(expected.includes('infinity'))valid&&=Number.parseFloat(actual)>1000000;
  else valid&&=Number.isFinite(Number.parseFloat(expected))&&Math.abs(Number.parseFloat(actual)-Number.parseFloat(expected))<.6;
  if(!valid)failures.push({family:element.closest('[data-family]')?.dataset.family,label,actual,expected,alpha});
 };
 for(const root of document.querySelectorAll('#component-preview .md-app-bar')) {
  const style=getComputedStyle(root);
  check(root,'bar height',style.height,variable(style,'_appbar-height'));
  check(root,'bar color',style.backgroundColor,variable(style,'_appbar-container'),true);
  for(const button of root.querySelectorAll('.md-icon-button')) {
   const leading=button.parentElement.classList.contains('md-appbar-row'),role=leading?'leading':'trailing',cs=getComputedStyle(button);
   const expected=variable(style,'_appbar-'+role+'-icon')||variable(style,leading?'md-sys-color-on-surface':'md-sys-color-on-surface-variant');
   check(button,role+' icon color',cs.color,expected,true);
   const svg=getComputedStyle(button.querySelector('svg'));
   check(button,role+' icon size',svg.width,variable(style,'_appbar-'+role+'-icon-size')||variable(style,'_appbar-icon-size')||variable(style,'md-comp-icon-button-small-icon-size'));
   const box=button.getBoundingClientRect(),glyph=button.querySelector('svg').getBoundingClientRect();
   check(button,'icon vertical centering',String(glyph.y+glyph.height/2),String(box.y+box.height/2));
  }
  const avatar=root.querySelector('.md-appbar-avatar');
  if(avatar) {
   const slot=avatar.parentElement.getBoundingClientRect(),image=avatar.getBoundingClientRect();
   check(avatar,'avatar size',String(image.width),variable(style,'_appbar-avatar-size'));
   check(avatar,'avatar horizontal centering',String(image.x+image.width/2),String(slot.x+slot.width/2));
   check(avatar,'avatar vertical centering',String(image.y+image.height/2),String(slot.y+slot.height/2));
   check(avatar,'avatar control footprint',String(slot.width),variable(style,'md-comp-icon-button-small-container-height'));
  }
 }
 for(const root of document.querySelectorAll('#component-preview .md-toolbar')) {
  const style=getComputedStyle(root),vertical=root.dataset.orientation==='vertical';
  check(root,'toolbar color',style.backgroundColor,variable(style,'_toolbar-container'),true);
  check(root,'toolbar thickness',vertical?style.width:style.height,variable(style,vertical?'_toolbar-width':'_toolbar-height'));
  for(const item of root.querySelectorAll('.md-toolbar-item:not([hidden])')) {
   const cs=getComputedStyle(item),icon=item.querySelector('svg');
   const target=getComputedStyle(item,'::after');
   check(item,'touch target width',String(Math.min(48,Math.max(item.offsetWidth,Number.parseFloat(target.width)||0,Number.parseFloat(target.minWidth)||0))),'48');
   check(item,'touch target height',String(Math.min(48,Math.max(item.offsetHeight,Number.parseFloat(target.height)||0,Number.parseFloat(target.minHeight)||0))),'48');
   check(item,'button container',cs.backgroundColor,variable(cs,'_toolbar-item-container'),true);
   if(icon)check(item,'button icon',getComputedStyle(icon).color,variable(cs,'_toolbar-item-icon'),true,Number(variable(cs,'_toolbar-item-icon-opacity')));
   if(item.classList.contains('md-button'))check(item,'button label',cs.color,variable(cs,'_toolbar-item-label'),true,Number(variable(cs,'_toolbar-item-label-opacity')));
   const layer=getComputedStyle(item,'::before'),opacity=variable(cs,'_toolbar-item-state-opacity')||'0';
   check(item,'button state opacity',layer.opacity,opacity);
  }
  const fab=root.parentElement.querySelector('.md-toolbar-fab');
  if(fab){const fs=getComputedStyle(fab);check(fab,'FAB width',fs.width,variable(fs,'_toolbar-fab-width'));check(fab,'FAB height',fs.height,variable(fs,'_toolbar-fab-height'));check(fab,'FAB color',fs.backgroundColor,variable(fs,'_toolbar-fab-container'),true);
   const pressed=fab.matches(':active,[data-pressed]')&&!fab.hasAttribute('data-button-cancelled'),focused=fab.matches(':focus-visible'),hovered=fab.matches(':hover')&&!fab.hasAttribute('data-touch');
   check(fab,'FAB elevation',variable(fs,'_elevation-level'),variable(fs,pressed||focused?'md-sys-elevation-level2':hovered?'md-sys-elevation-level3':'_toolbar-fab-elevation'));
  }
 }
 return{checked,failures};
}

// Check the composed page, where nested cards can inherit the wrong color role.
export function checkHandoffRendering() {
  const failures = [], cards = [];
  const rgb = value => {
    if (value.startsWith('#')) return [1,3,5].map(i => Number.parseInt(value.slice(i,i+2),16)/255);
    const values = (value.match(/-?\d*\.?\d+(?:e[+-]?\d+)?/g) || []).map(Number).slice(0,3);
    return value.startsWith('rgb') ? values.map(v => v/255) : values;
  };
  const luminance = values => values.map(v => v <= .04045 ? v/12.92 : ((v+.055)/1.055)**2.4).reduce((n,v,i) => n + v*[.2126,.7152,.0722][i],0);
  for (const node of document.querySelectorAll('.tokens-handoff-grid .md-card')) {
    const style = getComputedStyle(node), foreground = rgb(style.color), background = rgb(style.backgroundColor),
      role = node.classList.contains('tokens-secondary-option') ? '--md-sys-color-on-secondary-container' : '--md-sys-color-on-surface',
      expected = rgb(style.getPropertyValue(role).trim()), a = luminance(foreground), b = luminance(background), ratio = (Math.max(a,b)+.05)/(Math.min(a,b)+.05),
      id = node.id || node.getAttribute('aria-labelledby');
    cards.push({id,ratio});
    if (foreground.length !== 3 || expected.length !== 3 || foreground.some((v,i) => Math.abs(v-expected[i])>.001)) failures.push({id,property:'foreground role',actual:style.color,expected:role});
    if (!Number.isFinite(ratio) || ratio < 4.5) failures.push({id,property:'text contrast',actual:ratio,expected:4.5});
  }
  return {checked:cards.length*2,cards,failures};
}
