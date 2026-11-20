# Navigation motion verification

The site rail's section drawer timing was measured from a reference recording of
the Material 3 website (`20261103-1101-01.0954632.mp4`, 898 × 1592, 30 fps,
SHA-256 `ab38456936c8c623422f31fbaf74da2b0e1578628c1766de82f1bd1f603707a6`).

## Measurements

Frames were extracted with Material Motion Lab's CLI and tracked along the
drawer's exposed edge (`edge-x`):

| Observation | Evidence |
| --- | --- |
| Attached edge stays fixed | Surface expands from the rail boundary; rail and window edges stay put. |
| Exposed edge settles without overshoot | Frames 17–25: x = 139.50, 273.50, 362.48, 412.49, 443.49, 462.50, 475.40, 482.40, 486.40 px (settled ≈ 486.38 px). |
| Surface reveal ≈ 250 ms | Last closed frame at 566.7 ms; settled by 833.3 ms. Standard Bézier fit: 237.5 ms (RMS error 8.15 px over 346.88 px travel). |
| Text enters after surface opens | Title ink absent at frame 21 (700 ms), faint at frame 22 (733.3 ms), opaque by frame 27 (900 ms). |
| Section changes fade sequentially | Outgoing title opaque at frame 63, gone by 65; frames 66–68 blank; incoming title visible at 69, opaque at 75 without translating. |

From this, the site drawer uses a 250 ms width transition with standard easing,
~75 ms text exit, 150 ms incoming-text delay, and 200 ms linear text entry, with
fixed content width to prevent reflow and transitions disabled under reduced
motion.

## Attached component surfaces

Underdamped springs on full-surface translation can overshoot and expose a gap
behind a bottom sheet or beside a modal drawer (unlike sheet height snapping,
which keeps its bottom edge fixed). `MaterialMotion.attachedTiming` therefore
uses the generated spatial duration with standard easing for modal-drawer and
sheet entry/exit translations, while sheet height snapping keeps its spatial
spring and release velocity.
