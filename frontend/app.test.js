const test = require("node:test");
const assert = require("node:assert/strict");

const {
  EDIT_TAB_INDEX,
  bindPriceInput,
  buildAssetCreatePayload,
  buildAssetFormState,
  buildAssetTypeFormState,
  buildAssetTypeSelectOptions,
  buildCalendarWeeks,
  buildEmptyAssetForm,
  buildEmptyAssetTypeForm,
  buildCreateSnapshotInput,
  buildDeleteSnapshotDialogModel,
  buildHomeOverflowActions,
  buildSaveSnapshotInput,
  deactivateManualField,
  findVerticalTargetAssetID,
  formatDateLabel,
  formatEditableNumber,
  formatTHB,
  isEditGridShortcutKey,
  isEditHorizontalShortcutKey,
  isValidPartialDecimal,
  renderAssetEditorCard,
  renderInteractiveToggle,
  renderAssetReorderPage,
  renderAssetTypeReorderPage,
  renderAssetTable,
  renderAssetTypeEditorCard,
  renderAssetTypeTable,
  buildGoalPayload,
  formatMonthYearLabel,
  renderDateControl,
  renderSelectControl,
  handleAssetManagementKeydown,
  shouldCloseAssetManagementModalOnEscape,
  shouldCloseSnapshotAssetModalOnEscape,
  resolveMonthStart,
  resolveAssetManagementView,
  buildReorderAssetPayloads,
  buildReorderAssetTypesPayload,
  ensureAssetReorderState,
  reorderVisibleRows,
  renderReorderFilterToggle,
  resolveDeleteDialogShortcutAction,
  resolveEditVerticalDirection,
  resolveEditHorizontalMove,
  resolveHomeShortcutAction,
  shouldCloseHomeMenuOnClick,
  scrollHomePage,
  sanitizePartialDecimal,
  shouldAcceptReorderDrag,
  startReorderDrag,
  state,
  shouldDeactivateManualFieldOnKey,
  shouldDiscardEditPageOnEscape,
  shouldActivateManualFieldOnEnter,
  todayISODate,
} = require("./app.js");

test("formatTHB uses parentheses for negative values", () => {
  assert.equal(formatTHB(-11800), "(THB 11,800.00)");
  assert.equal(formatTHB(11800), "THB 11,800.00");
});

test("formatDateLabel matches home page date format", () => {
  assert.equal(formatDateLabel("2026-04-12"), "12 Apr 2026");
  assert.equal(formatDateLabel("2026-04-12T00:00:00Z"), "12 Apr 2026");
});

test("renderSelectControl renders hidden input and selected option label", () => {
  const markup = renderSelectControl({
    id: "snapshot-select",
    value: 1,
    options: [
      { value: 0, label: "12 Apr 2026" },
      { value: 1, label: "12 Mar 2026" },
    ],
  });

  assert.match(markup, /id="snapshot-select"/);
  assert.match(markup, /12 Mar 2026/);
  assert.match(markup, /data-control-select-option="snapshot-select"/);
});

test("renderDateControl renders custom date trigger without native input", () => {
  const markup = renderDateControl({
    id: "snapshot-date-input",
    value: "2026-04-12",
  });

  assert.match(markup, /id="snapshot-date-input"/);
  assert.match(markup, /12 Apr 2026/);
  assert.doesNotMatch(markup, /type="date"/);
  assert.match(markup, /data-direction="-12">&lt;&lt;</);
  assert.match(markup, /data-direction="-1">&lt;</);
  assert.match(markup, /data-direction="1">&gt;</);
  assert.match(markup, /data-direction="12">&gt;&gt;</);
});

test("calendar helpers build a full monday-first month grid", () => {
  assert.equal(resolveMonthStart("2026-04-12"), "2026-04-01");
  assert.equal(formatMonthYearLabel("2026-04-01"), "Apr 2026");
  assert.equal(formatMonthYearLabel("2027-04-01"), "Apr 2027");

  const days = buildCalendarWeeks("2026-04-01", "2026-04-12");
  assert.equal(days.length, 42);
  assert.equal(days[0].isoDate, "2026-03-30");
  assert.equal(days[13].isoDate, "2026-04-12");
  assert.equal(days[13].isCurrentMonth, true);
});

test("todayISODate returns yyyy-mm-dd", () => {
  assert.match(todayISODate(), /^\d{4}-\d{2}-\d{2}$/);
});

