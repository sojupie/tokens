import {checkCarousel} from './check-carousel.mjs';
run.onclick=async()=>{run.disabled=true;result.textContent='Running';try{result.textContent=JSON.stringify(await checkCarousel(document.querySelector('#frames')),null,2);}catch(e){result.textContent=e.stack;}run.disabled=false;};
