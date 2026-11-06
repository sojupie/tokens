type Clock<Timer> = {
  now(): number;
  setTimeout(callback: () => void, delay: number): Timer;
  clearTimeout(timer: Timer | undefined): void;
};
type Edges = Pick<DOMRectReadOnly, 'left' | 'right' | 'top' | 'bottom'>;
type Size = { width: number; height: number };

// Snackbar timing is renderer behavior; Material's captured tokens contain no timeout.
export function snackbarTimer<Timer>(close: () => void, clock: Clock<Timer>) {
  let timer: Timer | undefined,
    remaining = 0,
    started = 0,
    active = false;
  const pauses = new Set<string>();
  const clear = () => {
    clock.clearTimeout(timer);
    timer = undefined;
  };
  const schedule = () => {
    if (!active || pauses.size || !remaining) return;
    started = clock.now();
    timer = clock.setTimeout(() => {
      active = false;
      close();
    }, remaining);
  };
  return {
    start(duration: number) {
      clear();
      pauses.clear();
      remaining = duration;
      active = duration > 0;
      schedule();
    },
    stop() {
      clear();
      active = false;
      pauses.clear();
    },
    pause(reason: string) {
      if (!pauses.size && active) {
        remaining = Math.max(0, remaining - (clock.now() - started));
        clear();
      }
      pauses.add(reason);
    },
    resume(reason: string) {
      if (!pauses.delete(reason)) return;
      if (active && !pauses.size) {
        if (remaining <= 0) {
          active = false;
          close();
        } else schedule();
      }
    },
  };
}
export function toolbarGap(available: number, widths: number[], min: number, max: number) {
  if (widths.length < 2) return 0;
  return Math.max(
    min,
    Math.min(
      max,
      (available - widths.reduce((sum, width) => sum + width, 0)) /
        (widths.length - 1),
    ),
  );
}
export function sliderValues(values: number[], index: number, min: number, max: number) {
  const result = values.map((value) => Math.max(min, Math.min(max, value)));
  if (result.length === 2) {
    if (index === 0) result[0] = Math.min(result[0], result[1]);
    else result[1] = Math.max(result[0], result[1]);
  }
  return result;
}
export function splitMenuPosition(trigger: Edges, group: Pick<Edges, "left" | "right">, menu: Size, viewport: Size, rtl: boolean, gap: number, margin: number) {
  const fits = (x: number) =>
    x >= margin && x + menu.width <= viewport.width - margin;
  // Prefer the trailing edge, then the other trailing edge and group edges.
  const edges = rtl
    ? [
        trigger.left,
        trigger.right - menu.width,
        group.left,
        group.right - menu.width,
      ]
    : [
        trigger.right - menu.width,
        trigger.left,
        group.right - menu.width,
        group.left,
      ];
  const left =
    edges.find(fits) ??
    Math.max(
      margin,
      Math.min(edges[0], viewport.width - menu.width - margin),
    );
  const below = trigger.bottom + gap,
    above = trigger.top - menu.height - gap;
  const top =
    below + menu.height <= viewport.height - margin
      ? below
      : above >= margin
        ? above
        : Math.max(
            margin,
            Math.min(below, viewport.height - menu.height - margin),
          );
  return { left, top };
}