const test=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');
const out={exports:{}};runInNewContext(readFileSync(`${__dirname}/static/search.js`,'utf8'),{module:out});
test('search examples match case-insensitively and preserve an empty query',()=>{assert.equal(out.exports.matches('Checkboxes',' CHECK '),true);assert.equal(out.exports.matches('Cards','button'),false);assert.equal(out.exports.matches('Cards',''),true);});

test('exact matching is independent from substring suggestions',()=>{
 const {matches}=out.exports;
 assert.equal(matches('Cards','c',true),false);assert.equal(matches('Cards','CARDS',true),true);assert.equal(matches('Cards','',true),true);
});
test('adaptive search stays docked at every width; explicit layouts stay explicit',()=>{
 const {searchLayout}=out.exports;
 assert.equal(searchLayout('adaptive',599),'docked');assert.equal(searchLayout('adaptive',600),'docked');
 assert.equal(searchLayout('docked',320),'docked');assert.equal(searchLayout('full-screen',900),'full-screen');
});
