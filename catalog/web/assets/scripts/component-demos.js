// Catalog example actions and announcements. Component mechanics live in Forge.
(() => {
  document.addEventListener("click", (event) => {
    const proxy = event.target.closest?.("[data-toolbar-proxy]");
    if (proxy && !proxy.disabled)
      proxy
        .closest(".md-toolbar-demo")
        .querySelector(
          '[data-toolbar-item="' + proxy.dataset.toolbarProxy + '"]',
        )
        .click();
    const tool = event.target.closest?.('[data-toolbar-item="0"]');
    if (tool)
      queueMicrotask(() => {
        const peer = tool
          .closest(".md-toolbar-demo")
          .querySelector('[data-toolbar-proxy="0"]');
        if (!peer) return;
        const selected = tool.getAttribute("aria-pressed") === "true";
        peer.setAttribute("aria-checked", String(selected));
        peer.querySelector("[data-menu-checked-icon]").hidden = !selected;
        peer.querySelector("[data-menu-unchecked-icon]").hidden = selected;
      });
  });
  document.addEventListener("md-navigation-select", (event) => {
    const group = event.target,
      label = event.detail.label,
      output = group.querySelector("output");
    if (output) output.textContent = label + " selected";
    const barDemo = group.closest("[data-navbar-demo]");
    if (barDemo) {
      for (const page of barDemo.querySelectorAll("[data-navbar-page]")) {
        page.hidden = page.dataset.navbarPage !== label;
        if (!page.hidden) {
          if (event.detail.reselected) page.scrollTop=0;
          else if (!matchMedia("(prefers-reduced-motion: reduce)").matches)
            page.animate([{opacity:0},{opacity:0,offset:0.2},{opacity:1}],MaterialMotion.timing(group,"default-effects"));
        }
      }
    }
    const frame = group.closest(".md-rail-frame");
    if (!frame) return;
    frame.querySelector("[data-navigation-title]").textContent = label;
    frame.querySelector("[data-navigation-description]").textContent =
      "Your " + label.toLowerCase() + " in this workspace.";
    const screen = frame.querySelector(".md-rail-screen");
    screen.scrollTop = 0;
    if (!matchMedia("(prefers-reduced-motion: reduce)").matches)
      screen.animate(
        [{ opacity: 0 }, { opacity: 1 }],
        MaterialMotion.timing(group, "fast-effects"),
      );
  });
  document.addEventListener("click", (event) => {
    const button = event.target.closest?.("[data-navigation-create]"),
      frame = button?.closest(".md-rail-frame");
    if (frame)
      frame.querySelector("[data-navigation-title]").textContent =
        "Create file";
  });
  for (const type of ["md-chip-select", "md-chip-action", "md-chip-remove"])
    document.addEventListener(type, (event) => {
      const chip = event.target,
        demo = chip.closest(".md-chip-demo"),
        output = demo?.querySelector("[data-chip-output]");
      if (!output) return;
      if (type === "md-chip-select")
        output.textContent = event.detail.selected ? "Selected" : "Unselected";
      if (type === "md-chip-action")
        output.textContent = event.detail.label + " activated";
      if (type === "md-chip-remove") {
        output.textContent =
          "Removed " + chip.querySelector(".md-chip-label").textContent;
        demo
          .closest("figure")
          ?.closest("#component-preview")
          ?.querySelector("select,button")
          ?.focus({ preventScroll: true });
      }
    });
  document.addEventListener("md-button-group-action", (event) => {
    const button = event.target,
      group = button.closest(".md-button-group"),
      output = group.closest(".md-button-group-demo")?.querySelector("output");
    if (group.hasAttribute("data-preview-compare")) {
      const input = group.parentElement.querySelector('[name="preview.compare"]');
      if (input.value !== button.dataset.previewComparison) {
        input.value = button.dataset.previewComparison;
        input.dispatchEvent(new Event("change", { bubbles: true }));
      }
      return;
    }
    if (!output) return;
    const selected = [
      ...group.querySelectorAll('button[aria-pressed="true"]'),
    ].map((item) => item.textContent.trim());
    output.textContent =
      group.dataset.control === "standard"
        ? button.textContent.trim() + " activated."
        : selected.length
          ? selected.join(", ") + " selected."
          : "No selection.";
  });
  document.addEventListener("md-snackbar-action", (event) => {
    const button = event.target,
      output = button.closest("[data-snackbar-demo]")?.querySelector("output");
    if (output)
      output.textContent = button.hasAttribute("data-snackbar-open")
        ? ""
        : button.hasAttribute("data-snackbar-action")
          ? "Message undone."
          : "Notification dismissed.";
  });
  document.addEventListener("md-fab-menu-select", (event) => {
    const button = event.target,
      output = button.closest(".md-fab-menu")?.querySelector("output");
    if (output) output.textContent = button.textContent.trim() + " selected.";
  });
  document.addEventListener("md-split-action", (event) => {
    const button = event.target,
      output = button.closest(".md-split-demo")?.querySelector("output");
    if (output)
      output.textContent = button.hasAttribute("data-split-primary")
        ? "Added to cart."
        : button.textContent + ".";
  });

  function searchExample(root, label, description) {
    for (const item of root.querySelectorAll('[data-search-message]')) {const selected=item.dataset.searchMessage===label;item.setAttribute('aria-pressed',String(selected));item.dataset.selected=String(selected);}
    root.querySelector('[data-search-page-title]').textContent=label;
    root.querySelector('[data-search-page-description]').textContent=description;
    if(root.querySelector('.md-search-view').dataset.layout==='adaptive') {
      root.dataset.messageOpen='true';
      root.querySelector('[data-search-page-title]').focus({preventScroll:true});
    }
  }
  document.addEventListener('md-search-select',event=>{
    const root=event.target.closest('#component-preview .md-search-demo');if(!root)return;
    searchExample(root,event.detail.label,'Selected from search results.');
    root.querySelector('[data-search-feedback]').textContent=event.detail.label+' opened';
  });
  document.addEventListener('md-search-profile',event=>{
    const root=event.target.closest('#component-preview .md-search-demo');if(root)root.querySelector('[data-search-feedback]').textContent='Profile selected';
  });
  document.addEventListener('click',event=>{
    const root=event.target.closest?.('#component-preview .md-search-demo');if(!root)return;
    const destination=event.target.closest('[data-search-mail-destination]');
    if(destination){root.querySelector('[data-search-folder-title]').textContent=destination.querySelector('.md-nav-label').textContent;return;}
    const message=event.target.closest('[data-search-message]');
    if(message){searchExample(root,message.dataset.searchMessage,'Review the latest component updates and examples.');return;}
    if(event.target.closest('[data-search-message-back]')){delete root.dataset.messageOpen;root.querySelector('[data-search-message][aria-pressed="true"]')?.focus({preventScroll:true});return;}
    if(event.target.closest('[data-search-compose],[data-search-reply]'))root.querySelector('[data-search-feedback]').textContent='Example mail composition: use search to explore the component.';
  });
  const previewDimensions = {};
  let restoringDimensions = false;
  function scrollFieldMinimumWidth(track) {
    const field = track.closest('[data-scroll-field]');
    const sizing = document.createElement('div');
    sizing.className = 'md-scroll-sizing';
    sizing.style.cssText = 'position:absolute;inline-size:max-content';
    for (const {label} of JSON.parse(field.dataset.scrollOptions)) {
      const option = document.createElement('span');
      option.textContent = label;
      sizing.append(option);
    }
    track.append(sizing);
    const style = getComputedStyle(track);
    const width = sizing.offsetWidth + parseFloat(style.paddingLeft) + parseFloat(style.paddingRight);
    sizing.remove();
    return Math.max(track.offsetWidth, width);
  }
  document.addEventListener('input', event => {
    const input = event.target;
    const control = input.closest?.('[data-preview-dimension]');
    if (!control || input.type !== 'range') return;
    const dimension = control.dataset.previewDimension;
    const family = control.closest('[data-preview-dimensions]').dataset.previewDimensions;
    const uncapped = dimension === 'max-width' && input.value === input.max;
    if (!restoringDimensions) previewDimensions[family + ':' + dimension] = uncapped ? 'none' : input.value;
    const preview = input.closest('#component-preview');
    preview.style.setProperty('--_preview-dimension-' + dimension, uncapped ? 'none' : input.value + 'px');
    let valueText = uncapped ? 'No limit' : input.value + ' px';
    if (dimension === 'composition-width') {
      const viewport = preview.querySelector('[data-search-viewport]');
      valueText = input.value === input.max
        ? 'Fit to available space'
        : Math.round(viewport.getBoundingClientRect().width) + ' px';
    }
    control.querySelector('output').value = valueText;
    input.setAttribute('aria-valuetext', valueText);
  });
  function applyPreviewDimensions(preview) {
    if (!preview.isConnected) return;
    const controls = preview.querySelector('[data-preview-dimensions]');
    const family = controls.dataset.previewDimensions;
    const width = controls.querySelector('[data-preview-dimension="width"] input');
    const height = controls.querySelector('[data-preview-dimension="height"] input');
    const cap = controls.querySelector('[data-preview-dimension="max-width"] input');
    /* Read the intrinsic floors from the rendered specimens before applying any saved dimensions. */
    if (cap) {
      const groups = Array.from(preview.querySelectorAll('.tokens-group-preview .md-button-group'), group => {
        group.style.inlineSize = 'max-content';
        group.style.maxInlineSize = 'none';
        const intrinsic = group.getBoundingClientRect().width;
        group.style.removeProperty('inline-size');
        group.style.removeProperty('max-inline-size');
        return intrinsic;
      });
      const minimum = Math.ceil(Math.max(0, ...groups));
      if (minimum) {
        // The comparison can include groups wider than the usual 800px ceiling.
        cap.max = width.max = String(Math.max(Number(width.max), minimum + 160));
        cap.min = width.min = String(minimum);
      }
    } else if (width) {
      preview.style.setProperty('--_preview-dimension-width', '0px');
      const widths = family === 'time'
        ? Array.from(preview.querySelectorAll('[data-time-inline]'), panel => panel.offsetWidth)
        : Array.from(preview.querySelectorAll('.tokens-scroll-field-specimen'), specimen => {
            const tracks = [...specimen.querySelectorAll('.md-scroll-track')];
            const gap = Number.parseFloat(getComputedStyle(specimen).columnGap) || 0;
            return tracks.reduce((total, track) => total + scrollFieldMinimumWidth(track), gap * (tracks.length - 1));
          });
      const minimum = Math.ceil(Math.max(0, ...widths));
      if (minimum) width.min = String(minimum);
      preview.style.removeProperty('--_preview-dimension-width');
    }
    if (height && family === 'scroll-field') {
      preview.style.setProperty('--_preview-dimension-height', '0px');
      const heights = Array.from(preview.querySelectorAll('.md-scroll-track'),
        track => Number.parseFloat(getComputedStyle(track, '::before').blockSize));
      const minimum = Math.ceil(Math.max(0, ...heights));
      if (minimum) height.min = String(minimum);
      preview.style.removeProperty('--_preview-dimension-height');
    }
    restoringDimensions = true;
    for (const control of controls.querySelectorAll('[data-preview-dimension]')) {
      const input = control.querySelector('input');
      const saved = previewDimensions[family + ':' + control.dataset.previewDimension];
      const initial = family === 'search' ? input.defaultValue : input.min;
      input.value = saved === 'none' ? input.max : saved ?? initial;
      input.dispatchEvent(new Event('input', {bubbles: true}));
    }
    restoringDimensions = false;
  }

  function initializePreviewDimensions(preview) {
    if (!preview?.hasAttribute('data-preview-sizing-pending')) return;
    applyPreviewDimensions(preview);
    preview.removeAttribute('data-preview-sizing-pending');

    // The preview can render with fallback metrics. Recheck the bounds when
    // the web font arrives, preserving choices made since the swap.
    if (preview.querySelector('[data-preview-dimensions]').dataset.previewDimensions !== 'search' &&
        document.fonts.status === 'loading') {
      document.fonts.ready.then(() => applyPreviewDimensions(preview));
    }
  }

  // Fragment styles are ready before insertion and sizing previews skip
  // HTMX attribute settling, so the measured bounds survive the swap.
  document.addEventListener('htmx:before:settle', event => {
    const target = event.target;
    initializePreviewDimensions(target.id === 'component-preview' ? target : target.querySelector?.('#component-preview'));
  });
  document.addEventListener('tokens:assets-ready', () => {
    initializePreviewDimensions(document.getElementById('component-preview'));
  });
  document.addEventListener('md-list-change',event=>{
    const host=event.target.closest('#component-preview .md-list-demo');if(!host)return;
    for(const output of host.querySelectorAll('output:not([data-list-status])'))output.textContent=event.detail.text;
  });
})();
