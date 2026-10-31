// Keep review framing outside the app viewport; no screenshot or image pipeline.
(() => {
 const controls = document.querySelector('.scene-review-controls');
 const apply = (settings) => {
  document.documentElement.toggleAttribute('data-scene-grid', settings.grid);
  document.documentElement.dataset.theme = settings.theme;
  document.documentElement.dataset.contrast = settings.contrast;
 };
 if (!controls) {
  window.addEventListener('message', (event) => {
   if (event.origin !== window.location.origin || event.source !== window.parent || event.data?.type !== 'scene-review') return;
   apply(event.data);
  });
  return;
 }
 const frames = [...document.querySelectorAll('.scene-review-viewport iframe')];
 const settings = () => ({type:'scene-review', grid:controls.querySelector('[data-review-grid]').checked, theme:controls.querySelector('[data-review-theme]').value, contrast:controls.querySelector('[data-review-contrast]').value});
 const send = (frame) => frame.contentWindow?.postMessage(settings(),window.location.origin);
 controls.addEventListener('change', () => {apply(settings());frames.forEach(send);});
 const observer = new ResizeObserver((entries) => {
  for (const {target, contentRect} of entries) {
   const frame=target.querySelector('iframe');
   frame.style.transform=`scale(${Math.min(1,contentRect.width/Number(frame.width))})`;
  }
 });
 for (const frame of frames) {observer.observe(frame.parentElement);frame.addEventListener('load',()=>send(frame));}
})();
