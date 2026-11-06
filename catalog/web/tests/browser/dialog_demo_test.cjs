const test=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');
const out={exports:{}};runInNewContext(readFileSync(`${__dirname}/../../assets/scripts/dialog-demo.js`,'utf8'),{module:out});
test('discard confirmation is needed only for unsaved edits',()=>{
 const state={committed:'Sample',input:{value:'Sample'}};
 assert.equal(out.exports.isDirty(state),false);state.input.value='Changed';assert.equal(out.exports.isDirty(state),true);state.committed='Changed';assert.equal(out.exports.isDirty(state),false);
});
