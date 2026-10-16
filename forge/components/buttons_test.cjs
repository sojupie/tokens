const { test } = require("node:test");
const assert = require("node:assert/strict");
const { readFileSync } = require("node:fs");
const { runInNewContext } = require("node:vm");
const source = readFileSync(`${__dirname}/static/buttons.js`, "utf8");

function harness(reduced = false) {
  const listeners = {},
    windows = {},
    animations = [];
  let observer,
    mediaChange,
    now = 0,
    timerID = 0;
  const timers = new Map();
  const advance = (ms) => {
    const end = now + ms;
    while (true) {
      const next = [...timers]
        .filter(([, t]) => t.at <= end)
        .sort((a, b) => a[1].at - b[1].at)[0];
      if (!next) break;
      now = next[1].at;
      timers.delete(next[0]);
      next[1].fn();
    }
    now = end;
  };
  class Element {
    constructor(tagName = "SPAN") {
      this.tagName = tagName;
      this.attrs = new Map();
      this.children = [];
      this.style = {
        setProperty(k, v) {
          this[k] = v;
        },
      };
      this.isConnected = true;
      this.className = "";
    }
    setAttribute(k, v) {
      this.attrs.set(k, String(v));
    }
    getAttribute(k) {
      return this.attrs.get(k) ?? null;
    }
    hasAttribute(k) {
      return this.attrs.has(k);
    }
    removeAttribute(k) {
      this.attrs.delete(k);
    }
    closest() {
      return this.isButton ? this : (this.parent?.closest() ?? null);
    }
    matches(selector) {
      if (selector === ".md-filter-chip") return !!this.isChip;
      return (
        !!this.disabled ||
        !!this.disabledChild ||
        this.getAttribute("aria-disabled") === "true"
      );
    }
    contains(other) {
      return other === this || this.children.some((c) => c.contains(other));
    }
    append(child) {
      child.parent = this;
      child.isConnected = this.isConnected;
      this.children.push(child);
    }
    remove() {
      this.isConnected = false;
      if (this.parent)
        this.parent.children = this.parent.children.filter((c) => c !== this);
    }
    querySelector(selector) {
      return (
        this.children.find((c) => c.className === selector.slice(1)) ??
        this.children.map((c) => c.querySelector(selector)).find(Boolean) ??
        null
      );
    }
    getBoundingClientRect() {
      return { left: 0, top: 0, width: 120, height: 40 };
    }
    animate(frames, options) {
      let resolve,
        reject,
        done = false;
      const finished = new Promise((a, b) => {
        resolve = a;
        reject = b;
      });
      const animation = {
        target: this,
        frames,
        options,
        finished,
        cancelled: false,
        complete() {
          if (!done) {
            done = true;
            resolve();
          }
        },
        cancel() {
          this.cancelled = true;
          if (!done) {
            done = true;
            reject(new Error("Cancelled"));
          }
        },
      };
      animations.push(animation);
      return animation;
    }
  }
  const media = {
    matches: reduced,
    addEventListener: (_, fn) => {
      mediaChange = fn;
    },
  };
  const document = {
    documentElement: new Element("HTML"),
    hidden: false,
    createElement: (tag) => new Element(tag),
    addEventListener: (type, fn) => {
      (listeners[type] ??= []).push(fn);
    },
  };
  const window = {
    addEventListener: (type, fn) => {
      windows[type] = fn;
    },
  };
  const getComputedStyle = (button) => ({
    opacity: "1",
    borderTopLeftRadius: "20px",
    borderTopRightRadius: "20px",
    borderBottomRightRadius: "20px",
    borderBottomLeftRadius: "20px",
    getPropertyValue: (name) =>
      button.noSpring && name.includes("spring")
        ? ""
        : ({
            "--_button-rest-shape":
              button.getAttribute("aria-pressed") === "true"
                ? "12px"
                : "calc(infinity * 1px)",
            "--_button-pressed-shape": "8px",
            "--_button-spring-damping": "0.9",
            "--_button-spring-stiffness": "1400",
            "--_button-state-opacity": "0.1",
            "--_button-state-color": "rgb(20, 30, 40)",
            "--_button-ripple-duration": "250ms",
            "--_button-effects-duration": "150ms",
            "--_button-easing": "cubic-bezier(0.2, 0, 0, 1)",
          }[name] ?? ""),
  });
  const module = { exports: {} };
  runInNewContext(source, {
    module,
    document,
    window,
    Element,
    Node: Element,
    MutationObserver: class {
      constructor(fn) {
        observer = fn;
      }
      observe() {}
    },
    matchMedia: (query) =>
      query.includes("forced-colors")
        ? { matches: false, addEventListener() {} }
        : media,
    performance: { now: () => now },
    setTimeout: (fn, delay) => {
      const id = ++timerID;
      timers.set(id, { fn, at: now + delay });
      return id;
    },
    clearTimeout: (id) => timers.delete(id),
    getComputedStyle,
  });
  const emit = (type, target, extra = {}) => {
    const event = {
      type,
      target,
      button: 0,
      pointerId: 1,
      isPrimary: true,
      clientX: 12,
      clientY: 8,
      ...extra,
    };
    for (const fn of listeners[type] ?? []) fn(event);
    return event;
  };
  const button = () => {
    const b = new Element("BUTTON");
    b.isButton = true;
    return b;
  };
  return {
    Element,
    button,
    emit,
    document,
    animations,
    advance,
    media,
    exports: module.exports,
    mutate: () => observer(),
    blur: () => windows.blur(),
    reduce: () => {
      media.matches = true;
      mediaChange();
    },
  };
}

