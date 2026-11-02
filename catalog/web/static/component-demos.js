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
      const input = group.parentElement.querySelector('input[name="preview.compare"]');
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
  document.addEventListener('input',event=>{
    if(!event.target.matches?.('[data-search-composition-width]'))return;
    const root=event.target.closest('.md-search-demo'),viewport=root.querySelector('[data-search-viewport]');
    viewport.style.setProperty('--_search-composition-width',event.target.value+'px');
    root.querySelector('[data-search-composition-width-label]').textContent=Math.round(viewport.getBoundingClientRect().width)+'px';
  });
  document.addEventListener('md-list-change',event=>{
    const host=event.target.closest('#component-preview .md-list-demo');if(!host)return;
    for(const output of host.querySelectorAll('output:not([data-list-status])'))output.textContent=event.detail.text;
  });
})();
