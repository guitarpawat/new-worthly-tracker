(function initAssetManagement(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const assetManagementUI = root.WorthlyAssetManagementUI || (typeof require !== "undefined" ? require("./asset_management_ui.js") : null);
  const assetManagement = factory(shared, assetManagementUI, root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = assetManagement;
  }
  root.WorthlyAssetManagement = assetManagement;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildAssetManagement(shared, ui, root) {
  const { escapeHTML, renderAppTitle, renderErrorState, runTransition, state } = shared;

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
    if (!page || !state.assetTypeForm || !state.assetForm) {
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
        ${renderAssetManagementContent(view, page)}
      </main>
      ${renderAssetManagementModal(page)}
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
          state.assetForm = ui.buildEmptyAssetForm(state.assetManagementPage);
        }
        if (nextView === "create_asset_type") {
          state.assetTypeForm = ui.buildEmptyAssetTypeForm();
        }
        if (nextView === "edit_asset") {
          state.assetForm = ui.buildEmptyAssetForm(state.assetManagementPage);
          state.assetManagementModal = null;
        }
        if (nextView === "edit_asset_type") {
          state.assetTypeForm = ui.buildEmptyAssetTypeForm();
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

    bindAssetManagementFilters(app);
    bindAssetManagementModal(app);

    for (const row of root.document.querySelectorAll("[data-asset-type-row-id]")) {
      row.addEventListener("click", () => {
        state.assetTypeForm = ui.buildAssetTypeFormState(state.assetManagementPage, Number(row.dataset.assetTypeRowId));
        state.assetTypeError = "";
        state.assetManagementView = "edit_asset_type";
        state.assetManagementModal = { kind: "asset_type" };
        renderAssetManagementPage(app);
      });
    }

    for (const row of root.document.querySelectorAll("[data-asset-row-id]")) {
      row.addEventListener("click", () => {
        state.assetForm = ui.buildAssetFormState(state.assetManagementPage, Number(row.dataset.assetRowId));
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
        state.assetManagementNotice = "";
      });
    }

    const activeInput = root.document.getElementById("asset-type-active-input");
    if (activeInput) {
      activeInput.addEventListener("change", (event) => {
        state.assetTypeForm.isActive = event.target.checked;
        state.assetManagementNotice = "";
      });
    }

    const resetButton = root.document.getElementById("asset-type-reset-button");
    if (resetButton) {
      resetButton.addEventListener("click", async () => {
        await runTransition(async () => {
          if (state.assetManagementView === "create_asset_type") {
            state.assetTypeForm = ui.buildEmptyAssetTypeForm();
            state.assetTypeError = "";
            renderAssetManagementPage(app);
            return;
          }

          state.assetTypeForm = ui.buildEmptyAssetTypeForm();
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
              ? await backend.UpdateAssetType(ui.buildAssetTypeUpdatePayload(state.assetTypeForm))
              : await backend.CreateAssetType(ui.buildAssetTypeCreatePayload(state.assetTypeForm));
            if (!isEditMode) {
              state.assetManagementNotice = `Created asset type ${state.assetTypeForm.name}.`;
            }
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
        state.assetManagementNotice = "";
      });
    }

    const typeSelect = root.document.getElementById("asset-type-select");
    if (typeSelect) {
      typeSelect.addEventListener("change", (event) => {
        state.assetForm.assetTypeID = Number(event.target.value);
        state.assetManagementNotice = "";
      });
    }

    const brokerInput = root.document.getElementById("asset-broker-input");
    if (brokerInput) {
      brokerInput.addEventListener("input", (event) => {
        state.assetForm.broker = event.target.value;
        state.assetManagementNotice = "";
      });
    }

    const cashInput = root.document.getElementById("asset-is-cash-input");
    if (cashInput) {
      cashInput.addEventListener("change", (event) => {
        state.assetForm.isCash = event.target.checked;
        if (state.assetForm.isCash) {
          state.assetForm.autoIncrement = "0.00";
        }
        state.assetManagementNotice = "";
        renderAssetManagementPage(app);
      });
    }

    const liabilityInput = root.document.getElementById("asset-is-liability-input");
    if (liabilityInput) {
      liabilityInput.addEventListener("change", (event) => {
        state.assetForm.isLiability = event.target.checked;
        state.assetManagementNotice = "";
      });
    }

    const activeInput = root.document.getElementById("asset-is-active-input");
    if (activeInput) {
      activeInput.addEventListener("change", (event) => {
        state.assetForm.isActive = event.target.checked;
        state.assetManagementNotice = "";
      });
    }

    const autoIncrementInput = root.document.getElementById("asset-auto-increment-input");
    if (autoIncrementInput) {
      ui.bindStandaloneDecimalInput(autoIncrementInput, {
        onChange: (value) => {
          state.assetForm.autoIncrement = value;
          state.assetManagementNotice = "";
        },
      });
    }

    const resetButton = root.document.getElementById("asset-reset-button");
    if (resetButton) {
      resetButton.addEventListener("click", async () => {
        await runTransition(async () => {
          if (state.assetManagementView === "create_asset") {
            state.assetForm = ui.buildEmptyAssetForm(state.assetManagementPage);
            state.assetError = "";
            renderAssetManagementPage(app);
            return;
          }

          state.assetForm = ui.buildEmptyAssetForm(state.assetManagementPage);
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
              ? await backend.UpdateAsset(ui.buildAssetUpdatePayload(state.assetForm))
              : await backend.CreateAsset(ui.buildAssetCreatePayload(state.assetForm));
            if (!isEditMode) {
              state.assetManagementNotice = `Created asset ${state.assetForm.name}.`;
            }
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

  function bindAssetManagementFilters(app) {
    bindAssetManagementFilter("asset-type-filter-active", "assetTypeActive", app);
    bindAssetManagementFilter("asset-filter-active", "assetActive", app);
    bindAssetManagementFilter("asset-filter-cash", "assetCash", app);
    bindAssetManagementFilter("asset-filter-liability", "assetLiability", app);
  }

  function bindAssetManagementFilter(elementID, stateKey, app) {
    const input = root.document.getElementById(elementID);
    if (!input) {
      return;
    }
    input.addEventListener("change", (event) => {
      state.assetManagementFilters[stateKey] = event.target.value || "all";
      renderAssetManagementPage(app);
    });
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

  function renderAssetManagementContent(view, page) {
    const notice = state.assetManagementNotice
      ? `<p class="success-copy success-banner">${escapeHTML(state.assetManagementNotice)}</p>`
      : "";
    switch (view) {
      case "create_asset":
        return `<section class="asset-management-centered-shell">${notice}${ui.renderAssetEditorCard(page, state.assetForm)}</section>`;
      case "edit_asset":
        return `<section class="asset-management-table-shell">${ui.renderAssetTable(page, state.assetForm)}</section>`;
      case "create_asset_type":
        return `<section class="asset-management-centered-shell">${notice}${ui.renderAssetTypeEditorCard(state.assetTypeForm)}</section>`;
      case "edit_asset_type":
        return `<section class="asset-management-table-shell">${ui.renderAssetTypeTable(page, state.assetTypeForm)}</section>`;
      case "reorder_asset":
        return resolveAssetReorder().renderAssetReorderPage();
      case "reorder_asset_type":
        return resolveAssetReorder().renderAssetTypeReorderPage();
      default:
        return `<section class="asset-management-centered-shell">${ui.renderAssetEditorCard(page, state.assetForm)}</section>`;
    }
  }

  function renderAssetManagementModal(page) {
    if (!state.assetManagementModal) {
      return "";
    }

    const body = state.assetManagementModal.kind === "asset_type"
      ? ui.renderAssetTypeEditorCard(state.assetTypeForm, { errorMessage: state.assetTypeError })
      : ui.renderAssetEditorCard(page, state.assetForm, { errorMessage: state.assetError });
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
          state.assetForm = ui.buildEmptyAssetForm(state.assetManagementPage);
        }
        if (state.assetManagementView === "edit_asset_type") {
          state.assetTypeForm = ui.buildEmptyAssetTypeForm();
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
      state.assetForm = ui.buildEmptyAssetForm(state.assetManagementPage);
    }
    if (state.assetManagementView === "edit_asset_type") {
      state.assetTypeForm = ui.buildEmptyAssetTypeForm();
    }
    renderAssetManagementPage(app);
    return true;
  }

  function shouldCloseAssetManagementModalOnEscape(options) {
    return options.key === "Escape" && options.hasAssetManagementModal;
  }

  return {
    ...ui,
    handleAssetManagementKeydown,
    renderAssetManagementPage,
    resolveAssetManagementView,
    shouldCloseAssetManagementModalOnEscape,
  };
}));
