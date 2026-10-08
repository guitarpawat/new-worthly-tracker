(function initAssetManagementDelete(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const api = factory(shared, root);
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  root.WorthlyAssetManagementDelete = api;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildAssetManagementDelete(shared, root) {
  const { state, escapeHTML } = shared;

  function selected() {
    const isType = state.assetManagementModal?.kind === "asset_type";
    const form = isType ? state.assetTypeForm : state.assetForm;
    const rows = isType ? state.assetManagementPage?.AssetTypes : state.assetManagementPage?.Assets;
    return { isType, form, row: rows?.find((row) => row.ID === form?.id), label: isType ? "Type" : "Asset" };
  }

  function editorOptions() {
    if (!["asset", "asset_type"].includes(state.assetManagementModal?.kind)) return {};
    const { isType, form, row, label } = selected();
    if (!form?.id) return {};
    const allowed = isType ? row?.AssetCount === 0 : row?.CanDelete === true;
    const reason = isType
      ? "This type contains assets. Move or delete them before deleting the type."
      : "This asset has snapshot records, including records in deleted snapshots, and cannot be deleted.";
    const button = `<button id="management-delete" class="button chart-delete-button" type="button" ${allowed ? "" : `disabled aria-label="Delete ${label}: ${escapeHTML(reason)}"`}>Delete ${label}</button>`;
    return {
      deleteAction: allowed ? button : `<span class="management-delete-hover" title="${escapeHTML(reason)}">${button}</span>`,
    };
  }

  function renderConfirmation() {
    const { row, label } = selected();
    const modal = state.assetManagementModal;
    return `<section class="panel asset-management-card">
      <div class="table-header"><h2>Delete ${label}</h2></div>
      <div class="asset-management-form"><p>Delete “${escapeHTML(row?.Name || "")}"?</p>
        ${modal.error ? `<p class="error-copy error-banner" role="alert">${escapeHTML(modal.error)}</p>` : ""}
        <div class="actions">
          <button id="management-delete-confirm" class="button chart-delete-button" type="button" ${modal.saving ? "disabled" : ""}>Delete ${label}</button>
          <button id="management-delete-cancel" class="button" type="button" ${modal.saving ? "disabled" : ""}>Cancel</button>
        </div>
      </div>
    </section>`;
  }

  function bind(refresh, reload) {
    const modal = state.assetManagementModal;
    if (!["asset", "asset_type"].includes(modal?.kind)) return;
    const byID = (id) => root.document.getElementById(id);
    byID("management-delete")?.addEventListener("click", () => {
      modal.confirmDelete = true; modal.error = ""; refresh(); byID("management-delete-cancel")?.focus();
    });
    byID("management-delete-cancel")?.addEventListener("click", () => {
      modal.confirmDelete = false; modal.error = ""; refresh(); byID("management-delete")?.focus();
    });
    byID("management-delete-confirm")?.addEventListener("click", async () => {
      if (modal.saving) return;
      const { isType, form, label } = selected();
      modal.saving = true; modal.error = ""; refresh();
      try {
        const backend = shared.resolveBackend();
        await (isType ? backend.DeleteAssetType(form.id) : backend.DeleteAsset(form.id));
        state.assetManagementModal = null;
        state.assetManagementNotice = `${label} deleted.`;
        await reload({ view: isType ? "edit_asset_type" : "edit_asset", preserveScroll: true });
        byID(`management-add-${isType ? "asset_type" : "asset"}`)?.focus({ preventScroll: true });
      } catch (error) {
        modal.saving = false; modal.error = error?.message || String(error); refresh();
      }
    });
  }

  return { editorOptions, renderConfirmation, bind };
}));
