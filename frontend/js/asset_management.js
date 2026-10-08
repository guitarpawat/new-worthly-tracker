(function initAssetManagement(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const assetManagementUI = root.WorthlyAssetManagementUI || (typeof require !== "undefined" ? require("./asset_management_ui.js") : null);
  const assetManagement = factory(shared, assetManagementUI, root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = assetManagement;
  }
  root.WorthlyAssetManagement = assetManagement;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildAssetManagement(shared, ui, root) {
  const customAllocations = root.WorthlyCustomAllocations || (typeof require !== "undefined" ? require("./custom_allocations.js") : null);
  const deletion = root.WorthlyAssetManagementDelete || (typeof require !== "undefined" ? require("./asset_management_delete.js") : null);
  const assetCharts = root.WorthlyAssetChartAllocations || (typeof require !== "undefined" ? require("./asset_chart_allocations.js") : null);
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

    const view = resolveAssetManagementView();
    state.assetManagementView = view;
    const scrollX = root.scrollX || 0;
    const scrollY = root.scrollY || 0;
    const focusedID = root.document.activeElement?.id;
    const dialogScrollTop = root.document.querySelector?.(".asset-management-dialog-backdrop")?.scrollTop;
    appRoot.innerHTML = `
      <main ${state.assetManagementModal ? "inert" : ""} class="app-layout asset-management-layout ${state.assetManagementModal ? "app-modal-open" : ""}">
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
    if (state.assetManagementModal) {
      const dialog = root.document.querySelector(".asset-management-dialog-shell");
      const previous = focusedID ? root.document.getElementById(focusedID) : null;
      const target = previous && dialog.contains(previous) && !previous.disabled
        ? previous
        : dialog.querySelector("input:not([type=hidden]):not(:disabled), #management-delete-cancel");
      target?.focus({ preventScroll: true });
    }
    root.scrollTo?.({ left: scrollX, top: scrollY, behavior: "instant" });
    const backdrop = root.document.querySelector?.(".asset-management-dialog-backdrop");
    if (backdrop && Number.isFinite(dialogScrollTop)) backdrop.scrollTop = dialogScrollTop;
  }

  function bindAssetManagementPage(app) {
    if (state.assetManagementView === "manage_chart") {
      customAllocations.bind(state.assetManagementPage, () => renderAssetManagementPage(app),
        () => app.loadAssetManagementPage({ view: "manage_chart", preserveScroll: true }));
    }
    for (const button of root.document.querySelectorAll("[data-asset-management-view]")) {
      button.addEventListener("click", (event) => {
        const nextView = event.currentTarget.dataset.assetManagementView;
        if (!nextView || event.currentTarget.disabled) {
          return;
        }

        state.assetTypeError = "";
        state.assetError = "";
        state.assetManagementView = nextView;
        state.assetManagementModal = null;
        state.assetForm = ui.buildEmptyAssetForm(state.assetManagementPage);
        state.assetTypeForm = ui.buildEmptyAssetTypeForm();
        renderAssetManagementPage(app);
        if (nextView === "manage_chart") root.document.getElementById("chart-add")?.focus({ preventScroll: true });
      });
    }

    if (state.assetManagementModal?.kind === "asset_type") bindAssetTypeForm(app);
    if (state.assetManagementModal?.kind === "asset") bindAssetForm(app);
    for (const kind of ["asset", "asset_type"]) {
      root.document.getElementById(`management-add-${kind}`)?.addEventListener("click", () => {
        state.assetForm = ui.buildEmptyAssetForm(state.assetManagementPage);
        state.assetTypeForm = ui.buildEmptyAssetTypeForm();
        state.assetError = "";
        state.assetTypeError = "";
        state.assetManagementModal = { kind };
        renderAssetManagementPage(app);
      });
    }
    if (state.assetManagementView === "reorder_asset_type") {
      resolveAssetReorder().bindAssetTypeReorderPage(app);
    }
    if (state.assetManagementView === "reorder_asset") {
      resolveAssetReorder().bindAssetReorderPage(app);
    }

    bindAssetManagementFilters(app);
    bindAssetManagementModal(app);
    deletion.bind(() => renderAssetManagementPage(app), (options) => app.loadAssetManagementPage(options));

    for (const row of root.document.querySelectorAll("[data-asset-type-row-id]")) {
      row.addEventListener("keydown", activateManagementRow);
      row.addEventListener("click", () => {
        state.assetTypeForm = ui.buildAssetTypeFormState(state.assetManagementPage, Number(row.dataset.assetTypeRowId));
        state.assetTypeError = "";
        state.assetManagementView = "edit_asset_type";
        state.assetManagementModal = { kind: "asset_type" };
        renderAssetManagementPage(app);
      });
    }

    for (const row of root.document.querySelectorAll("[data-asset-row-id]")) {
      row.addEventListener("keydown", activateManagementRow);
      row.addEventListener("click", () => {
        state.assetForm = ui.buildAssetFormState(state.assetManagementPage, Number(row.dataset.assetRowId));
        state.assetError = "";
        state.assetManagementView = "edit_asset";
        state.assetManagementModal = { kind: "asset" };
        renderAssetManagementPage(app);
      });
    }
  }

  function activateManagementRow(event) {
    if ((event.key === "Enter" || event.key === " ") && !event.repeat && !event.isComposing) {
      event.preventDefault(); event.currentTarget.click();
    }
  }

  function bindAssetTypeForm(app) {
    const nameInput = root.document.getElementById("asset-type-name-input");
    if (nameInput) {
      ui.bindAssetTypeNameSubmit(nameInput);
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
          closeAssetEditor(app);
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
                ? { view: "edit_asset_type", selectedAssetTypeID: result.ID, preserveScroll: true }
                : { view: "edit_asset_type", preserveScroll: true },
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
    assetCharts.bind(state.assetForm, () => renderAssetManagementPage(app));
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
          closeAssetEditor(app);
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
            if (isEditMode) {
              const error = assetCharts.validate(state.assetForm);
              if (error) throw new Error(error);
            }
            const result = state.assetForm.id > 0
              ? await backend.UpdateAsset(ui.buildAssetUpdatePayload(state.assetForm))
              : await backend.CreateAsset(ui.buildAssetCreatePayload(state.assetForm));
            if (!isEditMode) {
              state.assetManagementNotice = `Created asset ${state.assetForm.name}.`;
            }
            await app.loadAssetManagementPage(
              isEditMode
                ? { view: "edit_asset", selectedAssetID: result.ID, preserveScroll: true }
                : { view: "edit_asset", preserveScroll: true },
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
    const view = options.view || (options.selectedAssetID ? "edit_asset" : options.selectedAssetTypeID ? "edit_asset_type" : state.assetManagementView) || "edit_asset";
    return view === "create_asset" ? "edit_asset" : view === "create_asset_type" ? "edit_asset_type" : view;
  }

  function renderAssetManagementSubnav(view, page) {
    const actions = [
      { id: "edit_asset", label: "Manage Asset", disabled: false },
      { id: "reorder_asset", label: "Reorder Asset", disabled: (page?.Assets?.length || 0) === 0 },
      { id: "edit_asset_type", label: "Manage Type", disabled: false },
      { id: "reorder_asset_type", label: "Reorder Type", disabled: (page?.AssetTypes?.length || 0) === 0 },
      { id: "manage_chart", label: "Manage Chart", disabled: false },
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

  function renderManagementToolbar(kind, label) {
    const disabled = kind === "asset" && !state.assetManagementPage.ActiveAssetTypes?.length;
    return `<section class="panel asset-management-intro"><div><h2>${label}</h2>
      <p class="subtle">Select a row to edit or delete. ${disabled ? "Add an active type before adding assets." : ""}</p></div>
      <button id="management-add-${kind}" class="button" type="button" ${disabled ? "disabled" : ""}>Add ${kind === "asset" ? "Asset" : "Type"}</button></section>`;
  }

  function renderAssetManagementContent(view, page) {
    const notice = state.assetManagementNotice
      ? `<p class="success-copy success-banner">${escapeHTML(state.assetManagementNotice)}</p>`
      : "";
    switch (view) {
      case "manage_chart":
        return customAllocations.render(page);
      case "edit_asset":
        return `${renderManagementToolbar("asset", "Assets")}${notice}<section class="asset-management-table-shell">${ui.renderAssetTable(page, state.assetForm)}</section>`;
      case "edit_asset_type":
        return `${renderManagementToolbar("asset_type", "Types")}${notice}<section class="asset-management-table-shell">${ui.renderAssetTypeTable(page, state.assetTypeForm)}</section>`;
      case "reorder_asset":
        return resolveAssetReorder().renderAssetReorderPage();
      case "reorder_asset_type":
        return resolveAssetReorder().renderAssetTypeReorderPage();
      default:
        return `${renderManagementToolbar("asset", "Assets")}<section class="asset-management-table-shell">${ui.renderAssetTable(page, state.assetForm)}</section>`;
    }
  }

  function closeAssetEditor(app) {
    const modal = state.assetManagementModal;
    if (modal?.saving) return;
    const isType = modal?.kind === "asset_type";
    const id = isType ? state.assetTypeForm.id : state.assetForm.id;
    state.assetManagementModal = null;
    state.assetForm = ui.buildEmptyAssetForm(state.assetManagementPage);
    state.assetTypeForm = ui.buildEmptyAssetTypeForm();
    state.assetError = "";
    state.assetTypeError = "";
    renderAssetManagementPage(app);
    const row = id ? root.document.querySelector?.(`[data-${isType ? "asset-type" : "asset"}-row-id="${id}"]`) : null;
    const add = root.document.getElementById(`management-add-${isType ? "asset_type" : "asset"}`);
    (row || add)?.focus?.({ preventScroll: true });
  }

  function renderAssetManagementModal(page) {
    if (!state.assetManagementModal) {
      return "";
    }

    const isChart = state.assetManagementModal.kind === "chart";
    const deleteOptions = deletion.editorOptions(page);
    const body = state.assetManagementModal.confirmDelete ? deletion.renderConfirmation() : isChart ? customAllocations.renderDialog() : state.assetManagementModal.kind === "asset_type"
      ? ui.renderAssetTypeEditorCard(state.assetTypeForm, { errorMessage: state.assetTypeError, secondaryButtonLabel: state.assetTypeForm.id ? "Close" : "Cancel", ...deleteOptions })
      : ui.renderAssetEditorCard(page, state.assetForm, { errorMessage: state.assetError, secondaryButtonLabel: state.assetForm.id ? "Close" : "Cancel", ...deleteOptions });
    return `
      <div class="dialog-backdrop asset-management-dialog-backdrop" data-asset-management-modal-close="backdrop">
        <section class="asset-management-dialog-shell ${isChart ? "chart-management-dialog" : ""}" role="dialog" aria-modal="true" ${isChart ? 'aria-labelledby="chart-dialog-title"' : ""}>
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
        if (state.assetManagementModal.kind === "chart") {
          customAllocations.close(() => renderAssetManagementPage(app));
          return;
        }
        closeAssetEditor(app);
      });
    }
  }

  function handleAssetManagementKeydown(event, app) {
    if (state.assetManagementModal && event.key === "Tab") {
      const dialog = root.document.querySelector(".asset-management-dialog-shell");
      const fields = Array.from(dialog.querySelectorAll(
        'input:not([type="hidden"]):not(:disabled), button:not(:disabled), [tabindex="0"]',
      )).filter((field) => field.getClientRects().length > 0);
      const index = fields.indexOf(root.document.activeElement);
      const next = index < 0 ? (event.shiftKey ? fields.length - 1 : 0)
        : (index + (event.shiftKey ? -1 : 1) + fields.length) % fields.length;
      event.preventDefault();
      fields[next]?.focus();
      return true;
    }
    if (!shouldCloseAssetManagementModalOnEscape({
      key: event.key,
      hasAssetManagementModal: Boolean(state.assetManagementModal),
    })) {
      return false;
    }

    event.preventDefault();
    if (state.assetManagementModal?.kind === "chart") {
      customAllocations.close(() => renderAssetManagementPage(app));
      return true;
    }
    closeAssetEditor(app);
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