test("spring follows captured damping/stiffness and reaches its destination", () => {
  const { springFrames, corners } = harness().exports;
  const a = springFrames(0.9, 1400),
    b = springFrames(0.6, 1400),
    c = springFrames(0.9, 700);
  assert.equal(a.samples[0], 0);
  assert.equal(a.samples.at(-1), 1);
  assert.ok(b.samples.some((x) => x > 1));
  assert.ok(c.duration > a.duration);
  assert.deepEqual(
    Array.from(corners("0px 16px 16px 0px", 120, 40)),
    [0, 16, 16, 0],
  );
  assert.deepEqual(
    Array.from(corners("calc(infinity * 1px)", 120, 40)),
    [20, 20, 20, 20],
  );
  assert.throws(() => springFrames(0, 1400));
});

test("mouse ink follows upstream geometry, timing and click release", async () => {
  const h = harness(),
    b = h.button();
  h.emit("pointerdown", b);
  const growth = h.animations.find((a) => a.frames[0].transform);
  assert.equal(growth.options.duration, 450);
  assert.equal(growth.frames[0].transform, "translate(0px, -4px) scale(1)");
  assert.ok(
    growth.frames[1].transform.startsWith("translate(48px, 8px) scale("),
  );
  const layer = b.querySelector(".md-button-ripple");
  assert.equal(layer.children[0].style.width, "24px");
  const entrance = h.animations.at(-1);
  assert.equal(entrance.options.duration, 105);
  assert.equal(entrance.options.easing, "linear");
  h.emit("pointerup", b);
  h.emit("click", b);
  h.advance(224);
  assert.equal(
    h.animations
      .filter((a) => a.frames[0].opacity !== undefined)
      .at(-1)
      .frames.at(-1).opacity,
    1,
  );
  h.advance(1);
  const fade = h.animations.at(-1);
  assert.equal(fade.options.duration, 375);
  assert.equal(fade.options.easing, "linear");
  assert.equal(growth.cancelled, false); // fading overlaps the 450ms expansion
  fade.complete();
  await Promise.resolve();
  assert.equal(b.children.length, 0);
});

test("touch waits 150ms and cancels scrolling without showing ink", () => {
  const h = harness(),
    b = h.button();
  h.emit("pointerdown", b, { pointerType: "touch" });
  h.advance(149);
  assert.equal(b.children.length, 0);
  h.emit("pointercancel", b, { pointerType: "touch" });
  h.advance(1000);
  assert.equal(b.children.length, 0);
  h.emit("pointerdown", b, { pointerType: "touch" });
  h.advance(150);
  assert.equal(b.children.length, 1);
  assert.ok(b.hasAttribute("data-touch"));
  h.emit("pointerout", b, { pointerType: "touch" });
  assert.ok(b.hasAttribute("data-pressed"));
  h.emit("pointerup", b, { pointerType: "touch" });
  h.emit("click", b);
  h.advance(225);
  assert.equal(h.animations.at(-1).options.duration, 375);
});

test("quick touch shows ink on touch-up and still meets minimum visibility", () => {
  const h = harness(),
    b = h.button();
  h.emit("pointerdown", b, { pointerType: "touch" });
  h.advance(40);
  h.emit("pointerup", b, { pointerType: "touch" });
  h.emit("click", b);
  assert.equal(b.children.length, 1);
  h.advance(224);
  assert.equal(
    h.animations.filter((a) => a.frames[0].opacity !== undefined).at(-1).options
      .duration,
    105,
  );
  h.advance(1);
  assert.equal(h.animations.at(-1).options.duration, 375);
});

