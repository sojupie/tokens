const test=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');
const out={exports:{}};runInNewContext(readFileSync(`${__dirname}/static/date-picker.js`,'utf8'),{module:out});const {parse,cells,addMonths,selectRange}=out.exports;
test('date input rejects impossible dates and respects supported years',()=>{assert.ok(parse('2028-02-29'));for(const value of ['2026-02-29','2026-04-31','2026-13-01','26-01-01','1899-12-31','2101-01-01'])assert.equal(parse(value),null,value);});
test('calendar includes complete weeks and month changes clamp leap days',()=>{assert.equal(cells('2026-10').length,35);assert.equal(cells('2026-08').length,42);assert.equal(cells('2026-02').length,28);assert.equal(addMonths('2028-02-29',12),'2029-02-28');assert.equal(addMonths('2026-01-31',1),'2026-02-28');assert.equal(addMonths('1900-01-01',-1),'1900-01-01');});
test('range selection restarts after completion and when the next date precedes the start',()=>{assert.equal(JSON.stringify(selectRange(null,null,'2026-10-21')),JSON.stringify({start:'2026-10-21',end:null}));assert.equal(JSON.stringify(selectRange('2026-10-21',null,'2026-10-26')),JSON.stringify({start:'2026-10-21',end:'2026-10-26'}));assert.equal(selectRange('2026-10-21',null,'2026-10-18').start,'2026-10-18');assert.equal(selectRange('2026-10-21','2026-10-26','2026-10-25').end,null);});
test('entry fields validate typed dates before replacing the committed selection',()=>{
 const helpers={},fields=['start','end'].map(part=>({dataset:{dateValue:part},value:part==='start'?'2028-02-29':'2028-03-02',parentElement:{setAttribute(){}},setAttribute(name,value){this[name]=value;}}));
 const state={fields,range:true,root:{querySelector:selector=>helpers[selector]??={}},month:'2026-10'};
 assert.equal(out.exports.readEntries(state),true);
 assert.equal(state.month,'2028-02');assert.equal(state.committed.start,'2028-02-29');
 fields[0].value='2026-02-29';
 assert.equal(out.exports.readEntries(state),false);assert.equal(fields[0]['aria-invalid'],'true');assert.equal(state.committed.start,'2028-02-29');
 fields[0].value='2028-03-03';assert.equal(out.exports.readEntries(state),false);assert.equal(fields[1]['aria-invalid'],'true');
 fields[0].value='';assert.equal(out.exports.readEntries(state),false);assert.equal(fields[0]['aria-invalid'],'true');
 fields[1].value='';assert.equal(out.exports.readEntries(state),true);assert.equal(state.committed.start,null);
});
test('typed dates accept common separators and normalize after entry',()=>{
 for(const input of ['2028-2-29','2028/02/29','02/29/2028','2.29.2028','2 29 2028'])assert.equal(out.exports.normalizeInput(input),'2028-02-29');
 for(const input of ['2/29/2026','13/1/2026','2/3/26','hello'])assert.equal(out.exports.normalizeInput(input),null);
});
test('range month windows provide continuous months at both year bounds',()=>{
 for(const month of ['1900-01','2026-10','2100-12']){
  const window=out.exports.monthWindow(month);assert.ok(window.includes(month));assert.equal(new Set(window).size,window.length);
  for(let i=1;i<window.length;i++)assert.equal(addMonths(window[i-1]+'-01',1).slice(0,7),window[i]);
 }
});
test('bounded full-screen calendar promotes input to a modal and restores its origin',()=>{
 const attributes=new Set(['data-date-inline']),events=[],panel={dataset:{variant:'full-screen',mode:'calendar'},querySelector:()=>({scrollTop:120}),open:true,hasAttribute:n=>attributes.has(n),removeAttribute:n=>attributes.delete(n),setAttribute:n=>attributes.add(n),close(){events.push('close');this.open=false;},showModal(){events.push('modal');this.open=true;}};
 const state={panel,variant:'full-screen',inline:true};
 out.exports.setMode(state,'input');assert.equal(state.calendarScroll.top,120);assert.equal(panel.dataset.mode,'input');assert.equal(attributes.has('data-date-inline'),false);assert.deepEqual(events,['close','modal']);
 out.exports.setMode(state,'calendar');assert.equal(panel.open,true);assert.equal(attributes.has('data-date-inline'),true);assert.equal(panel.dataset.variant,'full-screen');
 const modalEvents=events.length;state.inline=false;
 out.exports.setMode(state,'input');out.exports.setMode(state,'calendar');assert.equal(events.length,modalEvents,'an already modal picker must retain its native dialog');
});
