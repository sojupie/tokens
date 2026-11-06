import {test} from 'node:test';
import assert from 'node:assert/strict';
import {shortName,shorten} from './short-css-names.ts';

test('short names exhaust each length without collisions, including existing short names',()=>{
  const names=Array.from({length:10000},(_,i)=>shortName(i));
  assert.equal(new Set(names).size,names.length);
  assert.equal(names[0].length,3);assert.equal(names[63].length,3);
  assert.equal(names[64].length,4);assert.equal(names[4159].length,4);
  assert.equal(names[4160].length,5);
  const corpus={pages:[{styles:[{text:':root{--a:red;--long-name:var(--a)}x{color:var(--long-name);content:"--long-name"}'}],inline:['color:var(--long-name)']}]};
  const result=shorten(corpus);
  assert.equal(result.mapping['--long-name'],'--a');
  assert.equal(result.mapping['--a'],'--b');
  assert.match(result.pages[0].styles[0].text,/--b:red;--a:var\(--b\)/);
  assert.match(result.pages[0].styles[0].text,/content:"--long-name"/);
  assert.equal(result.pages[0].inline[0],'color:var(--a)');
});

test('mapping is stable across file order and rewrites at-rules and references',()=>{
  const a={styles:[{text:'@property --xx{syntax:"<color>";inherits:true;initial-value:red}'}],inline:[]};
  const b={styles:[{text:'@supports (--xx:red){x{color:var(--xx)}}'}],inline:[]};
  assert.deepEqual(shorten({pages:[a,b]}).mapping,shorten({pages:[b,a]}).mapping);
  assert.match(shorten({pages:[a,b]}).pages[0].styles[0].text,/@property --a/);
  assert.throws(()=>shorten({pages:[{styles:[{text:':root{--é:red}'}],inline:[]}]}));
});

test('other dashed identifiers and quoted text are preserved',()=>{
  const text='x{--long:red;color:var(/* comment */ --long);anchor-name:--anchor;content:"var(--long)";animation-name:--animation}@supports(anchor-name:--anchor){x{color:var(--long)}}';
  const result=shorten({pages:[{styles:[{text}],inline:[]}]});
  assert.deepEqual(result.mapping,{'--long':'--a'});
  assert.match(result.pages[0].styles[0].text,/anchor-name:--anchor/);
  assert.match(result.pages[0].styles[0].text,/animation-name:--animation/);
  assert.match(result.pages[0].styles[0].text,/content:"var\(--long\)"/);
});

test('style queries use the same property names as declarations and var references',()=>{
  const text='x{--_state:blue}@container style(/* query */ --_state){x{color:var(--_state)}}';
  const result=shorten({pages:[{styles:[{text}],inline:[]}]});
  const name=result.mapping['--_state'];
  assert(result.pages[0].styles[0].text.includes(`style(/* query */ ${name})`));
  assert(result.pages[0].styles[0].text.includes(`${name}:blue`));
  assert(result.pages[0].styles[0].text.includes(`var(${name})`));
});
