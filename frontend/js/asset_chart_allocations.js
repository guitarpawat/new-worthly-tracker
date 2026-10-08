(function initAssetChartAllocations(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const allocations = root.WorthlyCustomAllocations || (typeof require !== "undefined" ? require("./custom_allocations.js") : null);
  const controls = root.WorthlyControls || (typeof require !== "undefined" ? require("./controls.js") : null);
  const api = factory(shared, allocations, controls, root);
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  root.WorthlyAssetChartAllocations = api;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildAssetChartAllocations(shared, allocations, controls, root) {
  const { escapeHTML } = shared;

  function buildDrafts(page, assetID) {
    return (page.CustomAllocationCharts || []).map((chart) => ({
      ChartID: chart.ID,
      Name: chart.Name,
      Excluded: (chart.ExcludedAssetIDs || []).includes(assetID),
      Entries: chart.Entries.filter((entry) => entry.AssetID === assetID).map((entry) => ({ ...entry })),
    }));
  }

  function buildPayload(form) {
    return form.chartAllocations.map((chart) => ({
      ChartID: chart.ChartID,
      Excluded: Boolean(chart.Excluded),
      Entries: chart.Excluded ? [] : chart.Entries.map((entry) => ({ ...entry, AssetID: form.id })),
    }));
  }

  function validate(form) {
    for (const chart of form.chartAllocations || []) {
      if (chart.Excluded) continue;
      const error = allocations.validateDraft(chart);
      if (error) {
        form.selectedChartID = chart.ChartID;
        return `${chart.Name}: ${error}`;
      }
    }
    return "";
  }

  function totalLabel(chart, assetID) {
    const total = allocations.assetTotal(chart.Entries, assetID);
    if (!Number.isFinite(total)) return "Check percentages";
    return `${(total / 100).toFixed(2)}% allocated · ${(Math.max(0, 10000 - total) / 100).toFixed(2)}% unallocated`;
  }

  function render(page, form) {
    if (!form.id) return "";
    const charts = form.chartAllocations || [];
    const selected = charts.find((chart) => chart.ChartID === form.selectedChartID) || charts[0];
    return `<section class="asset-chart-tags">
      <h3>Chart Tags</h3>
      <p class="subtle">Assign categories for each chart. Percentages apply to every snapshot; any remainder is Unallocated.</p>
      ${charts.length ? `<label class="field-inline asset-chart-select"><span class="field-label">Chart</span>${controls.renderSelectControl({
        id: "asset-chart-select", value: selected.ChartID, ariaLabel: "Chart tags to edit",
        buttonClass: "snapshot-select asset-management-select",
        options: charts.map((chart) => ({ value: chart.ChartID, label: chart.Name })),
      })}</label>` : ""}
      ${charts.length ? charts.map((chart, index) => {
        if (chart.ChartID !== selected.ChartID) return "";
        const suggestions = [...new Set((page.CustomAllocationCharts?.find((item) => item.ID === chart.ChartID)?.Entries || [])
          .map((entry) => entry.Category))];
        return `<fieldset class="asset-chart-group"><legend>${escapeHTML(chart.Name)}</legend>
          <label class="asset-chart-inclusion">
            <span class="field-label">Include in chart</span>
            <span class="toggle-field"><input id="asset-chart-included-${index}" class="toggle-field-input" type="checkbox"
              data-asset-chart-included="${index}" ${chart.Excluded ? "" : "checked"}>
              <span class="mock-toggle" aria-hidden="true"><span class="mock-toggle-thumb"></span></span></span>
          </label>
          ${chart.Excluded ? '<p class="subtle">Excluded from this chart’s values and percentages. Previously saved tags are kept.</p>' : `
          <datalist id="asset-chart-suggestions-${index}">${suggestions.map((name) => `<option value="${escapeHTML(name)}"></option>`).join("")}</datalist>
          ${chart.Entries.map((entry, row) => `<div class="asset-chart-row">
            <label class="field-inline"><span class="field-label">Category</span>
              <input id="asset-chart-category-${index}-${row}" class="form-input" maxlength="100" list="asset-chart-suggestions-${index}"
                data-asset-chart-category="${index}:${row}" value="${escapeHTML(entry.Category)}" placeholder="e.g. USA"></label>
            <label class="field-inline"><span class="field-label">Allocation (%)</span>
              <input id="asset-chart-percentage-${index}-${row}" class="form-input numeric" inputmode="decimal"
                data-asset-chart-percentage="${index}:${row}" value="${escapeHTML(entry.Percentage)}" placeholder="e.g. 50"></label>
            <button type="button" class="button" data-asset-chart-remove="${index}:${row}" aria-label="Remove ${escapeHTML(entry.Category || "category")}">Remove</button>
          </div>`).join("")}
          <div class="asset-chart-footer"><button id="asset-chart-add-${index}" type="button" class="button button-small" data-asset-chart-add="${index}">Add Tag</button>
            <span class="subtle" data-asset-chart-total="${index}">${escapeHTML(totalLabel(chart, form.id))}</span></div>`}
        </fieldset>`;
      }).join("") : '<p class="subtle">Create a chart in Manage Chart to add tags to this asset.</p>'}
    </section>`;
  }

  function bind(form, refresh) {
    for (const input of root.document.querySelectorAll("[data-asset-chart-included]")) {
      input.addEventListener("change", () => {
        form.chartAllocations[Number(input.dataset.assetChartIncluded)].Excluded = !input.checked;
        refresh();
        root.document.getElementById(input.id)?.focus({ preventScroll: true });
      });
    }
    root.document.getElementById("asset-chart-select")?.addEventListener("change", (event) => {
      form.selectedChartID = Number(event.target.value);
      refresh();
      root.document.getElementById("asset-chart-select-trigger")?.focus({ preventScroll: true });
    });
    const findEntry = (key) => {
      const [chart, row] = key.split(":").map(Number);
      return form.chartAllocations[chart].Entries[row];
    };
    for (const input of root.document.querySelectorAll("[data-asset-chart-category]")) {
      input.addEventListener("input", () => { findEntry(input.dataset.assetChartCategory).Category = input.value; });
    }
    for (const input of root.document.querySelectorAll("[data-asset-chart-percentage]")) {
      input.addEventListener("input", () => {
        findEntry(input.dataset.assetChartPercentage).Percentage = input.value;
        const index = Number(input.dataset.assetChartPercentage.split(":")[0]);
        const chart = form.chartAllocations[index];
        const total = root.document.querySelector(`[data-asset-chart-total="${index}"]`);
        total.textContent = totalLabel(chart, form.id);
      });
    }
    for (const button of root.document.querySelectorAll("[data-asset-chart-add]")) {
      button.addEventListener("click", () => {
        const index = Number(button.dataset.assetChartAdd);
        const entries = form.chartAllocations[index].Entries;
        entries.push({ AssetID: form.id, Category: "", Percentage: "" });
        refresh();
        root.document.getElementById(`asset-chart-category-${index}-${entries.length - 1}`)?.focus({ preventScroll: true });
      });
    }
    for (const button of root.document.querySelectorAll("[data-asset-chart-remove]")) {
      button.addEventListener("click", () => {
        const [index, row] = button.dataset.assetChartRemove.split(":").map(Number);
        form.chartAllocations[index].Entries.splice(row, 1);
        refresh();
        root.document.getElementById(`asset-chart-add-${index}`)?.focus({ preventScroll: true });
      });
    }
  }

  return { buildDrafts, buildPayload, validate, render, bind };
}));
