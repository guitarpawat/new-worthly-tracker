const test = require("node:test");
const assert = require("node:assert/strict");
const assetCharts = require("./js/asset_chart_allocations.js");
const ui = require("./js/asset_management_ui.js");
const management = require("./js/asset_management.js");
const { state } = require("./js/shared.js");

function samplePage() {
  return {
    Assets: [{ ID: 1, Name: "Fund 1", AssetTypeID: 1, IsActive: true, AutoIncrement: "0" }],
    AssetTypes: [{ ID: 1, Name: "Funds", IsActive: true }],
    ActiveAssetTypes: [{ ID: 1, Name: "Funds" }],
    CustomAllocationCharts: [{ ID: 7, Name: "Country", Entries: [
      { AssetID: 1, Category: "USA", Percentage: "70" },
      { AssetID: 2, Category: "China", Percentage: "100" },
    ] }, { ID: 8, Name: "Type", Entries: [] }],
  };
}

test("Edit Asset drafts only its own tags, and Save Asset includes each chart", () => {
  const page = samplePage();
  const form = ui.buildAssetFormState(page, 1);
  assert.equal(form.chartAllocations.length, 2);
  assert.equal(form.chartAllocations[0].Entries.length, 1);
  form.chartAllocations[0].Entries[0].Percentage = "50";
  assert.equal(page.CustomAllocationCharts[0].Entries[0].Percentage, "70");
  const payload = ui.buildAssetUpdatePayload(form);
  assert.equal(payload.ChartAllocations[0].Entries[0].Percentage, "50");
  assert.deepEqual(payload.ChartAllocations[1], { ChartID: 8, Excluded: false, Entries: [] });
  assert.equal(assetCharts.validate(form), "");
  form.chartAllocations[0].Entries.push({ AssetID: 1, Category: "Japan", Percentage: "51" });
  assert.match(assetCharts.validate(form), /^Country:.*100%/);
  const fresh = ui.buildAssetFormState(page, 1);
  assert.equal(fresh.chartAllocations[0].Entries[0].Percentage, "70", "reopening discards cancelled draft");
});

test("chart tags appear only in Edit Asset, using the existing fields and buttons", () => {
  const page = samplePage();
  const form = ui.buildAssetFormState(page, 1);
  const markup = ui.renderAssetEditorCard(page, form);
  assert.match(markup, /Chart Tags/);
  assert.match(markup, /<legend>Country<\/legend>/);
  assert.doesNotMatch(markup, /<legend>Type<\/legend>/);
  assert.match(markup, /id="asset-chart-select"/);
  assert.match(markup, /70.00% allocated · 30.00% unallocated/);
  assert.match(markup, /class="form-input numeric"/);
  assert.match(markup, /Add Tag/);
  assert.doesNotMatch(ui.renderAssetEditorCard(page, ui.buildEmptyAssetForm(page)), /Chart Tags/);
});

test("assets default to included; exclusion is per chart and retains saved tags", (t) => {
  const previous = global.document;
  t.after(() => { global.document = previous; });
  const page = samplePage();
  const form = ui.buildAssetFormState(page, 1);
  assert.equal(form.chartAllocations[0].Excluded, false);
  assert.match(assetCharts.render(page, form), /data-asset-chart-included="0" checked/);
  const input = { id: "asset-chart-included-0", dataset: { assetChartIncluded: "0" }, checked: false,
    addEventListener(event, callback) { this.change = callback; } };
  global.document = {
    querySelectorAll: (selector) => selector === "[data-asset-chart-included]" ? [input] : [],
    getElementById: (id) => id === "asset-chart-select" ? null : { focus() {} },
  };
  assetCharts.bind(form, () => {});
  input.change();
  assert.equal(form.chartAllocations[0].Excluded, true);
  assert.equal(form.chartAllocations[1].Excluded, false);
  assert.equal(form.chartAllocations[0].Entries[0].Category, "USA");
  const markup = assetCharts.render(page, form);
  assert.match(markup, /Previously saved tags are kept/);
  assert.doesNotMatch(markup, /data-asset-chart-category=/);
  assert.deepEqual(assetCharts.buildPayload(form)[0], { ChartID: 7, Excluded: true, Entries: [] });
  input.checked = true;
  input.change();
  assert.equal(assetCharts.buildPayload(form)[0].Entries[0].Category, "USA");
  page.CustomAllocationCharts[0].ExcludedAssetIDs = [1];
  assert.equal(ui.buildAssetFormState(page, 1).chartAllocations[0].Excluded, true);
});

