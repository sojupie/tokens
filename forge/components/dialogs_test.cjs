const test=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');const out={exports:{}};runInNewContext(readFileSync(`${__dirname}/static/dialogs.js`,'utf8'),{module:out});const {scrollEdges}=out.exports;
test('dialog elevation follows content hidden at each scroll edge',()=>{assert.deepEqual({...scrollEdges(0,100,300)},{start:false,end:true});assert.deepEqual({...scrollEdges(50,100,300)},{start:true,end:true});assert.deepEqual({...scrollEdges(200,100,300)},{start:true,end:false});assert.deepEqual({...scrollEdges(0,100,100)},{start:false,end:false});assert.equal(scrollEdges(199.5,100,300).end,false);});
test('discard confirmation is needed only for unsaved edits',()=>{
 const state={committed:'Sample',input:{value:'Sample'}};
 assert.equal(out.exports.isDirty(state),false);state.input.value='Changed';assert.equal(out.exports.isDirty(state),true);state.committed='Changed';assert.equal(out.exports.isDirty(state),false);
});
