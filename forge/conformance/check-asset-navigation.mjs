export async function checkAssetNavigation(load, assert) {
  // srcdoc has an about: URL, which HTMX correctly treats as a different origin
  // for boosted links. Exercise navigation in an actual HTTP document.
  const { w, d } = await load(null, { path: "/md3/changes", real: true });
  const requested = (name) =>
    [...d.scripts].some((script) =>
      new URL(script.src || w.location.href).pathname.endsWith("/" + name),
    );
  assert(
    !requested("date-picker.js") &&
      !requested("runtime/navigation.js") &&
      !requested("artwork.js"),
    "Change log omits unrelated component behavior",
  );
  assert(
    !d.querySelector('link[href*="/components.css"]'),
    "Change log omits the complete component stylesheet",
  );
  const theme = d.querySelector("[data-app-theme]");
  async function follow(path, query = "") {
    let link = [...d.querySelectorAll("a[href]")].find(
      (link) => new URL(link.href).pathname === path,
    );
    if (!link) throw new Error("Missing navigation link: " + path);
    const swapped = new Promise((resolve, reject) => {
      const timeout = setTimeout(
        () => reject(new Error("HTMX navigation timed out: " + path)),
        8000,
      );
      d.addEventListener(
        "htmx:afterSettle",
        () => {
          clearTimeout(timeout);
          resolve();
        },
        { once: true },
      );
    });
    // HTMX reads a boosted link's URL when processing its markup. Reprocess a
    // fresh link when selecting a test-only query rather than changing its cache.
    const next = link.cloneNode(true);
    next.search = query;
    link.replaceWith(next);
    link = next;
    w.htmx.process(link);
    link.click();
    await swapped;
    await w.TokenAssets.ready;
    assert(
      d === w.document && d.querySelector("[data-app-theme]") === theme,
      "HTMX retains the document and preserved header",
    );
  }
  await follow(
    "/md3/components/navigation-rail",
    "?preview.compare=interactive",
  );
  assert(
    requested("runtime/navigation.js") && !requested("date-picker.js"),
    "HTMX loads the newly visited rail without loading pickers",
  );
  const saved = [...d.querySelectorAll("[data-navigation-item]")].find(
    (item) => item.textContent.trim() === "Saved",
  );
  saved.click();
  assert(
    d.querySelector("[data-navigation-title]").textContent === "Saved",
    "Lazily loaded rail behavior handles selection",
  );
  await follow("/md3/changes");
  assert(
    !d.querySelector("link[data-component-stylesheet]"),
    "Leaving a playground removes its page stylesheet",
  );
  await follow(
    "/md3/components/date-pickers",
    "?preview.compare=interactive&preview.variant=modal",
  );
  assert(
    requested("date-picker.js"),
    "HTMX loads date-picker behavior on first use",
  );
  d.querySelector("[data-date-open]").click();
  assert(
    d.querySelector(".md-date-panel").open,
    "Date picker opens after lazy initialization",
  );
  const closed = new Promise((resolve) =>
    d
      .querySelector(".md-date-panel")
      .addEventListener("close", resolve, { once: true }),
  );
  d.querySelector("[data-date-cancel]").click();
  await closed;
  await follow("/md3/changes");
  await follow(
    "/md3/components/navigation-rail",
    "?preview.compare=interactive",
  );
  assert(
    [...d.scripts].filter((script) =>
      script.src.includes("/runtime/navigation.js?"),
    ).length === 1,
    "Revisiting a component does not register its script twice",
  );
  assert(
    !w.motionErrors.length,
    "Conditional assets and HTMX navigation have no script errors",
  );
}
