const test = require("node:test");
const assert = require("node:assert/strict");
const management = require("./js/asset_management.js");
const { state, loadAssetManagementPage } = require("./app.js");

test("Tab and Shift+Tab stay inside both asset editors, skipping hidden controls", (t) => {
  const originalDocument = global.document;
  const originalModal = state.assetManagementModal;
  t.after(() => { global.document = originalDocument; state.assetManagementModal = originalModal; });
  const fields = Array.from({ length: 3 }, () => ({
    getClientRects: () => [1],
    focus() { global.document.activeElement = this; },
  }));
  const hidden = { getClientRects: () => [], focus() { assert.fail("hidden control focused"); } };
  global.document = {
    activeElement: null,
    querySelector: () => ({ querySelectorAll: () => [...fields, hidden] }),
  };
  for (const kind of ["asset", "asset_type"]) {
    state.assetManagementModal = { kind };
    for (const [from, shiftKey, expected] of [[null, false, 0], [0, false, 1], [2, false, 0], [0, true, 2], [null, true, 2]]) {
      global.document.activeElement = fields[from] || null;
      let prevented = false;
      assert.equal(management.handleAssetManagementKeydown({ key: "Tab", shiftKey, preventDefault() { prevented = true; } }), true);
      assert.equal(prevented, true);
      assert.equal(global.document.activeElement, fields[expected]);
    }
  }
  state.assetManagementModal = null;
  assert.equal(management.handleAssetManagementKeydown({ key: "Tab" }), false);
});

test("reload after asset save retains list and scroll until replacement is ready", async (t) => {
  const originals = { document: global.document, go: global.go, scrollX: global.scrollX, scrollY: global.scrollY, scrollTo: global.scrollTo };
  const originalState = { ...state };
  const originalRender = management.renderAssetManagementPage;
  t.after(() => { Object.assign(global, originals); Object.assign(state, originalState); management.renderAssetManagementPage = originalRender; });
  let replacements = 0;
  global.document = { getElementById: () => ({ set innerHTML(value) { replacements++; } }) };
  global.scrollX = 0;
  global.scrollY = 850;
  let restored;
  global.scrollTo = (position) => { restored = position; };
  global.go = { app: { App: { GetHomePage() {}, GetAssetManagementPage: async () => ({ Assets: [], AssetTypes: [] }) } } };
  management.renderAssetManagementPage = () => { global.scrollY = 0; };
  await loadAssetManagementPage({ view: "edit_asset", preserveScroll: true });
  assert.equal(replacements, 0, "loading screen must not collapse list");
  assert.deepEqual(restored, { left: 0, top: 850, behavior: "instant" });
  assert.equal(state.assetManagementModal, null);
});

test("opening and rerendering an editor focuses its form without moving the list", (t) => {
  const originals = { document: global.document, scrollX: global.scrollX, scrollY: global.scrollY, scrollTo: global.scrollTo };
  const originalState = { ...state };
  t.after(() => { Object.assign(global, originals); Object.assign(state, originalState); });
  const focusCalls = [];
  const firstInput = { focus: (options) => focusCalls.push(["first", options]) };
  const previousInput = { focus: (options) => focusCalls.push(["previous", options]) };
  const dialog = { contains: (element) => element === previousInput, querySelector: () => firstInput };
  let markup;
  const appRoot = { set innerHTML(value) { markup = value; global.scrollY = 0; } };
  global.document = {
    activeElement: {},
    getElementById: (id) => id === "app" ? appRoot : id === "focused-field" ? previousInput : null,
    querySelector: () => dialog,
    querySelectorAll: () => [],
  };
  global.scrollX = 0;
  global.scrollY = 850;
  global.scrollTo = ({ top }) => { global.scrollY = top; };
  state.assetManagementPage = { Assets: [], AssetTypes: [] };
  state.assetManagementView = "edit_asset";
  state.assetManagementModal = { kind: "asset" };
  state.assetForm = management.buildEmptyAssetForm(state.assetManagementPage);
  state.assetTypeForm = management.buildEmptyAssetTypeForm();
  management.renderAssetManagementPage({});
  assert.match(markup, /<main inert/);
  assert.deepEqual(focusCalls[0], ["first", { preventScroll: true }]);
  assert.equal(global.scrollY, 850);
  global.document.activeElement = { id: "focused-field" };
  management.renderAssetManagementPage({});
  assert.deepEqual(focusCalls[1], ["previous", { preventScroll: true }]);
  assert.equal(global.scrollY, 850);
});

test("new asset and asset type pages focus their first field", (t) => {
  const originalDocument = global.document;
  const originalState = { ...state };
  t.after(() => { global.document = originalDocument; Object.assign(state, originalState); });
  let focused;
  const inputs = Object.fromEntries(["asset-name-input", "asset-type-name-input"].map((id) => [id, {
    addEventListener() {}, focus() { focused = id; },
  }]));
  global.document = {
    activeElement: null,
    getElementById: (id) => id === "app" ? { innerHTML: "" } : inputs[id],
    querySelectorAll: () => [],
  };
  state.assetManagementPage = { Assets: [], AssetTypes: [] };
  state.assetManagementModal = null;
  state.assetForm = management.buildEmptyAssetForm(state.assetManagementPage);
  state.assetTypeForm = management.buildEmptyAssetTypeForm();
  for (const [view, field] of [["create_asset", "asset-name-input"], ["create_asset_type", "asset-type-name-input"]]) {
    state.assetManagementView = view;
    management.renderAssetManagementPage({});
    assert.equal(focused, field);
  }
});

test("asset type name Enter saves, excluding IME, repeats, and modified keys", (t) => {
  const originalDocument = global.document;
  t.after(() => { global.document = originalDocument; });
  let clicks = 0;
  let prevented = 0;
  let keydown;
  global.document = { getElementById: () => ({ click() { clicks++; } }) };
  management.bindAssetTypeNameSubmit({ addEventListener(type, handler) { keydown = handler; } });
  keydown({ key: "Enter", preventDefault() { prevented++; } });
  assert.equal(clicks, 1);
  assert.equal(prevented, 1);
  for (const options of [{ key: "Tab" }, { isComposing: true }, { repeat: true }, { ctrlKey: true }, { altKey: true }, { metaKey: true }, { shiftKey: true }]) {
    keydown({ key: "Enter", preventDefault() { assert.fail("unexpected interception"); }, ...options });
  }
  assert.equal(clicks, 1);
});

test("snapshot creation popups focus their first field and preserve the draft", async (t) => {
  const modal = require("./js/snapshot_asset_modal.js");
  const originals = { document: global.document, go: global.go };
  const originalState = { ...state };
  t.after(() => { Object.assign(global, originals); Object.assign(state, originalState); });
  let focused;
  let renders = 0;
  const draft = { rows: [{ assetID: 42, currentPrice: "123.45" }] };
  state.editDraft = draft;
  state.isTransitioning = false;
  global.document = { getElementById: (id) => ({ focus() { focused = id; } }) };
  global.go = { app: { App: { GetHomePage() {}, GetAssetManagementPage: async () => ({ Assets: [], AssetTypes: [] }) } } };
  for (const [kind, field] of [["asset", "asset-name-input"], ["asset_type", "asset-type-name-input"]]) {
    await modal.openSnapshotAssetModal(kind, {}, () => { renders++; });
    assert.equal(focused, field);
    assert.equal(state.snapshotAssetModal.kind, kind);
    assert.equal(state.editDraft, draft);
  }
  assert.equal(renders, 2);
});
