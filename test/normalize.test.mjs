import test from "node:test";
import assert from "node:assert/strict";
import { normalizeSnapshot } from "../tools/m3-site-harvester/normalize.mjs";

test("normalizes canonical md token names and preserves evidence", () => {
  const result = normalizeSnapshot({
    source: { kind: "fixture", url: "https://example.test" },
    tokenSets: [{
      name: "Button - Color - Filled",
      lists: [{
        sectionIndex: 0,
        tokens: [{
          displayName: "Button filled container color",
          tokenName: "md.comp.button.filled.container.color",
          valueKind: "color",
          value: "#6750A4",
          previewClass: null,
        }],
      }],
    }],
  });

  assert.deepEqual(result.tokens[0], {
    name: "comp.button.filled.container.color",
    displayName: "Button filled container color",
    tokenName: "md.comp.button.filled.container.color",
    valueKind: "color",
    value: "#6750A4",
    previewClass: null,
    sourceSet: "Button - Color - Filled",
    sourceSection: 0,
  });
});

test("falls back to a stable name when the viewer omits a canonical name", () => {
  const result = normalizeSnapshot({
    source: {},
    tokenSets: [{
      name: "Button - Size - Medium",
      lists: [{
        sectionIndex: 2,
        tokens: [{ displayName: "Button medium shape round", valueKind: "shape", value: null }],
      }],
    }],
  });

  assert.equal(result.tokens[0].name, "button.medium.shape.round");
});

test("normalizes token-table graph records without collapsing context variants", () => {
  const result = normalizeSnapshot({
    schemaVersion: 2,
    source: { kind: "m3-token-table", tokenTableUrl: "https://example.test/TOKEN_TABLE.json" },
    tokenSets: [{
      name: "md.comp.button.filled",
      displayName: "Button - Color - Filled",
      tokens: [{
        name: "md.comp.button.filled.container.color",
        displayName: "Button filled container color",
        valueType: "COLOR",
        directValues: [],
        contextualValues: [
          { contextTags: ["dark"], references: [{ tokenName: "md.sys.color.primary" }] },
          { contextTags: ["light"], references: [{ tokenName: "md.sys.color.primary" }] },
        ],
        sourceId: "token-id",
      }],
    }],
  });

  assert.equal(result.schemaVersion, 2);
  assert.equal(result.tokens[0].name, "comp.button.filled.container.color");
  assert.equal(result.tokens[0].contextualValues.length, 2);
  assert.deepEqual(result.tokens[0].contextualValues.map((context) => context.contextTags), [["dark"], ["light"]]);
});
