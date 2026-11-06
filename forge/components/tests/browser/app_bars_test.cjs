const test = require('node:test'), assert = require('node:assert/strict');
const {readFileSync} = require('node:fs'), {runInNewContext} = require('node:vm');
test('an application scroll container updates its app bar without a playground host', () => {
 const states=[],bar={dataset:{appbarScrollTarget:'messages'},toggleAttribute:(name,on)=>states.push([name,on])};
 let scroll;const document={querySelectorAll:()=>[bar],addEventListener:(name,fn)=>scroll=fn};
 runInNewContext(readFileSync(`${__dirname}/../../behaviors/app-bars.js`,'utf8'),{document});
 scroll({target:{id:'messages',scrollTop:45}});scroll({target:{id:'other',scrollTop:12}});scroll({target:{id:'messages',scrollTop:0}});
 assert.deepEqual(states,[['data-scrolled',true],['data-scrolled',false]]);
});
