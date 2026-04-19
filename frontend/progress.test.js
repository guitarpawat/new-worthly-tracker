const test = require("node:test");
const assert = require("node:assert/strict");

const {
  buildAllocationChartConfig,
  buildAllocationRows,
  buildTrendChartConfig,
  formatAllocationPercent,
  formatChartAxisValue,
  formatChartTooltipValue,
  goalStatusToneClass,
  normalizeProgressDateSelection,
  PROGRESS_CHART_COLORS,
  renderAllocationModal,
  renderAllocationTotalRow,
  renderProgressHeroActions,
  renderProjectionSelector,
  resolveAllocationTotalValue,
  resolveProgressView,
  state,
} = require("./app.js");

test("buildTrendChartConfig keeps projection only for total net worth mode", () => {
  const page = {
    TrendPoints: [
      { SnapshotDate: "2026-01-12", TotalCurrent: 10000, ProfitRate: 0.1, CashRatio: 0.2 },
      { SnapshotDate: "2026-02-12", TotalCurrent: 11000, ProfitRate: 0.11, CashRatio: 0.21 },
    ],
    ProjectionPoints: [
      { SnapshotDate: "2026-02-12", TotalCurrent: 11000 },
      { SnapshotDate: "2026-03-12", TotalCurrent: 12000 },
    ],
    Goals: [
      { Name: "Visible", TargetAmount: 11500 },
      { Name: "Too High", TargetAmount: 20000 },
    ],
  };

  const netWorthConfig = buildTrendChartConfig(page, "net_worth", 6);
  const profitConfig = buildTrendChartConfig(page, "profit_rate", 6);

  assert.equal(netWorthConfig.type, "line");
  assert.equal(netWorthConfig.data.datasets.length, 3);
  assert.equal(profitConfig.data.datasets.length, 1);
  assert.equal(netWorthConfig.options.interaction.axis, "x");
  assert.deepEqual(netWorthConfig.data.labels, ["12 Jan 2026", "12 Feb 2026", "Mar 2026"]);
  assert.equal(netWorthConfig.data.datasets[0].pointHitRadius, 24);
  assert.equal(netWorthConfig.data.datasets[2].label, "Visible Goal");
  assert.equal(netWorthConfig.data.datasets[1].borderColor, "#c5c1b9");
  assert.equal(netWorthConfig.data.datasets[2].borderColor, "#275a86");
  assert.equal(netWorthConfig.data.datasets[2].borderWidth, 2);
  assert.deepEqual(netWorthConfig.data.datasets[2].borderDash, [12, 4]);
  assert.equal(netWorthConfig.options.scales.y.ticks.callback(12000), "12,000.00");
  assert.equal(netWorthConfig.options.plugins.tooltip.callbacks.label({
    dataset: { label: "Projection" },
    parsed: { y: 12000 },
  }), "Projection: THB 12,000.00");
});

test("renderProjectionSelector includes all supported projection month options", () => {
  const markup = renderProjectionSelector(6);

  assert.match(markup, /Projection Time/);
  assert.match(markup, /data-value="6"/);
  assert.match(markup, /custom-select-option custom-select-option-active/);
  assert.match(markup, /data-value="12"/);
  assert.match(markup, /data-value="18"/);
  assert.match(markup, /data-value="24"/);
  assert.match(markup, /data-value="36"/);
});

test("buildAllocationChartConfig builds a pie chart from allocation rows", () => {
  const config = buildAllocationChartConfig({
    AllocationSnapshots: [
      {
        SnapshotDate: "2026-04-12",
        ByAssetType: [
          { Name: "Investment", Value: 30000 },
          { Name: "Credit Card", Value: -12000 },
        ],
        ByAsset: [],
        ByCategory: [
          { Name: "Non Cash Asset", Value: 30000 },
          { Name: "Liabilities", Value: 12000 },
        ],
      },
    ],
  }, "2026-04-12", "asset_type");

  assert.equal(config.type, "pie");
  assert.deepEqual(config.data.labels, ["Investment", "Credit Card"]);
  assert.deepEqual(config.data.datasets[0].data, [30000, -12000]);
  assert.equal(config.options.plugins.tooltip.callbacks.title(), "");
  assert.equal(config.options.plugins.tooltip.callbacks.label({
    label: "Investment",
    raw: 30000,
    dataIndex: 0,
  }), "Investment: THB 30,000.00");
  assert.ok(PROGRESS_CHART_COLORS.length >= 8);
});

test("buildAllocationChartConfig uses fixed colors for category allocation mode", () => {
  const config = buildAllocationChartConfig({
    AllocationSnapshots: [
      {
        SnapshotDate: "2026-04-12",
        ByAssetType: [],
        ByAsset: [],
        ByCategory: [
          { Name: "Cash", Value: 15000 },
          { Name: "Liabilities", Value: -12000 },
          { Name: "Non Cash Asset", Value: 30000 },
        ],
      },
    ],
  }, "2026-04-12", "category");

  assert.deepEqual(config.data.datasets[0].backgroundColor, [
    "#7fa882",
    "#c27a7a",
    "#c9a44a",
  ]);
  assert.deepEqual(config.data.datasets[0].data, [15000, -12000, 30000]);
});

test("buildAllocationRows keeps signed values for legend rendering", () => {
  const rows = buildAllocationRows({
    AllocationSnapshots: [
      {
        SnapshotDate: "2026-04-12",
        ByAssetType: [
          { Name: "Investment", Value: 30000 },
          { Name: "Credit Card", Value: -12000 },
        ],
        ByAsset: [],
        ByCategory: [
          { Name: "Non Cash Asset", Value: 30000 },
          { Name: "Liabilities", Value: 12000 },
        ],
      },
    ],
  }, "2026-04-12", "asset_type");

  assert.deepEqual(rows, [
    { name: "Investment", value: 30000 },
    { name: "Credit Card", value: -12000 },
  ]);
});