test("sanitizePartialDecimal removes invalid characters and keeps two decimals", () => {
  assert.equal(sanitizePartialDecimal("12a3.456b"), "123.45");
  assert.equal(sanitizePartialDecimal("--5.2.7"), "-5.27");
  assert.equal(sanitizePartialDecimal("THB -1,234.50"), "-1234.50");
});

test("isValidPartialDecimal accepts only partial decimal states used by edit inputs", () => {
  assert.equal(isValidPartialDecimal(""), true);
  assert.equal(isValidPartialDecimal("-"), true);
  assert.equal(isValidPartialDecimal("123.45"), true);
  assert.equal(isValidPartialDecimal("123.456"), false);
  assert.equal(isValidPartialDecimal("12a"), false);
});

test("formatEditableNumber always keeps two decimal places", () => {
  assert.equal(formatEditableNumber(10), "10.00");
  assert.equal(formatEditableNumber(-5.2), "-5.20");
});

test("buildSaveSnapshotInput forces cash bought price to zero", () => {
  const payload = buildSaveSnapshotInput({
    SnapshotID: 2,
    SnapshotDate: "2026-04-12",
    Groups: [
      {
        AssetTypeName: "Cash",
        Rows: [
          {
            AssetID: 1,
            IsCash: true,
            BoughtPrice: 9999,
            CurrentPrice: 1200,
            Remarks: "Cash row",
          },
        ],
      },
      {
        AssetTypeName: "Investment",
        Rows: [
          {
            AssetID: 2,
            IsCash: false,
            BoughtPrice: 1500,
            CurrentPrice: 1700,
            Remarks: "ETF",
          },
        ],
      },
    ],
  });

  assert.deepEqual(payload, {
    SnapshotID: 2,
    SnapshotDate: "2026-04-12",
    Items: [
      {
        AssetID: 1,
        BoughtPrice: 0,
        CurrentPrice: 1200,
        Remarks: "Cash row",
      },
      {
        AssetID: 2,
        BoughtPrice: 1500,
        CurrentPrice: 1700,
        Remarks: "ETF",
      },
    ],
  });
});

test("buildCreateSnapshotInput omits snapshot id and forces cash bought price to zero", () => {
  const payload = buildCreateSnapshotInput({
    SnapshotDate: "2026-04-12",
    Groups: [
      {
        AssetTypeName: "Cash",
        Rows: [
          {
            AssetID: 1,
            IsCash: true,
            BoughtPrice: 9999,
            CurrentPrice: 1200,
            Remarks: "Cash row",
          },
        ],
      },
      {
        AssetTypeName: "Investment",
        Rows: [
          {
            AssetID: 2,
            IsCash: false,
            BoughtPrice: 1500,
            CurrentPrice: 1700,
            Remarks: "ETF",
          },
        ],
      },
    ],
  });

  assert.deepEqual(payload, {
    SnapshotDate: "2026-04-12",
    Items: [
      {
        AssetID: 1,
        BoughtPrice: 0,
        CurrentPrice: 1200,
        Remarks: "Cash row",
      },
      {
        AssetID: 2,
        BoughtPrice: 1500,
        CurrentPrice: 1700,
        Remarks: "ETF",
      },
    ],
  });
});

test("buildHomeOverflowActions keeps future pages in menu and moves delete there", () => {
  const actions = buildHomeOverflowActions({
    HasSnapshot: true,
    SnapshotDate: "2026-04-12",
  });

  assert.deepEqual(actions.map((action) => action.id), [
    "asset_management",
    "progress",
    "delete",
  ]);
  assert.equal(actions[0].disabled, false);
  assert.equal(actions[0].note, "Manage assets and asset types");
  assert.equal(actions[1].disabled, false);
  assert.equal(actions[1].label, "Progress & Goals");
  assert.equal(actions[1].note, "Summary table, chart, and goal projection");
  assert.equal(actions[2].disabled, false);
  assert.equal(actions[2].label, "Delete Snapshot");
  assert.equal(actions[2].note, "12 Apr 2026");
});

test("resolveAssetManagementView prefers explicit child page selection", () => {
  state.assetManagementView = "create_asset";

  assert.equal(resolveAssetManagementView({ view: "create_asset" }), "create_asset");
  assert.equal(resolveAssetManagementView({ selectedAssetID: 9 }), "edit_asset");
  assert.equal(resolveAssetManagementView({ selectedAssetTypeID: 4 }), "edit_asset_type");
  assert.equal(resolveAssetManagementView({ view: "reorder_asset" }), "reorder_asset");
  assert.equal(resolveAssetManagementView({}), "create_asset");
});

