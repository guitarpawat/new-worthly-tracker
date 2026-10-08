const test = require("node:test");
const assert = require("node:assert/strict");
const editor = require("./js/custom_allocations.js");
const controls = require("./js/custom_allocation_controls.js");
const charts = require("./js/progress_chart_logic.js");
const home = require("./js/home.js");
const progress = require("./js/progress_goals.js");
const { state } = require("./js/shared.js");

test("percentage validation keeps each asset and chart independent", () => {
  const row = (AssetID, Category, Percentage) => ({ AssetID, Category, Percentage });
  const valid = { Name: "Country", Entries: [row(1, "USA", "33.33"), row(1, "China", "66.67"), row(2, "China", "100")] };
  assert.equal(editor.validateDraft(valid), "");
  assert.equal(editor.assetTotal(valid.Entries, 1), 10000);
  for (const [name, entries] of [
    ["over 100", [row(1, "USA", "33.34"), row(1, "China", "66.67")]],
    ["duplicate", [row(1, "USA", "25"), row(1, " usa ", "25")]],
    ["reserved", [row(1, "Unallocated", "25")]],
    ["missing category", [row(1, " ", "25")]],
    ["precision", [row(1, "USA", "99.999")]],
    ["empty", [row(1, "USA", "")]],
    ["nonfinite", [row(1, "USA", "Infinity")]],
    ["zero", [row(1, "USA", "0")]],
    ["negative", [row(1, "USA", "-1")]],
  ]) assert.notEqual(editor.validateDraft({ Name: "Country", Entries: entries }), "", name);
  assert.equal(editor.validateDraft({ Name: "Empty chart", Entries: [] }), "");
  assert.notEqual(editor.validateDraft({ Name: " ", Entries: [] }), "");
});

test("draft edits preserve the saved chart until Save", () => {
  const chart = { ID: 1, Name: "Country", Entries: [{ AssetID: 1, Category: "USA", Percentage: "50" }] };
  const draft = editor.createDraft(chart);
  draft.Name = "Changed";
  assert.equal(draft.Entries, undefined, "chart management must not submit asset assignments");
  assert.equal(chart.Name, "Country");
  assert.equal(chart.Entries[0].Percentage, "50");
});

const breakdowns = [{ ID: 7, Name: "Country <&>", Rows: [{ Name: "USA", Value: "50" }, { Name: "Unallocated", Value: "50" }] }];

