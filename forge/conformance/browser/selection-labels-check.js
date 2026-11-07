const button=document.querySelector('#run'),out=document.querySelector('#result'),samples=document.querySelector('#samples');
button.onclick=async()=>{button.disabled=true;const checks=[];const assert=(ok,message)=>{if(!ok)throw Error(message);checks.push(message)};
try{for(const slug of ['checkbox','switch']){samples.replaceChildren();const frame=document.createElement('iframe');samples.append(frame);await new Promise(resolve=>{frame.onload=resolve;frame.src=`/md3/components/${slug}?preview.compare=interactive`});const w=frame.contentWindow,d=frame.contentDocument;await w.TokenAssets.ready;
const input=d.querySelector(`#component-preview .md-${slug}>input`),label=d.getElementById(input.getAttribute('aria-labelledby'));
assert(label&&label.control===input,`${slug}: accessible name refers to its visible label`);assert(label.getBoundingClientRect().width>0,`${slug}: label is visible`);const initial=input.checked;label.click();assert(input.checked!==initial,`${slug}: clicking the text toggles the input`);
}out.textContent=JSON.stringify({passed:checks.length,checks},null,2);
}catch(error){out.textContent=JSON.stringify({passed:checks.length,checks,error:error.stack},null,2)}finally{button.disabled=false}};
