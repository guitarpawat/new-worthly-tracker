const shared = typeof module !== "undefined" && module.exports
  ? require("./js/shared.js")
  : window.WorthlyShared;
const controls = typeof module !== "undefined" && module.exports
  ? require("./js/controls.js")
  : window.WorthlyControls;
const home = typeof module !== "undefined" && module.exports
  ? require("./js/home.js")
  : window.WorthlyHome;
const snapshotForm = typeof module !== "undefined" && module.exports
  ? require("./js/snapshot_form.js")
  : window.WorthlySnapshotForm;
const snapshotFormLogic = typeof module !== "undefined" && module.exports
  ? require("./js/snapshot_form_logic.js")
  : window.WorthlySnapshotFormLogic;
const assetManagement = typeof module !== "undefined" && module.exports
  ? require("./js/asset_management.js")
  : window.WorthlyAssetManagement;
const assetReorder = typeof module !== "undefined" && module.exports
  ? require("./js/asset_reorder.js")
  : window.WorthlyAssetReorder;
const progressChartLogic = typeof module !== "undefined" && module.exports
  ? require("./js/progress_chart_logic.js")
  : window.WorthlyProgressChartLogic;
const progressGoals = typeof module !== "undefined" && module.exports
  ? require("./js/progress_goals.js")
  : window.WorthlyProgressGoals;

const { renderErrorState, resolveBackend, state } = shared;

function appContext() {
  return {
    loadAssetManagementPage,
    loadEditSnapshotPage,
    loadHomePage,
    loadNewSnapshotPage,
    loadProgressPage,
  };
}

async function loadHomePage() {
  state.currentView = "home";
  state.deleteDialog = null;
  const root = document.getElementById("app");
  root.innerHTML = '<div class="loading-state">Loading latest snapshot...</div>';

  try {
    const backend = resolveBackend();
    const page = await backend.GetHomePage(state.offset);
    state.currentPage = page;
    home.renderHomePage(page, appContext());
  } catch (error) {
    root.innerHTML = renderErrorState(
      "Unable to load records",
      error?.message || String(error),
    );
  }
}

async function loadEditSnapshotPage() {
  await loadSnapshotFormPage({
    title: "Loading snapshot editor...",
    errorTitle: "Unable to load snapshot editor",
    load: (backend) => backend.GetEditSnapshotPage(state.offset),
  });
}

async function loadNewSnapshotPage() {
  await loadSnapshotFormPage({
    title: "Loading new snapshot...",
    errorTitle: "Unable to load new snapshot page",
    load: (backend) => backend.GetNewSnapshotPage(),
  });
}

async function loadAssetManagementPage(options = {}) {
  state.currentView = "asset_management";
  state.deleteDialog = null;
  state.assetManagementModal = null;
  state.snapshotAssetModal = null;
  const root = document.getElementById("app");
  root.innerHTML = '<div class="loading-state">Loading asset management...</div>';

  try {
    const backend = resolveBackend();
    const page = await backend.GetAssetManagementPage();
    const view = assetManagement.resolveAssetManagementView(options);
    const selectedAssetTypeID = options.selectedAssetTypeID || 0;
    const selectedAssetID = options.selectedAssetID || 0;
    state.assetManagementPage = page;
    state.assetManagementView = view;
    state.assetTypeError = "";
    state.assetError = "";
    state.assetTypeForm = view === "create_asset_type"
      ? assetManagement.buildEmptyAssetTypeForm()
      : assetManagement.buildAssetTypeFormState(page, selectedAssetTypeID);
    state.assetForm = view === "create_asset"
      ? assetManagement.buildEmptyAssetForm(page)
      : assetManagement.buildAssetFormState(page, selectedAssetID);
    assetManagement.renderAssetManagementPage(appContext());
  } catch (error) {
    root.innerHTML = renderErrorState(
      "Unable to load asset management",
      error?.message || String(error),
    );
  }
}

async function loadProgressPage(filter = state.progressFilter || {}) {
  state.currentView = "progress";
  state.deleteDialog = null;
  state.goalModal = null;
  const root = document.getElementById("app");
  root.innerHTML = '<div class="loading-state">Loading progress and goals...</div>';

  try {
    const backend = resolveBackend();
    const page = await backend.GetProgressPage(filter);
    state.progressPage = page;
    state.progressFilter = page.Filter || filter || {};
    if (!page.AllocationSnapshots?.some((item) => item.SnapshotDate === state.progressAllocationDate)) {
      state.progressAllocationDate = page.AllocationSnapshots?.[page.AllocationSnapshots.length - 1]?.SnapshotDate || "";
    }
    progressGoals.renderProgressPage(appContext());
  } catch (error) {
    root.innerHTML = renderErrorState(
      "Unable to load progress and goals",
      error?.message || String(error),
    );
  }
}

async function loadSnapshotFormPage(config) {
  const root = document.getElementById("app");
  root.innerHTML = `<div class="loading-state">${shared.escapeHTML(config.title)}</div>`;

  try {
    const backend = resolveBackend();
    const page = await config.load(backend);
    snapshotForm.initializeDraft(page);
    snapshotForm.renderEditPage(appContext());
  } catch (error) {
    root.innerHTML = renderErrorState(
      config.errorTitle,
      error?.message || String(error),
    );
  }
}

function handleGlobalKeydown(event) {
  if (controls.handleControlKeydown(event)) {
    return;
  }

  if (event.defaultPrevented || event.ctrlKey || event.metaKey || event.altKey) {
    return;
  }

  if (state.currentView === "home") {
    home.handleHomeKeydown(event, appContext());
    return;
  }

  if (state.currentView === "edit") {
    snapshotForm.handleEditKeydown(event, appContext());
    return;
  }

  if (state.currentView === "asset_management") {
    assetManagement.handleAssetManagementKeydown(event, appContext());
  }
}

if (typeof window !== "undefined") {
  document.addEventListener("click", async (event) => {
    controls.handleControlClick(event);

    const target = event.target instanceof Element ? event.target : null;
    const homeMenu = document.getElementById("home-menu");
    if (home.shouldCloseHomeMenuOnClick({
      currentView: state.currentView,
      isMenuOpen: Boolean(homeMenu?.open),
      clickedInsideMenu: Boolean(target?.closest("#home-menu")),
    })) {
      homeMenu.open = false;
      document.body.classList.remove("overlay-open");
      const appLayout = document.querySelector(".app-layout");
      if (appLayout) {
        appLayout.classList.remove("app-overlay-open");
      }
    }

    const homeLink = target
      ? target.closest("[data-app-home-link]")
      : null;
    if (!homeLink) {
      return;
    }

    event.preventDefault();
    await loadHomePage();
  });

  document.addEventListener("keydown", handleGlobalKeydown);

  window.addEventListener("beforeunload", (event) => {
    if (state.currentView !== "edit" || !state.hasUnsavedChanges) {
      return;
    }

    event.preventDefault();
    event.returnValue = "";
  });

  window.addEventListener("DOMContentLoaded", async () => {
    await loadHomePage();
  });
}

if (typeof module !== "undefined" && module.exports) {
  module.exports = {
    ...shared,
    ...controls,
    ...home,
    ...snapshotForm,
    ...snapshotFormLogic,
    ...assetManagement,
    ...assetReorder,
    ...progressChartLogic,
    ...progressGoals,
  };
}
