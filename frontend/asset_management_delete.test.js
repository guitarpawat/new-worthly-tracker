const test = require("node:test");
const assert = require("node:assert/strict");
const deletion = require("./js/asset_management_delete.js");
const { state } = require("./js/shared.js");

test("delete actions disable blocked assets and types without warning text", (t) => {
  const original = { ...state };
  t.after(() => Object.assign(state, original));
  state.assetManagementPage = { Assets: [{ ID: 1, Name: "Fund", CanDelete: false }], AssetTypes: [{ ID: 2, Name: "Funds", AssetCount: 1 }] };
  state.assetForm = { id: 1 };
  state.assetTypeForm = { id: 2 };
  state.assetManagementModal = { kind: "asset" };
  assert.match(deletion.editorOptions().deleteAction, /disabled/);
  assert.match(deletion.editorOptions().deleteAction, /title="This asset has snapshot records/);
  assert.equal(deletion.editorOptions().deleteHint, undefined);
  state.assetManagementPage.Assets[0].CanDelete = true;
  assert.doesNotMatch(deletion.editorOptions().deleteAction, /disabled/);
  state.assetManagementModal.kind = "asset_type";
  assert.match(deletion.editorOptions().deleteAction, /disabled/);
  assert.match(deletion.editorOptions().deleteAction, /title="This type contains assets/);
  assert.equal(deletion.editorOptions().deleteHint, undefined);
  state.assetManagementPage.AssetTypes[0].AssetCount = 0;
  assert.doesNotMatch(deletion.editorOptions().deleteAction, /disabled/);
});

test("delete confirms, preserves a failed dialog, and reloads after success", async (t) => {
  const original = { ...state };
  const globals = { document: global.document, go: global.go };
  t.after(() => { Object.assign(state, original); Object.assign(global, globals); });
  const fields = new Map();
  const field = (id) => {
    if (!fields.has(id)) fields.set(id, { listeners: {}, addEventListener(event, fn) { this.listeners[event] = fn; }, focus() {} });
    return fields.get(id);
  };
  global.document = { getElementById: field };
  state.assetManagementPage = { Assets: [{ ID: 1, Name: "Fund", CanDelete: true }] };
  state.assetForm = { id: 1 };
  state.assetManagementModal = { kind: "asset" };
  let reloaded;
  let calls = 0;
  global.go = { app: { App: { GetHomePage() {}, async DeleteAsset() { calls++; throw new Error("now has records"); } } } };
  deletion.bind(() => {}, async (options) => { reloaded = options; });
  field("management-delete").listeners.click();
  assert.equal(state.assetManagementModal.confirmDelete, true);
  assert.equal(calls, 0);
  await field("management-delete-confirm").listeners.click();
  assert.equal(state.assetManagementModal.error, "now has records");
  global.go.app.App.DeleteAsset = async () => {};
  await field("management-delete-confirm").listeners.click();
  assert.equal(state.assetManagementModal, null);
  assert.deepEqual(reloaded, { view: "edit_asset", preserveScroll: true });
});
