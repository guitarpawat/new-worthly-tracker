(function initProgressGoals(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const progressChartLogic = root.WorthlyProgressChartLogic || (typeof require !== "undefined" ? require("./progress_chart_logic.js") : null);
  const controls = root.WorthlyControls || (typeof require !== "undefined" ? require("./controls.js") : null);
  const progressGoals = factory(shared, progressChartLogic, controls, root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = progressGoals;
  }
  root.WorthlyProgressGoals = progressGoals;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildProgressGoals(shared, progressChartLogic, controls, root) {
  const {
    escapeHTML,
    formatDateLabel,
    formatTHB,
    parseEditableNumber,
    renderAppTitle,
    renderErrorState,
    rootErrorBanner,
    runTransition,
    sanitizePartialDecimal,
    state,
  } = shared;
  const {
    ALLOCATION_MODES,
    CHART_MODES,
    PROGRESS_CHART_COLORS,
    PROGRESS_VIEWS,
    PROJECTION_OPTIONS,
    buildAllocationChartConfig,
    buildAllocationRows,
    buildTrendChartConfig,
    goalStatusToneClass,
    normalizeProgressDateSelection,
    resolveCategoryChartColor,
    resolveChartModeLabel,
    resolveProgressView,
  } = progressChartLogic;

  let trendChartInstance = null;
  let allocationChartInstance = null;

  function renderProgressPage(app) {
    const page = state.progressPage;
    const appRoot = root.document.getElementById("app");
    destroyProgressCharts();

    if (!page) {
      appRoot.innerHTML = renderErrorState("Unable to load progress", "Progress page state is empty.");
      return;
    }

    if (!page.HasData) {
      appRoot.innerHTML = `
        <main class="app-layout">
          <section class="hero">
            <div>
              <h1>${renderAppTitle("Progress & Goals")}</h1>
            </div>
          </section>
          <section class="panel progress-empty">
            <p class="eyebrow">Progress</p>
            <h2>No progress data yet</h2>
            <p>Add at least one snapshot first, then this page will show trend analysis, allocation, and goal projection.</p>
          </section>
        </main>
      `;
      syncProgressOverlayState();
      return;
    }

    const view = resolveProgressView({ view: state.progressView });
    const chartMode = state.progressChartMode || "net_worth";
    const projectionMonths = state.progressProjectionMonths || 6;
    const allocationMode = state.progressAllocationMode || "asset_type";
    const allocationDate = normalizeProgressDateSelection(
      page.AllocationSnapshots,
      state.progressAllocationDate || page.Filter?.EndDate,
    );
    state.progressView = view;
    state.progressAllocationDate = allocationDate;

    appRoot.innerHTML = `
      <main class="app-layout ${state.goalModal || state.progressAllocationModal ? "app-overlay-open" : ""}">
        <section class="hero progress-hero">
          <div>
            <h1>${renderAppTitle("Progress & Goals")}</h1>
          </div>
          ${renderProgressHeroActions(page.Filter?.StartDate, page.Filter?.EndDate)}
        </section>
        ${renderProgressSubnav(view)}
        ${renderProgressContent(page, view, chartMode, allocationMode, allocationDate, projectionMonths)}
      </main>
      ${renderAllocationModal(page)}
      ${renderGoalModal()}
    `;

    bindProgressPage(app);
    renderProgressCharts(page, view, chartMode, allocationMode, allocationDate, projectionMonths);
    syncProgressOverlayState();
  }

  function renderDateFilterControl(label, inputID, selectedDate) {
    return `
      <label class="field-inline progress-date-control">
        <span class="field-label">${escapeHTML(label)}</span>
        ${controls.renderDateControl({
          id: inputID,
          value: selectedDate,
          buttonClass: "snapshot-select snapshot-select-button progress-date-select",
        })}
      </label>
    `;
  }

  function renderProgressHeroActions(startDate, endDate) {
    return `
      <div class="actions progress-hero-actions">
        ${renderDateFilterControl("Start Date", "progress-start-date", startDate)}
        ${renderDateFilterControl("End Date", "progress-end-date", endDate)}
      </div>
    `;
  }

  function renderProgressSubnav(view) {
    return `
      <section class="panel asset-management-subnav">
        ${PROGRESS_VIEWS.map((item) => `
          <button
            class="button asset-management-subnav-button ${view === item.id ? "asset-management-subnav-button-active" : ""}"
            type="button"
            data-progress-view="${item.id}"
          >${escapeHTML(item.label)}</button>
        `).join("")}
      </section>
    `;
  }

  function renderProgressContent(page, view, chartMode, allocationMode, allocationDate, projectionMonths) {
    switch (view) {
      case "summary":
        return renderSummaryPage(page, allocationMode, allocationDate);
      case "trend":
      default:
        return renderTrendPage(page, chartMode, projectionMonths);
    }
  }

  function renderTrendPage(page, chartMode, projectionMonths) {
    return `
      <section class="panel progress-chart-panel">
        <div class="table-header">
          <h2>Trend Chart</h2>
          <span>${escapeHTML(resolveChartModeLabel(chartMode))}</span>
        </div>
        <div class="progress-mode-row">
          <div class="progress-chip-group">
            ${CHART_MODES.map((mode) => `
              <button
                class="button progress-chip ${chartMode === mode.id ? "progress-chip-active" : ""}"
                type="button"
                data-progress-chart-mode="${mode.id}"
              >${escapeHTML(mode.label)}</button>
            `).join("")}
          </div>
          ${chartMode === "net_worth" ? renderProjectionSelector(projectionMonths) : ""}
        </div>
        <div class="progress-chart-shell">
          <canvas id="progress-trend-chart" class="progress-chart-canvas" aria-label="${escapeHTML(resolveChartModeLabel(chartMode))}"></canvas>
        </div>
      </section>
      <section class="panel progress-goals-panel">
        <div class="table-header">
          <h2>Goals</h2>
          <button id="progress-add-goal" class="button button-primary button-small" type="button">Add Goal</button>
        </div>
        ${renderGoalsTable(page.GoalEstimates)}
      </section>
    `;
  }

  function renderSummaryPage(page, allocationMode, allocationDate) {
    const selectedDate = state.progressAllocationModal?.snapshotDate || "";
    return `
      <section class="panel table-panel">
        <div class="table-header">
          <h2>Summary Table</h2>
          <span>${page.TrendPoints.length} snapshot(s)</span>
        </div>
        <p class="subtle progress-note progress-summary-hint">Click a snapshot row to open its allocation chart.</p>
        <div class="table-scroll">
          <table class="progress-table">
            <thead>
              <tr>
                <th>Date</th>
                <th class="numeric">Bought</th>
                <th class="numeric">Net Worth</th>
                <th class="numeric">Profit</th>
                <th class="numeric">% Profit</th>
                <th class="numeric">Cash</th>
                <th class="numeric">Non Cash</th>
                <th class="numeric">% Cash</th>
              </tr>
            </thead>
            <tbody>
              ${page.TrendPoints.map((point) => renderProgressRow(point, selectedDate)).join("")}
            </tbody>
          </table>
        </div>
      </section>
    `;
  }

  function renderAllocationModal(page) {
    const modal = state.progressAllocationModal;
    if (!modal) {
      return "";
    }

    const selectedPoint = (page.TrendPoints || []).find((point) => point.SnapshotDate === modal.snapshotDate);
    if (!selectedPoint) {
      return "";
    }

    const allocationRows = buildAllocationRows(page, selectedPoint.SnapshotDate, modal.mode || "asset_type");
    return `
      <div class="dialog-backdrop" data-progress-allocation-close="backdrop">
        <section class="panel dialog-card progress-allocation-dialog" role="dialog" aria-modal="true">
          <div class="table-header">
            <h2>Allocation</h2>
            <span>${escapeHTML(formatDateLabel(selectedPoint.SnapshotDate))}</span>
          </div>
          <div class="progress-allocation-controls">
            <div class="progress-chip-group">
              ${ALLOCATION_MODES.map((mode) => `
                <button
                  class="button progress-chip ${modal.mode === mode.id ? "progress-chip-active" : ""}"
                  type="button"
                  data-progress-allocation-mode="${mode.id}"
                >${escapeHTML(mode.label)}</button>
              `).join("")}
            </div>
            <button id="progress-allocation-close" class="button progress-allocation-close-button" type="button">Close</button>
          </div>
          <div class="progress-allocation-content">
            <div class="progress-chart-shell progress-pie-shell">
              <canvas id="progress-allocation-chart" class="progress-chart-canvas" aria-label="Allocation pie chart"></canvas>
            </div>
            <aside class="progress-allocation-sidebar">
              ${renderAllocationLegend(allocationRows, modal.mode)}
              ${renderAllocationTotalRow(selectedPoint, allocationRows, modal.mode)}
            </aside>
          </div>
        </section>
      </div>
    `;
  }

  function renderAllocationLegend(rows, mode) {
    if (!rows.length) {
      return '<p class="subtle">No allocation rows for the selected snapshot.</p>';
    }

    const total = rows.reduce((sum, row) => sum + Number(row.value || 0), 0);

    return `
      <div class="progress-allocation-legend">
        <article class="progress-allocation-legend-row progress-allocation-legend-header">
          <span></span>
          <span class="progress-allocation-name">Name</span>
          <span class="progress-allocation-percent">%</span>
          <span class="progress-allocation-value">Value</span>
        </article>
        ${rows.map((row, index) => `
          <article class="progress-allocation-legend-row">
            <span class="progress-allocation-swatch" style="background:${mode === "category" ? resolveCategoryChartColor(row.name) : PROGRESS_CHART_COLORS[index % PROGRESS_CHART_COLORS.length]};"></span>
            <span class="progress-allocation-name">${escapeHTML(row.name)}</span>
            <span class="progress-allocation-percent ${row.value >= 0 ? "positive" : "negative"}">${escapeHTML(formatAllocationPercent(row.value, total))}</span>
            <span class="progress-allocation-value ${row.value >= 0 ? "positive" : "negative"}">${escapeHTML(formatTHB(row.value))}</span>
          </article>
        `).join("")}
      </div>
    `;
  }

  function renderAllocationTotalRow(selectedPoint, rows, mode) {
    return `
      <div class="progress-allocation-legend progress-allocation-total-list">
        <article class="progress-allocation-legend-row progress-allocation-total-row">
          <span class="progress-allocation-swatch progress-allocation-total-swatch"></span>
          <span class="progress-allocation-name">Total Value</span>
          <span class="progress-allocation-percent"></span>
          <span class="progress-allocation-value">${escapeHTML(formatTHB(resolveAllocationTotalValue(selectedPoint, rows, mode)))}</span>
        </article>
      </div>
    `;
  }

  function resolveAllocationTotalValue(point, rows, mode) {
    if (mode === "category") {
      return rows.reduce((sum, row) => sum + Number(row.value || 0), 0);
    }
    return Number(point.TotalCurrent || 0);
  }

  function formatAllocationPercent(value, total) {
    if (total === 0) {
      return "0.00%";
    }
    return `${((Number(value || 0) / total) * 100).toFixed(2)}%`;
  }

  function renderGoalsTable(goalEstimates) {
    if (!goalEstimates.length) {
      return '<p class="subtle">No goals yet. Add one to see projected reach dates from the selected range.</p>';
    }

    return `
      <div class="table-scroll">
        <table class="progress-goals-table">
          <thead>
            <tr>
              <th>Name</th>
              <th class="numeric">Target</th>
              <th>Target Date</th>
              <th>Estimate</th>
              <th class="numeric">Remaining</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            ${goalEstimates.map((goal) => `
              <tr>
                <td>${escapeHTML(goal.Name)}</td>
                <td class="numeric">${escapeHTML(formatTHB(goal.TargetAmount))}</td>
                <td>${goal.TargetDate ? escapeHTML(formatDateLabel(goal.TargetDate)) : '<span class="muted">N/A</span>'}</td>
                <td>${goal.EstimatedDate ? escapeHTML(formatDateLabel(goal.EstimatedDate)) : '<span class="muted">N/A</span>'}</td>
                <td class="numeric">${escapeHTML(formatTHB(goal.RemainingValue))}</td>
                <td><span class="metric-delta ${goalStatusToneClass(goal.Status)}">${escapeHTML(goal.Status)}</span></td>
                <td class="progress-goal-actions">
                  <button class="button button-small" type="button" data-goal-edit="${goal.GoalID}">Edit</button>
                  <button class="button button-small button-danger" type="button" data-goal-delete="${goal.GoalID}">Delete</button>
                </td>
              </tr>
            `).join("")}
          </tbody>
        </table>
      </div>
    `;
  }

  function renderProgressRow(point, selectedDate) {
    return `
      <tr class="${point.SnapshotDate === selectedDate ? "progress-summary-row-active" : ""}" data-progress-allocation-date="${escapeHTML(point.SnapshotDate)}">
        <td>${escapeHTML(formatDateLabel(point.SnapshotDate))}</td>
        <td class="numeric">${escapeHTML(formatTHB(point.TotalBought))}</td>
        <td class="numeric">${escapeHTML(formatTHB(point.TotalCurrent))}</td>
        <td class="numeric ${point.TotalProfit >= 0 ? "positive" : "negative"}">${escapeHTML(formatTHB(point.TotalProfit))}</td>
        <td class="numeric ${point.ProfitRate >= 0 ? "positive" : "negative"}">${escapeHTML((point.ProfitRate * 100).toFixed(2))}%</td>
        <td class="numeric">${escapeHTML(formatTHB(point.TotalCash))}</td>
        <td class="numeric">${escapeHTML(formatTHB(point.TotalNonCash))}</td>
        <td class="numeric">${escapeHTML((point.CashRatio * 100).toFixed(2))}%</td>
      </tr>
    `;
  }

  function renderGoalModal() {
    if (!state.goalModal) {
      return "";
    }

    if (state.goalModal.mode === "delete") {
      return `
        <div class="dialog-backdrop" data-goal-modal-close="backdrop">
          <section class="panel dialog-card progress-goal-dialog" role="dialog" aria-modal="true">
            <p class="eyebrow">Delete Goal</p>
            <h2>${escapeHTML(state.goalModal.form.name)}</h2>
            <p class="dialog-copy">This will soft delete the goal and remove it from future progress projections.</p>
            <div class="dialog-actions">
              <button id="progress-goal-cancel" class="button" type="button">Cancel</button>
              <button id="progress-goal-delete" class="button button-danger" type="button">Delete</button>
            </div>
          </section>
        </div>
      `;
    }

    const isEdit = state.goalModal.mode === "edit";
    return `
      <div class="dialog-backdrop" data-goal-modal-close="backdrop">
        <section class="panel dialog-card progress-goal-dialog" role="dialog" aria-modal="true">
          <p class="eyebrow">${isEdit ? "Edit Goal" : "Add Goal"}</p>
          <h2>${isEdit ? "Update Goal" : "Create Goal"}</h2>
          ${state.goalModal.error ? `<p class="error-copy error-banner">${escapeHTML(state.goalModal.error)}</p>` : ""}
          <div class="progress-goal-form">
            <label class="field-inline">
              <span class="field-label">Goal Name</span>
              <input id="progress-goal-name" class="form-input" type="text" value="${escapeHTML(state.goalModal.form.name)}" />
            </label>
            <label class="field-inline">
              <span class="field-label">Target Amount</span>
              <input id="progress-goal-amount" class="form-input" type="text" inputmode="decimal" value="${escapeHTML(state.goalModal.form.targetAmount)}" />
            </label>
            <label class="field-inline">
              <span class="field-label">Target Date</span>
              <div class="progress-goal-date-row">
                <div class="progress-goal-date-shell">
                  ${controls.renderDateControl({
                    id: "progress-goal-date",
                    value: state.goalModal.form.targetDate,
                    buttonClass: "snapshot-select snapshot-select-button progress-date-select progress-goal-date-button",
                    placeholder: "No target date",
                    allowClear: true,
                  })}
                </div>
                <button id="progress-goal-date-clear" class="button button-small" type="button">Clear</button>
              </div>
            </label>
          </div>
          <div class="dialog-actions">
            <button id="progress-goal-cancel" class="button" type="button">Cancel</button>
            <button id="progress-goal-save" class="button button-primary" type="button">${isEdit ? "Save Goal" : "Create Goal"}</button>
          </div>
        </section>
      </div>
    `;
  }

  function bindProgressPage(app) {
    for (const button of root.document.querySelectorAll("[data-progress-view]")) {
      button.addEventListener("click", () => {
        state.progressView = button.dataset.progressView;
        renderProgressPage(app);
      });
    }

    for (const inputID of ["progress-start-date", "progress-end-date"]) {
      const input = root.document.getElementById(inputID);
      if (!input) {
        continue;
      }
      input.addEventListener("change", async () => loadProgressPageWithSelectedFilter(app));
    }

    for (const button of root.document.querySelectorAll("[data-progress-chart-mode]")) {
      button.addEventListener("click", () => {
        state.progressChartMode = button.dataset.progressChartMode;
        renderProgressPage(app);
      });
    }

    const projectionSelect = root.document.getElementById("progress-projection-months");
    if (projectionSelect) {
      projectionSelect.addEventListener("change", () => {
        state.progressProjectionMonths = Number(projectionSelect.value);
        renderProgressPage(app);
      });
    }

    for (const button of root.document.querySelectorAll("[data-progress-allocation-mode]")) {
      button.addEventListener("click", () => {
        state.progressAllocationMode = button.dataset.progressAllocationMode;
        if (state.progressAllocationModal) {
          state.progressAllocationModal.mode = state.progressAllocationMode;
        }
        renderProgressPage(app);
      });
    }

    for (const row of root.document.querySelectorAll("[data-progress-allocation-date]")) {
      row.addEventListener("click", () => {
        state.progressAllocationDate = row.dataset.progressAllocationDate;
        state.progressAllocationModal = {
          snapshotDate: state.progressAllocationDate,
          mode: state.progressAllocationMode || "asset_type",
        };
        renderProgressPage(app);
      });
    }

    for (const element of root.document.querySelectorAll("[data-progress-allocation-close]")) {
      element.addEventListener("click", (event) => {
        if (event.target !== event.currentTarget) {
          return;
        }
        closeAllocationModal(app);
      });
    }

    const closeAllocationButton = root.document.getElementById("progress-allocation-close");
    if (closeAllocationButton) {
      closeAllocationButton.addEventListener("click", () => closeAllocationModal(app));
    }

    const addGoalButton = root.document.getElementById("progress-add-goal");
    if (addGoalButton) {
      addGoalButton.addEventListener("click", () => {
        state.goalModal = { mode: "create", form: buildGoalModalForm(), error: "" };
        renderProgressPage(app);
      });
    }

    for (const button of root.document.querySelectorAll("[data-goal-edit]")) {
      button.addEventListener("click", () => openGoalModal(app, Number(button.dataset.goalEdit), "edit"));
    }
    for (const button of root.document.querySelectorAll("[data-goal-delete]")) {
      button.addEventListener("click", () => openGoalModal(app, Number(button.dataset.goalDelete), "delete"));
    }

    bindGoalModal(app);
  }

  function openGoalModal(app, goalID, mode) {
    const goal = (state.progressPage?.GoalEstimates || []).find((item) => item.GoalID === goalID);
    if (!goal) {
      return;
    }
    state.goalModal = {
      mode,
      form: buildGoalModalForm(goal),
      error: "",
    };
    renderProgressPage(app);
  }

  function bindGoalModal(app) {
    if (!state.goalModal) {
      return;
    }

    for (const element of root.document.querySelectorAll("[data-goal-modal-close]")) {
      element.addEventListener("click", (event) => {
        if (event.target !== event.currentTarget) {
          return;
        }
        closeGoalModal(app);
      });
    }

    const cancelButton = root.document.getElementById("progress-goal-cancel");
    if (cancelButton) {
      cancelButton.addEventListener("click", () => closeGoalModal(app));
      cancelButton.focus();
    }

    const nameInput = root.document.getElementById("progress-goal-name");
    if (nameInput) {
      nameInput.addEventListener("input", () => {
        state.goalModal.form.name = nameInput.value;
      });
    }

    const amountInput = root.document.getElementById("progress-goal-amount");
    if (amountInput) {
      amountInput.addEventListener("input", () => {
        amountInput.value = sanitizePartialDecimal(amountInput.value);
        state.goalModal.form.targetAmount = amountInput.value;
      });
    }

    const dateInput = root.document.getElementById("progress-goal-date");
    if (dateInput) {
      dateInput.addEventListener("change", () => {
        state.goalModal.form.targetDate = dateInput.value;
      });
    }

    const clearDateButton = root.document.getElementById("progress-goal-date-clear");
    if (clearDateButton) {
      clearDateButton.addEventListener("click", () => {
        state.goalModal.form.targetDate = "";
        controls.setControlValue("progress-goal-date", "");
      });
    }

    const saveButton = root.document.getElementById("progress-goal-save");
    if (saveButton) {
      saveButton.addEventListener("click", async () => submitGoalForm(app));
    }

    const deleteButton = root.document.getElementById("progress-goal-delete");
    if (deleteButton) {
      deleteButton.addEventListener("click", async () => deleteGoal(app));
    }
  }

  function renderProgressCharts(page, view, chartMode, allocationMode, allocationDate, projectionMonths) {
    if (!root.document) {
      return;
    }
    if (view === "trend") {
      const canvas = root.document.getElementById("progress-trend-chart");
      if (canvas) {
        trendChartInstance = renderChart(canvas, buildTrendChartConfig(page, chartMode, projectionMonths));
      }
    }
    if (view === "summary") {
      const allocationCanvas = root.document.getElementById("progress-allocation-chart");
      if (allocationCanvas) {
        allocationChartInstance = renderChart(allocationCanvas, buildAllocationChartConfig(page, allocationDate, allocationMode));
      }
    }
  }

  function renderChart(canvas, config) {
    const Chart = root.Chart;
    if (!canvas || !Chart) {
      return null;
    }
    return new Chart(canvas, config);
  }

  function destroyProgressCharts() {
    if (trendChartInstance && typeof trendChartInstance.destroy === "function") {
      trendChartInstance.destroy();
    }
    if (allocationChartInstance && typeof allocationChartInstance.destroy === "function") {
      allocationChartInstance.destroy();
    }
    trendChartInstance = null;
    allocationChartInstance = null;
  }

  function buildGoalModalForm(goal) {
    return {
      id: Number(goal?.GoalID || goal?.ID || 0),
      name: goal?.Name || "",
      targetAmount: goal ? Number(goal.TargetAmount || 0).toFixed(2) : "",
      targetDate: goal?.TargetDate || "",
    };
  }

  function buildGoalPayload(form) {
    return {
      Name: form.name.trim(),
      TargetAmount: parseEditableNumber(form.targetAmount),
      TargetDate: form.targetDate || "",
    };
  }

  function closeGoalModal(app) {
    state.goalModal = null;
    renderProgressPage(app);
  }

  function closeAllocationModal(app) {
    state.progressAllocationModal = null;
    renderProgressPage(app);
  }

  async function loadProgressPageWithSelectedFilter(app) {
    const startDate = root.document.getElementById("progress-start-date")?.value || "";
    const endDate = root.document.getElementById("progress-end-date")?.value || "";
    await runTransition(async () => {
      state.progressFilter = { StartDate: startDate, EndDate: endDate };
      await app.loadProgressPage(state.progressFilter);
    });
  }

  async function submitGoalForm(app) {
    await runTransition(async () => {
      try {
        const backend = shared.resolveBackend();
        const payload = buildGoalPayload(state.goalModal.form);
        if (state.goalModal.mode === "edit") {
          await backend.UpdateGoal({ ID: state.goalModal.form.id, ...payload });
        } else {
          await backend.CreateGoal(payload);
        }

        state.goalModal = null;
        await app.loadProgressPage(state.progressFilter || state.progressPage?.Filter || {});
      } catch (error) {
        state.goalModal.error = error?.message || String(error);
        renderProgressPage(app);
      }
    });
  }

  async function deleteGoal(app) {
    await runTransition(async () => {
      try {
        const backend = shared.resolveBackend();
        await backend.DeleteGoal({ ID: state.goalModal.form.id });
        state.goalModal = null;
        await app.loadProgressPage(state.progressFilter || state.progressPage?.Filter || {});
      } catch (error) {
        state.goalModal = null;
        renderProgressPage(app);
        rootErrorBanner(error?.message || String(error));
      }
    });
  }

  function syncProgressOverlayState() {
    if (!root.document?.body) {
      return;
    }
    root.document.body.classList.toggle("overlay-open", Boolean(state.goalModal || state.progressAllocationModal));
  }

  function renderProjectionSelector(projectionMonths) {
    return `
      <label class="field-inline progress-projection-control">
        <span class="field-label">Projection Time</span>
        ${controls.renderSelectControl({
          id: "progress-projection-months",
          value: projectionMonths,
          buttonClass: "snapshot-select progress-date-select",
          ariaLabel: "Projection time",
          options: PROJECTION_OPTIONS.map((months) => ({
            value: months,
            label: `${months} months`,
          })),
        })}
      </label>
    `;
  }

  return {
    buildGoalModalForm,
    buildGoalPayload,
    closeAllocationModal,
    formatAllocationPercent,
    loadProgressPageWithSelectedFilter,
    renderAllocationModal,
    renderAllocationTotalRow,
    renderProgressHeroActions,
    renderProgressPage,
    renderProjectionSelector,
    resolveAllocationTotalValue,
  };
}));