test("shouldCloseAssetManagementModalOnEscape closes only edit popup escape", () => {
  assert.equal(shouldCloseAssetManagementModalOnEscape({
    key: "Escape",
    hasAssetManagementModal: true,
  }), true);

  assert.equal(shouldCloseAssetManagementModalOnEscape({
    key: "Escape",
    hasAssetManagementModal: false,
  }), false);

  assert.equal(shouldCloseAssetManagementModalOnEscape({
    key: "Enter",
    hasAssetManagementModal: true,
  }), false);
});

test("reorderVisibleRows keeps inactive rows in place for active-only mode", () => {
  const rows = [
    { ID: 3, IsActive: false },
    { ID: 1, IsActive: true },
    { ID: 4, IsActive: true },
  ];

  const reordered = reorderVisibleRows(rows, 4, 1, true);

  assert.deepEqual(reordered.map((row) => row.ID), [3, 4, 1]);
});

test("buildReorderAssetTypesPayload uses visible active rows when filter is enabled", () => {
  state.assetManagementPage = {
    AssetTypes: [
      { ID: 2, Name: "Cash", IsActive: false, Ordering: 1, AssetCount: 1 },
      { ID: 1, Name: "Investment", IsActive: true, Ordering: 2, AssetCount: 2 },
      { ID: 3, Name: "Property", IsActive: true, Ordering: 3, AssetCount: 0 },
    ],
    Assets: [],
  };

  const reorder = ensureAssetReorderState(state.assetManagementPage);
  reorder.typeRows = [
    { ID: 2, Name: "Cash", IsActive: false, Ordering: 1, AssetCount: 1 },
    { ID: 3, Name: "Property", IsActive: true, Ordering: 3, AssetCount: 0 },
    { ID: 1, Name: "Investment", IsActive: true, Ordering: 2, AssetCount: 2 },
  ];

  assert.deepEqual(buildReorderAssetTypesPayload(reorder), {
    OrderedIDs: [3, 1],
    ActiveOnly: true,
  });
});

test("buildReorderAssetPayloads emits one payload per visible asset type group", () => {
  state.assetManagementPage = {
    AssetTypes: [
      { ID: 1, Name: "Investment", IsActive: true, Ordering: 1, AssetCount: 3 },
      { ID: 2, Name: "Cash", IsActive: true, Ordering: 2, AssetCount: 1 },
    ],
    Assets: [
      { ID: 3, AssetTypeID: 1, AssetTypeName: "Investment", Name: "Old", IsActive: false, Ordering: 1, Broker: "IBKR" },
      { ID: 4, AssetTypeID: 1, AssetTypeName: "Investment", Name: "Growth", IsActive: true, Ordering: 2, Broker: "KKP" },
      { ID: 1, AssetTypeID: 1, AssetTypeName: "Investment", Name: "ETF", IsActive: true, Ordering: 3, Broker: "KKP" },
      { ID: 2, AssetTypeID: 2, AssetTypeName: "Cash", Name: "Wallet", IsActive: true, Ordering: 1, Broker: "SCB" },
    ],
  };

  const reorder = ensureAssetReorderState(state.assetManagementPage);

  assert.deepEqual(buildReorderAssetPayloads(reorder), [
    { AssetTypeID: 1, OrderedIDs: [4, 1], ActiveOnly: true },
    { AssetTypeID: 2, OrderedIDs: [2], ActiveOnly: true },
  ]);
});

test("render reorder pages show expected headings", () => {
  state.assetManagementPage = {
    AssetTypes: [
      { ID: 1, Name: "Investment", IsActive: true, Ordering: 1, AssetCount: 2 },
    ],
    Assets: [
      { ID: 1, Name: "ETF", AssetTypeID: 1, AssetTypeName: "Investment", Broker: "KKP", IsCash: false, IsActive: true, Ordering: 1, AutoIncrement: 1000 },
    ],
  };
  ensureAssetReorderState(state.assetManagementPage);

  assert.match(renderAssetTypeReorderPage(), /Reorder Asset Types/);
  const markup = renderAssetReorderPage();
  assert.match(markup, /Reorder Assets/);
  assert.match(markup, /data-asset-reorder-toggle="1"/);
  assert.match(markup, /Hidden/);
});

