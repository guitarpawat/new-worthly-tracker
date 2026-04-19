(function initSnapshotAssetModal(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const snapshotAssetModal = factory(shared, root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = snapshotAssetModal;
  }
  root.WorthlySnapshotAssetModal = snapshotAssetModal;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildSnapshotAssetModal(shared, root) {
  const { runTransition, state } = shared;

  function resolveAssetManagement() {
    const module = root.WorthlyAssetManagement || (typeof require !== "undefined" ? require("./asset_management.js") : null);
    if (!module) {
      throw new Error("Asset management module is not available");
    }
    return module;
  }

  function renderSnapshotAssetModal() {
    const modal = state.snapshotAssetModal;
    if (!modal) {
      return "";
    }

    const body = modal.kind === "asset_type"
      ? resolveAssetManagement().renderAssetTypeEditorCard(modal.assetTypeForm, { errorMessage: modal.error, secondaryButtonLabel: "Close" })
      : resolveAssetManagement().renderAssetEditorCard(modal.page, modal.assetForm, { errorMessage: modal.error, secondaryButtonLabel: "Close" });

    return `
      <div class="dialog-backdrop asset-management-dialog-backdrop" data-snapshot-asset-modal-close="backdrop">
        <section class="asset-management-dialog-shell" role="dialog" aria-modal="true">
          ${body}
        </section>
      </div>
    `;
  }

  function bindSnapshotAssetModal(options) {
    const { app, renderEditPage, addCreatedAssetToDraft, markEditDirty } = options;
    const modal = state.snapshotAssetModal;
    if (!modal) {
      return;
    }

    for (const element of root.document.querySelectorAll("[data-snapshot-asset-modal-close]")) {
      element.addEventListener("click", (event) => {
        if (event.target !== event.currentTarget) {
          return;
        }
        state.snapshotAssetModal = null;
        renderEditPage(app);
      });
    }

    if (modal.kind === "asset_type") {
      bindSnapshotAssetTypeModal(app, modal, renderEditPage);
      return;
    }

    bindSnapshotAssetModalForm(app, modal, renderEditPage, addCreatedAssetToDraft, markEditDirty);
  }

  function bindSnapshotAssetTypeModal(app, modal, renderEditPage) {
    const assetManagement = resolveAssetManagement();

    const nameInput = root.document.getElementById("asset-type-name-input");
    if (nameInput) {
      nameInput.addEventListener("input", (event) => {
        modal.assetTypeForm.name = event.target.value;
      });
    }

    const activeInput = root.document.getElementById("asset-type-active-input");
    if (activeInput) {
      activeInput.addEventListener("change", (event) => {
        modal.assetTypeForm.isActive = event.target.checked;
      });
    }

    const closeButton = root.document.getElementById("asset-type-reset-button");
    if (closeButton) {
      closeButton.addEventListener("click", () => {
        state.snapshotAssetModal = null;
        renderEditPage(app);
      });
    }

    const saveButton = root.document.getElementById("asset-type-save-button");
    if (saveButton) {
      saveButton.addEventListener("click", async () => {
        await runTransition(async () => {
          modal.error = "";
          try {
            const backend = shared.resolveBackend();
            await backend.CreateAssetType(assetManagement.buildAssetTypeCreatePayload(modal.assetTypeForm));
            state.editNotice = `Created asset type ${modal.assetTypeForm.name}.`;
            state.snapshotAssetModal = null;
            renderEditPage(app);
          } catch (error) {
            modal.error = error?.message || String(error);
            renderEditPage(app);
          }
        });
      });
    }
  }

  function bindSnapshotAssetModalForm(app, modal, renderEditPage, addCreatedAssetToDraft, markEditDirty) {
    const assetManagement = resolveAssetManagement();

    const nameInput = root.document.getElementById("asset-name-input");
    if (nameInput) {
      nameInput.addEventListener("input", (event) => {
        modal.assetForm.name = event.target.value;
      });
    }

    const typeSelect = root.document.getElementById("asset-type-select");
    if (typeSelect) {
      typeSelect.addEventListener("change", (event) => {
        modal.assetForm.assetTypeID = Number(event.target.value);
      });
    }

    const brokerInput = root.document.getElementById("asset-broker-input");
    if (brokerInput) {
      brokerInput.addEventListener("input", (event) => {
        modal.assetForm.broker = event.target.value;
      });
    }

    const cashInput = root.document.getElementById("asset-is-cash-input");
    if (cashInput) {
      cashInput.addEventListener("change", (event) => {
        modal.assetForm.isCash = event.target.checked;
        if (modal.assetForm.isCash) {
          modal.assetForm.autoIncrement = "0.00";
        }
        renderEditPage(app);
      });
    }

    const liabilityInput = root.document.getElementById("asset-is-liability-input");
    if (liabilityInput) {
      liabilityInput.addEventListener("change", (event) => {
        modal.assetForm.isLiability = event.target.checked;
      });
    }

    const activeInput = root.document.getElementById("asset-is-active-input");
    if (activeInput) {
      activeInput.addEventListener("change", (event) => {
        modal.assetForm.isActive = event.target.checked;
      });
    }

    const autoIncrementInput = root.document.getElementById("asset-auto-increment-input");
    if (autoIncrementInput) {
      assetManagement.bindStandaloneDecimalInput(autoIncrementInput, {
        onChange: (value) => {
          modal.assetForm.autoIncrement = value;
        },
      });
    }

    const closeButton = root.document.getElementById("asset-reset-button");
    if (closeButton) {
      closeButton.addEventListener("click", () => {
        state.snapshotAssetModal = null;
        renderEditPage(app);
      });
    }

    const saveButton = root.document.getElementById("asset-save-button");
    if (saveButton) {
      saveButton.addEventListener("click", async () => {
        await runTransition(async () => {
          modal.error = "";
          try {
            const backend = shared.resolveBackend();
            const result = await backend.CreateAsset(assetManagement.buildAssetCreatePayload(modal.assetForm));
            if (modal.assetForm.isActive) {
              addCreatedAssetToDraft(modal, result.ID);
              markEditDirty();
            }
            state.editNotice = `Created asset ${modal.assetForm.name}.`;
            state.snapshotAssetModal = null;
            renderEditPage(app);
          } catch (error) {
            modal.error = error?.message || String(error);
            renderEditPage(app);
          }
        });
      });
    }
  }

  async function openSnapshotAssetModal(kind, app, renderEditPage) {
    await runTransition(async () => {
      state.editError = "";
      try {
        const backend = shared.resolveBackend();
        const page = await backend.GetAssetManagementPage();
        const assetManagement = resolveAssetManagement();
        state.snapshotAssetModal = kind === "asset_type"
          ? {
            kind,
            page,
            assetTypeForm: assetManagement.buildEmptyAssetTypeForm(),
            error: "",
          }
          : {
            kind,
            page,
            assetForm: assetManagement.buildEmptyAssetForm(page),
            error: "",
          };
        renderEditPage(app);
      } catch (error) {
        state.editError = error?.message || String(error);
        renderEditPage(app);
      }
    });
  }

  function shouldCloseSnapshotAssetModalOnEscape(options) {
    return options.key === "Escape" && options.hasSnapshotAssetModal;
  }

  return {
    bindSnapshotAssetModal,
    openSnapshotAssetModal,
    renderSnapshotAssetModal,
    shouldCloseSnapshotAssetModalOnEscape,
  };
}));