test("chart dropdown switches visible tags without discarding other charts, and validation selects the invalid chart", (t) => {
  const previous = global.document;
  t.after(() => { global.document = previous; });
  const page = samplePage();
  const form = ui.buildAssetFormState(page, 1);
  let change;
  const select = { addEventListener(event, handler) { change = handler; } };
  global.document = {
    querySelectorAll: () => [],
    getElementById: (id) => id === "asset-chart-select" ? select : { focus() {} },
  };
  let renders = 0;
  assetCharts.bind(form, () => { renders++; });
  form.chartAllocations[0].Entries[0].Percentage = "60";
  change({ target: { value: "8" } });
  const markup = assetCharts.render(page, form);
  assert.match(markup, /<legend>Type<\/legend>/);
  assert.doesNotMatch(markup, /<legend>Country<\/legend>/);
  assert.equal(form.chartAllocations[0].Entries[0].Percentage, "60");
  assert.equal(renders, 1);
  form.chartAllocations[0].Entries[0].Percentage = "101";
  assert.match(assetCharts.validate(form), /^Country:/);
  assert.equal(form.selectedChartID, 7);
});

test("adding and removing tags keeps the asset draft and restores keyboard focus", (t) => {
  const previous = global.document;
  t.after(() => { global.document = previous; });
  const form = ui.buildAssetFormState(samplePage(), 1);
  const button = (dataset) => ({ dataset, addEventListener(event, handler) { this.click = handler; } });
  const add = button({ assetChartAdd: "1" });
  const remove = button({ assetChartRemove: "0:0" });
  let focused;
  let renders = 0;
  global.document = {
    querySelectorAll: (selector) => selector === "[data-asset-chart-add]" ? [add] : selector === "[data-asset-chart-remove]" ? [remove] : [],
    getElementById: (id) => id === "asset-chart-select" ? null : ({ focus() { focused = id; } }),
  };
  assetCharts.bind(form, () => { renders++; });
  add.click();
  assert.equal(form.chartAllocations[1].Entries.length, 1);
  assert.equal(focused, "asset-chart-category-1-0");
  remove.click();
  assert.equal(form.chartAllocations[0].Entries.length, 0);
  assert.equal(focused, "asset-chart-add-0");
  assert.equal(form.name, "Fund 1");
  assert.equal(renders, 2);
});

test("management menu uses the requested left-to-right order", (t) => {
  const originalDocument = global.document;
  const originalState = { ...state };
  t.after(() => { global.document = originalDocument; Object.assign(state, originalState); });
  let markup;
  global.document = {
    getElementById: (id) => id === "app" ? { set innerHTML(value) { markup = value; } } : null,
    querySelectorAll: () => [],
    activeElement: {},
  };
  state.assetManagementPage = samplePage();
  state.assetManagementView = "create_asset";
  state.assetManagementModal = null;
  state.assetForm = ui.buildEmptyAssetForm(state.assetManagementPage);
  state.assetTypeForm = ui.buildEmptyAssetTypeForm();
  management.renderAssetManagementPage({});
  assert.match(markup, /Manage Chart/);
  const nav = markup.match(/data-asset-management-view="([^"]+)"/g);
  assert.deepEqual(nav, ["edit_asset", "reorder_asset", "edit_asset_type", "reorder_asset_type", "manage_chart"].map((id) => `data-asset-management-view="${id}"`));
  assert.doesNotMatch(markup, /data-asset-management-view="custom_allocations"/);
});