test("renderReorderFilterToggle uses the shared toggle markup", () => {
  const markup = renderReorderFilterToggle("reorder-filter", true);

  assert.match(markup, /toggle-field/);
  assert.match(markup, /Active only/);
  assert.match(markup, /checked/);
});

test("startReorderDrag sets dataTransfer payload for browser drag and drop", () => {
  let captured = null;
  const event = {
    dataTransfer: {
      effectAllowed: "",
      setData(type, value) {
        captured = { type, value };
      },
    },
  };

  startReorderDrag(event, { id: 7, scope: "asset_type" });

  assert.deepEqual(state.assetReorder.drag, { id: 7, scope: "asset_type" });
  assert.equal(event.dataTransfer.effectAllowed, "move");
  assert.deepEqual(captured, { type: "text/plain", value: "7" });
  assert.equal(shouldAcceptReorderDrag(state.assetReorder.drag, "asset_type"), true);
  assert.equal(shouldAcceptReorderDrag(state.assetReorder.drag, "asset:1"), false);
});

test("buildEmptyAssetTypeForm starts with active enabled", () => {
  assert.deepEqual(buildEmptyAssetTypeForm(), {
    id: 0,
    name: "",
    isActive: true,
  });
});

test("buildAssetTypeFormState loads selected row", () => {
  const form = buildAssetTypeFormState({
    AssetTypes: [
      { ID: 2, Name: "Investment", IsActive: false },
    ],
  }, 2);

  assert.deepEqual(form, {
    id: 2,
    name: "Investment",
    isActive: false,
  });
});

test("buildEmptyAssetForm defaults to first active asset type", () => {
  const form = buildEmptyAssetForm({
    ActiveAssetTypes: [
      { ID: 2, Name: "Investment" },
    ],
  });

  assert.equal(form.id, 0);
  assert.equal(form.assetTypeID, 2);
  assert.equal(form.autoIncrement, "0.00");
});

test("buildAssetFormState loads selected asset", () => {
  const form = buildAssetFormState({
    Assets: [
      {
        ID: 3,
        Name: "US Index Fund",
        AssetTypeID: 2,
        Broker: "IBKR",
        IsCash: false,
        IsActive: true,
        AutoIncrement: 5200,
      },
    ],
  }, 3);

  assert.equal(form.id, 3);
  assert.equal(form.name, "US Index Fund");
  assert.equal(form.assetTypeID, 2);
  assert.equal(form.autoIncrement, "5200.00");
});

test("buildAssetTypeSelectOptions keeps selected inactive type available in edit mode", () => {
  const options = buildAssetTypeSelectOptions({
    ActiveAssetTypes: [
      { ID: 1, Name: "Cash" },
    ],
    AssetTypes: [
      { ID: 1, Name: "Cash", IsActive: true },
      { ID: 2, Name: "Investment", IsActive: false },
    ],
  }, 2);

  assert.deepEqual(options, [
    { id: 1, label: "Cash" },
    { id: 2, label: "Investment (Inactive)" },
  ]);
});

test("buildAssetCreatePayload forces cash auto increment to zero", () => {
  const payload = buildAssetCreatePayload({
    name: "Emergency Fund",
    assetTypeID: 1,
    broker: "SCB",
    isCash: true,
    isLiability: false,
    isActive: true,
    autoIncrement: "6500.00",
  });

  assert.deepEqual(payload, {
    Name: "Emergency Fund",
    AssetTypeID: 1,
    Broker: "SCB",
    IsCash: true,
    IsLiability: false,
    IsActive: true,
    AutoIncrement: 0,
  });
});

test("renderInteractiveToggle uses checkbox state instead of fixed on off classes", () => {
  const markup = renderInteractiveToggle("asset-type-active-input", true);

  assert.match(markup, /type="checkbox" checked/);
  assert.match(markup, /class="mock-toggle"/);
  assert.doesNotMatch(markup, /mock-toggle-on|mock-toggle-off/);
});

