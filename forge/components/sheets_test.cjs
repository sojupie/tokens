const {test}=require('node:test');
const assert=require('node:assert/strict');
const {readFileSync}=require('node:fs');
const {runInNewContext}=require('node:vm');
const exported={exports:{}};
runInNewContext(readFileSync(`${__dirname}/static/sheets.js`,'utf8'),{module:exported});
const {sheetTarget,sheetExit}=exported.exports;

test('sheet dragging respects positional thresholds and available anchors',()=>{
 assert.equal(sheetTarget('expanded',360,384,56,0,56,.125),'expanded');
 assert.equal(sheetTarget('expanded',300,384,56,0,56,.125),'minimized');
 assert.equal(sheetTarget('expanded',20,384,56,0,56,.125),'hidden');
 assert.equal(sheetTarget('minimized',100,56,56,0,56,.125),'minimized');
 assert.equal(sheetTarget('minimized',120,56,56,0,56,.125),'expanded');
});
test('sheet flings follow the last direction even after a reverse drag',()=>{
 assert.equal(sheetTarget('expanded',370,384,56,.2,56,.125),'minimized');
 assert.equal(sheetTarget('minimized',55,56,56,.2,56,.125),'hidden');
 assert.equal(sheetTarget('minimized',55,56,56,-.2,56,.125),'expanded');
 assert.equal(sheetTarget('expanded',20,384,56,-.2,56,.125),'expanded');
});
test('side sheet enter and exit direction follows logical edge and text direction',()=>{
 assert.equal(sheetExit(true,false,false),'translateY(100%)');
 assert.equal(sheetExit(false,false,false),'translateX(100%)');
 assert.equal(sheetExit(false,true,false),'translateX(-100%)');
 assert.equal(sheetExit(false,false,true),'translateX(-100%)');
 assert.equal(sheetExit(false,true,true),'translateX(100%)');
});
