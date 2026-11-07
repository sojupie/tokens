const list=document.querySelector('.md-list');
for(const label of ['First item','Second item','Third item']){
 const row=document.createElement('li');row.dataset.listRow='';row.dataset.listLabel=label;
 row.innerHTML=`<div class="md-list-row"><button class="md-list-item"><span class="md-list-body"><span class="md-list-label">${label}</span></span></button><button class="md-icon-button md-list-grip" data-list-grip aria-pressed="false" aria-label="Move ${label}">=</button></div>`;list.append(row);
}