test("renderAssetTable highlights selected row without inline dropdown editor", () => {
  const markup = renderAssetTable({
    Assets: [
      {
        ID: 4,
        Name: "Cash Buffer",
        AssetTypeID: 1,
        AssetTypeName: "Cash",
        Broker: "SCB",
        IsCash: true,
        IsActive: true,
        AutoIncrement: 0,
      },
      {
        ID: 8,
        Name: "US Index Fund",
        AssetTypeID: 2,
        AssetTypeName: "Investment",
        Broker: "IBKR",
        IsCash: false,
        IsActive: true,
        AutoIncrement: 5200,
      },
    ],
    ActiveAssetTypes: [
      { ID: 1, Name: "Cash" },
      { ID: 2, Name: "Investment" },
    ],
  }, {
    id: 4,
    name: "Cash Buffer",
    assetTypeID: 1,
    broker: "SCB",
    isCash: true,
    isActive: true,
    autoIncrement: "0.00",
  });

  assert.match(
    markup,
    /management-row-active[\s\S]*data-asset-row-id="4"[\s\S]*data-asset-row-id="8"/,
  );
  assert.match(markup, /delta-positive[\s\S]*Yes/);
  assert.match(markup, /delta-negative[\s\S]*No/);
  assert.doesNotMatch(markup, /management-inline-row|mock-toggle-on|mock-toggle-off/);
});

test("renderAssetTypeTable highlights selected row without inline dropdown editor", () => {
  const markup = renderAssetTypeTable({
    AssetTypes: [
      { ID: 2, Name: "Investment", AssetCount: 4, IsActive: true },
      { ID: 3, Name: "Credit Card", AssetCount: 1, IsActive: false },
    ],
  }, {
    id: 2,
    name: "Investment",
    isActive: true,
  });

  assert.match(
    markup,
    /management-row-active[\s\S]*data-asset-type-row-id="2"[\s\S]*data-asset-type-row-id="3"/,
  );
  assert.match(markup, /delta-positive[\s\S]*Yes/);
  assert.match(markup, /delta-negative[\s\S]*No/);
  assert.doesNotMatch(markup, /management-inline-row/);
});

test("renderAssetEditorCard keeps toggles left and actions right in one footer row", () => {
  const markup = renderAssetEditorCard({
    ActiveAssetTypes: [
      { ID: 1, Name: "Cash" },
    ],
  }, {
    id: 0,
    name: "",
    assetTypeID: 1,
    broker: "",
    isCash: false,
    isActive: true,
    autoIncrement: "0.00",
  });

  assert.match(
    markup,
    /asset-management-form-footer asset-management-form-footer-wide[\s\S]*asset-management-toggle-row[\s\S]*Cash Asset[\s\S]*Liability[\s\S]*asset-management-card-actions[\s\S]*Create Asset[\s\S]*Reset/,
  );
  assert.doesNotMatch(markup, />Active</);
});

test("renderAssetEditorCard supports custom close label for snapshot modal reuse", () => {
  const markup = renderAssetEditorCard({
    AssetTypes: [{ ID: 1, Name: "Cash", IsActive: true }],
  }, {
    id: 0,
    name: "",
    assetTypeID: 1,
    broker: "",
    isCash: false,
    isActive: true,
    autoIncrement: "0.00",
  }, {
    secondaryButtonLabel: "Close",
  });

  assert.match(markup, /Create Asset/);
  assert.match(markup, /id="asset-reset-button" class="button" type="button">Close</);
});

test("renderAssetTypeEditorCard keeps active toggle left and actions right in one footer row", () => {
  const markup = renderAssetTypeEditorCard({
    id: 0,
    name: "",
    isActive: true,
  });

  assert.match(
    markup,
    /asset-management-form-footer[\s\S]*asset-management-form-footer-left[\s\S]*asset-management-card-actions[\s\S]*Create Type[\s\S]*Reset/,
  );
  assert.doesNotMatch(markup, />Active</);
});

test("render edit forms keep active toggle visible", () => {
  const assetMarkup = renderAssetEditorCard({
    ActiveAssetTypes: [{ ID: 1, Name: "Cash" }],
  }, {
    id: 7,
    name: "Wallet",
    assetTypeID: 1,
    broker: "",
    isCash: true,
    isLiability: false,
    isActive: true,
    autoIncrement: "0.00",
  });
  const assetTypeMarkup = renderAssetTypeEditorCard({
    id: 5,
    name: "Cash",
    isActive: true,
  });

  assert.match(assetMarkup, />Active</);
  assert.match(assetTypeMarkup, />Active</);
});

test("renderAssetTypeEditorCard supports custom close label for snapshot modal reuse", () => {
  const markup = renderAssetTypeEditorCard({
    id: 0,
    name: "",
    isActive: true,
  }, {
    secondaryButtonLabel: "Close",
  });

  assert.match(markup, /Create Type/);
  assert.match(markup, /id="asset-type-reset-button" class="button" type="button">Close</);
});

