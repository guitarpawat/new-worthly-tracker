(function initControls(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const controls = factory(shared, root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = controls;
  }
  root.WorthlyControls = controls;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildControls(shared, root) {
  const { escapeHTML, formatDateLabel } = shared;
  const WEEKDAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
  let openControlID = "";

  function renderSelectControl(config) {
    const {
      id,
      value,
      options,
      buttonClass = "snapshot-select",
      panelClass = "",
      placeholder = "Select",
      ariaLabel = "",
      disabled = false,
      tabIndex,
    } = config;
    const normalizedValue = value === undefined || value === null ? "" : String(value);
    const selectedOption = (options || []).find((option) => String(option.value) === normalizedValue);
    const triggerLabel = selectedOption?.label || placeholder;
    const tabIndexAttr = tabIndex === undefined ? "" : ` tabindex="${tabIndex}"`;
    const disabledAttr = disabled ? " disabled" : "";
    const labelAttr = ariaLabel ? ` aria-label="${escapeHTML(ariaLabel)}"` : "";

    return `
      <div class="custom-control custom-select-control" data-control-id="${escapeHTML(id)}" data-control-kind="select">
        <input id="${escapeHTML(id)}" type="hidden" value="${escapeHTML(normalizedValue)}" />
        <button
          id="${escapeHTML(id)}-trigger"
          class="${escapeHTML(buttonClass)} custom-control-trigger"
          type="button"
          aria-haspopup="listbox"
          aria-expanded="false"
          data-control-trigger="${escapeHTML(id)}"${labelAttr}${tabIndexAttr}${disabledAttr}
        >
          <span class="custom-control-trigger-label">${escapeHTML(triggerLabel)}</span>
          <span class="custom-control-chevron" aria-hidden="true"></span>
        </button>
        <div id="${escapeHTML(id)}-panel" class="custom-control-panel custom-select-panel ${escapeHTML(panelClass)}" hidden role="listbox">
          ${(options || []).length
            ? (options || []).map((option) => renderSelectOption(id, normalizedValue, option)).join("")
            : '<div class="custom-control-empty">No options available</div>'}
        </div>
      </div>
    `;
  }

  function renderDateControl(config) {
    const {
      id,
      value,
      buttonClass = "snapshot-select snapshot-select-button",
      placeholder = "Select date",
      tabIndex,
      allowClear = false,
    } = config;
    const normalizedValue = value || "";
    const tabIndexAttr = tabIndex === undefined ? "" : ` tabindex="${tabIndex}"`;
    const viewMonth = resolveMonthStart(normalizedValue || todayISODate());

    return `
      <div
        class="custom-control custom-date-control"
        data-control-id="${escapeHTML(id)}"
        data-control-kind="date"
        data-control-view-month="${escapeHTML(viewMonth)}"
        data-control-placeholder="${escapeHTML(placeholder)}"
        data-control-allow-clear="${allowClear ? "true" : "false"}"
      >
        <input id="${escapeHTML(id)}" type="hidden" value="${escapeHTML(normalizedValue)}" />
        <button
          id="${escapeHTML(id)}-trigger"
          class="${escapeHTML(buttonClass)} custom-control-trigger"
          type="button"
          aria-haspopup="dialog"
          aria-expanded="false"
          data-control-trigger="${escapeHTML(id)}"${tabIndexAttr}
        >
          <span class="custom-control-trigger-label">${escapeHTML(normalizedValue ? formatDateLabel(normalizedValue) : placeholder)}</span>
          <span class="custom-control-chevron" aria-hidden="true"></span>
        </button>
        <div id="${escapeHTML(id)}-panel" class="custom-control-panel custom-date-panel" hidden role="dialog">
          ${renderDatePanelBody(id, normalizedValue, viewMonth, allowClear, placeholder)}
        </div>
      </div>
    `;
  }

  function renderSelectOption(controlID, selectedValue, option) {
    const normalizedValue = String(option.value);
    const selectedClass = normalizedValue === selectedValue ? " custom-select-option-active" : "";
    const disabledAttr = option.disabled ? " disabled" : "";
    return `
      <button
        class="custom-select-option${selectedClass}"
        type="button"
        data-control-select-option="${escapeHTML(controlID)}"
        data-value="${escapeHTML(normalizedValue)}"
        data-label="${escapeHTML(option.label)}"${disabledAttr}
      >${escapeHTML(option.label)}</button>
    `;
  }

  function renderDatePanelBody(controlID, value, viewMonth, allowClear, placeholder) {
    const days = buildCalendarWeeks(viewMonth, value);
    return `
      <div class="custom-date-header">
        <div class="custom-date-nav-group">
          <button class="button button-small custom-date-nav" type="button" data-control-date-nav="${escapeHTML(controlID)}" data-direction="-12">&lt;&lt;</button>
          <button class="button button-small custom-date-nav" type="button" data-control-date-nav="${escapeHTML(controlID)}" data-direction="-1">&lt;</button>
        </div>
        <span class="custom-date-title">${escapeHTML(formatMonthYearLabel(viewMonth))}</span>
        <div class="custom-date-nav-group custom-date-nav-group-end">
          <button class="button button-small custom-date-nav" type="button" data-control-date-nav="${escapeHTML(controlID)}" data-direction="1">&gt;</button>
          <button class="button button-small custom-date-nav" type="button" data-control-date-nav="${escapeHTML(controlID)}" data-direction="12">&gt;&gt;</button>
        </div>
      </div>
      <div class="custom-date-weekdays">
        ${WEEKDAY_LABELS.map((label) => `<span>${escapeHTML(label)}</span>`).join("")}
      </div>
      <div class="custom-date-grid">
        ${days.map((day) => renderCalendarDay(controlID, day, value)).join("")}
      </div>
      <div class="custom-date-actions">
        <button class="button button-small" type="button" data-control-date-today="${escapeHTML(controlID)}">Today</button>
        ${allowClear ? `<button class="button button-small" type="button" data-control-date-clear="${escapeHTML(controlID)}">Clear</button>` : ""}
      </div>
    `;
  }

  function renderCalendarDay(controlID, day, selectedValue) {
    const classes = [
      "custom-date-day",
      day.isCurrentMonth ? "" : "custom-date-day-outside",
      day.isoDate === selectedValue ? "custom-date-day-selected" : "",
      day.isoDate === todayISODate() ? "custom-date-day-today" : "",
    ].filter(Boolean).join(" ");

    return `
      <button
        class="${classes}"
        type="button"
        data-control-date-value="${escapeHTML(controlID)}"
        data-value="${escapeHTML(day.isoDate)}"
      >${day.dayOfMonth}</button>
    `;
  }

  function handleControlClick(event) {
    const target = event.target instanceof root.Element ? event.target : null;
    if (!target) {
      return false;
    }

    const trigger = target.closest("[data-control-trigger]");
    if (trigger) {
      event.preventDefault();
      toggleControl(trigger.dataset.controlTrigger);
      trigger.focus();
      return true;
    }

    const selectOption = target.closest("[data-control-select-option]");
    if (selectOption) {
      event.preventDefault();
      commitSelectValue(selectOption.dataset.controlSelectOption, selectOption.dataset.value, selectOption.dataset.label || selectOption.textContent || "");
      return true;
    }

    const dateNav = target.closest("[data-control-date-nav]");
    if (dateNav) {
      event.preventDefault();
      shiftDateControlMonth(dateNav.dataset.controlDateNav, Number(dateNav.dataset.direction || 0));
      return true;
    }

    const dateValue = target.closest("[data-control-date-value]");
    if (dateValue) {
      event.preventDefault();
      commitDateValue(dateValue.dataset.controlDateValue, dateValue.dataset.value || "");
      return true;
    }

    const todayButton = target.closest("[data-control-date-today]");
    if (todayButton) {
      event.preventDefault();
      commitDateValue(todayButton.dataset.controlDateToday, todayISODate());
      return true;
    }

    const clearButton = target.closest("[data-control-date-clear]");
    if (clearButton) {
      event.preventDefault();
      commitDateValue(clearButton.dataset.controlDateClear, "");
      return true;
    }

    if (openControlID && !target.closest(`[data-control-id="${openControlID}"]`)) {
      closeOpenControl();
    }

    return false;
  }

  function handleControlKeydown(event) {
    const target = event.target instanceof root.Element ? event.target : null;
    const trigger = target?.closest("[data-control-trigger]");
    if (trigger) {
      return handleTriggerKeydown(event, trigger.dataset.controlTrigger || "");
    }

    if (target) {
      const option = target.closest("[data-control-select-option]");
      if (option) {
        return handleSelectOptionKeydown(event, option);
      }
    }

    if (!openControlID || event.key !== "Escape") {
      if (!openControlID || !isArrowNavigationKey(event.key)) {
        return false;
      }

      const openControlElement = findControl(openControlID);
      if (!openControlElement || openControlElement.dataset.controlKind !== "select") {
        return false;
      }

      event.preventDefault();
      focusFromOpenControl(event.key === "ArrowDown" ? 1 : -1);
      return true;
    }

    if (target && !target.closest(`[data-control-id="${openControlID}"]`)) {
      return false;
    }

    event.preventDefault();
    const triggerToFocus = findControlTrigger(openControlID);
    closeOpenControl();
    triggerToFocus?.focus();
    return true;
  }

  function handleTriggerKeydown(event, controlID) {
    const control = findControl(controlID);
    if (!control) {
      return false;
    }

    const kind = control.dataset.controlKind;
    if (kind === "select") {
      if ((event.key === "Enter" || event.key === " ") && openControlID !== controlID) {
        event.preventDefault();
        openControl(controlID);
        focusSelectedOrFirstOption(controlID);
        return true;
      }
      if ((event.key === "ArrowDown" || event.key === "ArrowUp") && openControlID !== controlID) {
        event.preventDefault();
        openControl(controlID);
        focusSelectedOrFirstOption(controlID, event.key === "ArrowUp" ? "last" : "first");
        return true;
      }
      if (openControlID === controlID && (event.key === "ArrowDown" || event.key === "ArrowUp")) {
        event.preventDefault();
        focusFromOpenControl(event.key === "ArrowDown" ? 1 : -1);
        return true;
      }
    }

    if ((event.key === "Enter" || event.key === " ") && openControlID !== controlID) {
      event.preventDefault();
      openControl(controlID);
      return true;
    }

    if (event.key === "Escape" && openControlID === controlID) {
      event.preventDefault();
      closeOpenControl();
      return true;
    }

    return false;
  }

  function handleSelectOptionKeydown(event, option) {
    const controlID = option.dataset.controlSelectOption || "";
    if (!controlID) {
      return false;
    }

    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      commitSelectValue(controlID, option.dataset.value || "", option.dataset.label || option.textContent || "");
      findControlTrigger(controlID)?.focus();
      return true;
    }
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      focusAdjacentOption(controlID, event.key === "ArrowDown" ? 1 : -1);
      return true;
    }
    if (event.key === "Home" || event.key === "End") {
      event.preventDefault();
      focusSelectedOrFirstOption(controlID, event.key === "Home" ? "first" : "last");
      return true;
    }
    if (event.key === "Escape") {
      event.preventDefault();
      closeOpenControl();
      findControlTrigger(controlID)?.focus();
      return true;
    }

    return false;
  }

  function setControlValue(controlID, value) {
    const control = findControl(controlID);
    const input = findControlInput(controlID);
    const triggerLabel = findControlTriggerLabel(controlID);
    if (!control || !input || !triggerLabel) {
      return;
    }

    input.value = value;
    if (control.dataset.controlKind === "date") {
      const placeholder = control.dataset.controlPlaceholder || "Select date";
      triggerLabel.textContent = value ? formatDateLabel(value) : placeholder;
      control.dataset.controlViewMonth = resolveMonthStart(value || todayISODate());
      const panel = findControlPanel(controlID);
      if (panel) {
        panel.innerHTML = renderDatePanelBody(
          controlID,
          value,
          control.dataset.controlViewMonth,
          control.dataset.controlAllowClear === "true",
          placeholder,
        );
      }
      return;
    }

    const selectedOption = control.querySelector(`[data-control-select-option="${controlID}"][data-value="${cssEscape(String(value))}"]`);
    if (selectedOption) {
      triggerLabel.textContent = selectedOption.dataset.label || selectedOption.textContent || "";
    }
    syncSelectedOptionState(controlID, value);
  }

  function buildCalendarWeeks(monthStartValue, selectedValue) {
    const monthStart = parseISODate(monthStartValue) || parseISODate(todayISODate());
    const firstWeekday = (monthStart.getUTCDay() + 6) % 7;
    const gridStart = new Date(Date.UTC(monthStart.getUTCFullYear(), monthStart.getUTCMonth(), 1 - firstWeekday));
    const days = [];
    for (let index = 0; index < 42; index += 1) {
      const current = new Date(Date.UTC(gridStart.getUTCFullYear(), gridStart.getUTCMonth(), gridStart.getUTCDate() + index));
      days.push({
        isoDate: formatISODate(current),
        dayOfMonth: current.getUTCDate(),
        isCurrentMonth: current.getUTCMonth() === monthStart.getUTCMonth(),
      });
    }
    return days;
  }

  function resolveMonthStart(value) {
    const parsed = parseISODate(value);
    const date = parsed || parseISODate(todayISODate());
    return `${date.getUTCFullYear()}-${String(date.getUTCMonth() + 1).padStart(2, "0")}-01`;
  }

  function addMonths(monthStartValue, delta) {
    const monthStart = parseISODate(monthStartValue) || parseISODate(todayISODate());
    return formatISODate(new Date(Date.UTC(monthStart.getUTCFullYear(), monthStart.getUTCMonth() + delta, 1)));
  }

  function formatMonthYearLabel(monthStartValue) {
    const parsed = parseISODate(monthStartValue) || parseISODate(todayISODate());
    return parsed.toLocaleDateString("en-GB", {
      month: "short",
      year: "numeric",
      timeZone: "UTC",
    });
  }

  function todayISODate() {
    const now = new Date();
    return formatISODate(new Date(Date.UTC(now.getFullYear(), now.getMonth(), now.getDate())));
  }

  function parseISODate(value) {
    const match = String(value || "").match(/^(\d{4})-(\d{2})-(\d{2})$/);
    if (!match) {
      return null;
    }
    const date = new Date(Date.UTC(Number(match[1]), Number(match[2]) - 1, Number(match[3])));
    return Number.isNaN(date.getTime()) ? null : date;
  }

  function formatISODate(date) {
    return [
      date.getUTCFullYear(),
      String(date.getUTCMonth() + 1).padStart(2, "0"),
      String(date.getUTCDate()).padStart(2, "0"),
    ].join("-");
  }

  function toggleControl(controlID) {
    if (!controlID) {
      return;
    }
    if (openControlID === controlID) {
      closeOpenControl();
      return;
    }
    openControl(controlID);
  }

  function openControl(controlID) {
    closeOpenControl();
    const control = findControl(controlID);
    if (!control) {
      return;
    }

    openControlID = controlID;
    control.classList.add("custom-control-open");
    const trigger = findControlTrigger(controlID);
    const panel = findControlPanel(controlID);
    if (trigger) {
      trigger.setAttribute("aria-expanded", "true");
    }
    if (panel) {
      panel.hidden = false;
    }
  }

  function closeOpenControl() {
    if (!openControlID) {
      return;
    }

    const control = findControl(openControlID);
    const trigger = findControlTrigger(openControlID);
    const panel = findControlPanel(openControlID);
    if (control) {
      control.classList.remove("custom-control-open");
    }
    if (trigger) {
      trigger.setAttribute("aria-expanded", "false");
    }
    if (panel) {
      panel.hidden = true;
    }
    openControlID = "";
  }

  function commitSelectValue(controlID, value, label) {
    const input = findControlInput(controlID);
    const triggerLabel = findControlTriggerLabel(controlID);
    if (!input || !triggerLabel) {
      return;
    }

    input.value = value;
    triggerLabel.textContent = label;
    syncSelectedOptionState(controlID, value);
    closeOpenControl();
    dispatchControlChange(input);
  }

  function commitDateValue(controlID, value) {
    setControlValue(controlID, value);
    closeOpenControl();
    const input = findControlInput(controlID);
    if (input) {
      dispatchControlChange(input);
    }
  }

  function shiftDateControlMonth(controlID, delta) {
    const control = findControl(controlID);
    const panel = findControlPanel(controlID);
    if (!control || !panel) {
      return;
    }
    const nextMonth = addMonths(control.dataset.controlViewMonth, delta);
    control.dataset.controlViewMonth = nextMonth;
    panel.innerHTML = renderDatePanelBody(
      controlID,
      findControlInput(controlID)?.value || "",
      nextMonth,
      control.dataset.controlAllowClear === "true",
      control.dataset.controlPlaceholder || "Select date",
    );
  }

  function syncSelectedOptionState(controlID, value) {
    for (const option of root.document.querySelectorAll(`[data-control-select-option="${controlID}"]`)) {
      option.classList.toggle("custom-select-option-active", option.dataset.value === String(value));
    }
  }

  function findSelectOptions(controlID) {
    return Array.from(root.document?.querySelectorAll(`[data-control-select-option="${cssEscape(controlID)}"]:not([disabled])`) || []);
  }

  function focusSelectedOrFirstOption(controlID, fallback = "first") {
    const options = findSelectOptions(controlID);
    if (!options.length) {
      return;
    }
    const input = findControlInput(controlID);
    const selected = options.find((option) => option.dataset.value === String(input?.value || ""));
    const target = selected || (fallback === "last" ? options[options.length - 1] : options[0]);
    target.focus();
  }

  function focusAdjacentOption(controlID, delta) {
    const options = findSelectOptions(controlID);
    if (!options.length) {
      return;
    }
    const active = root.document?.activeElement;
    const currentIndex = options.findIndex((option) => option === active);
    const nextIndex = currentIndex < 0
      ? (delta < 0 ? options.length - 1 : 0)
      : Math.max(0, Math.min(options.length - 1, currentIndex + delta));
    options[nextIndex].focus();
  }

  function focusFromOpenControl(delta) {
    if (!openControlID) {
      return;
    }

    const active = root.document?.activeElement;
    const openControlElement = findControl(openControlID);
    const isInsideOpenControl = Boolean(active && openControlElement?.contains(active));
    if (!isInsideOpenControl || active?.hasAttribute("data-control-trigger")) {
      focusSelectedOrFirstOption(openControlID, delta < 0 ? "last" : "first");
      return;
    }

    focusAdjacentOption(openControlID, delta);
  }

  function isArrowNavigationKey(key) {
    return key === "ArrowDown" || key === "ArrowUp";
  }

  function findControl(controlID) {
    return root.document?.querySelector(`[data-control-id="${cssEscape(controlID)}"]`) || null;
  }

  function findControlInput(controlID) {
    return root.document?.getElementById(controlID) || null;
  }

  function findControlTrigger(controlID) {
    return root.document?.getElementById(`${controlID}-trigger`) || null;
  }

  function findControlTriggerLabel(controlID) {
    return findControlTrigger(controlID)?.querySelector(".custom-control-trigger-label") || null;
  }

  function findControlPanel(controlID) {
    return root.document?.getElementById(`${controlID}-panel`) || null;
  }

  function dispatchControlChange(input) {
    input.dispatchEvent(new root.Event("input", { bubbles: true }));
    input.dispatchEvent(new root.Event("change", { bubbles: true }));
  }

  function cssEscape(value) {
    return String(value).replace(/["\\]/g, "\\$&");
  }

  return {
    addMonths,
    buildCalendarWeeks,
    formatMonthYearLabel,
    handleControlClick,
    handleControlKeydown,
    renderDateControl,
    renderSelectControl,
    resolveMonthStart,
    setControlValue,
    todayISODate,
  };
}));
