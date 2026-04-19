(function initAssetManagementUI(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const controls = root.WorthlyControls || (typeof require !== "undefined" ? require("./controls.js") : null);
  const assetManagementUI = factory(shared, controls, root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = assetManagementUI;
  }
  root.WorthlyAssetManagementUI = assetManagementUI;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildAssetManagementUI(shared, controls, root) {
  const {
    escapeHTML,
    formatEditableNumber,
    formatTHB,
    isAllowedControlKey,
    isValidPartialDecimal,
    parseEditableNumber,
    previewNumericValue,
    sanitizePartialDecimal,
    state,
  } = shared;

  function renderAssetTypeEditorCard(assetTypeForm, options = {}) {
    const errorMessage = options.errorMessage ?? state.assetTypeError;
    const secondaryButtonLabel = options.secondaryButtonLabel || (assetTypeForm.id > 0 ? "Close" : "Reset");
    const showActiveToggle = assetTypeForm.id > 0;
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
              ${showActiveToggle
                ? `
                  <label class="field-inline field-inline-toggle">
                    <span class="field-label">Active</span>
                    ${renderInteractiveToggle("asset-type-active-input", assetTypeForm.isActive)}
                  </label>
                `
                : ""}
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
    const showActiveToggle = assetForm.id > 0;
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
                <span class="field-label">Liability</span>
                ${renderInteractiveToggle("asset-is-liability-input", assetForm.isLiability)}
              </label>
              ${showActiveToggle
                ? `
                  <label class="field-inline field-inline-toggle">
                    <span class="field-label">Active</span>
                    ${renderInteractiveToggle("asset-is-active-input", assetForm.isActive)}
                  </label>
                `
                : ""}
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
    const rows = filterAssetTypes(page.AssetTypes || []);
    return `
      <section class="panel table-panel">
        <div class="table-header">
          <h2>Asset Types</h2>
          <span>${rows.length} of ${page.AssetTypes.length} type(s)</span>
        </div>
        <div class="asset-management-filter-bar">
          ${renderAssetTypeFilterControl()}
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
              ${rows.map((assetType) => renderAssetTypeTableRow(assetType, assetTypeForm)).join("")}
            </tbody>
          </table>
        </div>
      </section>
    `;
  }

  function renderAssetTable(page, assetForm) {
    const rows = filterAssets(page.Assets || []);
    return `
      <section class="panel table-panel">
        <div class="table-header">
          <h2>Assets</h2>
          <span>${rows.length} of ${page.Assets.length} asset(s)</span>
        </div>
        <div class="asset-management-filter-bar asset-management-filter-bar-wide">
          ${renderAssetFilterControl("asset-filter-active", "Active", state.assetManagementFilters.assetActive, [
            { value: "all", label: "All" },
            { value: "active", label: "Active" },
            { value: "inactive", label: "Inactive" },
          ])}
          ${renderAssetFilterControl("asset-filter-cash", "Cash", state.assetManagementFilters.assetCash, [
            { value: "all", label: "All" },
            { value: "cash", label: "Cash" },
            { value: "non_cash", label: "Non Cash" },
          ])}
          ${renderAssetFilterControl("asset-filter-liability", "Liability", state.assetManagementFilters.assetLiability, [
            { value: "all", label: "All" },
            { value: "liability", label: "Liability" },
            { value: "non_liability", label: "Non Liability" },
          ])}
        </div>
        <div class="table-scroll">
          <table class="management-table asset-management-assets-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Type</th>
                <th>Broker</th>
                <th>Cash</th>
                <th>Liability</th>
                <th>Active</th>
                <th class="numeric">Auto Increment</th>
              </tr>
            </thead>
            <tbody>
              ${rows.map((asset) => renderAssetTableRow(asset, assetForm)).join("")}
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

  function renderAssetTableRow(asset, assetForm) {
    return `
      <tr class="${assetForm.id === asset.ID ? "management-row-active" : ""}" data-asset-row-id="${asset.ID}">
        <td>${escapeHTML(asset.Name)}</td>
        <td>${escapeHTML(asset.AssetTypeName)}</td>
        <td>${escapeHTML(asset.Broker)}</td>
        <td>${renderStatusPill(asset.IsCash)}</td>
        <td>${renderStatusPill(asset.IsLiability)}</td>
        <td>${renderStatusPill(asset.IsActive)}</td>
        <td class="numeric">${formatTHB(asset.AutoIncrement)}</td>
      </tr>
    `;
  }

  function buildEmptyAssetTypeForm() {
    return { id: 0, name: "", isActive: true };
  }

  function buildAssetTypeFormState(page, selectedID) {
    const row = (page?.AssetTypes || []).find((item) => item.ID === selectedID);
    if (!row) {
      return buildEmptyAssetTypeForm();
    }
    return { id: row.ID, name: row.Name, isActive: row.IsActive };
  }

  function buildEmptyAssetForm(page) {
    return {
      id: 0,
      name: "",
      assetTypeID: page?.ActiveAssetTypes?.[0]?.ID || 0,
      broker: "",
      isCash: false,
      isLiability: false,
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
      isLiability: row.IsLiability,
      isActive: row.IsActive,
      autoIncrement: formatEditableNumber(row.AutoIncrement),
    };
  }

  function buildAssetTypeCreatePayload(form) {
    return { Name: form.name, IsActive: form.isActive };
  }

  function buildAssetTypeUpdatePayload(form) {
    return { ID: form.id, Name: form.name, IsActive: form.isActive };
  }

  function buildAssetCreatePayload(form) {
    return {
      Name: form.name,
      AssetTypeID: form.assetTypeID,
      Broker: form.broker,
      IsCash: form.isCash,
      IsLiability: form.isLiability,
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
      IsLiability: form.isLiability,
      IsActive: form.isActive,
      AutoIncrement: form.isCash ? 0 : parseEditableNumber(form.autoIncrement),
    };
  }

  function renderAssetTypeFilterControl() {
    return renderAssetFilterControl("asset-type-filter-active", "Active", state.assetManagementFilters.assetTypeActive, [
      { value: "all", label: "All" },
      { value: "active", label: "Active" },
      { value: "inactive", label: "Inactive" },
    ]);
  }

  function renderAssetFilterControl(id, label, value, options) {
    return `
      <label class="field-inline asset-management-filter-field">
        <span class="field-label">${escapeHTML(label)}</span>
        ${controls.renderSelectControl({
          id,
          value,
          buttonClass: "snapshot-select asset-management-filter-select",
          ariaLabel: label,
          options,
        })}
      </label>
    `;
  }

  function filterAssetTypes(rows) {
    return rows.filter((row) => matchesActiveFilter(row.IsActive, state.assetManagementFilters.assetTypeActive));
  }

  function filterAssets(rows) {
    return rows.filter((row) => (
      matchesActiveFilter(row.IsActive, state.assetManagementFilters.assetActive)
      && matchesBooleanFilter(row.IsCash, state.assetManagementFilters.assetCash, "cash")
      && matchesBooleanFilter(row.IsLiability, state.assetManagementFilters.assetLiability, "liability")
    ));
  }

  function matchesActiveFilter(value, filter) {
    if (filter === "active") {
      return value;
    }
    if (filter === "inactive") {
      return !value;
    }
    return true;
  }

  function matchesBooleanFilter(value, filter, trueLabel) {
    if (filter === trueLabel) {
      return value;
    }
    if (filter === `non_${trueLabel}`) {
      return !value;
    }
    return true;
  }

  function buildAssetTypeSelectOptions(page, selectedID) {
    const options = (page?.ActiveAssetTypes || []).map((item) => ({ id: item.ID, label: item.Name }));
    if (selectedID <= 0 || options.some((item) => item.id === selectedID)) {
      return options;
    }
    const inactiveSelected = (page?.AssetTypes || []).find((item) => item.ID === selectedID);
    if (!inactiveSelected) {
      return options;
    }
    return [...options, { id: inactiveSelected.ID, label: `${inactiveSelected.Name} (Inactive)` }];
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
    buildAssetTypeCreatePayload,
    buildAssetTypeSelectOptions,
    buildAssetTypeUpdatePayload,
    buildAssetUpdatePayload,
    buildEmptyAssetForm,
    buildEmptyAssetTypeForm,
    renderAssetEditorCard,
    renderAssetTable,
    renderAssetTypeEditorCard,
    renderAssetTypeTable,
    renderInteractiveToggle,
  };
}));
