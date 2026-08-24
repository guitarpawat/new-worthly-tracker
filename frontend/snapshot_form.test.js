const test = require("node:test");
const assert = require("node:assert/strict");

const { state } = require("./js/shared.js");
const {
  bindPriceInput,
  keepTabbedEditFieldVisible,
} = require("./js/snapshot_form_logic.js");

test("bindPriceInput defers profit and percentage refresh until blur in shared snapshot forms", (t) => {
  const originalDocument = global.document;
  t.after(() => {
    global.document = originalDocument;
    state.editDraft = null;
  });

  const listeners = new Map();
  const profitCell = { className: "numeric neutral", textContent: "THB 0.00" };
  const percentCell = { className: "numeric neutral", textContent: "0.00%" };
  global.document = {
    getElementById(id) {
      return id === "profit-value-1" ? profitCell : percentCell;
    },
  };
  state.editDraft = {
    Mode: "new",
    Groups: [{ Rows: [{ AssetID: 1, IsCash: false, BoughtPrice: 100, CurrentPrice: 100 }] }],
  };

  const input = createFakeInput({
    classNames: ["edit-current-input"],
    dataset: { assetId: "1" },
  });
  input.addEventListener = (name, handler) => {
    listeners.set(name, handler);
  };
  bindPriceInput(input, {
    onChange: (row, value) => {
      row.CurrentPrice = value;
    },
    onDirty: () => {},
  });

  input.value = "150";
  listeners.get("input")({ target: input });
  assert.equal(profitCell.textContent, "THB 0.00");
  assert.equal(percentCell.textContent, "0.00%");

  listeners.get("blur")({ target: input });
  assert.equal(profitCell.textContent, "THB 50.00");
  assert.equal(percentCell.textContent, "50.00%");
});

test("keepTabbedEditFieldVisible scrolls down for a newly focused offscreen input", (t) => {
  const originalDocument = global.document;
  const originalInnerHeight = global.innerHeight;
  const originalRequestAnimationFrame = global.requestAnimationFrame;
  const originalScrollBy = global.scrollBy;
  t.after(() => {
    global.document = originalDocument;
    global.innerHeight = originalInnerHeight;
    global.requestAnimationFrame = originalRequestAnimationFrame;
    global.scrollBy = originalScrollBy;
    state.editDraft = null;
  });

  const scrollCalls = [];
  global.document = {
    activeElement: {
      matches: () => true,
      getBoundingClientRect: () => ({ top: 700, bottom: 740 }),
    },
    documentElement: { clientHeight: 600 },
  };
  global.innerHeight = 600;
  global.requestAnimationFrame = (callback) => callback();
  global.scrollBy = (options) => scrollCalls.push(options);
  state.editDraft = { Mode: "new" };

  keepTabbedEditFieldVisible({ key: "Tab", shiftKey: false });

  assert.deepEqual(scrollCalls, [{ top: 480, behavior: "smooth" }]);
});

function createFakeInput({ classNames = [], dataset = {} } = {}) {
  const classes = new Set(classNames);
  return {
    value: "",
    dataset,
    selectionStart: 0,
    selectionEnd: 0,
    classList: {
      contains: (name) => classes.has(name),
    },
    addEventListener: () => {},
    select() {},
  };
}
