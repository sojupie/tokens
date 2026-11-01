// Build-time readiness: fonts, procedural geometry and a fixed motion sample.
(async () => {
 await document.fonts.ready;
 await Promise.all([...document.images].map(image => image.decode()));
 const started = performance.now();
 while ([...document.querySelectorAll('.md-loading path')].some(path => !path.getAttribute('d')) && performance.now()-started < 10000) await new Promise(resolve => requestAnimationFrame(resolve));
 for (const animation of document.getAnimations()) { animation.pause(); animation.currentTime = 400; }
 await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
 document.documentElement.setAttribute('data-scene-ready', '');
})();
