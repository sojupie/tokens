const { test } = require('node:test');
const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { runInNewContext } = require('node:vm');

// Exercise the actual destination key handler; browser checks cover DOM visibility
// and whether focus really moves on the rendered navigation components.
function navigation(options = {}) {
  const listeners = [];
  let active;
  const group = {
    classList: { contains: name => name === 'md-navigation-bar' && !!options.horizontal },
    querySelectorAll: () => items,
    direction: options.rtl ? 'rtl' : 'ltr',
  };
  const items = (options.items || [{}, {}, {}, {}, { hidden: true }, { hidden: true }]).map(flags => ({
    ...flags,
    closest(selector) {
      if (selector === '[data-navigation-item]') return this;
      if (selector === '[data-navigation-group]') return group;
      if (selector === '[inert]') return this.inert ? {} : null;
      return null;
    },
    getClientRects() { return this.hidden ? [] : [{}]; },
    matches() { return !!(this.disabled || this.ariaDisabled); },
    focus() { active = this; },
  }));
  const globals = {
    document: { readyState: 'loading', addEventListener(type, listener) {
      if (type === 'keydown') listeners.push(listener);
    } },
    ResizeObserver: class { observe() {} },
    getComputedStyle: element => ({ direction: element.direction || 'ltr', visibility: element.visibility || 'visible' }),
  };
  runInNewContext(readFileSync(`${__dirname}/behaviors/utils.js`, 'utf8'), globals);
  runInNewContext(readFileSync(`${__dirname}/behaviors/navigation.js`, 'utf8'), globals);
  return {
    key(index, key) {
      active = items[index];
      let prevented = false;
      listeners[0]({ target: active, key, preventDefault() { prevented = true; } });
      return { index: items.indexOf(active), prevented };
    },
  };
}

test('collapsed rail wraps visible destinations and Home/End exclude hidden items', () => {
  const nav = navigation();
  assert.equal(nav.key(3, 'ArrowDown').index, 0);
  assert.equal(nav.key(0, 'ArrowUp').index, 3);
  assert.equal(nav.key(0, 'End').index, 3);
  assert.equal(nav.key(3, 'Home').index, 0);
});

test('expanded rail includes newly visible secondary destinations', () => {
  const nav = navigation({ items: Array.from({ length: 6 }, () => ({})) });
  assert.equal(nav.key(3, 'ArrowDown').index, 4);
  assert.equal(nav.key(0, 'End').index, 5);
});

test('destination traversal skips disabled, inert and visibility-hidden controls', () => {
  const nav = navigation({ items: [{}, { disabled: true }, { ariaDisabled: true }, { inert: true }, { visibility: 'hidden' }, {}] });
  assert.equal(nav.key(0, 'ArrowDown').index, 5);
  assert.equal(nav.key(5, 'ArrowUp').index, 0);
});

test('horizontal RTL navigation retains direction and skips unavailable destinations', () => {
  const nav = navigation({ horizontal: true, rtl: true, items: [{}, { hidden: true }, {}, {}] });
  assert.equal(nav.key(0, 'ArrowLeft').index, 2);
  assert.equal(nav.key(0, 'ArrowRight').index, 3);
  assert.deepEqual(nav.key(0, 'ArrowDown'), { index: 0, prevented: false });
});

test('unavailable event targets do not consume keys or move focus', () => {
  const nav = navigation({ items: [{ disabled: true }] });
  assert.deepEqual(nav.key(0, 'Home'), { index: 0, prevented: false });
});

function drawerMotion({rtl=false,reduced=false}={}) {
  const pending=[];
  const globals={module:{exports:{}},matchMedia:()=>({matches:reduced}),
    getComputedStyle:()=>({direction:rtl?'rtl':'ltr',transform:'matrix(1, 0, 0, 1, -80, 0)'}),
    MaterialMotion:{timing:()=>({duration:400,easing:'linear'})},
  };
  runInNewContext(readFileSync(`${__dirname}/behaviors/navigation.js`,'utf8'),globals);
  const drawer={open:false,dataset:{},listeners:[],closes:0,
    addEventListener(type,fn){if(type==='close')this.listeners.push(fn);},
    showModal(){this.open=true;},
    close(){this.open=false;this.closes++;this.listeners.splice(0).forEach(fn=>fn());},
    animate(frames,timing){
      let finish,reject;
      const finished=new Promise((resolve,no)=>{finish=resolve;reject=no;});
      // Native animations only expose a rejected finished promise to a consumer.
      finished.catch(()=>{});
      const animation={frames,timing,finished,cancel(){reject(Error('cancelled'));}};
      pending.push({animation,finish});return animation;
    },
  };
  return {drawer,pending,...globals.module.exports};
}
test('modal drawer slides from the logical start edge and stays open until exit completes',async()=>{
  for(const rtl of [false,true]) {
    const {drawer,pending,openDrawer,closeDrawer}=drawerMotion({rtl});
    openDrawer(drawer);
    assert.equal(pending[0].animation.frames[0].transform,rtl?'translateX(100%)':'translateX(-100%)');
    const exit=closeDrawer(drawer);
    assert.equal(drawer.open,true);
    assert.equal(drawer.dataset.drawerClosing,'');
    assert.equal(pending[1].animation.frames[0].transform,'matrix(1, 0, 0, 1, -80, 0)');
    await closeDrawer(drawer);
    assert.equal(pending.length,2,'repeated dismiss does not restart exit');
    pending[1].finish();await exit;
    assert.equal(drawer.open,false);
    assert.equal(Object.hasOwn(drawer.dataset,'drawerClosing'),false);
  }
});
test('reopening cancels a pending drawer exit; reduced motion closes immediately',async()=>{
  const {drawer,pending,openDrawer,closeDrawer}=drawerMotion();
  openDrawer(drawer);const exit=closeDrawer(drawer);openDrawer(drawer);
  await exit;
  assert.equal(drawer.open,true);
  assert.equal(drawer.closes,0);
  const reduced=drawerMotion({reduced:true});
  reduced.openDrawer(reduced.drawer);await reduced.closeDrawer(reduced.drawer);
  assert.equal(reduced.drawer.open,false);
  assert.equal(reduced.pending.length,0);
});
