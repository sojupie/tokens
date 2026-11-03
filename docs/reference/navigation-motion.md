# Navigation motion verification

The site rail's section drawer follows the supplied M3 website recording
`20261103-1101-01.0954632.mp4` (898 × 1592, 30 fps).
SHA-256: `ab38456936c8c623422f31fbaf74da2b0e1578628c1766de82f1bd1f603707a6`.
These are site renderer choices based on that example, not additional captured
Material component tokens.

## Measurements

Material Motion Lab's native CLI imported a temporary copy, extracted individual
frames and tracked the drawer's exposed edge with `edge-x`. The recording and
original project data were preserved. The analysis project was
`9420bb7a-4b16-4718-9ef3-9be88e05f2f9` under `/tmp/nav-reference-analysis`.

| Observation | Evidence |
| --- | --- |
| The attached edge stays fixed | Surface expands from the rail boundary; the rail and window edges do not translate. |
| Exposed edge approaches its resting position without meaningful overshoot | Frames 17–25: x = 139.50, 273.50, 362.48, 412.49, 443.49, 462.50, 475.40, 482.40, 486.40 video px. Settled x ≈ 486.38. |
| Surface reveal is approximately 250ms | Last closed frame at 566.7ms; edge is effectively settled by 833.3ms. A duration fit of the captured standard Bézier gives 237.5ms, RMS error 8.15 video px over 346.88px travel. Frame quantization and event onset limit precision. |
| Text enters after the surface starts | Title ink is absent at frame 21 (700ms), faint at frame 22 (733.3ms), and effectively opaque at frame 27 (900ms). |
| Section changes are sequential, rather than overlapping text | Old title is opaque at frame 63, nearly absent at 65; frames 66–68 are blank. Incoming title is visible at 69 and effectively opaque at 75. The incoming title stays in place. |

The site uses a 250ms width transition with the captured standard easing,
approximately 75ms text exit, a 150ms incoming-text delay and a 200ms linear text
entry. Its content retains a fixed width while the surface expands, preventing
line reflow. CSS transitions can reverse without restarting from the endpoints.
Reduced motion removes these transitions.

## Attached component surfaces

An underdamped spring applied to the whole surface's translation can travel
past its resting position. This exposes a gap behind a bottom sheet or beside a
modal drawer. It is a different constraint from a spring animating sheet height:
height changes retain the fixed bottom inset.

`MaterialMotion.attachedTiming` keeps the existing generated spatial duration
and uses the captured standard easing for modal-drawer and sheet entry/exit
translations. Sheet height snapping keeps its spatial spring and release
velocity. The drawer recording does not establish a sheet-specific duration.
