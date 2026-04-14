(function initShared(root, factory) {
  const shared = factory(root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = shared;
  }
  root.WorthlyShared = shared;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildShared(root) {
  const state = {
    offset: 0,
    currentPage: null,
    currentView: "home",
    isTransitioning: false,
    onboardingError: "",
    editError: "",
    editDraft: null,
    editOriginalSignature: "",
    removedRows: [],
    hasUnsavedChanges: false,
    deleteDialog: null,
    assetManagementPage: null,
    assetManagementView: "create_asset",
    assetTypeForm: null,
    assetForm: null,
    assetTypeError: "",
    assetError: "",
    assetManagementModal: null,
    snapshotAssetModal: null,
    progressPage: null,
    progressFilter: null,
    progressView: "trend",
    progressChartMode: "net_worth",
    progressProjectionMonths: 6,
    progressAllocationMode: "asset_type",
    progressAllocationDate: "",
    progressAllocationModal: null,
    goalModal: null,
  };

  const EDIT_TAB_INDEX = {
    date: 1,
    current: 2,
    save: 3,
  };

  function renderErrorState(title, message) {
    return `
      <section class="panel panel-error">
        <p class="eyebrow">Error</p>
        <h1>${escapeHTML(title)}</h1>
        <p>${escapeHTML(message)}</p>
      </section>
    `;
  }

  function renderAppTitle(secondary = "") {
    if (!secondary) {
      return '<button class="brand-link" type="button" data-app-home-link>Worthly Tracker</button>';
    }

    return `<button class="brand-link" type="button" data-app-home-link>Worthly Tracker</button> <span class="heading-separator">&gt;</span> <span class="heading-secondary">${escapeHTML(secondary)}</span>`;
  }

  async function logHomeAction(action, offset = state.offset) {
    try {
      const backend = resolveBackend();
      if (typeof backend.LogHomeAction === "function") {
        await backend.LogHomeAction(action, offset);
      }
    } catch (_) {
      // Logging must not block UI interaction.
    }
  }

  async function runTransition(work) {
    if (state.isTransitioning) {
      return;
    }

    state.isTransitioning = true;
    try {
      await work();
    } finally {
      state.isTransitioning = false;
    }
  }

  function snapshotCount() {
    return state.currentPage?.SnapshotOptions?.length || 0;
  }

  function resolveBackend() {
    const candidates = [
      root?.go?.app?.App,
      root?.go?.main?.App,
    ];

    for (const candidate of candidates) {
      if (candidate && typeof candidate.GetHomePage === "function") {
        return candidate;
      }
    }

    throw new Error("Wails backend bindings are not available");
  }

  function clickIfEnabled(id) {
    const element = root.document?.getElementById(id);
    if (isEnabledElement(element)) {
      element.click();
    }
  }

  function scrollHomePage(top) {
    const scrollTarget = typeof window !== "undefined" && typeof window.scrollBy === "function"
      ? window
      : root;
    if (typeof scrollTarget.scrollBy !== "function") {
      return;
    }

    scrollTarget.scrollBy({
      top,
      behavior: "smooth",
    });
  }

  function isEnabledElement(element) {
    return Boolean(element) && !element.disabled;
  }

  function isGeneralTypingTarget(target) {
    if (!(target instanceof root.Element)) {
      return false;
    }

    if (target.isContentEditable) {
      return true;
    }

    const tagName = target.tagName;
    if (tagName === "TEXTAREA" || tagName === "SELECT") {
      return true;
    }
    if (tagName !== "INPUT") {
      return false;
    }

    return !target.readOnly;
  }

  function parseNumberInput(value) {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }

  function isAllowedControlKey(event) {
    return [
      "Backspace",
      "Delete",
      "Tab",
      "ArrowLeft",
      "ArrowRight",
      "ArrowUp",
      "ArrowDown",
      "Home",
      "End",
      "Enter",
    ].includes(event.key);
  }

  function previewNumericValue(input, insertedText) {
    const start = input.selectionStart ?? input.value.length;
    const end = input.selectionEnd ?? input.value.length;
    return `${input.value.slice(0, start)}${insertedText}${input.value.slice(end)}`;
  }

  function isValidPartialDecimal(value) {
    return /^-?\d*(\.\d{0,2})?$/.test(value);
  }

  function sanitizePartialDecimal(value) {
    let sanitized = String(value).replace(/[^\d.-]/g, "");
    if (sanitized.includes("-")) {
      sanitized = `${sanitized.startsWith("-") ? "-" : ""}${sanitized.replaceAll("-", "")}`;
    }

    const dotIndex = sanitized.indexOf(".");
    if (dotIndex >= 0) {
      sanitized = `${sanitized.slice(0, dotIndex + 1)}${sanitized.slice(dotIndex + 1).replaceAll(".", "")}`;
    }

    const match = sanitized.match(/^-?\d*(\.\d{0,2})?/);
    return match ? match[0] : "";
  }

  function parseEditableNumber(value) {
    if (value === "" || value === "-" || value === "." || value === "-.") {
      return 0;
    }

    return parseNumberInput(value);
  }

  function formatEditableNumber(value) {
    return Number(value || 0).toFixed(2);
  }

  function formatTHB(value) {
    const amount = Number(value);
    const formatted = `THB ${Math.abs(amount).toLocaleString("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    })}`;
    return amount < 0 ? `(${formatted})` : formatted;
  }

  function formatPercent(value) {
    return `${(Number(value) * 100).toFixed(2)}%`;
  }

  function formatDateLabel(value) {
    if (!value) {
      return "";
    }

    const normalized = String(value).includes("T") ? String(value) : `${value}T00:00:00Z`;
    const date = new Date(normalized);
    if (Number.isNaN(date.getTime())) {
      return String(value);
    }

    return date.toLocaleDateString("en-GB", {
      day: "2-digit",
      month: "short",
      year: "numeric",
      timeZone: "UTC",
    });
  }

  function formatDeltaTHB(value) {
    const amount = `THB ${Math.abs(Number(value)).toLocaleString("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    })}`;
    if (value > 0) {
      return `+${amount}`;
    }
    if (value < 0) {
      return `-${amount}`;
    }
    return amount;
  }

  function formatDeltaPercent(value) {
    const amount = `${(Math.abs(Number(value)) * 100).toFixed(2)}%`;
    if (value > 0) {
      return `+${amount}`;
    }
    if (value < 0) {
      return `-${amount}`;
    }
    return amount;
  }

  function cloneValue(value) {
    return JSON.parse(JSON.stringify(value));
  }

  function escapeHTML(value) {
    return String(value)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function rootErrorBanner(message) {
    const appRoot = root.document?.getElementById("app");
    if (!appRoot) {
      return;
    }

    const existing = appRoot.querySelector(".floating-error");
    if (existing) {
      existing.remove();
    }

    const banner = root.document.createElement("section");
    banner.className = "panel panel-error floating-error";
    banner.innerHTML = `
      <p class="eyebrow">Error</p>
      <p class="error-copy error-banner">${escapeHTML(message)}</p>
    `;
    appRoot.prepend(banner);
  }

  return {
    EDIT_TAB_INDEX,
    clickIfEnabled,
    cloneValue,
    escapeHTML,
    formatDateLabel,
    formatDeltaPercent,
    formatDeltaTHB,
    formatEditableNumber,
    formatPercent,
    formatTHB,
    isAllowedControlKey,
    isEnabledElement,
    isGeneralTypingTarget,
    isValidPartialDecimal,
    logHomeAction,
    parseEditableNumber,
    parseNumberInput,
    previewNumericValue,
    renderAppTitle,
    renderErrorState,
    resolveBackend,
    rootErrorBanner,
    runTransition,
    sanitizePartialDecimal,
    scrollHomePage,
    snapshotCount,
    state,
  };
}));
