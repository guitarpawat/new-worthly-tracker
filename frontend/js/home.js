(function initHome(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const controls = root.WorthlyControls || (typeof require !== "undefined" ? require("./controls.js") : null);
  const home = factory(shared, controls, root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = home;
  }
  root.WorthlyHome = home;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildHome(shared, controls, root) {
  const {
    clickIfEnabled,
    escapeHTML,
    formatDateLabel,
    formatDeltaPercent,
    formatDeltaTHB,
    formatPercent,
    formatTHB,
    isEnabledElement,
    isGeneralTypingTarget,
    logHomeAction,
    renderAppTitle,
    rootErrorBanner,
    runTransition,
    scrollHomePage,
    snapshotCount,
    state,
  } = shared;

  function syncHomeOverlayState() {
    if (!root.document?.body) {
      return;
    }

    const homeMenu = root.document.getElementById("home-menu");
    const hasOverlay = Boolean(state.deleteDialog) || Boolean(homeMenu?.open);
    root.document.body.classList.toggle("overlay-open", hasOverlay);
  }

  function renderHomePage(page, app) {
    const appRoot = root.document.getElementById("app");
    if (!page || !page.HasSnapshot) {
      appRoot.innerHTML = `
        <main class="app-layout">
          <section class="hero">
            <div>
              <p class="eyebrow">Worthly Tracker</p>
              <p class="hero-copy">No snapshots yet. Start a new snapshot to add your first assets and records.</p>
            </div>
          </section>
          <section class="panel onboarding">
            <p class="eyebrow">Onboarding</p>
            <h2>Set up your data</h2>
            <p>Create your first snapshot to start entering data. You can add existing assets, create new assets, and create new asset types from that page.</p>
            <p>You can also insert demo data to explore the latest snapshot, grouped holdings, and previous snapshot comparison immediately.</p>
            ${state.onboardingError ? `<p class="error-copy error-banner">${escapeHTML(state.onboardingError)}</p>` : ""}
            <div class="actions">
              <button id="insert-demo-data-button" class="button button-primary" type="button">Insert Demo Data</button>
              <button id="new-onboarding-button" class="button" type="button">New</button>
            </div>
          </section>
        </main>
      `;

      const insertButton = root.document.getElementById("insert-demo-data-button");
      if (insertButton) {
        insertButton.addEventListener("click", async () => {
          await runTransition(async () => {
            insertButton.disabled = true;
            state.onboardingError = "";
            void logHomeAction("insert_demo_data_click", state.offset);

            try {
              const backend = shared.resolveBackend();
              await backend.InsertDemoData();
              state.offset = 0;
              await app.loadHomePage();
            } catch (error) {
              state.onboardingError = error?.message || String(error);
              await app.loadHomePage();
            } finally {
              insertButton.disabled = false;
            }
          });
        });
      }

      const newOnboardingButton = root.document.getElementById("new-onboarding-button");
      if (newOnboardingButton) {
        newOnboardingButton.addEventListener("click", async () => {
          await runTransition(async () => {
            void logHomeAction("new_snapshot_click", state.offset);
            await app.loadNewSnapshotPage();
          });
        });
      }

      syncHomeOverlayState();

      return;
    }

    state.onboardingError = "";
    const metrics = buildMetrics(page);
    const groups = page.Groups.map((group) => `
      <section class="panel table-panel">
        <div class="table-header">
          <div>
            <h2>${escapeHTML(group.AssetTypeName)}</h2>
          </div>
          <div class="home-group-header-side">
            ${renderGroupSummary(group)}
            <span class="home-group-count">${group.Summary?.AssetCount || 0} asset(s)</span>
          </div>
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
            </colgroup>
            <thead>
              <tr>
                <th>Name</th>
                <th>Broker</th>
                <th class="numeric">Bought Price</th>
                <th class="numeric">Current Price</th>
                <th class="numeric">Profit</th>
                <th class="numeric">% Profit</th>
                <th>Notes</th>
              </tr>
            </thead>
            <tbody>
              ${group.Rows.map(renderRow).join("")}
            </tbody>
          </table>
        </div>
      </section>
    `).join("");

    appRoot.innerHTML = `
      <main class="app-layout ${state.deleteDialog ? "app-overlay-open" : ""}">
        <section class="hero">
          <div>
            <h1>${renderAppTitle()}</h1>
          </div>
          <div class="actions">
            <button id="previous-button" class="button" type="button" ${page.CanNavigateBack ? "" : "disabled"}>Prev</button>
            ${renderSnapshotSelector(page.SnapshotOptions || [])}
            <button id="next-button" class="button" type="button" ${page.CanNavigateForward ? "" : "disabled"}>Next</button>
            <button id="latest-button" class="button" type="button" ${state.offset === 0 ? "disabled" : ""}>Latest</button>
            <button id="new-button" class="button button-primary" type="button">New</button>
            <button id="edit-button" class="button button-secondary" type="button">Edit</button>
            ${renderHomeOverflowMenu(page)}
          </div>
        </section>
        <section class="panel summary-strip">
          <div class="summary-strip-grid">
            ${metrics.map(renderMetric).join("")}
          </div>
        </section>
        ${groups}
      </main>
      ${renderDeleteSnapshotDialog(page)}
    `;

    bindHomeNavigation(page, app);
    syncHomeOverlayState();
  }

  function renderHomeOverflowMenu(page) {
    const actions = buildHomeOverflowActions(page);
    return `
      <details id="home-menu" class="home-menu">
        <summary class="button home-menu-trigger" aria-label="More actions">
          <span class="home-menu-icon" aria-hidden="true">
            <span></span>
            <span></span>
            <span></span>
          </span>
        </summary>
        <div class="home-menu-panel">
          ${actions.map((action) => `
            <button
              class="home-menu-item ${action.tone === "danger" ? "home-menu-item-danger" : ""}"
              type="button"
              data-menu-action="${action.id}"
              ${action.disabled ? "disabled" : ""}
            >
              <span class="home-menu-item-label">${escapeHTML(action.label)}</span>
              ${action.note ? `<span class="home-menu-item-note">${escapeHTML(action.note)}</span>` : ""}
            </button>
          `).join("")}
        </div>
      </details>
    `;
  }

  function buildHomeOverflowActions(page) {
    return [
      {
        id: "asset_management",
        label: "Asset Management",
        note: "Manage assets and asset types",
        disabled: false,
      },
      {
        id: "progress",
        label: "Progress & Goals",
        note: "Summary table, chart, and goal projection",
        disabled: false,
      },
      {
        id: "delete",
        label: "Delete Snapshot",
        note: page?.HasSnapshot ? formatDateLabel(page.SnapshotDate) : "",
        disabled: !page?.HasSnapshot,
        tone: "danger",
      },
    ];
  }

  function renderDeleteSnapshotDialog(page) {
    if (!state.deleteDialog || !page?.HasSnapshot) {
      return "";
    }

    const dialog = buildDeleteSnapshotDialogModel(page);
    return `
      <div class="dialog-backdrop" data-delete-dialog-close="backdrop">
        <section
          class="panel dialog-card"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-dialog-title"
          aria-describedby="delete-dialog-body"
        >
          <p class="eyebrow">Delete Snapshot</p>
          <h2 id="delete-dialog-title">${escapeHTML(dialog.title)}</h2>
          <p id="delete-dialog-body" class="dialog-copy">${escapeHTML(dialog.body)}</p>
          <div class="dialog-actions">
            <button id="delete-dialog-cancel" class="button" type="button">Cancel</button>
            <button id="delete-dialog-confirm" class="button button-danger" type="button">Delete</button>
          </div>
        </section>
      </div>
    `;
  }

  function buildDeleteSnapshotDialogModel(page) {
    const snapshotDate = page?.SnapshotDate ? formatDateLabel(page.SnapshotDate) : "";
    return {
      title: snapshotDate || "Delete snapshot",
      body: snapshotDate
        ? `This will soft delete snapshot ${snapshotDate} and its record rows. You can still keep older history.`
        : "This will soft delete the selected snapshot and its record rows.",
    };
  }

  function bindHomeNavigation(page, app) {
    const previousButton = root.document.getElementById("previous-button");
    if (previousButton) {
      previousButton.addEventListener("click", async () => {
        await runTransition(async () => {
          const count = snapshotCount();
          const nextOffset = state.offset + 1;
          if (count > 0 && nextOffset >= count) {
            return;
          }
          void logHomeAction("older_snapshot_click", nextOffset);
          state.offset = nextOffset;
          await app.loadHomePage();
        });
      });
    }

    const nextButton = root.document.getElementById("next-button");
    if (nextButton) {
      nextButton.addEventListener("click", async () => {
        if (state.offset === 0) {
          return;
        }

        await runTransition(async () => {
          const nextOffset = state.offset - 1;
          void logHomeAction("newer_snapshot_click", nextOffset);
          state.offset = nextOffset;
          await app.loadHomePage();
        });
      });
    }

    const latestButton = root.document.getElementById("latest-button");
    if (latestButton) {
      latestButton.addEventListener("click", async () => {
        if (state.offset === 0) {
          return;
        }

        await runTransition(async () => {
          void logHomeAction("latest_snapshot_click", 0);
          state.offset = 0;
          await app.loadHomePage();
        });
      });
    }

    const editButton = root.document.getElementById("edit-button");
    if (editButton) {
      editButton.addEventListener("click", async () => {
        await runTransition(async () => {
          void logHomeAction("edit_snapshot_click", state.offset);
          await app.loadEditSnapshotPage();
        });
      });
    }

    const newButton = root.document.getElementById("new-button");
    if (newButton) {
      newButton.addEventListener("click", async () => {
        await runTransition(async () => {
          void logHomeAction("new_snapshot_click", state.offset);
          await app.loadNewSnapshotPage();
        });
      });
    }

    const homeMenu = root.document.getElementById("home-menu");
    for (const button of root.document.querySelectorAll("[data-menu-action]")) {
      button.addEventListener("click", async (event) => {
        const action = event.currentTarget.dataset.menuAction;
        if (!action) {
          return;
        }

        if (action === "asset_management") {
          closeHomeOverflowMenu();
          await runTransition(async () => {
            void logHomeAction("asset_management_click", state.offset);
            await app.loadAssetManagementPage();
          });
          return;
        }

        if (action === "progress") {
          closeHomeOverflowMenu();
          await runTransition(async () => {
            void logHomeAction("progress_click", state.offset);
            await app.loadProgressPage();
          });
          return;
        }

        if (action !== "delete") {
          closeHomeOverflowMenu();
          rootErrorBanner("This page is planned but not available in the current MVP.");
          return;
        }

        openDeleteSnapshotDialog(page, app);
      });
    }

    if (homeMenu) {
      homeMenu.addEventListener("toggle", () => {
        const appLayout = root.document.querySelector(".app-layout");
        if (appLayout) {
          appLayout.classList.toggle("app-overlay-open", homeMenu.open);
        }
        syncHomeOverlayState();

        if (!homeMenu.open) {
          return;
        }

        for (const menu of root.document.querySelectorAll(".home-menu")) {
          if (menu !== homeMenu) {
            menu.open = false;
          }
        }
      });
    }

    bindSnapshotSelector(page, app);
    bindDeleteSnapshotDialog(page, app);
  }

  function closeHomeOverflowMenu() {
    const menu = root.document.getElementById("home-menu");
    if (menu) {
      menu.open = false;
    }
    const appLayout = root.document.querySelector(".app-layout");
    if (appLayout) {
      appLayout.classList.remove("app-overlay-open");
    }
    syncHomeOverlayState();
  }

  function openDeleteSnapshotDialog(page, app) {
    state.deleteDialog = {
      snapshotID: page.SnapshotID,
      snapshotDate: page.SnapshotDate,
      offset: state.offset,
    };
    closeHomeOverflowMenu();
    renderHomePage(page, app);
  }

  function closeDeleteSnapshotDialog(app) {
    if (!state.deleteDialog || !state.currentPage) {
      state.deleteDialog = null;
      syncHomeOverlayState();
      return;
    }

    state.deleteDialog = null;
    renderHomePage(state.currentPage, app);
  }

  function bindDeleteSnapshotDialog(page, app) {
    if (!state.deleteDialog) {
      return;
    }

    const cancelButton = root.document.getElementById("delete-dialog-cancel");
    if (cancelButton) {
      cancelButton.addEventListener("click", () => {
        closeDeleteSnapshotDialog(app);
      });
      cancelButton.focus();
    }

    const confirmButton = root.document.getElementById("delete-dialog-confirm");
    if (confirmButton) {
      confirmButton.addEventListener("click", async () => {
        await confirmDeleteSnapshot(page, app);
      });
    }

    for (const element of root.document.querySelectorAll("[data-delete-dialog-close]")) {
      element.addEventListener("click", (event) => {
        if (event.target !== event.currentTarget) {
          return;
        }
        closeDeleteSnapshotDialog(app);
      });
    }
  }

  async function confirmDeleteSnapshot(page, app) {
    await runTransition(async () => {
      void logHomeAction("delete_snapshot_click", state.offset);
      try {
        const backend = shared.resolveBackend();
        const result = await backend.DeleteSnapshot({
          SnapshotID: page.SnapshotID,
          Offset: state.offset,
        });
        state.deleteDialog = null;
        state.offset = result.Offset;
        await app.loadHomePage();
      } catch (error) {
        state.deleteDialog = null;
        renderHomePage(page, app);
        rootErrorBanner(error?.message || String(error));
      }
    });
  }

  function bindSnapshotSelector(page, app) {
    const options = page.SnapshotOptions || [];
    if (options.length === 0) {
      return;
    }

    const select = root.document.getElementById("snapshot-select");
    if (select) {
      select.addEventListener("change", async (event) => {
        await runTransition(async () => {
          const nextOffset = Number(event.target.value);
          const count = snapshotCount();
          if (count > 0 && (nextOffset < 0 || nextOffset >= count)) {
            return;
          }
          void logHomeAction("snapshot_select_change", nextOffset);
          state.offset = nextOffset;
          await app.loadHomePage();
        });
      });
    }
  }

  function handleHomeKeydown(event, app) {
    if (state.deleteDialog) {
      const dialogAction = resolveDeleteDialogShortcutAction({
        key: event.key,
        isEditableTarget: isGeneralTypingTarget(event.target),
      });
      if (dialogAction) {
        event.preventDefault();
        if (dialogAction === "cancel") {
          closeDeleteSnapshotDialog(app);
        }
        return true;
      }
    }

    const action = resolveHomeShortcutAction({
      key: event.key,
      canNavigateBack: Boolean(state.currentPage?.CanNavigateBack),
      canNavigateForward: Boolean(state.currentPage?.CanNavigateForward),
      canEditSnapshot: Boolean(state.currentPage?.HasSnapshot),
      canCreateSnapshot: isEnabledElement(root.document.getElementById("new-button")),
      canJumpToLatest: state.offset > 0,
      isEditableTarget: isGeneralTypingTarget(event.target),
    });
    if (!action) {
      return false;
    }

    event.preventDefault();
    if (action === "previous") {
      clickIfEnabled("previous-button");
      return true;
    }
    if (action === "next") {
      clickIfEnabled("next-button");
      return true;
    }
    if (action === "edit") {
      clickIfEnabled("edit-button");
      return true;
    }
    if (action === "new") {
      clickIfEnabled("new-button");
      return true;
    }
    if (action === "latest") {
      clickIfEnabled("latest-button");
      return true;
    }
    if (action === "scroll_up") {
      scrollHomePage(-240);
      return true;
    }
    if (action === "scroll_down") {
      scrollHomePage(240);
      return true;
    }

    return false;
  }

  function resolveHomeShortcutAction({
    key,
    canNavigateBack,
    canNavigateForward,
    canEditSnapshot,
    canCreateSnapshot,
    canJumpToLatest,
    isEditableTarget,
  }) {
    if (isEditableTarget) {
      return null;
    }

    switch (key) {
      case "ArrowLeft":
      case "a":
      case "A":
        return canNavigateBack ? "previous" : null;
      case "ArrowRight":
      case "d":
      case "D":
        return canNavigateForward ? "next" : null;
      case "e":
      case "E":
        return canEditSnapshot ? "edit" : null;
      case "+":
      case "Add":
      case "n":
      case "N":
        return canCreateSnapshot ? "new" : null;
      case "l":
      case "L":
        return canJumpToLatest ? "latest" : null;
      case "w":
      case "W":
        return "scroll_up";
      case "s":
      case "S":
        return "scroll_down";
      default:
        return null;
    }
  }

  function resolveDeleteDialogShortcutAction({ key, isEditableTarget }) {
    if (isEditableTarget) {
      return null;
    }

    if (key === "Escape") {
      return "cancel";
    }

    return null;
  }

  function shouldCloseHomeMenuOnClick({ currentView, isMenuOpen, clickedInsideMenu }) {
    if (currentView !== "home") {
      return false;
    }
    if (!isMenuOpen) {
      return false;
    }
    return !clickedInsideMenu;
  }

  function buildMetrics(page) {
    const comparison = page.Comparison || null;
    return [
      metric("Bought Price", formatTHB(page.Summary.TotalBought), comparison?.BoughtChange),
      metric("Current Profit", formatTHB(page.Summary.TotalProfit), comparison?.ProfitChange),
      metric("% Profit", formatPercent(page.Summary.TotalProfitRate), comparison?.ProfitRateChange, true),
      metric("Total Cash", formatTHB(page.Summary.TotalCash), comparison?.CashChange),
      metric("Total Non Cash", formatTHB(page.Summary.TotalNonCash), comparison?.NonCashChange),
      metric("% Cash / Total", formatPercent(page.Summary.CashRatio), comparison?.CashRatioChange, true),
      metric("Current Net Worth", formatTHB(page.Summary.TotalCurrent), comparison?.CurrentChange),
    ];
  }

  function metric(label, value, delta, isPercent = false) {
    return { label, value, delta, isPercent };
  }

  function renderMetric(metricData) {
    const deltaText = metricData.delta === undefined || metricData.delta === null
      ? ""
      : metricData.isPercent
        ? formatDeltaPercent(metricData.delta)
        : formatDeltaTHB(metricData.delta);

    const deltaClass = metricData.delta > 0
      ? "delta-positive"
      : metricData.delta < 0
        ? "delta-negative"
        : "delta-neutral";

    return `
      <div class="summary-item">
        <p class="metric-label">${escapeHTML(metricData.label)}</p>
        <p class="metric-value">${escapeHTML(metricData.value)}</p>
        ${deltaText ? `<p class="metric-delta ${deltaClass}">${escapeHTML(deltaText)}</p>` : ""}
      </div>
    `;
  }

  function renderSnapshotSelector(options) {
    if (!options || options.length === 0) {
      return "";
    }

    const start = Math.max(state.offset - 12, 0);
    const end = Math.min(state.offset + 13, options.length);
    const visibleOptions = options.slice(start, end);

    return `
      <div class="snapshot-selector-inline">
        ${controls.renderSelectControl({
          id: "snapshot-select",
          value: state.offset,
          ariaLabel: "Snapshot date",
          options: visibleOptions.map((option) => ({
            value: option.Offset,
            label: option.Label,
          })),
        })}
      </div>
    `;
  }

  function renderGroupSummary(group) {
    const summary = group.Summary || {};
    return `
      <span class="home-group-count">${escapeHTML(formatTHB(summary.TotalCurrent || 0))}</span>
    `;
  }

  function renderRow(row) {
    const boughtValue = row.ProfitApplicable ? formatTHB(row.BoughtPrice) : "N/A";
    const profitValue = row.ProfitApplicable ? formatTHB(row.Profit) : "N/A";
    const percentValue = row.ProfitApplicable ? formatPercent(row.ProfitPercent) : "N/A";
    const tone = row.ProfitApplicable
      ? row.Profit > 0
        ? "positive"
        : row.Profit < 0
          ? "negative"
          : "neutral"
      : "muted";

    return `
      <tr>
        <td>${escapeHTML(row.AssetName)}</td>
        <td>${escapeHTML(row.Broker || "-")}</td>
        <td class="numeric ${row.ProfitApplicable ? "" : "muted"}">${boughtValue}</td>
        <td class="numeric">${formatTHB(row.CurrentPrice)}</td>
        <td class="numeric ${tone}">${profitValue}</td>
        <td class="numeric ${tone}">${percentValue}</td>
        <td>${escapeHTML(row.Remarks || "-")}</td>
      </tr>
    `;
  }

  return {
    buildDeleteSnapshotDialogModel,
    buildHomeOverflowActions,
    handleHomeKeydown,
    renderHomePage,
    resolveDeleteDialogShortcutAction,
    resolveHomeShortcutAction,
    shouldCloseHomeMenuOnClick,
    scrollHomePage,
  };
}));
