// Canvas is an optional renderer for the same token-driven hero choreography.
(() => {
  function turnProgress(time, duration, easing) {
    const t = Math.max(0, Math.min(1, time / duration));
    if (t === 0 || t === 1 || easing === "linear") return t;
    const points = easing
      .match(/^cubic-bezier\(([^)]+)\)$/)?.[1]
      .split(",")
      .map(Number);
    if (!points || points.length !== 4 || !points.every(Number.isFinite))
      throw new Error("Unsupported artwork turn easing");
    const sample = (s, a, b) =>
      3 * (1 - s) * (1 - s) * s * a + 3 * (1 - s) * s * s * b + s * s * s;
    let low = 0,
      high = 1;
    for (let i = 0; i < 16; i++) {
      const mid = (low + high) / 2;
      if (sample(mid, points[0], points[2]) < t) low = mid;
      else high = mid;
    }
    return sample((low + high) / 2, points[1], points[3]);
  }
  function trackValues(track, scene, progress) {
    return track[scene].map(
      (from, i) =>
        from + (track[(scene + 1) % track.length][i] - from) * progress,
    );
  }
  function tileBounds(width, height, gap, columns, rows) {
    const boxes = [];
    let x = 0;
    columns.forEach((fraction, col) => {
      const w =
        ((width - 2 * gap) * fraction) / columns.reduce((a, b) => a + b, 0);
      let y = 0;
      rows[col].forEach((part, row) => {
        const h =
          ((height - 2 * gap) * part) / rows[col].reduce((a, b) => a + b, 0);
        boxes.push({ x, y, width: w, height: h, col, row });
        y += h + gap;
      });
      x += w + gap;
    });
    return boxes;
  }
  if (typeof module !== "undefined")
    module.exports = { turnProgress, trackValues, tileBounds };
  if (typeof document === "undefined") return;
  globalThis.TokenArtworkCanvas = (root, model) => {
    const canvas = root.querySelector("canvas"),
      ctx = canvas.getContext("2d");
    if (!ctx) return null; // Keep the HTML tiles and let the DOM renderer run.
    const {
      tracks,
      spring,
      shapeSpring,
      corners,
      turnDuration,
      turnEasing,
      loop,
      progress,
      cornerStates,
    } = model;
    const tileCorners = Array.from({ length: 9 }, (_, i) =>
      cornerStates(tracks, Math.floor(i / 3), i % 3, corners),
    );
    const tiles = [...root.querySelectorAll(".tokens-art-tile")];
    let width = 0,
      height = 0,
      gap = 8,
      colors = [],
      running = false,
      elapsed = 0,
      started = 0,
      raf = 0,
      timer = 0,
      disposed = false;
    const now = () => elapsed + (running ? performance.now() - started : 0);
    function readColors() {
      // Resolve color-mix and theme aliases through CSS only when appearance changes.
      const probe = document.createElement("span");
      probe.hidden = true;
      colors = tiles.map((tile) => {
        tile.append(probe);
        return ["a", "b", "c"].map((key) => {
          probe.style.color = `var(--art-${key})`;
          return getComputedStyle(probe).color;
        });
      });
      probe.remove();
    }
    const renderer = {
      draw(time) {
        if (!width || !height || disposed) return;
        const sceneLength = loop / 4,
          phase = ((time % loop) + loop) % loop,
          scene = Math.floor(phase / sceneLength),
          within = phase % sceneLength;
        const spatial = progress(spring, within),
          shape = progress(shapeSpring, within),
          turn = turnProgress(within, turnDuration, turnEasing);
        const columns = trackValues(tracks[0], scene, spatial),
          rows = tracks
            .slice(1)
            .map((track) => trackValues(track, scene, spatial));
        ctx.clearRect(0, 0, width, height);
        for (const [index, box] of tileBounds(
          width,
          height,
          gap,
          columns,
          rows,
        ).entries()) {
          const { x, y, width: w, height: h, col, row } = box;
          const radii = trackValues(tileCorners[index], scene, shape).map(
            (value) => Math.max(0, value),
          );
          const phase = (col + row) % 4,
            from = Math.floor(((scene + phase) % 4) / 2) * 90,
            to = Math.floor(((((scene + 1) % 4) + phase) % 4) / 2) * 90;
          const angle = ((from + (to - from) * turn) * Math.PI) / 180,
            c = Math.abs(Math.cos(angle)),
            s = Math.abs(Math.sin(angle));
          const rw = w * c + h * s,
            rh = w * s + h * c;
          ctx.save();
          ctx.beginPath();
          ctx.roundRect(x, y, w, h, radii);
          ctx.clip();
          ctx.translate(x + w / 2, y + h / 2);
          ctx.rotate(angle);
          const bands = ctx.createLinearGradient(-rw / 2, 0, rw / 2, 0),
            palette = colors[index];
          [
            [0, palette[0]],
            [1 / 3, palette[0]],
            [1 / 3, palette[1]],
            [2 / 3, palette[1]],
            [2 / 3, palette[2]],
            [1, palette[2]],
          ].forEach(([stop, color]) => bands.addColorStop(stop, color));
          ctx.fillStyle = bands;
          ctx.fillRect(-rw / 2, -rh / 2, rw, rh);
          const direction = (125 * Math.PI) / 180,
            dx = Math.sin(direction),
            dy = -Math.cos(direction),
            length = Math.abs(rw * dx) + Math.abs(rh * dy);
          const glaze = ctx.createLinearGradient(
            (-dx * length) / 2,
            (-dy * length) / 2,
            (dx * length) / 2,
            (dy * length) / 2,
          );
          glaze.addColorStop(0, "rgba(255,255,255,0)");
          glaze.addColorStop(1, "rgba(255,255,255,.16)");
          ctx.fillStyle = glaze;
          ctx.fillRect(-rw / 2, -rh / 2, rw, rh);
          ctx.restore();
        }
      },
      play() {
        if (running || disposed) return;
        running = true;
        started = performance.now();
        frame();
      },
      pause() {
        if (running) elapsed = now();
        running = false;
        cancelAnimationFrame(raf);
        clearTimeout(timer);
      },
      dispose() {
        renderer.pause();
        disposed = true;
        resize.disconnect();
        appearance.disconnect();
        window.removeEventListener("resize", measure);
        root.removeAttribute("data-canvas-ready");
      },
    };
    function frame() {
      if (!running || disposed) return;
      const time = now();
      renderer.draw(time);
      const within = time % (loop / 4),
        settled = Math.max(
          spring?.duration || 0,
          shapeSpring?.duration || 0,
          turnDuration,
        );
      // Settled scenes have no visual changes; sleep until the next shared beat.
      if (within >= settled) timer = setTimeout(frame, loop / 4 - within);
      else raf = requestAnimationFrame(frame);
    }
    function measure() {
      const box = root.getBoundingClientRect();
      width = box.width;
      height = box.height;
      if (width <= 0 || height <= 0) return;
      // Bound the decorative bitmap density; the profile records the effective DPR.
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      canvas.width = Math.max(1, Math.round(width * dpr));
      canvas.height = Math.max(1, Math.round(height * dpr));
      ctx.setTransform(
        canvas.width / width,
        0,
        0,
        canvas.height / height,
        0,
        0,
      );
      gap = parseFloat(getComputedStyle(root).gap);
      renderer.draw(now());
    }
    const resize = new ResizeObserver(measure),
      appearance = new MutationObserver(() => {
        readColors();
        renderer.draw(now());
      });
    readColors();
    root.setAttribute("data-canvas-ready", "");
    measure();
    resize.observe(root);
    appearance.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["data-theme", "data-contrast", "style"],
    });
    window.addEventListener("resize", measure);
    return renderer;
  };
})();
