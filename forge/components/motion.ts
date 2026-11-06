import { springState, springFrames, springTrack, cssTiming } from "./helpers/springs.ts";

/* Material spring sampling and popup exit retention. Shared by the compiler and browser. */
type MotionRole = `${'fast' | 'default' | 'slow'}-${'spatial' | 'effects'}`;
declare global { var MaterialMotion: ReturnType<typeof createMotion>; }
function createMotion() {
  function timing(element: Element, role: MotionRole = 'fast-spatial') {
    const style = getComputedStyle(element);
    const duration = parseFloat(style.getPropertyValue(`--_motion-${role}-duration`));
    const easing = style.getPropertyValue(`--_motion-${role}-easing`).trim();
    if (!Number.isFinite(duration) || !easing) throw new Error(`Missing generated ${role} motion`);
    return { duration, easing };
  }
  // Edge-attached surfaces must never translate past their resting edge.
  // Keep the generated spatial duration, with the captured bounded easing.
  function attachedTiming(element: Element, role: MotionRole = 'default-spatial') {
    const duration = timing(element, role).duration;
    const easing = getComputedStyle(element).getPropertyValue('--md-sys-motion-easing-standard').trim();
    if (!easing) throw new Error('Missing captured standard easing for attached motion');
    return { duration, easing };
  }
  // An exit copy needs its used geometry and paint, not hundreds of unrelated
  // computed CSS properties (each read can resolve layout in some browsers).
  const exitProperties=('position top right bottom left display box-sizing width height min-width max-width min-height max-height margin padding border border-radius background color box-shadow opacity overflow transform transform-origin translate rotate scale flex flex-direction flex-wrap align-items align-self align-content justify-content gap grid-template-columns grid-template-rows grid-auto-flow grid-column grid-row order font letter-spacing text-align white-space text-overflow word-break overflow-wrap visibility vertical-align fill fill-rule stroke stroke-width stroke-linecap stroke-linejoin clip-path appearance contain isolation z-index pointer-events direction writing-mode --_button-container --_button-background-state-color --_button-background-state-opacity --_button-state-color --_button-state-opacity --_button-ripple-color --_elevation-level --_shadow-color --_menu-container --_field-menu-label-color').split(' ');
  const exits = new Map<HTMLElement, () => void>();
  function clearExit(element: HTMLElement) { exits.get(element)?.(); }
  function retainExit(element: HTMLElement, mirror = false) {
    if (exits.has(element) || CSS.supports('overlay', 'auto') || matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    const copy = element.cloneNode(true) as HTMLElement, originals = [element, ...(mirror ? element.querySelectorAll<HTMLElement | SVGElement>('*') : [])], copies = [copy, ...(mirror ? copy.querySelectorAll<HTMLElement | SVGElement>('*') : [])];
    const animations: Animation[] = []; let timer: ReturnType<typeof setTimeout> | undefined, cleaned = false;
    const cleanup = () => { if (cleaned) return; cleaned = true; clearTimeout(timer); for (const a of animations) a.cancel(); copy.remove(); if (exits.get(element) === cleanup) exits.delete(element); };
    exits.set(element, cleanup);
    try {
      copy.removeAttribute('id'); for (const child of copy.querySelectorAll('[id]')) child.removeAttribute('id');
      copy.removeAttribute('popover'); copy.setAttribute('data-menu-exit-copy', ''); copy.setAttribute('aria-hidden', 'true'); copy.inert = true;
      originals.forEach((node, i) => {
        const style = getComputedStyle(node);
        // Do not enumerate computed styles: the publication contributes thousands
        // of inherited tokens to every descendant's CSSStyleDeclaration.
        const declarations=exitProperties.map(property=>`${property}:${style.getPropertyValue(property)};`);
        copies[i].style.cssText=declarations.join('')+'transition:none!important;';
      });
      copy.style.setProperty('pointer-events', 'none', 'important'); copy.style.setProperty('z-index', '2147483647');
      (element.closest('dialog[open]') || document.body).append(copy);
      copy.scrollTop = element.scrollTop; copy.scrollLeft = element.scrollLeft;
      if (mirror) {
        originals.forEach((node, i) => { for (const original of node.getAnimations()) {
          const effect = original.effect as KeyframeEffect | null;
          if (effect?.target !== node) continue;
          const animation = copies[i].animate(effect.getKeyframes(), effect.getTiming());
          animation.currentTime = original.currentTime ?? 0; animations.push(animation);
        } });
      } else {
        const style = getComputedStyle(element);
        animations.push(copy.animate([{ opacity: style.opacity }, { opacity: 0 }], { ...timing(element, 'fast-effects'), fill: 'forwards' }));
        animations.push(copy.animate([{ scale: style.scale || '1' }, { scale: '.8' }], { ...timing(element), fill: 'forwards' }));
      }
      const duration = Math.max(0, ...animations.map(a => Number(a.effect!.getComputedTiming().endTime)));
      timer = setTimeout(cleanup, duration + 50);
      Promise.all(animations.map(a => a.finished)).then(cleanup, cleanup);
    } catch { cleanup(); }
  }
  const api = { springState, springFrames, springTrack, cssTiming, timing, attachedTiming, retainExit, clearExit };
  if (typeof document !== 'undefined' && typeof matchMedia !== 'undefined') matchMedia('(prefers-reduced-motion: reduce)').addEventListener?.('change', event => { if (event.matches) for (const cleanup of [...exits.values()]) cleanup(); });
  return api;
}
const api = createMotion();
if (typeof module !== 'undefined') module.exports = api;
globalThis.MaterialMotion = api;