test("long holds fade immediately after click, and context menus end held ink", () => {
  for (const event of ["click", "contextmenu", "pointercancel", "pointerout"]) {
    const h = harness(),
      b = h.button();
    h.emit("pointerdown", b);
    h.advance(300);
    h.emit(event, b);
    assert.equal(h.animations.at(-1).options.duration, 375, event);
    assert.equal(b.hasAttribute("data-pressed"), false, event);
  }
});

test("a repeated press restarts at its new position and invalidates old release timers", async () => {
  const h = harness(),
    b = h.button();
  h.emit("pointerdown", b);
  h.emit("pointerup", b);
  h.emit("click", b);
  const layer = b.children[0],
    oldGrowth = h.animations.find((a) => a.frames[0].transform);
  h.advance(100);
  h.emit("pointerdown", b, { clientX: 90, clientY: 30 });
  assert.equal(b.children[0], layer);
  assert.equal(b.children.length, 1);
  assert.equal(oldGrowth.cancelled, true);
  const growth = h.animations.filter((a) => a.frames[0].transform).at(-1);
  assert.equal(growth.frames[0].transform, "translate(78px, 18px) scale(1)");
  h.advance(125); // previous release deadline
  assert.equal(
    h.animations.filter((a) => a.frames[0].opacity !== undefined).at(-1).options
      .duration,
    105,
  );
  h.emit("pointerup", b);
  h.emit("click", b);
  h.advance(100);
  assert.equal(h.animations.at(-1).options.duration, 375);
  h.animations.at(-1).complete();
  await Promise.resolve();
  assert.equal(b.children.length, 0);
});

test("keyboard-synthesized clicks create centered ink and toggle selection once", () => {
  const h = harness(),
    b = h.button();
  b.setAttribute("data-button-toggle", "");
  b.setAttribute("aria-pressed", "false");
  h.emit("keydown", b, { key: " " });
  assert.equal(b.children.length, 0);
  h.emit("keyup", b, { key: " " });
  h.emit("click", b, { detail: 0 });
  const growth = h.animations.find((a) => a.frames[0].transform);
  assert.equal(growth.frames[0].transform, "translate(48px, 8px) scale(1)");
  assert.equal(b.getAttribute("aria-pressed"), "true");
  h.advance(225);
  assert.equal(h.animations.at(-1).options.duration, 375);
});

test("reduced motion preserves static feedback and disabled controls reject activation", () => {
  const h = harness(true),
    b = h.button();
  h.emit("pointerdown", b);
  assert.ok(b.hasAttribute("data-pressed"));
  assert.equal(h.animations.length, 0);
  h.emit("pointerup", b);
  h.emit("click", b);
  assert.equal(b.children.length, 0);
  b.disabled = true;
  let prevented = false,
    stopped = false;
  h.emit("click", b, {
    preventDefault() {
      prevented = true;
    },
    stopImmediatePropagation() {
      stopped = true;
    },
  });
  assert.ok(prevented && stopped);
});

test("blur, disablement and reduced motion clear delayed and visible feedback", () => {
  for (const action of ["blur", "disable", "reduce"]) {
    const h = harness(),
      b = h.button();
    h.emit("pointerdown", b);
    if (action === "blur") h.blur();
    else if (action === "reduce") h.reduce();
    else {
      b.disabled = true;
      h.mutate();
    }
    h.advance(1000);
    assert.equal(b.children.length, 0);
    assert.equal(b.hasAttribute("data-pressed"), false);
    assert.ok(h.animations.every((a) => a.cancelled));
  }
  const h = harness(),
    b = h.button();
  h.emit("pointerdown", b, { pointerType: "touch" });
  h.blur();
  h.advance(150);
  assert.equal(b.children.length, 0);
});

test("shared chip feedback handles native input activation once", () => {
  const h = harness(),
    b = h.button();
  b.tagName = "LABEL";
  b.isChip = true;
  b.noSpring = true;
  const input = new h.Element("INPUT");
  b.append(input);
  h.emit("pointerdown", b);
  h.emit("pointerup", b);
  h.emit("click", b); // label event waits for native forwarding
  assert.equal(h.animations.at(-1).options.duration, 105);
  h.emit("click", input, { detail: 1 });
  h.advance(225);
  assert.equal(h.animations.at(-1).options.duration, 375);
  assert.equal(b.hasAttribute("aria-pressed"), false);
});
