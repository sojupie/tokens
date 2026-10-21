import postcss from "postcss";
import { compactCSS } from "./css-output.mjs";

// Native HTML compositions require their child families as well as their shell.
export const bundleGroups = {
  common: [],
  navigation: ["navigation"],
  chips: ["chip"],
  radio: ["radio"],
  checkbox: ["checkbox"],
  switch: ["switch"],
  tabs: ["tab"],
  segmented: ["segment"],
  sliders: ["slider"],
  progress: ["progress", "slider"],
  loading: ["loading", "progress", "slider"],
  appbar: ["appbar"],
  toolbar: ["toolbar"],
  sheets: ["sheet", "list"],
  split: ["split"],
  "button-group": ["button-group"],
  "fab-menu": ["fab-menu"],
  lists: ["list"],
  carousel: ["carousel"],
  search: ["search", "navigation", "list", "checkbox"],
  date: ["date", "list"],
  time: ["time"],
  dialog: ["checkbox", "radio"],
  snackbar: ["snackbar"],
  tooltip: ["tooltip"],
};

function classGroup(name) {
  for (const [pattern, group] of [
    [/^md-(?:navigation|nav-|rail-)/, "navigation"],
    // Catalog filter pills use the native filter-chip anatomy on every page.
    [/^md-(?:filter-chip|chip)/, "common"],
    [/^md-(?:assist|suggestion|input)-chip/, "chip"],
    [/^md-fab-menu/, "fab-menu"],
    [/^md-button-group|^md-group-/, "button-group"],
    [/^md-(?:bottom-sheet|side-sheet|sheet)/, "sheet"],
    [/^md-segment/, "segment"],
    [/^md-tab/, "tab"],
    [/^md-app-bar/, "appbar"],
    ...[
      "radio",
      "checkbox",
      "switch",
      "slider",
      "progress",
      "loading",
      "toolbar",
      "split",
      "list",
      "carousel",
      "search",
      "date",
      "time",
      "snackbar",
      "tooltip",
    ].map((group) => [new RegExp("^md-" + group), group]),
  ])
    if (pattern.test(name)) return group;
  return "common";
}

export function componentBundles(css) {
  const parsed = postcss.parse(css);
  function filter(container, allowed, common) {
    const result = container.clone({ nodes: [] });
    for (const node of container.nodes ?? []) {
      if (node.type === "rule") {
        const groups = new Set(
          [...node.selector.matchAll(/\.([a-zA-Z_][\w-]*)/g)].map((m) =>
            classGroup(m[1]),
          ),
        );
        // Keep selector lists intact. Optional families can share a rule, but
        // it must remain in each family's source position (not move ahead of
        // later rules, especially reduced-motion overrides).
        const shared = !groups.size || groups.has("common");
        if (shared ? common : [...groups].some(group => allowed.has(group)))
          result.append(node.clone());
      } else if (node.type === "atrule" && node.name === "property") {
        if (node.params.startsWith("--_time-") ? allowed.has("time") : common)
          result.append(node.clone());
      } else if (node.nodes) {
        const child = filter(node, allowed, common);
        if (child.nodes.length) result.append(child);
      } else if (common) result.append(node.clone());
    }
    return result;
  }
  // Shared primitives precede family styling. Each sheet retains its own rule
  // order; computed-style conformance checks cover interactions across sheets.
  const groups = {
    ...bundleGroups,
    all: [...new Set(Object.values(bundleGroups).flat())],
  };
  return new Map(
    Object.entries(groups).map(([name, groups]) => [
      name,
      compactCSS(filter(parsed, new Set(groups), name === "common").toString()),
    ]),
  );
}