test("buildDeleteSnapshotDialogModel uses the formatted snapshot date", () => {
  const dialog = buildDeleteSnapshotDialogModel({
    HasSnapshot: true,
    SnapshotDate: "2026-04-12",
  });

  assert.equal(dialog.title, "12 Apr 2026");
  assert.equal(
    dialog.body,
    "This will soft delete snapshot 12 Apr 2026 and its record rows. You can still keep older history.",
  );
});

test("shouldCloseSnapshotAssetModalOnEscape closes only snapshot popup escape", () => {
  assert.equal(shouldCloseSnapshotAssetModalOnEscape({
    key: "Escape",
    hasSnapshotAssetModal: true,
  }), true);

  assert.equal(shouldCloseSnapshotAssetModalOnEscape({
    key: "Escape",
    hasSnapshotAssetModal: false,
  }), false);

  assert.equal(shouldCloseSnapshotAssetModalOnEscape({
    key: "Enter",
    hasSnapshotAssetModal: true,
  }), false);
});

test("buildGoalPayload trims the name and parses amount", () => {
  assert.deepEqual(buildGoalPayload({
    name: "  FIRE Number  ",
    targetAmount: "1500000.50",
    targetDate: "2028-12-31",
  }), {
    Name: "FIRE Number",
    TargetAmount: 1500000.5,
    TargetDate: "2028-12-31",
  });
});

test("resolveHomeShortcutAction maps supported home page keys", () => {
  assert.equal(resolveHomeShortcutAction({
    key: "ArrowLeft",
    canNavigateBack: true,
    canNavigateForward: false,
    canEditSnapshot: true,
    canCreateSnapshot: false,
    isEditableTarget: false,
  }), "previous");

  assert.equal(resolveHomeShortcutAction({
    key: "A",
    canNavigateBack: true,
    canNavigateForward: true,
    canEditSnapshot: true,
    canCreateSnapshot: false,
    canJumpToLatest: true,
    isEditableTarget: false,
  }), "previous");

  assert.equal(resolveHomeShortcutAction({
    key: "ArrowRight",
    canNavigateBack: true,
    canNavigateForward: true,
    canEditSnapshot: true,
    canCreateSnapshot: false,
    isEditableTarget: false,
  }), "next");

  assert.equal(resolveHomeShortcutAction({
    key: "D",
    canNavigateBack: true,
    canNavigateForward: true,
    canEditSnapshot: true,
    canCreateSnapshot: false,
    canJumpToLatest: true,
    isEditableTarget: false,
  }), "next");

  assert.equal(resolveHomeShortcutAction({
    key: "E",
    canNavigateBack: true,
    canNavigateForward: true,
    canEditSnapshot: true,
    canCreateSnapshot: false,
    isEditableTarget: false,
  }), "edit");

  assert.equal(resolveHomeShortcutAction({
    key: "+",
    canNavigateBack: true,
    canNavigateForward: true,
    canEditSnapshot: true,
    canCreateSnapshot: true,
    canJumpToLatest: true,
    isEditableTarget: false,
  }), "new");

  assert.equal(resolveHomeShortcutAction({
    key: "Add",
    canNavigateBack: true,
    canNavigateForward: true,
    canEditSnapshot: true,
    canCreateSnapshot: true,
    canJumpToLatest: true,
    isEditableTarget: false,
  }), "new");

  assert.equal(resolveHomeShortcutAction({
    key: "N",
    canNavigateBack: true,
    canNavigateForward: true,
    canEditSnapshot: true,
    canCreateSnapshot: true,
    canJumpToLatest: true,
    isEditableTarget: false,
  }), "new");

  assert.equal(resolveHomeShortcutAction({
    key: "L",
    canNavigateBack: true,
    canNavigateForward: true,
    canEditSnapshot: true,
    canCreateSnapshot: false,
    canJumpToLatest: true,
    isEditableTarget: false,
  }), "latest");

  assert.equal(resolveHomeShortcutAction({
    key: "W",
    canNavigateBack: true,
    canNavigateForward: true,
    canEditSnapshot: true,
    canCreateSnapshot: false,
    canJumpToLatest: true,
    isEditableTarget: false,
  }), "scroll_up");

  assert.equal(resolveHomeShortcutAction({
    key: "s",
    canNavigateBack: true,
    canNavigateForward: true,
    canEditSnapshot: true,
    canCreateSnapshot: false,
    canJumpToLatest: true,
    isEditableTarget: false,
  }), "scroll_down");

  assert.equal(resolveHomeShortcutAction({
    key: "E",
    canNavigateBack: true,
    canNavigateForward: true,
    canEditSnapshot: true,
    canCreateSnapshot: true,
    canJumpToLatest: true,
    isEditableTarget: true,
  }), null);
});

