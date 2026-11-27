// Captured-source checks remain optional in the quick synthetic-fixture lane.
// The explicit captured lane must never silently skip its publication checks.
export const capturedAssets = process.env.TOKEN_ASSETS_DIR;
if (process.env.TOKEN_ASSETS_REQUIRED === '1' && !capturedAssets) {
  throw new Error('Captured-data verification requires TOKEN_ASSETS_DIR.');
}