test("Custom is a single menu trigger, even before selecting a custom chart", () => {
  assert.equal(controls.render([], "asset", "home"), "");
  const initial = controls.render(breakdowns, "asset", "home");
  assert.match(initial, /custom-control-trigger-label">Custom<\/span>/);
  assert.match(initial, /hidden role="listbox"/);
  assert.match(initial, /home-menu-panel custom-allocation-menu/);
  assert.doesNotMatch(initial, /data-home-allocation-mode=/);
  const selected = controls.render(breakdowns, "custom:7", "home");
  assert.match(selected, /role="listbox"/);
  assert.match(selected, /Country &lt;&amp;&gt;/);
  assert.equal(controls.normalizeMode("custom:7", breakdowns), "custom:7");
  assert.equal(controls.normalizeMode("custom:99", breakdowns), "asset_type");
});

test("Home and Progress show the same saved custom breakdown", (t) => {
  const original = { ...state };
  t.after(() => Object.assign(state, original));
  const date = "2026-01-01";
  const page = { HasSnapshot: true, SnapshotDate: date, Groups: [], CustomAllocations: breakdowns };
  const snapshot = home.buildHomeAllocationSnapshot(page);
  const progressPage = { AllocationSnapshots: [snapshot], TrendPoints: [{ SnapshotDate: date, TotalCurrent: "100" }] };
  assert.deepEqual(charts.buildAllocationRows(progressPage, date, "custom:7"), [{ name: "USA", value: 50 }, { name: "Unallocated", value: 50 }]);
  state.homeAllocationModal = { mode: "custom:7" };
  state.progressAllocationModal = { mode: "custom:7", snapshotDate: date };
  assert.match(home.renderHomeAllocationModal(page), /home-custom-allocation/);
  assert.match(progress.renderAllocationModal(progressPage), /progress-custom-allocation/);
  assert.deepEqual(charts.buildAllocationRows(progressPage, date, "custom:99"), []);
});

test("negative custom values use signed bars; positive values use pie", () => {
  const page = { AllocationSnapshots: [{ SnapshotDate: "2026-01-01", CustomAllocations: breakdowns }] };
  assert.equal(charts.buildAllocationChartConfig(page, "2026-01-01", "custom:7").type, "pie");
  const negativePage = { AllocationSnapshots: [{ SnapshotDate: "2026-01-01", CustomAllocations: [{ ID: 7, Rows: [{ Name: "Debt", Value: "-25" }] }] }] };
  const config = charts.buildAllocationChartConfig(negativePage, "2026-01-01", "custom:7");
  assert.equal(config.type, "bar");
  assert.equal(config.options.indexAxis, "y");
  assert.deepEqual(config.data.datasets[0].data, [-25]);
});

test("custom chart totals count only included assets on Home and Progress", (t) => {
  const original = { ...state };
  t.after(() => Object.assign(state, original));
  const rows = [{ name: "USA", value: 70 }, { name: "Unallocated", value: 30 }];
  assert.equal(progress.resolveAllocationTotalValue({ TotalCurrent: 500 }, rows, "custom:7"), 100);
  assert.equal(progress.resolveAllocationTotalValue({ TotalCurrent: 500 }, rows, "asset"), 500);
  state.homeAllocationModal = { mode: "custom:7" };
  const markup = home.renderHomeAllocationModal({ HasSnapshot: true, SnapshotDate: "2026-01-01", Groups: [], Summary: { TotalCurrent: 500 },
    CustomAllocations: [{ ID: 7, Name: "Country", Rows: rows.map((row) => ({ Name: row.name, Value: row.value })) }],
  });
  assert.match(markup, /THB 100.00/);
  assert.doesNotMatch(markup, /THB 500.00/);
});

test("Manage Chart lists names and Add Chart, with form fields only in its popup", (t) => {
  const original = state.customAllocationEditor;
  t.after(() => { state.customAllocationEditor = original; });
  const markup = editor.render({ CustomAllocationCharts: [{ ID: 7, Name: "Country" }] });
  assert.match(markup, /Add Chart/);
  assert.match(markup, /data-chart-row="7"/);
  assert.doesNotMatch(markup, /custom-allocation-name|<form/);
  state.customAllocationEditor = { draft: { ID: 7, Name: '<img src=x onerror="bad()">' }, error: "Chart already exists" };
  const dialog = editor.renderDialog();
  assert.match(dialog, /Edit Chart/);
  assert.match(dialog, /asset-management-form/);
  assert.match(dialog, /error-copy error-banner/);
  assert.match(dialog, /button chart-delete-button/);
  assert.doesNotMatch(dialog, /Allocation \(.*%|data-custom-add|<img/);
});

function editorHarness(t, savedChart) {
  const originals = { document: global.document, go: global.go };
  const originalState = { ...state };
  t.after(() => { Object.assign(global, originals); Object.assign(state, originalState); });
  const elements = new Map();
  const element = (id) => {
    if (!elements.has(id)) elements.set(id, {
      listeners: {}, dataset: {}, value: "", textContent: "", focus() { this.focused = true; },
      addEventListener(name, handler) { this.listeners[name] = handler; },
    });
    return elements.get(id);
  };
  const page = { Assets: [], CustomAllocationCharts: savedChart ? [savedChart] : [] };
  const rows = () => page.CustomAllocationCharts.map((chart) => {
    const row = element(`row-${chart.ID}`); row.dataset.chartRow = String(chart.ID); return row;
  });
  global.document = {
    getElementById: element,
    querySelectorAll: (selector) => selector === "[data-chart-row]" ? rows() : [],
    querySelector: (selector) => rows().find((row) => selector.includes(`"${row.dataset.chartRow}"`)),
  };
  state.assetManagementPage = page;
  state.customAllocationEditor = null;
  state.assetManagementModal = null;
  const refresh = () => { editor.render(page); editor.bind(page, refresh, async () => refresh()); };
  refresh();
  const emit = (id, event, extra = {}) => element(id).listeners[event]({ target: element(id), preventDefault() {}, ...extra });
  return { element, emit, page, refresh };
}

test("Add opens a popup, rows open edit with keyboard, and closing discards the name draft", (t) => {
  const harness = editorHarness(t, { ID: 7, Name: "Country", Entries: [] });
  assert.equal(state.customAllocationEditor, null);
  harness.emit("chart-add", "click");
  assert.equal(state.customAllocationEditor.draft.ID, 0);
  assert.equal(state.assetManagementModal.kind, "chart");
  assert.equal(harness.element("custom-allocation-name").focused, true);
  harness.emit("chart-dialog-close", "click");
  assert.equal(state.assetManagementModal, null);
  harness.emit("row-7", "keydown", { key: "Enter" });
  harness.element("custom-allocation-name").value = "Modified";
  harness.emit("custom-allocation-name", "input");
  assert.equal(state.customAllocationEditor.draft.Name, "Modified");
  harness.emit("chart-dialog-close", "click");
  assert.equal(state.customAllocationEditor, null);
  assert.equal(harness.element("row-7").focused, true);
  harness.emit("row-7", "click");
  assert.equal(state.customAllocationEditor.draft.Name, "Country");
});

test("form submission saves and closes the popup; failed saves keep it and its draft", async (t) => {
  const harness = editorHarness(t);
  let payload;
  global.go = { app: { App: { GetHomePage() {}, async SaveCustomAllocationChart(chart) {
    payload = structuredClone(chart);
    harness.page.CustomAllocationCharts = [{ ...chart, ID: 8 }];
    return 8;
  } } } };
  harness.emit("chart-add", "click");
  harness.element("custom-allocation-name").value = "Country";
  harness.emit("custom-allocation-name", "input");
  await harness.emit("custom-allocation-form", "submit");
  assert.equal(payload.Name, "Country");
  assert.equal(state.customAllocationEditor, null);
  assert.equal(state.assetManagementModal, null);
  assert.equal(state.chartManagementNotice, "Chart saved.");
  harness.emit("row-8", "click");
  global.go.app.App.SaveCustomAllocationChart = async () => { throw new Error("write failed"); };
  harness.element("custom-allocation-name").value = "Keep this draft";
  harness.emit("custom-allocation-name", "input");
  await harness.emit("custom-allocation-form", "submit");
  assert.equal(state.customAllocationEditor.draft.Name, "Keep this draft");
  assert.equal(state.customAllocationEditor.error, "write failed");
  assert.equal(state.customAllocationEditor.saving, false);
});

test("delete requires confirmation and Cancel returns to the edit popup", async (t) => {
  const harness = editorHarness(t, { ID: 7, Name: "Country", Entries: [] });
  let deleted = 0;
  global.go = { app: { App: { GetHomePage() {}, async DeleteCustomAllocationChart(id) { deleted = id; } } } };
  harness.emit("row-7", "click");
  harness.emit("custom-allocation-delete", "click");
  assert.equal(state.customAllocationEditor.confirmDelete, true);
  assert.equal(deleted, 0);
  harness.emit("custom-allocation-cancel-delete", "click");
  assert.equal(state.customAllocationEditor.confirmDelete, false);
  harness.emit("custom-allocation-delete", "click");
  await harness.emit("custom-allocation-confirm-delete", "click");
  assert.equal(deleted, 7);
  assert.equal(state.assetManagementModal, null);
});

test("custom chart selection restores keyboard focus to the dropdown", (t) => {
  const original = global.document;
  t.after(() => { global.document = original; });
  let change;
  let selected;
  let focused = false;
  global.document = { getElementById: (id) => id.endsWith("-trigger")
    ? { focus() { focused = true; } }
    : { addEventListener(event, handler) { assert.equal(event, "change"); change = handler; } } };
  controls.bind("home", (value) => { selected = value; });
  change({ target: { value: "custom:7" } });
  assert.equal(selected, "custom:7");
  assert.equal(focused, true);
});

test("chart menu keyboard selection preserves Custom label while normal selects use the selected name", (t) => {
  const sharedControls = require("./js/controls.js");
  const originals = { document: global.document, Element: global.Element };
  t.after(() => Object.assign(global, originals));
  class Option {
    dataset = { controlSelectOption: "chart", value: "custom:7", label: "Country" };
    classList = { toggle() {} };
    closest(selector) { return selector === "[data-control-select-option]" ? this : null; }
  }
  global.Element = Option;
  const option = new Option();
  const label = { textContent: "" };
  const changes = [];
  const input = { value: "", dispatchEvent(event) { changes.push(event.type); } };
  let focused = false;
  const trigger = { querySelector: () => label, focus() { focused = true; } };
  const control = { dataset: { controlKind: "select", controlFixedLabel: "Custom" }, querySelector: () => option };
  global.document = {
    querySelector: () => control,
    querySelectorAll: () => [option],
    getElementById: (id) => id === "chart" ? input : trigger,
  };
  for (const [fixedLabel, expected] of [["Custom", "Custom"], ["", "Country"]]) {
    control.dataset.controlFixedLabel = fixedLabel;
    sharedControls.setControlValue("chart", "custom:7");
    assert.equal(label.textContent, expected);
    assert.equal(sharedControls.handleControlKeydown({ target: option, key: "Enter", preventDefault() {} }), true);
    assert.equal(label.textContent, expected);
    assert.equal(input.value, "custom:7");
  }
  assert.deepEqual(changes, ["input", "change", "input", "change"]);
  assert.equal(focused, true);
});