test("resolveDeleteDialogShortcutAction maps enter and escape", () => {
  assert.equal(resolveDeleteDialogShortcutAction({
    key: "Escape",
    isEditableTarget: false,
  }), "cancel");

  assert.equal(resolveDeleteDialogShortcutAction({
    key: "Enter",
    isEditableTarget: false,
  }), null);

  assert.equal(resolveDeleteDialogShortcutAction({
    key: "Enter",
    isEditableTarget: true,
  }), null);
});

test("shouldCloseHomeMenuOnClick closes only for outside clicks on the home page", () => {
  assert.equal(shouldCloseHomeMenuOnClick({
    currentView: "home",
    isMenuOpen: true,
    clickedInsideMenu: false,
  }), true);

  assert.equal(shouldCloseHomeMenuOnClick({
    currentView: "home",
    isMenuOpen: true,
    clickedInsideMenu: true,
  }), false);

  assert.equal(shouldCloseHomeMenuOnClick({
    currentView: "edit",
    isMenuOpen: true,
    clickedInsideMenu: false,
  }), false);
});

test("resolveEditHorizontalMove follows row navigation rules", () => {
  assert.equal(resolveEditHorizontalMove({
    key: "A",
    fieldKey: "current",
    isCash: false,
    isNotesEditable: false,
  }), "bought");

  assert.equal(resolveEditHorizontalMove({
    key: "D",
    fieldKey: "current",
    isCash: false,
    isNotesEditable: false,
  }), "notes");

  assert.equal(resolveEditHorizontalMove({
    key: "A",
    fieldKey: "current",
    isCash: true,
    isNotesEditable: false,
  }), null);

  assert.equal(resolveEditHorizontalMove({
    key: "D",
    fieldKey: "current",
    isCash: true,
    isNotesEditable: false,
  }), "notes");

  assert.equal(resolveEditHorizontalMove({
    key: "D",
    fieldKey: "notes",
    isCash: false,
    isNotesEditable: true,
  }), null);
});

test("resolveEditVerticalDirection maps W and S unless notes are actively editable", () => {
  assert.equal(resolveEditVerticalDirection({
    key: "W",
    isNotesEditable: false,
  }), -1);

  assert.equal(resolveEditVerticalDirection({
    key: "s",
    isNotesEditable: false,
  }), 1);

  assert.equal(resolveEditVerticalDirection({
    key: "W",
    isNotesEditable: true,
  }), null);
});

test("findVerticalTargetAssetID keeps the same field and skips cash rows for bought price", () => {
  const draft = {
    Groups: [
      {
        Rows: [
          { AssetID: 1, IsCash: false },
          { AssetID: 2, IsCash: true },
        ],
      },
      {
        Rows: [
          { AssetID: 3, IsCash: false },
        ],
      },
    ],
  };

  assert.equal(findVerticalTargetAssetID(draft, 1, "current", 1), 2);
  assert.equal(findVerticalTargetAssetID(draft, 1, "bought", 1), 3);
  assert.equal(findVerticalTargetAssetID(draft, 3, "bought", -1), 1);
  assert.equal(findVerticalTargetAssetID(draft, 1, "bought", -1), null);
});

test("EDIT_TAB_INDEX keeps date first, current rows second, save last", () => {
  assert.deepEqual(EDIT_TAB_INDEX, {
    date: 1,
    current: 2,
    save: 3,
  });
});

test("isEditHorizontalShortcutKey matches A and D only", () => {
  assert.equal(isEditHorizontalShortcutKey("A"), true);
  assert.equal(isEditHorizontalShortcutKey("d"), true);
  assert.equal(isEditHorizontalShortcutKey("L"), false);
  assert.equal(isEditHorizontalShortcutKey("1"), false);
});

test("isEditGridShortcutKey includes horizontal and vertical edit navigation keys", () => {
  assert.equal(isEditGridShortcutKey("A"), true);
  assert.equal(isEditGridShortcutKey("s"), true);
  assert.equal(isEditGridShortcutKey("W"), true);
  assert.equal(isEditGridShortcutKey("E"), false);
  assert.equal(isEditGridShortcutKey("Enter"), false);
});

