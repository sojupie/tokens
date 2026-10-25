const { test } = require("node:test"),
  assert = require("node:assert/strict"),
  { readFileSync } = require("node:fs"),
  { runInNewContext } = require("node:vm");
const moduleOutput = { exports: {} };
runInNewContext(readFileSync(__dirname + "/recorder.js", "utf8"), {
  module: moduleOutput,
  performance: { timeOrigin: 1000 },
});
const { MotionRecorder, bezier, series, summarize } = moduleOutput.exports;
test("recorder shares one clock for input and frame samples and stops sampling", () => {
  let now = 10,
    callback,
    cancelled;
  const r = new MotionRecorder({
    readFrame: () => ({ x: 3 }),
    now: () => now,
    requestFrame: (f) => ((callback = f), 1),
    cancelFrame: (id) => (cancelled = id),
  });
  r.start();
  now = 30;
  r.event("release");
  callback(35);
  now = 45;
  const data = r.stop();
  assert.equal(data.epoch, 1010);
  assert.equal(data.events[1].t, 20);
  assert.equal(data.frames[0].t, 25);
  assert.equal(data.duration, 35);
  assert.equal(cancelled, 1);
  assert.equal(r.stop(), data);
});
test("signed progress retains overshoot and reversal in either travel direction", () => {
  for (const direction of [1, -1]) {
    const s = summarize(
      [
        { t: 0, value: 0 },
        { t: 50, value: 120 * direction },
        { t: 100, value: 100 * direction },
      ],
      { from: 0, to: 100 * direction },
    );
    assert.ok(Math.abs(s.overshoot - 0.2) < 1e-10);
    assert.equal(s.samples[1].progress, 1.2);
    assert.equal(s.rmsProgressError, null);
  }
});
test("reference comparison uses supplied duration and curve, never a fixed MD3 default", () => {
  const curve = [0.2, 0, 0, 1],
    samples = Array.from({ length: 11 }, (_, i) => ({
      t: 200 + i * 50,
      value: 10 + 80 * bezier(i / 10, curve),
    })),
    s = summarize(samples, {
      start: 200,
      end: 700,
      from: 10,
      to: 90,
      easing: curve,
    });
  assert.equal(s.duration, 500);
  assert.ok(s.rmsProgressError < 1e-12);
  assert.ok(s.maxProgressError < 1e-12);
  assert.throws(() => summarize(samples, { from: 1, to: 1 }));
});
test("track selection uses finite values and explicit time alignment", () => {
  const s = series(
    {
      frames: [
        { t: 10, values: { label: 2 } },
        { t: 20, values: { label: null } },
        { t: 30, values: { label: 5 } },
      ],
    },
    "label",
    10,
  );
  assert.equal(s.length, 2);
  assert.equal(s[0].t, 0);
  assert.equal(s[1].value, 5);
  assert.notEqual(s[0].segment, s[1].segment);
});
test("recorder retains sampling failure rather than silently claiming complete frames", () => {
  let callback;
  const r = new MotionRecorder({
    readFrame: () => {
      throw Error("missing measurement");
    },
    now: () => 10,
    requestFrame: (f) => ((callback = f), 1),
    cancelFrame: () => {},
  });
  r.start();
  callback(20);
  assert.equal(r.result.samplingError, "missing measurement");
  assert.equal(r.active, null);
});