test("buildAllocationRows supports category mode as a separate allocation view", () => {
  const rows = buildAllocationRows({
    AllocationSnapshots: [
      {
        SnapshotDate: "2026-04-12",
        ByAssetType: [],
        ByAsset: [],
        ByCategory: [
          { Name: "Non Cash Asset", Value: 30000 },
          { Name: "Cash", Value: 15000 },
          { Name: "Liabilities", Value: -12000 },
        ],
      },
    ],
  }, "2026-04-12", "category");

  assert.deepEqual(rows, [
    { name: "Non Cash Asset", value: 30000 },
    { name: "Cash", value: 15000 },
    { name: "Liabilities", value: -12000 },
  ]);
});

test("resolveAllocationTotalValue uses chart rows for category view and snapshot total otherwise", () => {
  assert.equal(resolveAllocationTotalValue(
    { TotalCurrent: 33000 },
    [{ name: "Cash", value: 15000 }, { name: "Liabilities", value: -12000 }],
    "category",
  ), 3000);

  assert.equal(resolveAllocationTotalValue(
    { TotalCurrent: 33000 },
    [{ name: "Investment", value: 30000 }],
    "asset_type",
  ), 33000);
});

test("renderAllocationModal renders a popup with selected mode and total value", () => {
  state.progressAllocationModal = {
    snapshotDate: "2026-04-12",
    mode: "category",
  };

  const markup = renderAllocationModal({
    TrendPoints: [
      { SnapshotDate: "2026-04-12", TotalCurrent: 33000 },
    ],
    AllocationSnapshots: [
      {
        SnapshotDate: "2026-04-12",
        ByAssetType: [],
        ByAsset: [],
        ByCategory: [
          { Name: "Non Cash Asset", Value: 30000 },
          { Name: "Cash", Value: 15000 },
          { Name: "Liabilities", Value: -12000 },
        ],
      },
    ],
  });

  assert.match(markup, /dialog-backdrop/);
  assert.match(markup, /Cash \/ Non-Cash \/ Liabilities/);
  assert.match(markup, /progress-chip-active/);
  assert.match(markup, /progress-allocation-total-row/);
  assert.match(markup, /progress-allocation-legend-header/);
  assert.match(markup, /Total Value/);
  assert.match(markup, /THB 33,000.00/);
  assert.match(markup, /<span class="progress-allocation-percent positive">90.91%<\/span>/);
  assert.match(markup, /<span class="progress-allocation-value positive">THB 30,000.00<\/span>/);

  state.progressAllocationModal = null;
});

test("renderAllocationTotalRow matches allocation legend styling", () => {
  const markup = renderAllocationTotalRow(
    { TotalCurrent: 33000 },
    [{ name: "Cash", value: 15000 }, { name: "Liabilities", value: -12000 }],
    "category",
  );

  assert.match(markup, /progress-allocation-legend-row progress-allocation-total-row/);
  assert.match(markup, /progress-allocation-total-swatch/);
  assert.match(markup, /Total Value/);
  assert.match(markup, /THB 3,000.00/);
});

test("renderProgressHeroActions shows date controls without apply button", () => {
  const markup = renderProgressHeroActions("2026-01-01", "2026-12-31", ["2026-01-01", "2026-12-31"]);

  assert.match(markup, /progress-start-date/);
  assert.match(markup, /progress-end-date/);
  assert.match(markup, /progress-quick-range/);
  assert.doesNotMatch(markup, /progress-apply-filter/);
  assert.doesNotMatch(markup, />Apply</);
});

test("formatAllocationPercent uses two decimal places", () => {
  assert.equal(formatAllocationPercent(30000, 33000), "90.91%");
  assert.equal(formatAllocationPercent(-12000, 33000), "-36.36%");
  assert.equal(formatAllocationPercent(0, 0), "0.00%");
});

test("chart formatters keep two decimal places for amount and percent", () => {
  assert.equal(formatChartAxisValue(12000, "net_worth"), "12,000.00");
  assert.equal(formatChartAxisValue(0.1234, "profit_rate"), "12.34%");
  assert.equal(formatChartTooltipValue(12000, "net_worth"), "THB 12,000.00");
  assert.equal(formatChartTooltipValue(0.1234, "cash_ratio"), "12.34%");
});

test("goalStatusToneClass maps positive and negative states to shared pills", () => {
  assert.equal(goalStatusToneClass("Reached"), "delta-positive");
  assert.equal(goalStatusToneClass("Projected"), "delta-caution");
  assert.equal(goalStatusToneClass("Behind target"), "delta-negative");
  assert.equal(goalStatusToneClass("Something else"), "delta-neutral");
});

test("normalizeProgressDateSelection falls back to latest available date", () => {
  const snapshots = [
    { SnapshotDate: "2026-03-12" },
    { SnapshotDate: "2026-04-12" },
  ];

  assert.equal(normalizeProgressDateSelection(snapshots, "2026-04-12"), "2026-04-12");
  assert.equal(normalizeProgressDateSelection(snapshots, "2026-01-12"), "2026-04-12");
});

test("resolveProgressView prefers explicit subpage selection", () => {
  state.progressView = "summary";

  assert.equal(resolveProgressView({ view: "trend" }), "trend");
  assert.equal(resolveProgressView({}), "summary");
});
