(function initSnapshotForm(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const logic = root.WorthlySnapshotFormLogic || (typeof require !== "undefined" ? require("./snapshot_form_logic.js") : null);
  const controls = root.WorthlyControls || (typeof require !== "undefined" ? require("./controls.js") : null);
  const snapshotAssetModal = root.WorthlySnapshotAssetModal || (typeof require !== "undefined" ? require("./snapshot_asset_modal.js") : null);
  const snapshotForm = factory(shared, logic, controls, snapshotAssetModal, root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = snapshotForm;
  }
  root.WorthlySnapshotForm = snapshotForm;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildSnapshotForm(shared, logic, controls, snapshotAssetModal, root) {
  const {
    EDIT_TAB_INDEX,
    cloneValue,
    escapeHTML,
    formatDateLabel,
    formatEditableNumber,
    formatPercent,
    formatTHB,
    isGeneralTypingTarget,
    logHomeAction,
    renderAppTitle,
    renderErrorState,
    runTransition,
    state,
  } = shared;
  const {
    bindManualField,
    bindPriceInput,
    calculateProfitFromRow,
    calculateProfitPercentFromRow,
    deactivateManualField,
    findEditRow,
    findVerticalTargetAssetID,
    focusEditField,
    isEditGridShortcutKey,
    isEditHorizontalShortcutKey,
    isManualFieldEditing,
    resolveEditHorizontalMove,
    resolveEditVerticalDirection,
    shouldActivateManualFieldOnEnter,
    shouldDeactivateManualFieldOnKey,
    shouldDiscardEditPageOnEscape,
  } = logic;

  function initializeDraft(page) {
    state.currentView = "edit";
    state.deleteDialog = null;
    state.editError = "";
    state.editNotice = "";
    state.snapshotAssetModal = null;
    state.editDraft = cloneValue(page);
    sortEditDraft(state.editDraft);
    state.editOriginalSignature = createEditSignature(state.editDraft);
    state.removedRows = [];
    state.hasUnsavedChanges = false;
  }

  function renderEditPage(app) {
    const appRoot = root.document.getElementById("app");
    const draft = state.editDraft;
    if (!draft) {
      appRoot.innerHTML = renderErrorState("Unable to open editor", "Edit state is missing.");
      return;
    }

    const isNewMode = draft.Mode === "new";
    const pageTitle = isNewMode ? "New Snapshot" : "Edit Snapshot";
    const saveLabel = isNewMode ? "Create" : "Save";

    appRoot.innerHTML = `
      <main class="app-layout ${state.snapshotAssetModal ? "app-modal-open" : ""}">
        <section class="hero">
          <div>
            <h1>${renderAppTitle(pageTitle)}</h1>
          </div>
          <div class="actions actions-edit">
            <div class="field-inline">
              <span class="field-label">Snapshot Date</span>
              ${controls.renderDateControl({
                id: "snapshot-date-input",
                value: draft.SnapshotDate,
                tabIndex: EDIT_TAB_INDEX.date,
              })}
            </div>
            <button id="edit-cancel-button" class="button" type="button" tabindex="-1">Cancel</button>
            <button id="save-snapshot-button" class="button button-primary" type="button" tabindex="${EDIT_TAB_INDEX.save}">${saveLabel}</button>
          </div>
        </section>
        <section class="panel editor-toolbar">
          <div class="editor-toolbar-row">
            <p class="toolbar-inline-copy">
              <span class="eyebrow toolbar-inline-label">Add Asset</span>
            </p>
            <div class="asset-picker">
              ${renderAvailableAssetSelector(draft.AvailableAssetGroups)}
              <button id="add-asset-button" class="button" type="button" tabindex="-1" ${hasAvailableAssets(draft) ? "" : "disabled"}>Add Existing Asset</button>
              <button id="create-asset-button" class="button" type="button" tabindex="-1">Create Asset</button>
              <button id="create-asset-type-button" class="button" type="button" tabindex="-1">Create Asset Type</button>
            </div>
          </div>
          ${renderRemovedRowsUndo(state.removedRows)}
          ${state.editNotice ? `<p class="success-copy success-banner">${escapeHTML(state.editNotice)}</p>` : ""}
          ${state.editError ? `<p class="error-copy error-banner">${escapeHTML(state.editError)}</p>` : ""}
        </section>
        ${draft.Groups.map(renderEditGroup).join("")}
      </main>
      ${snapshotAssetModal.renderSnapshotAssetModal()}
    `;

    bindEditPage(app);
  }

  function bindEditPage(app) {
    const dateInput = root.document.getElementById("snapshot-date-input");
    if (dateInput) {
      dateInput.addEventListener("change", () => {
        state.editDraft.SnapshotDate = dateInput.value;
        markEditDirty();
      });
    }

    for (const input of root.document.querySelectorAll(".edit-bought-input")) {
      bindPriceInput(input, {
        onChange: (row, value) => {
          if (!row.IsCash) {
            row.BoughtPrice = value;
          }
        },
        onDirty: markEditDirty,
      });
    }

    for (const input of root.document.querySelectorAll(".edit-current-input")) {
      bindPriceInput(input, {
        onChange: (row, value) => {
          row.CurrentPrice = value;
        },
        onDirty: markEditDirty,
      });
    }

    for (const input of root.document.querySelectorAll(".edit-remarks-input")) {
      bindManualField(input);
      input.addEventListener("input", (event) => {
        const row = findEditRow(Number(event.target.dataset.assetId));
        if (!row) {
          return;
        }
        row.Remarks = event.target.value;
        markEditDirty();
      });
    }

    for (const input of root.document.querySelectorAll(".manual-field")) {
      bindManualField(input);
    }

    for (const button of root.document.querySelectorAll(".remove-row-button")) {
      button.addEventListener("click", async (event) => {
        const assetID = Number(event.currentTarget.dataset.assetId);
        await runTransition(async () => {
          void logHomeAction("remove_asset_from_snapshot_click", state.offset);
          removeAssetFromDraft(assetID);
          markEditDirty();
          renderEditPage(app);
        });
      });
    }

    for (const button of root.document.querySelectorAll(".undo-remove-button")) {
      button.addEventListener("click", async (event) => {
        const assetID = Number(event.currentTarget.dataset.assetId);
        await runTransition(async () => {
          void logHomeAction("undo_remove_asset_click", state.offset);
          undoRemovedAsset(assetID);
          markEditDirty();
          renderEditPage(app);
        });
      });
    }

    const addAssetButton = root.document.getElementById("add-asset-button");
    if (addAssetButton) {
      addAssetButton.addEventListener("click", async () => {
        const select = root.document.getElementById("available-asset-select");
        if (!select || !select.value) {
          state.editError = "Select an asset before adding it to the snapshot.";
          renderEditPage(app);
          return;
        }

        await runTransition(async () => {
          void logHomeAction("add_asset_to_snapshot_click", state.offset);
          state.editError = "";
          addAssetToDraft(Number(select.value));
          markEditDirty();
          renderEditPage(app);
        });
      });
    }

    const createAssetButton = root.document.getElementById("create-asset-button");
    if (createAssetButton) {
      createAssetButton.addEventListener("click", async () => {
        await snapshotAssetModal.openSnapshotAssetModal("asset", app, renderEditPage);
      });
    }

    const createAssetTypeButton = root.document.getElementById("create-asset-type-button");
    if (createAssetTypeButton) {
      createAssetTypeButton.addEventListener("click", async () => {
        await snapshotAssetModal.openSnapshotAssetModal("asset_type", app, renderEditPage);
      });
    }

    snapshotAssetModal.bindSnapshotAssetModal({
      app,
      renderEditPage,
      addCreatedAssetToDraft,
      markEditDirty,
    });

    const cancelButton = root.document.getElementById("edit-cancel-button");
    if (cancelButton) {
      cancelButton.addEventListener("click", async () => {
        await attemptCancelEditPage(app);
      });
    }

    const saveButton = root.document.getElementById("save-snapshot-button");
    if (saveButton) {
      saveButton.addEventListener("click", async () => {
        await runTransition(async () => {
          state.editError = "";
          const backend = shared.resolveBackend();
          const isNewMode = state.editDraft?.Mode === "new";
          void logHomeAction(isNewMode ? "create_snapshot_click" : "save_snapshot_click", state.offset);

          try {
            const result = isNewMode
              ? await backend.CreateSnapshot(buildCreateSnapshotInput(state.editDraft))
              : await backend.SaveSnapshot(buildSaveSnapshotInput(state.editDraft));
            state.offset = result.Offset;
            state.editDraft = null;
            state.removedRows = [];
            state.hasUnsavedChanges = false;
            await app.loadHomePage();
          } catch (error) {
            state.editError = error?.message || String(error);
            renderEditPage(app);
          }
        });
      });
    }
  }

  function handleEditKeydown(event, app) {
    if (snapshotAssetModal.shouldCloseSnapshotAssetModalOnEscape({
      key: event.key,
      hasSnapshotAssetModal: Boolean(state.snapshotAssetModal),
    })) {
      event.preventDefault();
      state.snapshotAssetModal = null;
      renderEditPage(app);
      return true;
    }

    const target = event.target;
    const navElement = target instanceof root.Element ? target.closest("[data-asset-id][data-nav-field]") : null;

    if (shouldDiscardEditPageOnEscape({
      key: event.key,
      isTypingTarget: isGeneralTypingTarget(target),
      isManualFieldEditing: isManualFieldEditing(target),
      isCurrentPriceTarget: Boolean(target instanceof root.Element && target.classList.contains("edit-current-input")),
    })) {
      event.preventDefault();
      void attemptCancelEditPage(app);
      return true;
    }

    if (!navElement) {
      return false;
    }

    const row = findEditRow(Number(navElement.dataset.assetId));
    if (!row) {
      return false;
    }

    const isNotesEditable = navElement.dataset.navField === "notes" && navElement.readOnly === false;
    if (shouldDeactivateManualFieldOnKey({
      key: event.key,
      isManualField: navElement.classList.contains("manual-field"),
      isReadOnly: navElement.readOnly,
    })) {
      event.preventDefault();
      deactivateManualField(navElement, { keepFocus: true });
      return true;
    }

    if (shouldActivateManualFieldOnEnter({
      key: event.key,
      isManualField: navElement.classList.contains("manual-field"),
      isDisabled: Boolean(navElement.disabled),
      isReadOnly: navElement.readOnly,
    })) {
      event.preventDefault();
      logic.activateManualField(navElement);
      return true;
    }

    const verticalDirection = resolveEditVerticalDirection({ key: event.key, isNotesEditable });
    if (verticalDirection !== null) {
      event.preventDefault();
      const targetAssetID = findVerticalTargetAssetID(
        state.editDraft,
        row.AssetID,
        navElement.dataset.navField,
        verticalDirection,
      );
      if (targetAssetID !== null) {
        focusEditField(targetAssetID, navElement.dataset.navField);
      }
      return true;
    }

    const nextField = resolveEditHorizontalMove({
      key: event.key,
      fieldKey: navElement.dataset.navField,
      isCash: row.IsCash,
      isNotesEditable,
    });
    if (!nextField) {
      return false;
    }

    event.preventDefault();
    focusEditField(row.AssetID, nextField);
    return true;
  }

  function buildSaveSnapshotInput(draft) {
    return {
      SnapshotID: draft.SnapshotID,
      SnapshotDate: draft.SnapshotDate,
      Items: buildSnapshotItemInputs(draft),
    };
  }

  function buildCreateSnapshotInput(draft) {
    return {
      SnapshotDate: draft.SnapshotDate,
      Items: buildSnapshotItemInputs(draft),
    };
  }

  function buildSnapshotItemInputs(draft) {
    return draft.Groups.flatMap((group) => group.Rows).map((row) => ({
      AssetID: row.AssetID,
      BoughtPrice: row.IsCash ? 0 : row.BoughtPrice,
      CurrentPrice: row.CurrentPrice,
      Remarks: row.Remarks || "",
    }));
  }

  function addAssetToDraft(assetID) {
    const { option, groupIndex, assetTypeName } = takeAvailableAssetOption(assetID);
    if (!option) {
      state.editError = "The selected asset is no longer available.";
      return;
    }

    const row = {
      AssetTypeOrdering: option.AssetTypeOrdering,
      AssetOrdering: option.AssetOrdering,
      AssetID: option.AssetID,
      AssetName: option.AssetName,
      Broker: option.Broker,
      IsCash: option.IsCash,
      IsActive: option.IsActive,
      BoughtPrice: 0,
      CurrentPrice: 0,
      Remarks: "",
    };

    removeEmptyAvailableGroup(groupIndex);
    insertRowIntoDraft(row, assetTypeName);
    removeFromRemovedRows(assetID);
  }

  function takeAvailableAssetOption(assetID) {
    const groups = state.editDraft?.AvailableAssetGroups || [];
    for (let groupIndex = 0; groupIndex < groups.length; groupIndex += 1) {
      const optionIndex = groups[groupIndex].Options.findIndex((option) => option.AssetID === assetID);
      if (optionIndex >= 0) {
        const [option] = groups[groupIndex].Options.splice(optionIndex, 1);
        return { option, groupIndex, assetTypeName: groups[groupIndex].AssetTypeName };
      }
    }
    return { option: null, groupIndex: -1, assetTypeName: "Other" };
  }

  function removeEmptyAvailableGroup(groupIndex) {
    if (groupIndex < 0) {
      return;
    }
    const group = state.editDraft.AvailableAssetGroups[groupIndex];
    if (group && group.Options.length === 0) {
      state.editDraft.AvailableAssetGroups.splice(groupIndex, 1);
    }
  }

  function insertRowIntoDraft(row, assetTypeName) {
    const groups = state.editDraft.Groups;
    const existingGroup = groups.find((group) => group.AssetTypeName === assetTypeName);
    if (existingGroup) {
      existingGroup.Rows.push(row);
    } else {
      groups.push({ AssetTypeName: assetTypeName, Rows: [row] });
    }

    sortEditDraft(state.editDraft);
  }

  function removeAssetFromDraft(assetID) {
    const groups = state.editDraft?.Groups || [];
    for (let groupIndex = 0; groupIndex < groups.length; groupIndex += 1) {
      const rowIndex = groups[groupIndex].Rows.findIndex((row) => row.AssetID === assetID);
      if (rowIndex < 0) {
        continue;
      }

      const assetTypeName = groups[groupIndex].AssetTypeName;
      const [row] = groups[groupIndex].Rows.splice(rowIndex, 1);
      if (groups[groupIndex].Rows.length === 0) {
        groups.splice(groupIndex, 1);
      }

      insertAvailableOption({
        AssetTypeOrdering: row.AssetTypeOrdering,
        AssetOrdering: row.AssetOrdering,
        AssetID: row.AssetID,
        AssetName: row.AssetName,
        Broker: row.Broker,
        IsCash: row.IsCash,
        IsActive: row.IsActive,
      }, assetTypeName);

      state.removedRows = [{ ...row, AssetTypeName: assetTypeName }, ...state.removedRows.filter((item) => item.AssetID !== row.AssetID)];
      sortEditDraft(state.editDraft);
      return;
    }
  }

  function undoRemovedAsset(assetID) {
    const removedIndex = state.removedRows.findIndex((item) => item.AssetID === assetID);
    if (removedIndex < 0) {
      return;
    }

    const [row] = state.removedRows.splice(removedIndex, 1);
    const availableGroups = state.editDraft.AvailableAssetGroups || [];
    for (let groupIndex = 0; groupIndex < availableGroups.length; groupIndex += 1) {
      const optionIndex = availableGroups[groupIndex].Options.findIndex((option) => option.AssetID === assetID);
      if (optionIndex >= 0) {
        availableGroups[groupIndex].Options.splice(optionIndex, 1);
        removeEmptyAvailableGroup(groupIndex);
        break;
      }
    }

    insertRowIntoDraft({
      AssetTypeOrdering: row.AssetTypeOrdering,
      AssetOrdering: row.AssetOrdering,
      AssetID: row.AssetID,
      AssetName: row.AssetName,
      Broker: row.Broker,
      IsCash: row.IsCash,
      IsActive: row.IsActive,
      BoughtPrice: row.BoughtPrice,
      CurrentPrice: row.CurrentPrice,
      Remarks: row.Remarks,
    }, row.AssetTypeName);
  }

  function insertAvailableOption(option, assetTypeName) {
    const groups = state.editDraft.AvailableAssetGroups;
    let group = groups.find((candidate) => candidate.AssetTypeName === assetTypeName);
    if (!group) {
      group = { AssetTypeName: assetTypeName, Options: [] };
      groups.push(group);
    }
    group.Options.push(option);
  }

  function removeFromRemovedRows(assetID) {
    state.removedRows = state.removedRows.filter((item) => item.AssetID !== assetID);
  }

  function markEditDirty() {
    sortEditDraft(state.editDraft);
    state.hasUnsavedChanges = createEditSignature(state.editDraft) !== state.editOriginalSignature;
  }

  function createEditSignature(draft) {
    if (!draft) {
      return "";
    }

    return JSON.stringify({
      SnapshotID: draft.SnapshotID,
      SnapshotDate: draft.SnapshotDate,
      Groups: draft.Groups.map((group) => ({
        AssetTypeName: group.AssetTypeName,
        Rows: group.Rows.map((row) => ({
          AssetTypeOrdering: row.AssetTypeOrdering,
          AssetOrdering: row.AssetOrdering,
          AssetID: row.AssetID,
          BoughtPrice: row.BoughtPrice,
          CurrentPrice: row.CurrentPrice,
          Remarks: row.Remarks,
        })),
      })),
      AvailableAssetGroups: draft.AvailableAssetGroups.map((group) => ({
        AssetTypeName: group.AssetTypeName,
        Options: group.Options.map((option) => option.AssetID),
      })),
    });
  }

  function sortEditDraft(draft) {
    if (!draft) {
      return;
    }

    draft.Groups = (draft.Groups || []).filter((group) => group.Rows.length > 0).sort(compareEditGroups);
    for (const group of draft.Groups) {
      group.Rows.sort(compareEditRows);
    }

    draft.AvailableAssetGroups = (draft.AvailableAssetGroups || []).filter((group) => group.Options.length > 0).sort(compareAvailableGroups);
    for (const group of draft.AvailableAssetGroups) {
      group.Options.sort(compareAvailableOptions);
    }
  }

  function compareEditGroups(left, right) {
    return compareNumbers(firstGroupOrder(left), firstGroupOrder(right)) || left.AssetTypeName.localeCompare(right.AssetTypeName);
  }

  function compareAvailableGroups(left, right) {
    return compareNumbers(firstAvailableGroupOrder(left), firstAvailableGroupOrder(right)) || left.AssetTypeName.localeCompare(right.AssetTypeName);
  }

  function compareEditRows(left, right) {
    return compareNumbers(left.AssetOrdering, right.AssetOrdering) || left.AssetName.localeCompare(right.AssetName);
  }

  function compareAvailableOptions(left, right) {
    return compareNumbers(left.AssetOrdering, right.AssetOrdering) || left.AssetName.localeCompare(right.AssetName);
  }

  function firstGroupOrder(group) {
    return group.Rows[0]?.AssetTypeOrdering ?? Number.MAX_SAFE_INTEGER;
  }

  function firstAvailableGroupOrder(group) {
    return group.Options[0]?.AssetTypeOrdering ?? Number.MAX_SAFE_INTEGER;
  }

  function compareNumbers(left, right) {
    return left - right;
  }

  async function attemptCancelEditPage(app) {
    if (!confirmLeaveEditMode()) {
      return;
    }

    await runTransition(async () => {
      void logHomeAction("cancel_snapshot_form_click", state.offset);
      state.editDraft = null;
      state.editError = "";
      state.snapshotAssetModal = null;
      state.removedRows = [];
      state.hasUnsavedChanges = false;
      await app.loadHomePage();
    });
  }

  function confirmLeaveEditMode() {
    return !state.hasUnsavedChanges || root.confirm("You have unsaved changes. Leave without saving?");
  }

  function renderAvailableAssetSelector(groups) {
    if (!groups || groups.length === 0) {
      return controls.renderSelectControl({
        id: "available-asset-select",
        value: "",
        disabled: true,
        tabIndex: -1,
        placeholder: "No additional assets",
        ariaLabel: "Available assets",
        options: [],
      });
    }

    return controls.renderSelectControl({
      id: "available-asset-select",
      value: "",
      ariaLabel: "Available assets",
      tabIndex: -1,
      placeholder: "Select asset...",
      options: groups.flatMap((group) => group.Options.map((option) => ({
        value: option.AssetID,
        label: `${option.AssetName}${option.Broker ? ` · ${option.Broker}` : ""}${option.IsActive ? "" : " · Inactive"} (${group.AssetTypeName})`,
      }))),
    });
  }

  function renderRemovedRowsUndo(removedRows) {
    if (!removedRows || removedRows.length === 0) {
      return "";
    }

    return `
      <div class="toolbar-divider" aria-hidden="true"></div>
      <div class="removed-rows-strip">
        <span class="field-label">Recently Removed</span>
        <div class="removed-rows-actions">
          ${removedRows.map((item) => `
            <button class="button button-small undo-remove-button" type="button" data-asset-id="${item.AssetID}" tabindex="-1">
              Undo ${escapeHTML(item.AssetName)}
            </button>
          `).join("")}
        </div>
      </div>
    `;
  }

  function hasAvailableAssets(draft) {
    return (draft?.AvailableAssetGroups || []).some((group) => group.Options.length > 0);
  }

  function addCreatedAssetToDraft(modal, assetID) {
    const typeMeta = (modal.page.AssetTypes || []).find((item) => item.ID === modal.assetForm.assetTypeID);
    const groupName = typeMeta?.Name || "Other";
    const existingGroup = (state.editDraft?.Groups || []).find((item) => item.AssetTypeName === groupName);
    const row = {
      AssetTypeOrdering: existingGroup?.Rows?.[0]?.AssetTypeOrdering ?? nextAssetTypeOrdering(),
      AssetOrdering: nextAssetOrdering(groupName),
      AssetID: assetID,
      AssetName: modal.assetForm.name.trim(),
      Broker: modal.assetForm.broker.trim(),
      IsCash: modal.assetForm.isCash,
      IsActive: modal.assetForm.isActive,
      BoughtPrice: 0,
      CurrentPrice: 0,
      Remarks: "",
    };
    insertRowIntoDraft(row, groupName);
  }

  function nextAssetTypeOrdering() {
    const values = (state.editDraft?.Groups || [])
      .flatMap((group) => group.Rows)
      .map((row) => row.AssetTypeOrdering);
    return values.length === 0 ? 1 : Math.max(...values) + 1;
  }

  function nextAssetOrdering(groupName) {
    const group = (state.editDraft?.Groups || []).find((item) => item.AssetTypeName === groupName);
    if (!group || group.Rows.length === 0) {
      return 1;
    }

    return Math.max(...group.Rows.map((row) => row.AssetOrdering)) + 1;
  }

  function renderEditGroup(group) {
    return `
      <section class="panel table-panel">
        <div class="table-header">
          <h2>${escapeHTML(group.AssetTypeName)}</h2>
          <span>${group.Rows.length} asset(s)</span>
        </div>
        <div class="table-scroll">
          <table>
            <colgroup>
              <col class="col-name" />
              <col class="col-broker" />
              <col class="col-bought" />
              <col class="col-current" />
              <col class="col-profit" />
              <col class="col-profit-percent" />
              <col class="col-notes" />
              <col class="col-actions" />
            </colgroup>
            <thead>
              <tr>
                <th>Name</th>
                <th>Broker</th>
                <th class="numeric">Bought Price</th>
                <th class="numeric">Current Price</th>
                <th class="numeric">% Profit</th>
                <th class="numeric">Profit</th>
                <th>Notes</th>
                <th class="numeric">Action</th>
              </tr>
            </thead>
            <tbody>
              ${group.Rows.map(renderEditRow).join("")}
            </tbody>
          </table>
        </div>
      </section>
    `;
  }

  function renderEditRow(row) {
    const profit = calculateProfitFromRow(row);
    const profitPercent = calculateProfitPercentFromRow(row);
    const tone = row.IsCash
      ? "muted"
      : profit > 0
        ? "positive"
        : profit < 0
          ? "negative"
          : "neutral";

    return `
      <tr data-asset-id="${row.AssetID}">
        <td>
          <div class="asset-name-cell">
            <span>${escapeHTML(row.AssetName)}</span>
            ${row.IsActive ? "" : '<span class="status-badge">Inactive</span>'}
          </div>
        </td>
        <td>${escapeHTML(row.Broker || "-")}</td>
        <td class="numeric">
          <input
            class="form-input form-input-table numeric edit-bought-input manual-field ${row.IsCash ? "input-muted" : "manual-field-locked"}"
            type="text"
            inputmode="decimal"
            value="${row.IsCash ? "" : escapeHTML(formatEditableNumber(row.BoughtPrice))}"
            data-asset-id="${row.AssetID}"
            data-nav-field="bought"
            ${row.IsCash ? 'disabled placeholder="N/A"' : 'readonly tabindex="-1"'}
            spellcheck="false"
          />
        </td>
        <td class="numeric">
          <input
            class="form-input form-input-table numeric edit-current-input"
            type="text"
            inputmode="decimal"
            value="${escapeHTML(formatEditableNumber(row.CurrentPrice))}"
            data-asset-id="${row.AssetID}"
            data-nav-field="current"
            tabindex="${EDIT_TAB_INDEX.current}"
            spellcheck="false"
          />
        </td>
        <td id="profit-percent-${row.AssetID}" class="numeric ${tone}">${row.IsCash ? '<span class="muted">N/A</span>' : formatPercent(profitPercent)}</td>
        <td id="profit-value-${row.AssetID}" class="numeric ${tone}">${row.IsCash ? '<span class="muted">N/A</span>' : formatTHB(profit)}</td>
        <td>
          <input
            class="form-input form-input-table edit-remarks-input manual-field manual-field-locked"
            type="text"
            value="${escapeHTML(row.Remarks || "")}"
            data-asset-id="${row.AssetID}"
            data-nav-field="notes"
            readonly
            tabindex="-1"
          />
        </td>
        <td class="numeric">
          <button class="button button-small remove-row-button" type="button" data-asset-id="${row.AssetID}" data-nav-field="remove" tabindex="-1">Remove</button>
        </td>
      </tr>
    `;
  }

  return {
    EDIT_TAB_INDEX,
    buildCreateSnapshotInput,
    buildSaveSnapshotInput,
    handleEditKeydown,
    initializeDraft,
    isEditGridShortcutKey,
    isEditHorizontalShortcutKey,
    renderEditPage,
    shouldCloseSnapshotAssetModalOnEscape: snapshotAssetModal.shouldCloseSnapshotAssetModalOnEscape,
  };
}));
