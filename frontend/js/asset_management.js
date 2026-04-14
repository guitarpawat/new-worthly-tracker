(function initAssetManagement(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const controls = root.WorthlyControls || (typeof require !== "undefined" ? require("./controls.js") : null);
  const assetManagement = factory(shared, controls, root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = assetManagement;
  }
  root.WorthlyAssetManagement = assetManagement;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildAssetManagement(shared, controls, root) {
  const {
    escapeHTML,
    formatEditableNumber,
    formatTHB,
    isAllowedControlKey,
    isValidPartialDecimal,
    parseEditableNumber,
    previewNumericValue,
    renderAppTitle,
    renderErrorState,
    runTransition,
    sanitizePartialDecimal,
    state,
  } = shared;

  function resolveAssetReorder() {
    const module = root.WorthlyAssetReorder || (typeof require !== "undefined" ? require("./asset_reorder.js") : null);
    if (!module) {
      throw new Error("Asset reorder module is not available");
    }
    return module;
  }

  function renderAssetManagementPage(app) {
    const appRoot = root.document.getElementById("app");
    const page = state.assetManagementPage;
    const assetTypeForm = state.assetTypeForm;
    const assetForm = state.assetForm;
    if (!page || !assetTypeForm || !assetForm) {
      appRoot.innerHTML = renderErrorState("Unable to open asset management", "Asset management state is missing.");
      return;
    }

    const view = state.assetManagementView || "create_asset";
    appRoot.innerHTML = `
      <main class="app-layout asset-management-layout ${state.assetManagementModal ? "app-modal-open" : ""}">
        <section class="hero">
          <div>
            <h1>${renderAppTitle("Manage Asset")}</h1>
          </div>
        </section>
        ${renderAssetManagementSubnav(view, page)}
        ${renderAssetManagementContent(view, page, assetTypeForm, assetForm)}
      </main>
      ${renderAssetManagementModal(page, assetTypeForm, assetForm)}
    `;

    bindAssetManagementPage(app);
  }

  function bindAssetManagementPage(app) {
    for (const button of root.document.querySelectorAll("[data-asset-management-view]")) {
      button.addEventListener("click", (event) => {
        const nextView = event.currentTarget.dataset.assetManagementView;
        if (!nextView || event.currentTarget.disabled) {
          return;
        }

        state.assetTypeError = "";
        state.assetError = "";
        state.assetManagementView = nextView;
        if (nextView === "create_asset") {
          state.assetForm = buildEmptyAssetForm(state.assetManagementPage);
        }
        if (nextView === "create_asset_type") {
          state.assetTypeForm = buildEmptyAssetTypeForm();
        }
        if (nextView === "edit_asset") {
          state.assetForm = buildEmptyAssetForm(state.assetManagementPage);
          state.assetManagementModal = null;
        }
        if (nextView === "edit_asset_type") {
          state.assetTypeForm = buildEmptyAssetTypeForm();
          state.assetManagementModal = null;
        }
        if (nextView === "reorder_asset" || nextView === "reorder_asset_type") {
          state.assetManagementModal = null;
        }
        renderAssetManagementPage(app);
      });
    }

    if (state.assetManagementView === "create_asset_type" || state.assetManagementView === "edit_asset_type") {
      bindAssetTypeForm(app);
    }
    if (state.assetManagementView === "create_asset" || state.assetManagementView === "edit_asset") {
      bindAssetForm(app);
    }
    if (state.assetManagementView === "reorder_asset_type") {
      resolveAssetReorder().bindAssetTypeReorderPage(app);
    }
    if (state.assetManagementView === "reorder_asset") {
      resolveAssetReorder().bindAssetReorderPage(app);
    }

    bindAssetManagementModal(app);

    for (const row of root.document.querySelectorAll("[data-asset-type-row-id]")) {
      row.addEventListener("click", () => {
        state.assetTypeForm = buildAssetTypeFormState(state.assetManagementPage, Number(row.dataset.assetTypeRowId));
        state.assetTypeError = "";
        state.assetManagementView = "edit_asset_type";
        state.assetManagementModal = { kind: "asset_type" };
        renderAssetManagementPage(app);
      });
    }

    for (const row of root.document.querySelectorAll("[data-asset-row-id]")) {
      row.addEventListener("click", () => {
        state.assetForm = buildAssetFormState(state.assetManagementPage, Number(row.dataset.assetRowId));
        state.assetError = "";
        state.assetManagementView = "edit_asset";
        state.assetManagementModal = { kind: "asset" };
        renderAssetManagementPage(app);
      });
    }
  }

  function bindAssetTypeForm(app) {
    const nameInput = root.document.getElementById("asset-type-name-input");
    if (nameInput) {
      nameInput.addEventListener("input", (event) => {
        state.assetTypeForm.name = event.target.value;
      });
    }

    const activeInput = root.document.getElementById("asset-type-active-input");
    if (activeInput) {
      activeInput.addEventListener("change", (event) => {
        state.assetTypeForm.isActive = event.target.checked;
      });
    }

    const resetButton = root.document.getElementById("asset-type-reset-button");
    if (resetButton) {
      resetButton.addEventListener("click", async () => {
        await runTransition(async () => {
          if (state.assetManagementView === "create_asset_type") {
            state.assetTypeForm = buildEmptyAssetTypeForm();
            state.assetTypeError = "";
            renderAssetManagementPage(app);
            return;
          }

          state.assetTypeForm = buildEmptyAssetTypeForm();
          state.assetTypeError = "";
          state.assetManagementModal = null;
          renderAssetManagementPage(app);
        });
      });
    }

    const saveButton = root.document.getElementById("asset-type-save-button");
    if (saveButton) {
      saveButton.addEventListener("click", async () => {
        await runTransition(async () => {
          state.assetTypeError = "";
          try {
            const backend = shared.resolveBackend();
            const isEditMode = state.assetTypeForm.id > 0;
            const result = state.assetTypeForm.id > 0
              ? await backend.UpdateAssetType(buildAssetTypeUpdatePayload(state.assetTypeForm))
              : await backend.CreateAssetType(buildAssetTypeCreatePayload(state.assetTypeForm));
            await app.loadAssetManagementPage(
              isEditMode
                ? { view: "edit_asset_type", selectedAssetTypeID: result.ID }
                : { view: "create_asset_type" },
            );
          } catch (error) {
            state.assetTypeError = error?.message || String(error);
            renderAssetManagementPage(app);
          }
        });
      });
    }
  }

  function bindAssetForm(app) {
    const nameInput = root.document.getElementById("asset-name-input");
    if (nameInput) {
      nameInput.addEventListener("input", (event) => {
        state.assetForm.name = event.target.value;
      });
    }

    const typeSelect = root.document.getElementById("asset-type-select");
    if (typeSelect) {
      typeSelect.addEventListener("change", (event) => {
        state.assetForm.assetTypeID = Number(event.target.value);
      });
    }

    const brokerInput = root.document.getElementById("asset-broker-input");
    if (brokerInput) {
      brokerInput.addEventListener("input", (event) => {
        state.assetForm.broker = event.target.value;
      });
    }

    const cashInput = root.document.getElementById("asset-is-cash-input");
    if (cashInput) {
      cashInput.addEventListener("change", (event) => {
        state.assetForm.isCash = event.target.checked;
        if (state.assetForm.isCash) {
          state.assetForm.autoIncrement = "0.00";
        }
        renderAssetManagementPage(app);
      });
    }

    const activeInput = root.document.getElementById("asset-is-active-input");
    if (activeInput) {
      activeInput.addEventListener("change", (event) => {
        state.assetForm.isActive = event.target.checked;
      });
    }

    const autoIncrementInput = root.document.getElementById("asset-auto-increment-input");
    if (autoIncrementInput) {
      bindStandaloneDecimalInput(autoIncrementInput, {
        onChange: (value) => {
          state.assetForm.autoIncrement = value;
        },
      });
    }

    const resetButton = root.document.getElementById("asset-reset-button");
    if (resetButton) {
      resetButton.addEventListener("click", async () => {
        await runTransition(async () => {
          if (state.assetManagementView === "create_asset") {
            state.assetForm = buildEmptyAssetForm(state.assetManagementPage);
            state.assetError = "";
            renderAssetManagementPage(app);
            return;
          }

          state.assetForm = buildEmptyAssetForm(state.assetManagementPage);
          state.assetError = "";
          state.assetManagementModal = null;
          renderAssetManagementPage(app);
        });
      });
    }

    const saveButton = root.document.getElementById("asset-save-button");
    if (saveButton) {
      saveButton.addEventListener("click", async () => {
        await runTransition(async () => {
          state.assetError = "";
          try {
            const backend = shared.resolveBackend();
            const isEditMode = state.assetForm.id > 0;
            const result = state.assetForm.id > 0
              ? await backend.UpdateAsset(buildAssetUpdatePayload(state.assetForm))
              : await backend.CreateAsset(buildAssetCreatePayload(state.assetForm));
            await app.loadAssetManagementPage(
              isEditMode
                ? { view: "edit_asset", selectedAssetID: result.ID }
                : { view: "create_asset" },
            );
          } catch (error) {
            state.assetError = error?.message || String(error);
            renderAssetManagementPage(app);
          }
        });
      });
    }
  }

  function resolveAssetManagementView(options = {}) {
    if (options.view) {
      return options.view;
    }
    if (options.selectedAssetID) {
      return "edit_asset";
    }
    if (options.selectedAssetTypeID) {
      return "edit_asset_type";
    }
    return state.assetManagementView || "create_asset";
  }

  function renderAssetManagementSubnav(view, page) {
    const actions = [
      { id: "create_asset", label: "Add New Asset", disabled: false },
      { id: "create_asset_type", label: "Add New Asset Type", disabled: false },
      { id: "edit_asset", label: "Edit Asset", disabled: (page?.Assets?.length || 0) === 0 },
      { id: "edit_asset_type", label: "Edit Asset Type", disabled: (page?.AssetTypes?.length || 0) === 0 },
      { id: "reorder_asset", label: "Reorder Asset", disabled: (page?.Assets?.length || 0) === 0 },
      { id: "reorder_asset_type", label: "Reorder Asset Type", disabled: (page?.AssetTypes?.length || 0) === 0 },
    ];

    return `
      <section class="panel asset-management-subnav">
        ${actions.map((action) => `
          <button
            class="button asset-management-subnav-button ${view === action.id ? "asset-management-subnav-button-active" : ""}"
            type="button"
            data-asset-management-view="${action.id}"
            ${action.disabled ? "disabled" : ""}
          >
            ${escapeHTML(action.label)}
          </button>
        `).join("")}
      </section>
    `;
  }

  function renderAssetManagementContent(view, page, assetTypeForm, assetForm) {
    switch (view) {
      case "create_asset":
        return `
          <section class="asset-management-centered-shell">
            ${renderAssetEditorCard(page, assetForm)}
          </section>
        `;
      case "edit_asset":
        return `
          <section class="asset-management-table-shell">
            ${renderAssetTable(page, assetForm)}
          </section>
        `;
      case "create_asset_type":
        return `
          <section class="asset-management-centered-shell">
            ${renderAssetTypeEditorCard(assetTypeForm)}
          </section>
        `;
      case "edit_asset_type":
        return `
          <section class="asset-management-table-shell">
            ${renderAssetTypeTable(page, assetTypeForm)}
          </section>
        `;
      case "reorder_asset":
        return resolveAssetReorder().renderAssetReorderPage();
      case "reorder_asset_type":
        return resolveAssetReorder().renderAssetTypeReorderPage();
      default:
        return `
          <section class="asset-management-centered-shell">
            ${renderAssetEditorCard(page, assetForm)}
          </section>
        `;
    }
  }

  function renderAssetManagementModal(page, assetTypeForm, assetForm) {
    if (!state.assetManagementModal) {
      return "";
    }

    const body = state.assetManagementModal.kind === "asset_type"
      ? renderAssetTypeEditorCard(assetTypeForm, { errorMessage: state.assetTypeError })
      : renderAssetEditorCard(page, assetForm, { errorMessage: state.assetError });

    return `
      <div class="dialog-backdrop asset-management-dialog-backdrop" data-asset-management-modal-close="backdrop">
        <section class="asset-management-dialog-shell" role="dialog" aria-modal="true">
          ${body}
        </section>
      </div>
    `;
  }

  function bindAssetManagementModal(app) {
    if (!state.assetManagementModal) {
      return;
    }

    for (const element of root.document.querySelectorAll("[data-asset-management-modal-close]")) {
      element.addEventListener("click", (event) => {
        if (event.target !== event.currentTarget) {
          return;
        }
        state.assetManagementModal = null;
        if (state.assetManagementView === "edit_asset") {
          state.assetForm = buildEmptyAssetForm(state.assetManagementPage);
        }
        if (state.assetManagementView === "edit_asset_type") {
          state.assetTypeForm = buildEmptyAssetTypeForm();
        }
        renderAssetManagementPage(app);
      });
    }
  }

  function handleAssetManagementKeydown(event, app) {
    if (!shouldCloseAssetManagementModalOnEscape({
      key: event.key,
      hasAssetManagementModal: Boolean(state.assetManagementModal),
    })) {
      return false;
    }

    event.preventDefault();
    state.assetManagementModal = null;
    if (state.assetManagementView === "edit_asset") {
      state.assetForm = buildEmptyAssetForm(state.assetManagementPage);
    }
    if (state.assetManagementView === "edit_asset_type") {
      state.assetTypeForm = buildEmptyAssetTypeForm();
    }
    renderAssetManagementPage(app);
    return true;
  }

  function shouldCloseAssetManagementModalOnEscape(options) {
    return options.key === "Escape" && options.hasAssetManagementModal;
  }

  function renderAssetTypeEditorCard(assetTypeForm, options = {}) {
    const errorMessage = options.errorMessage ?? state.assetTypeError;
    const secondaryButtonLabel = options.secondaryButtonLabel || (assetTypeForm.id > 0 ? "Close" : "Reset");
    return `
      <section class="panel asset-management-card asset-management-centered-card">
        <div class="table-header">
          <h2>Asset Type Form</h2>
        </div>
        <div class="asset-management-form">
          <label class="field-inline">
            <span class="field-label">Type Name</span>
            <input id="asset-type-name-input" class="form-input" type="text" value="${escapeHTML(assetTypeForm.name)}" />
          </label>
          ${errorMessage ? `<p class="error-copy error-banner">${escapeHTML(errorMessage)}</p>` : ""}
          <div class="asset-management-form-footer">
            <div class="asset-management-form-footer-left">
              <label class="field-inline field-inline-toggle">
                <span class="field-label">Active</span>
                ${renderInteractiveToggle("asset-type-active-input", assetTypeForm.isActive)}
              </label>
            </div>
            <div class="actions asset-management-card-actions">
              <button id="asset-type-save-button" class="button button-primary" type="button">${assetTypeForm.id > 0 ? "Save Type" : "Create Type"}</button>
              <button id="asset-type-reset-button" class="button" type="button">${secondaryButtonLabel}</button>
            </div>
          </div>
        </div>
      </section>
    `;
  }

  function renderAssetEditorCard(page, assetForm, options = {}) {
    const errorMessage = options.errorMessage ?? state.assetError;
    const secondaryButtonLabel = options.secondaryButtonLabel || (assetForm.id > 0 ? "Close" : "Reset");
    const assetTypeOptions = buildAssetTypeSelectOptions(page, assetForm.assetTypeID);
    return `
      <section class="panel asset-management-card asset-management-centered-card">
        <div class="table-header">
          <h2>Asset Form</h2>
        </div>
        <div class="asset-management-form asset-management-form-two-col">
          <label class="field-inline">
            <span class="field-label">Asset Name</span>
            <input id="asset-name-input" class="form-input" type="text" value="${escapeHTML(assetForm.name)}" />
          </label>
          <label class="field-inline">
            <span class="field-label">Asset Type</span>
            ${controls.renderSelectControl({
              id: "asset-type-select",
              value: assetForm.assetTypeID,
              buttonClass: "snapshot-select asset-management-select",
              ariaLabel: "Asset type",
              options: assetTypeOptions.map((option) => ({
                value: option.id,
                label: option.label,
              })),
            })}
          </label>
          <label class="field-inline">
            <span class="field-label">Broker</span>
            <input id="asset-broker-input" class="form-input" type="text" value="${escapeHTML(assetForm.broker)}" />
          </label>
          <label class="field-inline">
            <span class="field-label">Auto Increment</span>
            <input id="asset-auto-increment-input" class="form-input numeric" type="text" inputmode="decimal" value="${escapeHTML(assetForm.autoIncrement)}" ${assetForm.isCash ? "disabled" : ""} />
          </label>
          ${errorMessage ? `<p class="error-copy error-banner asset-management-error">${escapeHTML(errorMessage)}</p>` : ""}
          <div class="asset-management-form-footer asset-management-form-footer-wide">
            <div class="asset-management-toggle-row">
              <label class="field-inline field-inline-toggle">
                <span class="field-label">Cash Asset</span>
                ${renderInteractiveToggle("asset-is-cash-input", assetForm.isCash)}
              </label>
              <label class="field-inline field-inline-toggle">
                <span class="field-label">Active</span>
                ${renderInteractiveToggle("asset-is-active-input", assetForm.isActive)}
              </label>
            </div>
            <div class="actions asset-management-card-actions">
              <button id="asset-save-button" class="button button-primary" type="button">${assetForm.id > 0 ? "Save Asset" : "Create Asset"}</button>
              <button id="asset-reset-button" class="button" type="button">${secondaryButtonLabel}</button>
            </div>
          </div>
        </div>
      </section>
    `;
  }

  function renderAssetTypeTable(page, assetTypeForm) {
    return `
      <section class="panel table-panel">
        <div class="table-header">
          <h2>Asset Types</h2>
          <span>${page.AssetTypes.length} type(s)</span>
        </div>
        <div class="table-scroll">
          <table class="management-table asset-management-types-table">
            <thead>
              <tr>
                <th>Name</th>
                <th class="numeric">Assets</th>
                <th>Active</th>
              </tr>
            </thead>
            <tbody>
              ${page.AssetTypes.map((assetType) => renderAssetTypeTableRow(assetType, assetTypeForm)).join("")}
            </tbody>
          </table>
        </div>
      </section>
    `;
  }

  function renderAssetTable(page, assetForm) {
    return `
      <section class="panel table-panel">
        <div class="table-header">
          <h2>Assets</h2>
          <span>${page.Assets.length} asset(s)</span>
        </div>
        <div class="table-scroll">
          <table class="management-table asset-management-assets-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Type</th>
                <th>Broker</th>
                <th>Cash</th>
                <th>Active</th>
                <th class="numeric">Auto Increment</th>
              </tr>
            </thead>
            <tbody>
              ${page.Assets.map((asset) => renderAssetTableRow(page, asset, assetForm)).join("")}
            </tbody>
          </table>
        </div>
      </section>
    `;
  }

  function renderAssetTypeTableRow(assetType, assetTypeForm) {
    return `
      <tr class="${assetTypeForm.id === assetType.ID ? "management-row-active" : ""}" data-asset-type-row-id="${assetType.ID}">
        <td>${escapeHTML(assetType.Name)}</td>
        <td class="numeric">${assetType.AssetCount}</td>
        <td>${renderStatusPill(assetType.IsActive)}</td>
      </tr>
    `;
  }

  function renderAssetTableRow(page, asset, assetForm) {
    return `
      <tr class="${assetForm.id === asset.ID ? "management-row-active" : ""}" data-asset-row-id="${asset.ID}">
        <td>${escapeHTML(asset.Name)}</td>
        <td>${escapeHTML(asset.AssetTypeName)}</td>
        <td>${escapeHTML(asset.Broker)}</td>
        <td>${renderStatusPill(asset.IsCash)}</td>
        <td>${renderStatusPill(asset.IsActive)}</td>
        <td class="numeric">${formatTHB(asset.AutoIncrement)}</td>
      </tr>
    `;
  }

  function buildEmptyAssetTypeForm() {
    return {
      id: 0,
      name: "",
      isActive: true,
    };
  }

  function buildAssetTypeFormState(page, selectedID) {
    const row = (page?.AssetTypes || []).find((item) => item.ID === selectedID);
    if (!row) {
      return buildEmptyAssetTypeForm();
    }

    return {
      id: row.ID,
      name: row.Name,
      isActive: row.IsActive,
    };
  }

  function buildEmptyAssetForm(page) {
    return {
      id: 0,
      name: "",
      assetTypeID: page?.ActiveAssetTypes?.[0]?.ID || 0,
      broker: "",
      isCash: false,
      isActive: true,
      autoIncrement: "0.00",
    };
  }

  function buildAssetFormState(page, selectedID) {
    const row = (page?.Assets || []).find((item) => item.ID === selectedID);
    if (!row) {
      return buildEmptyAssetForm(page);
    }

    return {
      id: row.ID,
      name: row.Name,
      assetTypeID: row.AssetTypeID,
      broker: row.Broker || "",
      isCash: row.IsCash,
      isActive: row.IsActive,
      autoIncrement: formatEditableNumber(row.AutoIncrement),
    };
  }

  function buildAssetTypeCreatePayload(form) {
    return {
      Name: form.name,
      IsActive: form.isActive,
    };
  }

  function buildAssetTypeUpdatePayload(form) {
    return {
      ID: form.id,
      Name: form.name,
      IsActive: form.isActive,
    };
  }

  function buildAssetCreatePayload(form) {
    return {
      Name: form.name,
      AssetTypeID: form.assetTypeID,
      Broker: form.broker,
      IsCash: form.isCash,
      IsActive: form.isActive,
      AutoIncrement: form.isCash ? 0 : parseEditableNumber(form.autoIncrement),
    };
  }

  function buildAssetUpdatePayload(form) {
    return {
      ID: form.id,
      Name: form.name,
      AssetTypeID: form.assetTypeID,
      Broker: form.broker,
      IsCash: form.isCash,
      IsActive: form.isActive,
      AutoIncrement: form.isCash ? 0 : parseEditableNumber(form.autoIncrement),
    };
  }

  function buildAssetTypeSelectOptions(page, selectedID) {
    const options = (page?.ActiveAssetTypes || []).map((item) => ({
      id: item.ID,
      label: item.Name,
    }));

    if (selectedID <= 0 || options.some((item) => item.id === selectedID)) {
      return options;
    }

    const inactiveSelected = (page?.AssetTypes || []).find((item) => item.ID === selectedID);
    if (!inactiveSelected) {
      return options;
    }

    return [
      ...options,
      {
        id: inactiveSelected.ID,
        label: `${inactiveSelected.Name} (Inactive)`,
      },
    ];
  }

  function renderStatusPill(isOn) {
    return `
      <span class="metric-delta asset-management-status-pill ${isOn ? "delta-positive" : "delta-negative"}">
        ${isOn ? "Yes" : "No"}
      </span>
    `;
  }

  function renderInteractiveToggle(id, isOn) {
    return `
      <label class="toggle-field" for="${id}">
        <input id="${id}" class="toggle-field-input" type="checkbox" ${isOn ? "checked" : ""} />
        <span class="mock-toggle" aria-hidden="true">
          <span class="mock-toggle-thumb"></span>
        </span>
      </label>
    `;
  }

  function bindStandaloneDecimalInput(input, config) {
    input.addEventListener("focus", () => {
      input.select();
    });

    input.addEventListener("keydown", (event) => {
      if (isAllowedControlKey(event) || event.ctrlKey || event.metaKey || event.altKey) {
        return;
      }

      if (event.key.length !== 1) {
        return;
      }

      const nextValue = previewNumericValue(input, event.key);
      if (!isValidPartialDecimal(nextValue)) {
        event.preventDefault();
      }
    });

    input.addEventListener("paste", (event) => {
      const clipboardText = event.clipboardData?.getData("text") || "";
      const nextValue = previewNumericValue(input, clipboardText);
      if (!isValidPartialDecimal(nextValue)) {
        event.preventDefault();
      }
    });

    input.addEventListener("input", (event) => {
      const sanitized = sanitizePartialDecimal(event.target.value);
      if (event.target.value !== sanitized) {
        event.target.value = sanitized;
      }
      config.onChange(sanitized);
    });

    input.addEventListener("blur", (event) => {
      const formatted = formatEditableNumber(parseEditableNumber(event.target.value));
      event.target.value = formatted;
      config.onChange(formatted);
    });
  }

  return {
    bindStandaloneDecimalInput,
    buildAssetCreatePayload,
    buildAssetFormState,
    buildAssetTypeFormState,
    buildAssetTypeSelectOptions,
    buildAssetTypeCreatePayload,
    buildEmptyAssetForm,
    buildEmptyAssetTypeForm,
    renderAssetEditorCard,
    renderAssetTable,
    renderAssetTypeEditorCard,
    renderAssetTypeTable,
    handleAssetManagementKeydown,
    renderAssetManagementPage,
    renderInteractiveToggle,
    resolveAssetManagementView,
    shouldCloseAssetManagementModalOnEscape,
  };
}));
