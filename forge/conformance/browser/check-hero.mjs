const container = document.querySelector("#frames"),
  result = document.querySelector("#result");
const wait = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const frames = (w) =>
  new Promise((resolve) =>
    w.requestAnimationFrame(() => w.requestAnimationFrame(resolve)),
  );
const percentile = (values, p) =>
  [...values].sort((a, b) => a - b)[
    Math.min(values.length - 1, Math.floor(values.length * p))
  ] ?? null;
let checks = [];
const assert = (condition, message) => {
  if (!condition) throw Error(message);
  checks.push(message);
};
async function load(mode, geometry = false, unavailable = false) {
  const frame = document.createElement("iframe");
  container.append(frame);
  const html = await fetch("/md3/?artwork=" + mode).then((r) => {
    if (!r.ok) throw Error("Homepage " + r.status);
    return r.text();
  });
  const bootstrap = `
 ${unavailable ? `HTMLCanvasElement.prototype.getContext=()=>null;` : ""}
 window.heroErrors=[];addEventListener('error',e=>heroErrors.push(e.message));
 window.heroCapture={draws:0,times:[],boxes:[],record:false,renderer:null};

 const media=new EventTarget(),nativeMedia=matchMedia.bind(window);let reduced=false;
 Object.defineProperties(media,{matches:{get:()=>reduced},media:{value:'(prefers-reduced-motion: reduce)'}});
 window.matchMedia=q=>q.includes('prefers-reduced-motion')?media:nativeMedia(q);
 window.heroReduced=value=>{reduced=value;media.dispatchEvent(new Event('change'));};
 ${geometry ? `let hidden=false;Object.defineProperty(document,'hidden',{get:()=>hidden});window.heroHidden=value=>{hidden=value;document.dispatchEvent(new Event('visibilitychange'));};` : ""}

 let factory;Object.defineProperty(window,'TokenArtworkCanvas',{configurable:true,get:()=>factory,set:value=>{
 factory=(...args)=>{const renderer=value(...args);if(!renderer)return renderer;const draw=renderer.draw;
 renderer.draw=time=>{const start=performance.now();heroCapture.boxes=[];draw(time);heroCapture.draws++;if(heroCapture.record)heroCapture.times.push(performance.now()-start);};heroCapture.renderer=renderer;return renderer;};
 }});
 ${geometry ? `const rect=CanvasRenderingContext2D.prototype.roundRect;CanvasRenderingContext2D.prototype.roundRect=function(...args){heroCapture.boxes.push(args);return rect.apply(this,args);};` : ""}
 `;
  const loaded = new Promise((resolve) => (frame.onload = resolve));
  frame.srcdoc = html.replace(
    "<head>",
    `<head><base href="${location.origin}/md3/"><script>${bootstrap}<\/script>`,
  );
  await loaded;
  const w = frame.contentWindow,
    d = frame.contentDocument;
  await w.TokenAssets.ready;
  await d.fonts.ready;
  frame.scrollIntoView({ block: "start" });
  await frames(w);
  await wait(100);
  return {
    frame,
    w,
    d,
    art: d.querySelector("[data-artwork]"),
    toggle: d.querySelector(".tokens-art-pause input"),
  };
}
function setPaused(page, paused) {
  page.toggle.checked = paused;
  page.toggle.dispatchEvent(new page.w.Event("change", { bubbles: true }));
}
async function check() {
  checks = [];
  container.replaceChildren();
  const dom = await load("dom"),
    canvas = await load("canvas", true),
    capture = canvas.w.heroCapture;
  assert(
    !dom.d.querySelector('script[src*="/artwork-canvas.js"]'),
    "DOM page omits Canvas script",
  );
  assert(
    canvas.art.getAnimations({ subtree: true }).length === 0,
    "Canvas has no DOM animation tracks",
  );
  assert(
    !!capture.renderer && canvas.art.hasAttribute("data-canvas-ready"),
    "Canvas renderer initializes",
  );
  const animations = dom.art.getAnimations({ subtree: true });
  animations.forEach((a) => a.pause());
  setPaused(canvas, true);
  await wait(100);
  for (const width of [960, 390]) {
    dom.frame.style.width = canvas.frame.style.width = width + "px";
    await frames(canvas.w);
    await wait(50);
    for (let time = 0; time < 8000; time += 125) {
      animations.forEach((a) => (a.currentTime = time));
      capture.renderer.draw(time);
      const grid = dom.art.getBoundingClientRect(),
        outer = dom.art.parentElement.getBoundingClientRect();
      assert(
        grid.left >= outer.left + 1 &&
          grid.top >= outer.top + 1 &&
          grid.right <= outer.right - 1 &&
          grid.bottom <= outer.bottom - 1,
        `DOM ${width}px: grid stays inside frame`,
      );
      const tiles = [...dom.art.querySelectorAll(".tokens-art-tile")];
      assert(capture.boxes.length === 9, "Canvas draws nine blocks");
      tiles.forEach((tile, i) => {
        const rect = tile.getBoundingClientRect(),
          surface = dom.w.getComputedStyle(
            tile.querySelector(".tokens-art-surface"),
          );
        const [x, y, w, h, radii] = capture.boxes[i];
        assert(
          Math.max(
            Math.abs(rect.left - grid.left - x),
            Math.abs(rect.top - grid.top - y),
            Math.abs(rect.width - w),
            Math.abs(rect.height - h),
          ) < 0.15,
          `DOM/Canvas ${width}px: block ${i} geometry matches`,
        );
        const expected = [
          "borderTopLeftRadius",
          "borderTopRightRadius",
          "borderBottomRightRadius",
          "borderBottomLeftRadius",
        ].map((p) => parseFloat(surface[p]));
        assert(
          expected.every((r, j) => Math.abs(r - radii[j]) < 0.01),
          "Directional corner samples match",
        );
      });
    }
    const element = canvas.d.querySelector("canvas"),
      ratio = Math.min(canvas.w.devicePixelRatio, 2);
    assert(
      Math.abs(
        element.width - canvas.art.getBoundingClientRect().width * ratio,
      ) <= 1,
      "Canvas backing buffer follows viewport and DPR",
    );
  }
  const images = new Set();
  for (const theme of ["light", "dark"])
    for (const contrast of ["default", "medium", "high"]) {
      canvas.d.documentElement.dataset.theme = theme;
      canvas.d.documentElement.dataset.contrast = contrast;
      await frames(canvas.w);
      capture.renderer.draw(250);
      const element = canvas.d.querySelector("canvas"),
        ctx = element.getContext("2d"),
        pixels = ctx.getImageData(0, 0, element.width, element.height).data;
      let checksum = 0;
      for (let i = 0; i < pixels.length; i += 68)
        checksum = (checksum * 31 + pixels[i]) >>> 0;
      images.add(checksum);
      for (const [x, y, w, h] of capture.boxes) {
        const scale = element.width / canvas.art.getBoundingClientRect().width;
        assert(
          ctx.getImageData(
            Math.floor((x + w / 2) * scale),
            Math.floor((y + h / 2) * scale),
            1,
            1,
          ).data[3] === 255,
          "Canvas paints block centers without gaps",
        );
      }
    }
  assert(images.size === 6, "Canvas redraws each theme/contrast appearance");
  canvas.art.scrollIntoView({ block: "center" });
  await wait(150);
  let count = capture.draws;
  await wait(150);
  assert(capture.draws === count, "Pause stops Canvas redraws");
  setPaused(canvas, false);
  await wait(150);
  assert(capture.draws > count, "Canvas resumes");
  canvas.w.heroReduced(true);
  await wait(60);
  count = capture.draws;
  await wait(150);
  assert(capture.draws === count, "Reduced motion stops Canvas redraws");
  canvas.w.heroReduced(false);
  await wait(80);
  canvas.w.heroHidden(true);
  await wait(60);
  count = capture.draws;
  await wait(150);
  assert(capture.draws === count, "Hidden document stops Canvas redraws");
  canvas.w.heroHidden(false);
  await wait(80);
  canvas.w.scrollTo(0, canvas.d.body.scrollHeight);
  await wait(100);
  count = capture.draws;
  await wait(200);
  assert(
    canvas.art.hasAttribute("data-paused") && capture.draws === count,
    "Offscreen Canvas stops redraws",
  );
  assert(
    !dom.w.heroErrors.length && !canvas.w.heroErrors.length,
    "Both renderers have no script errors",
  );
  // A real HTTP document exercises HTMX's same-origin navigation and teardown.
  container.replaceChildren();
  const live = document.createElement("iframe");
  container.append(live);
  await new Promise((resolve) => {
    live.onload = resolve;
    live.src = "/md3/?artwork=canvas";
  });
  const w = live.contentWindow,
    d = live.contentDocument;
  await w.TokenAssets.ready;
  const old = d.querySelector("canvas"),
    ctx = old.getContext("2d"),
    clear = ctx.clearRect;
  let draws = 0;
  ctx.clearRect = function (...args) {
    draws++;
    return clear.apply(this, args);
  };
  async function follow(selector, query = "") {
    let link = d.querySelector(selector);
    const next = link.cloneNode(true);
    next.search = query;
    link.replaceWith(next);
    w.htmx.process(next);
    const swapped = new Promise((resolve, reject) => {
      const timer = setTimeout(
        () => reject(Error("Hero HTMX navigation timed out")),
        8000,
      );
      d.addEventListener(
        "htmx:afterSettle",
        () => {
          clearTimeout(timer);
          resolve();
        },
        { once: true },
      );
    });
    next.click();
    await swapped;
    await w.TokenAssets.ready;
  }
  await follow('.tokens-nav-link[href="/md3/components"]');
  count = draws;
  await wait(250);
  assert(
    draws === count && !old.isConnected,
    "HTMX disposes the outgoing Canvas loop",
  );
  await follow(".tokens-brand", "?artwork=canvas");
  assert(
    d.querySelector("[data-canvas-ready]") && d.querySelector("canvas") !== old,
    "HTMX recreates the Canvas renderer on revisit",
  );
  assert(
    [...d.scripts].filter((s) => s.src.includes("/artwork-canvas.js?"))
      .length === 1,
    "Canvas script loads once across revisits",
  );
  container.replaceChildren();
  const fallback = await load("canvas", false, true);
  assert(
    !fallback.art.hasAttribute("data-canvas-ready") &&
      fallback.art.getAnimations({ subtree: true }).length === 22,
    "Unavailable Canvas falls back to the DOM animation",
  );
  return { passed: checks.length, checks };
}
async function cpu() {
  const response = await fetch("/profile/cpu", { cache: "no-store" });
  return response.ok ? response.json() : null;
}
async function measure(page, mode, state, duration) {
  const { w, d, art, toggle } = page;
  if (state === "paused" && !toggle.checked) setPaused(page, true);
  if (state !== "paused" && toggle.checked) setPaused(page, false);
  w.scrollTo(0, state === "offscreen" ? d.body.scrollHeight : 0);
  await wait(1000);
  assert(
    art.hasAttribute("data-paused") === (state !== "active"),
    `Profile ${mode}/${state}: playback state matches`,
  );
  const gaps = [],
    capture = w.heroCapture,
    draws = capture.draws;
  capture.times = [];
  capture.record = true;
  const longFrames = [],
    observer = w.PerformanceObserver.supportedEntryTypes.includes(
      "long-animation-frame",
    )
      ? new w.PerformanceObserver((list) =>
          longFrames.push(...list.getEntries()),
        )
      : null;
  observer?.observe({ type: "long-animation-frame" });
  const before = await cpu(),
    start = performance.now();
  let prior, raf;
  function tick(now) {
    if (prior !== undefined) gaps.push(now - prior);
    prior = now;
    raf = w.requestAnimationFrame(tick);
  }
  raf = w.requestAnimationFrame(tick);
  await wait(duration);
  w.cancelAnimationFrame(raf);
  const end = performance.now(),
    after = await cpu();
  capture.record = false;
  observer?.disconnect();
  return {
    mode,
    state,
    artworkSize: [art.clientWidth, art.clientHeight],
    durationMs: end - start,
    frameIntervalMedian: percentile(gaps.slice(1), 0.5),
    frameIntervalP95: percentile(gaps.slice(1), 0.95),
    framesOver25ms: gaps.filter((g) => g > 25).length,
    longFrames: longFrames.length,
    draws: capture.draws - draws,
    canvasSubmissionP95: percentile(capture.times, 0.95),
    runningAnimations: art
      .getAnimations({ subtree: true })
      .filter((a) => a.playState === "running").length,
    paused: art.hasAttribute("data-paused"),
    processCPUPercent:
      before && after
        ? Object.fromEntries(
            Object.keys(before).map((pid) => [
              pid,
              ((after[pid] - before[pid]) / (end - start)) * 100,
            ]),
          )
        : null,
  };
}
async function profile() {
  container.replaceChildren();
  const samples = [];
  for (const mode of ["dom", "canvas", "canvas", "dom"]) {
    container.replaceChildren();
    const page = await load(mode);
    for (const state of ["paused", "active", "offscreen"]) {
      result.textContent = `Profiling ${mode}: ${state} (${samples.length + 1}/12)`;
      samples.push(
        await measure(page, mode, state, state === "active" ? 16000 : 8000),
      );
    }
  }
  return {
    userAgent: navigator.userAgent,
    devicePixelRatio,
    canvasDPRCap: 2,
    samples,
  };
}
for (const [id, run] of [
  ["check", check],
  ["profile", profile],
])
  document.getElementById(id).onclick = async () => {
    document.querySelectorAll("button").forEach((b) => (b.disabled = true));
    result.textContent = "Running";
    try {
      result.textContent = JSON.stringify(await run(), null, 2);
    } catch (error) {
      result.textContent = error.stack;
    }
    document.querySelectorAll("button").forEach((b) => (b.disabled = false));
  };
