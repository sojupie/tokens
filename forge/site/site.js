// Keep a viewport-height tool workspace below the measured navigation height.
{
 const header=document.getElementById('tokens-site-header');
 if(header)new ResizeObserver(([entry])=>document.documentElement.style.setProperty('--site-header-height',`${entry.borderBoxSize[0].blockSize}px`)).observe(header);
}
