// Shared source for generated CSS timing and browser spring sampling.
export function springState(damping: number, stiffness: number, time: number, from = 0, to = 1, velocity = 0) {
  if (!(damping > 0) || !(stiffness > 0)) throw new Error('Invalid Material spring');
  const omega = Math.sqrt(stiffness), x = from - to;
  if (damping < 1) {
    const decay = damping * omega, frequency = omega * Math.sqrt(1 - damping * damping);
    const a = x, b = (velocity + decay * x) / frequency;
    const cosine = Math.cos(frequency * time), sine = Math.sin(frequency * time), envelope = Math.exp(-decay * time);
    const displacement = a * cosine + b * sine;
    return { value: to + envelope * displacement, velocity: envelope * ((b * frequency - decay * a) * cosine - (a * frequency + decay * b) * sine) };
  }
  if (damping === 1) {
    const b = velocity + omega * x, envelope = Math.exp(-omega * time);
    return { value: to + envelope * (x + b * time), velocity: envelope * (b - omega * (x + b * time)) };
  }
  const root = Math.sqrt(damping * damping - 1), a = -omega * (damping - root), b = -omega * (damping + root);
  const c = (velocity - b * x) / (a - b), d = x - c;
  return { value: to + c * Math.exp(a * time) + d * Math.exp(b * time), velocity: a * c * Math.exp(a * time) + b * d * Math.exp(b * time) };
}
export function springFrames(damping: number, stiffness: number, velocity = 0) {
  const samples = [0];
  for (let i = 1; i <= 600; i++) {
    samples.push(springState(damping, stiffness, i / 60, 0, 1, velocity).value);
    if (i > 2 && Math.abs(1 - samples[i]) < .001 && Math.abs(samples[i] - samples[i - 1]) < .001) break;
  }
  samples[samples.length - 1] = 1;
  return { samples, duration: (samples.length - 1) * 1000 / 60 };
}
export function springTrack(damping: number, stiffness: number, from: number[], to: number[], velocity = from.map(() => 0)) {
  const samples = [from.slice()], epsilon = from.map((n, i) => Math.max(1, Math.abs(to[i] - n)) * .001);
  for (let i = 1; i <= 600; i++) {
    const next = from.map((n, j) => springState(damping, stiffness, i / 60, n, to[j], velocity[j]).value);
    samples.push(next);
    if (i > 2 && next.every((n, j) => Math.abs(to[j] - n) < epsilon[j] && Math.abs(n - samples[i - 1][j]) < epsilon[j])) break;
  }
  samples[samples.length - 1] = to.slice();
  return { samples, duration: (samples.length - 1) * 1000 / 60 };
}
export function cssTiming(damping: number, stiffness: number) {
  const { samples, duration } = springFrames(damping, stiffness);
  return { duration: `${duration.toFixed(3)}ms`, easing: `linear(${samples.map(n => Number(n.toFixed(6))).join(', ')})` };
}