test("shouldActivateManualFieldOnEnter only unlocks read-only manual fields", () => {
  assert.equal(shouldActivateManualFieldOnEnter({
    key: "Enter",
    isManualField: true,
    isDisabled: false,
    isReadOnly: true,
  }), true);

  assert.equal(shouldActivateManualFieldOnEnter({
    key: "Enter",
    isManualField: true,
    isDisabled: true,
    isReadOnly: true,
  }), false);

  assert.equal(shouldActivateManualFieldOnEnter({
    key: "E",
    isManualField: true,
    isDisabled: false,
    isReadOnly: true,
  }), true);

  assert.equal(shouldActivateManualFieldOnEnter({
    key: "A",
    isManualField: true,
    isDisabled: false,
    isReadOnly: true,
  }), false);
});

test("shouldDeactivateManualFieldOnKey exits active manual editing on enter or escape", () => {
  assert.equal(shouldDeactivateManualFieldOnKey({
    key: "Escape",
    isManualField: true,
    isReadOnly: false,
  }), true);

  assert.equal(shouldDeactivateManualFieldOnKey({
    key: "Enter",
    isManualField: true,
    isReadOnly: false,
  }), true);

  assert.equal(shouldDeactivateManualFieldOnKey({
    key: "Escape",
    isManualField: true,
    isReadOnly: true,
  }), false);
});

test("shouldDiscardEditPageOnEscape only triggers outside active field editing", () => {
  assert.equal(shouldDiscardEditPageOnEscape({
    key: "Escape",
    isTypingTarget: false,
    isManualFieldEditing: false,
    isCurrentPriceTarget: false,
  }), true);

  assert.equal(shouldDiscardEditPageOnEscape({
    key: "Escape",
    isTypingTarget: true,
    isManualFieldEditing: false,
    isCurrentPriceTarget: false,
  }), false);

  assert.equal(shouldDiscardEditPageOnEscape({
    key: "Escape",
    isTypingTarget: false,
    isManualFieldEditing: true,
    isCurrentPriceTarget: false,
  }), false);

  assert.equal(shouldDiscardEditPageOnEscape({
    key: "Escape",
    isTypingTarget: true,
    isManualFieldEditing: false,
    isCurrentPriceTarget: true,
  }), true);
});

test("scrollHomePage calls window.scrollBy when available", () => {
  const calls = [];
  const originalWindow = global.window;
  global.window = {
    scrollBy: (options) => {
      calls.push(options);
    },
  };

  scrollHomePage(240);

  assert.deepEqual(calls, [{
    top: 240,
    behavior: "smooth",
  }]);

  global.window = originalWindow;
});

test("bindPriceInput does not block edit shortcut keys on numeric fields", () => {
  state.currentView = "edit";
  const listeners = new Map();
  const input = createFakeInput({
    classNames: ["manual-field", "edit-bought-input"],
    dataset: { assetId: "1" },
  });
  input.addEventListener = (name, handler) => {
    listeners.set(name, handler);
  };

  bindPriceInput(input, {
    onChange: () => {},
  });

  const keydown = listeners.get("keydown");
  assert.ok(keydown, "expected keydown handler");

  for (const key of ["E", "Escape", "W", "S", "A", "D", "Enter"]) {
    const event = createFakeKeyboardEvent(key);
    keydown(event);
    assert.equal(event.defaultPrevented, false, `expected ${key} not to be blocked`);
  }

  state.currentView = "home";
});

test("deactivateManualField can preserve focus on the same field", () => {
  const field = createFakeInput({ classNames: ["manual-field"] });
  field.readOnly = false;

  deactivateManualField(field, { keepFocus: true });

  assert.equal(field.readOnly, true);
  assert.equal(field.classList.contains("manual-field-locked"), true);
  assert.equal(field.focusCalls, 1);
});

function createFakeInput({ classNames = [], dataset = {} } = {}) {
  const classes = new Set(classNames);
  return {
    value: "",
    dataset,
    disabled: false,
    readOnly: true,
    selectionStart: 0,
    selectionEnd: 0,
    focusCalls: 0,
    classList: {
      contains: (name) => classes.has(name),
      add: (name) => classes.add(name),
      remove: (name) => classes.delete(name),
    },
    addEventListener: () => {},
    focus() {
      this.focusCalls += 1;
    },
    blur() {},
    select() {},
  };
}

function createFakeKeyboardEvent(key) {
  return {
    key,
    ctrlKey: false,
    metaKey: false,
    altKey: false,
    defaultPrevented: false,
    preventDefault() {
      this.defaultPrevented = true;
    },
  };
}
