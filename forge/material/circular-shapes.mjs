// A circular source shape is a semantic value, not an animation coordinate.
// Resolve it against captured Material geometry only for the rendered app theme.
// DTCG and ordinary CSS exports keep the original shape representation.
export function circularShapeRadius(name, values) {
  if (!name.startsWith('md-comp-')) return 'calc(infinity * 1px)';
  name = name.replace(/-ignore$/, '');
  const exceptions = {
    'md-comp-list-expand-trailing-icon-shape': ['md-comp-icon-button-small-container-height'],
    'md-comp-list-reveal-list-item-icon-button-container-shape': ['md-comp-icon-button-small-container-height'],
    'md-comp-nav-bar-item-active-indicator-shape': ['md-comp-nav-bar-item-horizontal-active-indicator-height'],
    'md-comp-nav-rail-item-active-indicator-shape': ['md-comp-nav-rail-item-horizontal-active-indicator-height'],
    'md-comp-outlined-segmented-button-shape': ['md-comp-outlined-segmented-button-container-height'],
    'md-comp-search-view-contained-docked-bar-shape': ['md-comp-search-view-docked-header-container-height'],
    'md-comp-slider-active-track-outer-corner-corner-size': ['md-comp-slider-active-track-height'],
    'md-comp-time-picker-clock-dial-shape': ['md-comp-time-picker-clock-dial-container-size'],
    'md-comp-toolbar-standard-container-shape': ['md-comp-toolbar-floating-container-height'],
    'md-comp-toolbar-vibrant-container-shape': ['md-comp-toolbar-floating-container-height'],
    'md-comp-menus-horizontal-container-shape': ['md-comp-menus-menu-item-height'],
    'md-comp-menus-horizontal-icon-only-menu-item-selected-shape': ['md-comp-menus-menu-item-height'],
    'md-comp-menus-horizontal-menu-item-selected-focused-shape': ['md-comp-menus-menu-item-height'],
    'md-comp-menus-horizontal-menu-item-selected-hovered-shape': ['md-comp-menus-menu-item-height'],
    'md-comp-menus-horizontal-menu-item-selected-pressed-shape': ['md-comp-menus-menu-item-height'],
  };
  // Selected button shapes use the same container as their unselected state.
  const stem = name.replace(/-(?:shape(?:-round|-square)?|corner-size)$/, '')
    .replace(/-selected-container$/, '-container');
  const dimensions = exceptions[name] ?? ['height', 'width', 'size', 'thickness']
    .map(suffix => `${stem}-${suffix}`).filter(key => key in values);
  if (!dimensions.length || dimensions.some(key => !(key in values)))
    throw new Error(`Missing finite circular geometry for ${name}`);
  const lengths = dimensions.map(key => /^(\d+(?:\.\d+)?)(px|rem)$/.exec(String(values[key])));
  if (lengths.some(length => !length) || lengths.some(length => length[2] !== lengths[0][2]))
    throw new Error(`Unsupported circular dimensions for ${name}`);
  return `${Math.min(...lengths.map(length => Number(length[1]))) / 2}${lengths[0][2]}`;
}
