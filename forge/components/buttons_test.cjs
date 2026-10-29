function runMotionScript(code, globals) {
  runInNewContext(readFileSync(require.resolve('./static/motion.cjs'), 'utf8'), globals);
  return runInNewContext(code, globals);
}
const { test } = require("node:test");
const assert = require("node:assert/strict");
const { readFileSync } = require("node:fs");
const { runInNewContext } = require("node:vm");
const source = readFileSync(`${__dirname}/static/buttons.js`, "utf8");

function harness(reduced = false, initialCheckboxes = []) {
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
      if (selector === ".md-slider-input") return !!this.isSlider;
      if (selector === '.md-radio,.md-checkbox,.md-switch') return !!this.isSelectionControl;
      if (selector === ".md-filter-chip,.md-assist-chip,.md-suggestion-chip,.md-input-chip,.md-radio,.md-checkbox,.md-switch,.md-segmented-button,.md-slider-input") return !!(this.isChip || this.isSelectionControl);
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
    querySelectorAll: () => initialCheckboxes.filter(input => input.pending),
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
    ...Object.fromEntries(["TopLeft","TopRight","BottomRight","BottomLeft"].map((name,index)=>[
      "border"+name+"Radius",button.style.borderRadius
        ? module.exports.corners(button.style.borderRadius,120,40)[index]+"px"
        : button.hasAttribute("data-pressed")?"8px":"20px"
    ])),
    paddingInlineStart: "24px",
    paddingInlineEnd: "24px",
    getPropertyValue: (name) =>
      button.noSpring && name.includes("spring")
        ? ""
        : ({
            "--_group-width-growth": "15%",
            "--_group-width-damping": "0.9",
            "--_group-width-stiffness": "1400",
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
  runMotionScript(source, {
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
    mutate: (records = []) => observer(records),
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
  assert.deepEqual(Array.from(corners("calc(infinity * 1px) 4px 4px calc(infinity * 1px)",120,40)),[20,4,4,20]);
  assert.throws(() => springFrames(0, 1400));
});

test("a held press returns from its pressed corners without restarting on click", async () => {
  const h = harness(),
    b = h.button();
  h.emit("pointerdown", b);
  h.animations.find((a) => a.frames[0].borderRadius).complete();
  await Promise.resolve();
  h.emit("pointerup", b);
  const shapes = h.animations.filter((a) => a.frames[0].borderRadius);
  const returning = shapes.at(-1);
  assert.equal(returning.frames[0].borderRadius, "8px 8px 8px 8px");
  assert.equal(returning.frames.at(-1).borderRadius, "20px 20px 20px 20px");
  assert.ok(
    returning.frames.some(
      (f) => parseFloat(f.borderRadius) > 8 && parseFloat(f.borderRadius) < 20,
    ),
  );
  h.emit("click", b);
  assert.equal(returning.cancelled, false);
  assert.equal(
    h.animations.filter((a) => a.frames[0].borderRadius).length,
    shapes.length,
  );
});

test("keyboard release also animates from the held pressed shape", async () => {
  const h = harness(),
    b = h.button();
  h.emit("keydown", b, { key: " " });
  h.animations.find((a) => a.frames[0].borderRadius).complete();
  await Promise.resolve();
  h.emit("keyup", b, { key: " " });
  const returning = h.animations.filter((a) => a.frames[0].borderRadius).at(-1);
  assert.equal(returning.frames[0].borderRadius, "8px 8px 8px 8px");
  assert.equal(returning.frames.at(-1).borderRadius, "20px 20px 20px 20px");
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

test("application selection changes retarget the release shape", () => {
  const h = harness(), b = h.button();
  b.setAttribute("aria-pressed", "false");
  h.emit("pointerdown", b);
  h.emit("pointerup", b);
  h.emit("click", b);
  const returning = h.animations.filter(a => a.frames[0].borderRadius).at(-1);
  assert.equal(returning.frames.at(-1).borderRadius, "20px 20px 20px 20px");
  b.setAttribute("aria-pressed", "true");
  h.mutate([{target:b, attributeName:"aria-pressed", oldValue:"false"}]);
  const selected = h.animations.filter(a => a.frames[0].borderRadius).at(-1);
  assert.equal(returning.cancelled, true);
  assert.equal(selected.frames.at(-1).borderRadius, "12px 12px 12px 12px");
  const count = h.animations.length;
  h.mutate([{target:b, attributeName:"aria-pressed", oldValue:"true"}]);
  assert.equal(h.animations.length, count);
  b.setAttribute("aria-pressed", "false");
  h.mutate([{target:b, attributeName:"aria-pressed", oldValue:"true"}]);
  const unselected = h.animations.filter(a => a.frames[0].borderRadius).at(-1);
  assert.equal(selected.cancelled, true);
  assert.equal(unselected.frames.at(-1).borderRadius, "20px 20px 20px 20px");
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


test("indeterminate checkboxes initialize on page load and HTMX swaps", () => {
  const input = () => ({ pending:true, indeterminate:false, removeAttribute(name) { assert.equal(name, "data-indeterminate"); this.pending=false; } });
  const first = input(), inputs = [first];
  const h = harness(false, inputs);
  assert.equal(first.indeterminate, true);
  assert.equal(first.pending, false);
  first.indeterminate = false;
  const next = input();inputs.push(next);
  h.emit("htmx:afterSwap", null, {detail:{target:{isConnected:false,querySelectorAll:()=>{throw new Error("detached swap target");}}}});
  assert.equal(next.indeterminate, true);
  assert.equal(first.indeterminate, false);
});


for (const family of ["checkbox", "switch", "segmented"]) {
 test(`${family} shares pointer and native keyboard ripple lifecycle`, () => {
  const h = harness(), label = h.button();
  label.tagName="LABEL"; label.isSelectionControl=true; label.noSpring=true;
  const input = new h.Element("INPUT"); label.append(input);
  let host=label;
  if(family==="switch") { host=new h.Element("SPAN");host.className="md-switch-handle";label.append(host); }
  h.emit("pointerdown", input);
  assert.equal(host.children.at(-1).className,"md-button-ripple");
  h.emit("pointerup", input);
  h.emit("click", label);
  h.emit("click", input, {detail:1});
  assert.equal(host.children.filter(e=>e.className==="md-button-ripple").length,1);
  h.advance(225);
  assert.equal(h.animations.at(-1).options.duration,375);
  h.emit("keydown",input,{key:" "});
  assert.equal(label.hasAttribute("data-pressed"),true);
  h.emit("keyup",input,{key:" "});
  h.emit("click",input,{detail:0});
  assert.equal(label.hasAttribute("data-pressed"),false);
  h.emit("pointerdown",input);
  h.emit("pointercancel",input);
  assert.equal(label.hasAttribute("data-pressed"),false);
  label.disabledChild=true;
  const count=h.animations.length;
  h.emit("pointerdown",input);h.emit("keydown",input,{key:" "});
  assert.equal(h.animations.length,count);
 });
}

test("navigation item contains its ripple within the indicator", () => {
 const h=harness(),button=h.button(); button.noSpring=true;
 const indicator=new h.Element("SPAN");indicator.className="md-nav-indicator";button.append(indicator);
 h.emit("pointerdown",button);
 assert.equal(indicator.children.at(-1).className,"md-button-ripple");
 h.emit("pointerup",button);h.emit("click",button);h.advance(225);
 assert.equal(h.animations.at(-1).options.duration,375);
});

test("slider arrow keys start and release ripple in the moving state-layer host",()=>{
 const h=harness(),control=h.button();control.isSlider=true;control.noSpring=true;
 const input=new h.Element("INPUT");control.append(input);
 const host=new h.Element("SPAN");host.className="md-slider-state-layer";control.append(host);
 h.emit("keydown",input,{key:"ArrowRight"});
 assert.equal(control.hasAttribute("data-pressed"),true);
 assert.equal(host.children.at(-1).className,"md-button-ripple");
 h.emit("keyup",input,{key:"ArrowRight"});h.advance(225);
 assert.equal(control.hasAttribute("data-pressed"),false);
 assert.equal(h.animations.at(-1).options.duration,375);
});


test("standard group width growth preserves total width and only borrows neighboring padding",()=>{
 const {groupWidths}=harness().exports;
 assert.deepEqual(Array.from(groupWidths([100,100,100],1,.15,[24,24,24])),[92.5,115,92.5]);
 assert.deepEqual(Array.from(groupWidths([100,100,100],0,.15,[24,24,24])),[115,85,100]);
 assert.deepEqual(Array.from(groupWidths([100,100,100],2,.15,[24,24,24])),[100,85,115]);
 assert.deepEqual(Array.from(groupWidths([100,100,100],1,.15,[4,24,10])),[96,108,96]);
 assert.deepEqual(Array.from(groupWidths([100],0,.15,[24])),[100]);
 assert.deepEqual(Array.from(groupWidths([100,100,100],-1,.15,[24,24,24])),[100,100,100]);
});


test("FAB menu reveal starts next to the trigger and reverses for dismissal",()=>{
 const {fabMenuDelays}=harness().exports;
 const spring={samples:[0,.1,.2,.3,.4,.5,.6,.7,.8,.9,1],duration:1000};
 assert.deepEqual(Array.from(fabMenuDelays(3,spring,true)),[900,500,200]);
 assert.deepEqual(Array.from(fabMenuDelays(3,spring,false)),[200,500,900]);
 assert.deepEqual(Array.from(fabMenuDelays(1,spring,true)),[500]);
});

test("compound input chip uses shared keyboard and pointer feedback",()=>{
 const h=harness(),root=h.button();root.tagName="SPAN";root.isChip=true;root.noSpring=true;
 const action=new h.Element("BUTTON");root.append(action);
 h.emit("keydown",action,{key:" "});assert.equal(root.hasAttribute("data-pressed"),true);
 h.emit("keyup",action,{key:" "});assert.equal(root.hasAttribute("data-pressed"),false);
 h.emit("pointerdown",action);assert.equal(root.children.at(-1).className,"md-button-ripple");
 h.emit("pointercancel",action);assert.equal(root.hasAttribute("data-pressed"),false);
 root.setAttribute("aria-disabled","true");const count=h.animations.length;h.emit("pointerdown",action);assert.equal(h.animations.length,count);
});

test('native radio input owns pointer and Space feedback inside its ripple host',()=>{
 const h=harness(),host=h.button();host.tagName='SPAN';host.isSelectionControl=true;host.noSpring=true;
 const input=new h.Element('INPUT');host.append(input);
 h.emit('pointerdown',input);h.emit('pointerup',input);h.emit('click',input,{detail:1});
 assert.equal(host.children.filter(x=>x.className==='md-button-ripple').length,1);
 h.advance(225);h.emit('keydown',input,{key:' '});assert.equal(host.hasAttribute('data-pressed'),true);
 h.emit('keyup',input,{key:' '});h.emit('click',input,{detail:0});assert.equal(host.hasAttribute('data-pressed'),false);
 host.disabledChild=true;const count=h.animations.length;h.emit('pointerdown',input);h.emit('keydown',input,{key:' '});assert.equal(h.animations.length,count);
});
test('Enter does not press native checkbox, switch or radio controls',()=>{
 const h=harness(),host=h.button();host.isSelectionControl=true;host.noSpring=true;host.tagName='SPAN';const input=new h.Element('INPUT');host.append(input);
 h.emit('keydown',input,{key:'Enter'});assert.equal(host.hasAttribute('data-pressed'),false);assert.equal(h.animations.length,0);
});

test("CSS-only shape transitions are not disturbed by spring target probes", () => {
 const h=harness(),b=h.button();b.noSpring=true;
 let writes=0;
 Object.defineProperty(b.style,'borderRadius',{get:()=>'',set:()=>{writes++;}});
 b.setAttribute('aria-expanded','true');
 h.mutate([{target:b,attributeName:'aria-expanded',oldValue:'false'}]);
 assert.equal(writes,0);
});


test("click releases feedback when focus changed the control class", async () => {
 const h=harness(),button=h.button();button.noSpring=true;
 h.emit("pointerdown",button);
 button.isButton=false;
 h.emit("pointerup",button);h.emit("click",button);h.advance(225);
 const fade=h.animations.at(-1);
 assert.equal(fade.options.duration,375);
 fade.complete();await Promise.resolve();await Promise.resolve();
 assert.equal(button.hasAttribute("data-rippling"),false);
 assert.equal(button.querySelector(".md-button-ripple"),null);
});


function widthGroupFixture(reduced=false) {
 const h=harness(reduced),group=new h.Element('DIV');group.dataset={control:'standard'};
 const buttons=Array.from({length:3},()=>h.button());
 group.querySelectorAll=()=>buttons;
 for(const button of buttons){
  group.append(button);button.style.width='';button.naturalWidth=100;
  button.closest=selector=>selector==='.md-button-group'?group:selector==='.md-button-ripple'?null:button;
  button.getBoundingClientRect=()=>({left:0,top:0,width:parseFloat(button.style.width)||button.naturalWidth,height:40});
 }
 const widths=()=>buttons.map(button=>button.style.width);
 const finish=async()=>{for(const animation of [...h.animations])if(!animation.cancelled)animation.complete();await new Promise(setImmediate);};
 return {...h,group,buttons,widths,finish};
}

test('button group retargets changed label width before release animations finish',async()=>{
 const h=widthGroupFixture(),button=h.buttons[1];
 h.emit('pointerdown',button);h.emit('pointerup',button);h.emit('click',button,{detail:1});
 // Release is still running when the application's click handler changes Play to Pause.
 const old=h.animations.filter(a=>!a.cancelled);
 button.naturalWidth=140;
 h.mutate([{type:'childList',target:button,addedNodes:[],removedNodes:[]}]);
 assert.deepEqual(h.widths(),['100px','140px','100px']);
 assert.ok(old.some(a=>a.cancelled));
 await h.finish();
 assert.deepEqual(h.widths(),['','','']);
 assert.equal(button.getBoundingClientRect().width,140);
 assert.equal(h.group.hasAttribute('data-width-animating'),false);
});

test('button group preserves held growth when content changes and ignores ripple nodes',async()=>{
 const h=widthGroupFixture(),button=h.buttons[1];h.emit('pointerdown',button);
 button.naturalWidth=140;
 const text={nodeType:3,parentElement:button};
 h.mutate([{type:'characterData',target:text}]);
 assert.deepEqual(h.widths(),['89.5px','161px','89.5px']);
 const count=h.animations.length,ripple=new h.Element();ripple.className='md-button-ripple';
 h.mutate([{type:'childList',target:button,addedNodes:[ripple],removedNodes:[]}]);
 assert.equal(h.animations.length,count);
 h.emit('pointercancel',button);await h.finish();
 assert.deepEqual(h.widths(),['','','']);
});

test('reduced motion restores the changed natural width immediately',()=>{
 const h=widthGroupFixture(true),button=h.buttons[1];h.emit('pointerdown',button);
 button.naturalWidth=140;h.mutate([{type:'childList',target:button}]);
 h.emit('pointercancel',button);
 assert.deepEqual(h.widths(),['','','']);
 assert.equal(button.getBoundingClientRect().width,140);
});


test('standard group width animation includes icon-only commands',async()=>{
 const h=widthGroupFixture();
 h.group.querySelectorAll=selector=>selector.includes('.md-icon-button')?h.buttons:[h.buttons[1]];
 h.emit('pointerdown',h.buttons[0]);
 assert.ok(h.buttons.every(button=>button.style.width!==''),'all group children participate');
 h.emit('pointerup',h.buttons[0]);await h.finish();
});
