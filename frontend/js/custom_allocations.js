(function initCustomAllocations(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const api = factory(shared, root);
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  root.WorthlyCustomAllocations = api;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildCustomAllocations(shared, root) {
  const { escapeHTML, state } = shared;

  function createDraft(chart) {
    return { ID: chart?.ID || 0, Name: chart?.Name || "" };
  }

  function percentageUnits(value) {
    const text = String(value).trim();
    if (!/^\d+(?:\.\d{1,2})?$/.test(text)) return NaN;
    return Math.round(Number(text) * 100);
  }

  function assetTotal(entries, assetID) {
    return entries.filter((entry) => entry.AssetID === assetID)
      .reduce((sum, entry) => sum + percentageUnits(entry.Percentage), 0);
  }

  function validateDraft(draft) {
    if (!draft.Name.trim()) return "Enter a chart name.";
    const seen = new Set();
    for (const entry of draft.Entries) {
      const category = entry.Category.trim().toLowerCase();
      if (!category) return "Enter a category for every allocation row.";
      if (category === "unallocated") return "Unallocated is reserved for the remaining percentage.";
      const key = `${entry.AssetID}:${category}`;
      if (seen.has(key)) return "Each category can appear only once per asset in a chart.";
      seen.add(key);
      const units = percentageUnits(entry.Percentage);
      if (!Number.isFinite(units) || units <= 0 || units > 10000) {
        return "Enter a percentage greater than 0 and at most 100, with up to 2 decimal places.";
      }
      if (assetTotal(draft.Entries, entry.AssetID) > 10000) return "Each asset’s allocations must total at most 100%.";
    }
    return "";
  }

  function render(page) {
    const charts = page.CustomAllocationCharts || [];
    return `<section class="panel asset-management-intro">
      <div><h2>Charts</h2><p class="subtle">Select a chart to rename or delete it. Manage its tags in Manage Asset.</p></div>
      <button id="chart-add" class="button" type="button">Add Chart</button>
    </section>
    ${state.chartManagementNotice ? `<p class="success-copy success-banner" role="status">${escapeHTML(state.chartManagementNotice)}</p>` : ""}
    <section class="panel table-panel">
      <div class="table-header"><h2>Chart Names</h2><span>${charts.length} chart(s)</span></div>
      ${charts.length ? `<div class="table-scroll"><table class="management-table">
        <thead><tr><th>Name</th></tr></thead>
        <tbody>${charts.map((chart) => `<tr data-chart-row="${chart.ID}" tabindex="0" role="button" aria-label="Edit ${escapeHTML(chart.Name)}">
          <td>${escapeHTML(chart.Name)}</td></tr>`).join("")}</tbody>
      </table></div>` : '<p class="chart-management-empty subtle">No charts yet. Choose Add Chart to create one.</p>'}
    </section>`;
  }

  function renderDialog() {
    const editor = state.customAllocationEditor;
    if (!editor) return "";
    const { draft } = editor;
    return `<section class="panel asset-management-card">
      <div class="table-header"><h2 id="chart-dialog-title">${editor.confirmDelete ? "Delete Chart" : draft.ID ? "Edit Chart" : "Add Chart"}</h2></div>
      <form id="custom-allocation-form" class="asset-management-form" novalidate>
        <fieldset class="chart-management-fields" ${editor.saving ? "disabled" : ""}>
          ${editor.confirmDelete ? `<p>Delete “${escapeHTML(editor.originalName)}”? It will no longer appear in Allocation or Manage Asset.</p>` : `
            <label class="field-inline"><span class="field-label">Chart Name</span>
              <input id="custom-allocation-name" class="form-input" maxlength="100" value="${escapeHTML(draft.Name)}" placeholder="e.g. Asset Allocation Country"></label>`}
          ${editor.error ? `<p id="custom-allocation-error" class="error-copy error-banner" role="alert" tabindex="-1">${escapeHTML(editor.error)}</p>` : ""}
          <div class="chart-dialog-actions">
            ${editor.confirmDelete ? `
              <button type="button" class="button chart-delete-button" id="custom-allocation-confirm-delete">Delete Chart</button>
              <button type="button" class="button" id="custom-allocation-cancel-delete">Cancel</button>` : `
              ${draft.ID ? '<button class="button chart-delete-button" type="button" id="custom-allocation-delete">Delete Chart</button>' : ""}
              <div class="actions asset-management-card-actions">
                <button class="button button-primary" type="submit">${editor.saving ? "Saving…" : draft.ID ? "Save Chart" : "Create Chart"}</button>
                <button class="button" type="button" id="chart-dialog-close">${draft.ID ? "Close" : "Cancel"}</button>
              </div>`}
          </div>
        </fieldset>
      </form>
    </section>`;
  }

  function close(refresh) {
    const editor = state.customAllocationEditor;
    if (editor?.saving) return;
    const returnFocusID = editor?.returnFocusID || "chart-add";
    const chartID = editor?.draft.ID;
    state.customAllocationEditor = null;
    state.assetManagementModal = null;
    refresh();
    const target = chartID ? root.document.querySelector(`[data-chart-row="${chartID}"]`) : null;
    (target || root.document.getElementById(returnFocusID))?.focus({ preventScroll: true });
  }

  function bind(page, refresh, reload) {
    const byID = (id) => root.document.getElementById(id);
    const onClick = (id, callback) => byID(id)?.addEventListener("click", callback);
    const open = (chart) => {
      state.customAllocationEditor = { draft: createDraft(chart), originalName: chart?.Name || "", error: "", returnFocusID: "chart-add" };
      state.assetManagementModal = { kind: "chart" };
      state.chartManagementNotice = "";
      refresh();
      byID("custom-allocation-name")?.focus({ preventScroll: true });
    };
    onClick("chart-add", () => open());
    for (const row of root.document.querySelectorAll("[data-chart-row]")) {
      const activate = () => open(page.CustomAllocationCharts.find((chart) => chart.ID === Number(row.dataset.chartRow)));
      row.addEventListener("click", activate);
      row.addEventListener("keydown", (event) => {
        if ((event.key === "Enter" || event.key === " ") && !event.repeat && !event.isComposing) {
          event.preventDefault(); activate();
        }
      });
    }
    const editor = state.customAllocationEditor;
    if (state.assetManagementModal?.kind !== "chart" || !editor) return;
    byID("custom-allocation-name")?.addEventListener("input", (event) => { editor.draft.Name = event.target.value; });
    onClick("chart-dialog-close", () => close(refresh));
    byID("custom-allocation-form")?.addEventListener("submit", async (event) => {
      event.preventDefault();
      if (editor.saving || editor.confirmDelete) return;
      if (!editor.draft.Name.trim()) {
        editor.error = "Enter a chart name."; refresh(); byID("custom-allocation-name")?.focus();
        return;
      }
      await mutate(() => shared.resolveBackend().SaveCustomAllocationChart(editor.draft), "Chart saved.");
    });
    onClick("custom-allocation-delete", () => {
      editor.confirmDelete = true; editor.error = ""; refresh(); byID("custom-allocation-cancel-delete")?.focus();
    });
    onClick("custom-allocation-cancel-delete", () => {
      editor.confirmDelete = false; refresh(); byID("custom-allocation-delete")?.focus({ preventScroll: true });
    });
    onClick("custom-allocation-confirm-delete", () => mutate(
      () => shared.resolveBackend().DeleteCustomAllocationChart(editor.draft.ID), "Chart deleted.",
    ));

    async function mutate(action, notice) {
      if (editor.saving) return;
      editor.saving = true;
      editor.error = "";
      refresh();
      try {
        await action();
        state.customAllocationEditor = null;
        state.assetManagementModal = null;
        state.chartManagementNotice = notice;
        await reload();
        byID("chart-add")?.focus({ preventScroll: true });
      } catch (error) {
        editor.error = error?.message || String(error);
        editor.saving = false;
        state.customAllocationEditor = editor;
        refresh();
      }
    }
  }

  return { render, renderDialog, close, bind, createDraft, percentageUnits, assetTotal, validateDraft };
}));
