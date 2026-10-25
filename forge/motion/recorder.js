// The recording lifecycle is adapted from Carousel Motion Lab's MotionRecorder.
// Browser samples describe DOM/CSS state, not physical display presentation.
(() => {
  class MotionRecorder {
    constructor({
      readFrame,
      now = () => performance.now(),
      requestFrame = (fn) => requestAnimationFrame(fn),
      cancelFrame = (id) => cancelAnimationFrame(id),
    }) {
      Object.assign(this, { readFrame, now, requestFrame, cancelFrame });
      this.active = null;
    }
    start(metadata = {}) {
      if (this.active) throw Error("Recording already started.");
      this.active = {
        schemaVersion: 1,
        kind: "browser-values",
        ...metadata,
        started: this.now(),
        events: [],
        frames: [],
      };
      this.active.epoch = performance.timeOrigin + this.active.started;
      this.event("recording-start");
      this.frame = this.requestFrame((t) => this.sample(t));
    }
    event(type, detail = {}) {
      if (this.active)
        this.active.events.push({
          ...detail,
          t: this.now() - this.active.started,
          type,
        });
    }
    sample(t) {
      if (!this.active) return;
      try {
        this.active.frames.push({
          t: t - this.active.started,
          values: this.readFrame(),
        });
      } catch (error) {
        this.active.samplingError = error.message;
        this.stop("sampling error");
        return;
      }
      if (t - this.active.started >= 60000) {
        this.stop("limit");
        return;
      }
      this.frame = this.requestFrame((next) => this.sample(next));
    }
    stop(reason = "manual") {
      if (!this.active) return this.result;
      this.event("recording-stop", { reason });
      this.cancelFrame(this.frame);
      this.active.duration = this.now() - this.active.started;
      this.result = this.active;
      this.active = null;
      return this.result;
    }
  }
  function instrument(config) {
    const elements = new Map(),
      definitions = new Map(),
      measurementIssues = new Map(),
      animationIDs = new WeakMap();
    let nextAnimation = 1;
    const allowed = [
      "x",
      "y",
      "width",
      "height",
      "opacity",
      "translateX",
      "translateY",
      "scaleX",
      "scaleY",
      "borderTopLeftRadius",
      "borderTopRightRadius",
      "borderBottomLeftRadius",
      "borderBottomRightRadius",
    ];
    for (const track of config.tracks) {
      if (!track.name || !allowed.includes(track.property))
        throw Error("Invalid track " + JSON.stringify(track));
      if (elements.has(track.name))
        throw Error("Duplicate track name " + track.name);
      const matches = document.querySelectorAll(track.selector);
      if (matches.length !== 1 && !(track.optional && matches.length === 0))
        throw Error(
          `${track.selector}: expected one element, found ${matches.length}`,
        );
      elements.set(track.name, matches[0]);
    }
    const recorder = new MotionRecorder({
      readFrame: () => {
        const values = {},
          cache = new Map();
        for (const track of config.tracks) {
          const el = track.optional
            ? document.querySelector(track.selector)
            : elements.get(track.name);
          if (!el?.isConnected) {
            values[track.name] = null;
            if (!track.optional)
              measurementIssues.set(track.name, "Tracked element was removed.");
            continue;
          }
          if (!cache.has(el))
            cache.set(el, {
              rect: el.getBoundingClientRect(),
              style: getComputedStyle(el),
            });
          const { rect, style } = cache.get(el),
            p = track.property;
          const matrix = [
            "translateX",
            "translateY",
            "scaleX",
            "scaleY",
          ].includes(p)
            ? new DOMMatrixReadOnly(style.transform)
            : null;
          const number =
            p === "translateX"
              ? matrix.m41
              : p === "translateY"
                ? matrix.m42
                : p === "scaleX"
                  ? Math.hypot(matrix.m11, matrix.m12, matrix.m13)
                  : p === "scaleY"
                    ? Math.hypot(matrix.m21, matrix.m22, matrix.m23)
                    : p in rect
                      ? rect[p]
                      : p.startsWith("border") && style[p].includes("%")
                        ? NaN
                        : parseFloat(style[p]);
          values[track.name] = Number.isFinite(number) ? number : null;
          if (!Number.isFinite(number))
            measurementIssues.set(
              track.name,
              "Computed value could not be measured as a numeric length or ratio; inspect percentage radii and unresolved values.",
            );
          for (const animation of el.getAnimations()) {
            // Keep each actual animation instance and its timing origin, not just
            // its declared easing. Input time and animation start are different.
            if (!animationIDs.has(animation)) {
              const id = nextAnimation++;
              animationIDs.set(animation, id);
              definitions.set(id, {
                id,
                selector: track.selector,
                timing: animation.effect?.getTiming(),
                keyframes: animation.effect?.getKeyframes(),
                observedAt: performance.now() - recorder.active.started,
              });
            }
            const definition = definitions.get(animationIDs.get(animation));
            definition.startTime =
              typeof animation.startTime === "number"
                ? animation.startTime - recorder.active.started
                : null;
          }
        }
        return values;
      },
    });
    const events = [
      "pointerdown",
      "pointerup",
      "pointercancel",
      "touchstart",
      "touchend",
      "keydown",
      "click",
    ];
    const handlers = events.map((type) => {
      const fn = (e) =>
        recorder.event(type, {
          key: e.key,
          pointerType: e.pointerType,
          target: e.target.tagName,
        });
      document.addEventListener(type, fn, true);
      return [type, fn];
    });
    recorder.start({
      url: location.href,
      title: document.title,
      tracks: config.tracks,
      userAgent: navigator.userAgent,
      viewport: {
        width: innerWidth,
        height: innerHeight,
        dpr: devicePixelRatio,
      },
      reducedMotion: matchMedia("(prefers-reduced-motion: reduce)").matches,
    });
    return {
      mark: (type, detail) => recorder.event(type, detail),
      stop: () => {
        for (const [type, fn] of handlers)
          document.removeEventListener(type, fn, true);
        const report = recorder.stop();
        report.animations = [...definitions.values()];
        report.measurementIssues = [...measurementIssues].map(
          ([track, reason]) => ({ track, reason }),
        );
        return report;
      },
    };
  }
  function bezier(x, points) {
    const [x1, y1, x2, y2] = points;
    let lo = 0,
      hi = 1;
    for (let i = 0; i < 32; i++) {
      const t = (lo + hi) / 2,
        v =
          3 * (1 - t) * (1 - t) * t * x1 + 3 * (1 - t) * t * t * x2 + t * t * t;
      if (v < x) lo = t;
      else hi = t;
    }
    const t = (lo + hi) / 2;
    return (
      3 * (1 - t) * (1 - t) * t * y1 + 3 * (1 - t) * t * t * y2 + t * t * t
    );
  }
  function series(report, name, origin = 0) {
    let segment = 0;
    const samples = [];
    for (const f of report.frames) {
      if (!Number.isFinite(f.values?.[name])) {
        segment++;
        continue;
      }
      samples.push({ t: f.t - origin, value: f.values[name], segment });
    }
    return samples;
  }
  function summarize(samples, { start, end, from, to, easing } = {}) {
    if (!samples.length) throw Error("No finite samples.");
    start ??= samples[0].t;
    end ??= samples.at(-1).t;
    from ??= samples[0].value;
    to ??= samples.at(-1).value;
    const selected = samples.filter((s) => s.t >= start && s.t <= end),
      span = to - from;
    if (!(end > start) || !span || !selected.length)
      throw Error("Choose distinct endpoints and an increasing time range.");
    let error = 0,
      maxError = 0;
    const normalized = selected.map((s) => {
      const progress = (s.value - from) / span,
        time = (s.t - start) / (end - start),
        expected = easing ? bezier(time, easing) : null;
      if (expected !== null) {
        const delta = progress - expected;
        error += delta * delta;
        maxError = Math.max(maxError, Math.abs(delta));
      }
      return { t: s.t, value: s.value, progress, expected };
    });
    return {
      start,
      end,
      duration: end - start,
      from,
      to,
      overshoot: Math.max(0, ...normalized.map((s) => s.progress - 1)),
      undershoot: Math.max(0, ...normalized.map((s) => -s.progress)),
      rmsProgressError: easing ? Math.sqrt(error / normalized.length) : null,
      maxProgressError: easing ? maxError : null,
      samples: normalized,
    };
  }
  const api = { MotionRecorder, instrument, bezier, series, summarize };
  if (typeof module !== "undefined") module.exports = api;
  else globalThis.MotionInspection = api;
})();
