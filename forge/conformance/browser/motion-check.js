import {checkMotion} from './check-motion.mjs';
const button=document.querySelector('#run'),result=document.querySelector('#result');
button.onclick=async()=>{button.disabled=true;result.textContent='Running';try{result.textContent=JSON.stringify(await checkMotion(document.querySelector('#frame')),null,2);}catch(error){result.textContent=error.stack;}button.disabled=false;};
