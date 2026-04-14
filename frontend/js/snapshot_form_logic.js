(function initSnapshotFormLogic(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const logic = factory(shared, root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = logic;
  }
  root.WorthlySnapshotFormLogic = logic;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildSnapshotFormLogic(shared, root) {
  const {
    formatEditableNumber,
    formatPercent,
    formatTHB,
    isAllowedControlKey,
    isValidPartialDecimal,
    parseEditableNumber,
    previewNumericValue,
    sanitizePartialDecimal,
    state,
  } = shared;

  function resolveEditHorizontalMove({ key, fieldKey, isCash, isNotesEditable }) {
    if ((key !== "a" && key !== "A" && key !== "d" && key !== "D") || isNotesEditable) {
      return null;
    }

    const rowFields = isCash ? ["current", "notes", "remove"] : ["bought", "current", "notes", "remove"];
    const currentIndex = rowFields.indexOf(fieldKey);
    if (currentIndex < 0) {
      return null;
    }

    const nextIndex = key === "a" || key === "A" ? currentIndex - 1 : currentIndex + 1;
    if (nextIndex < 0 || nextIndex >= rowFields.length) {
      return null;
    }

    return rowFields[nextIndex];
  }

  function isEditHorizontalShortcutKey(key) {
    return key === "a" || key === "A" || key === "d" || key === "D";
  }

  function isEditGridShortcutKey(key) {
    return isEditHorizontalShortcutKey(key) || key === "w" || key === "W" || key === "s" || key === "S";
  }

  function isEditShortcutKey(key) {
    return isEditGridShortcutKey(key) || key === "e" || key === "E" || key === "Enter" || key === "Escape";
  }

  function resolveEditVerticalDirection({ key, isNotesEditable }) {
    if (isNotesEditable) {
      return null;
    }
    if (key === "w" || key === "W") {
      return -1;
    }
    if (key === "s" || key === "S") {
      return 1;
    }
    return null;
  }

  function findVerticalTargetAssetID(draft, currentAssetID, fieldKey, direction) {
    if (!draft || direction === 0) {
      return null;
    }

    const rows = draft.Groups.flatMap((group) => group.Rows);
    const currentIndex = rows.findIndex((row) => row.AssetID === currentAssetID);
    if (currentIndex < 0) {
      return null;
    }

    for (let index = currentIndex + direction; index >= 0 && index < rows.length; index += direction) {
      if (isFieldAvailableForRow(rows[index], fieldKey)) {
        return rows[index].AssetID;
      }
    }

    return null;
  }

  function isFieldAvailableForRow(row, fieldKey) {
    if (!row) {
      return false;
    }
    if (fieldKey === "bought") {
      return !row.IsCash;
    }
    return fieldKey === "current" || fieldKey === "notes" || fieldKey === "remove";
  }

  function shouldActivateManualFieldOnEnter({ key, isManualField, isDisabled, isReadOnly }) {
    return (key === "Enter" || key === "e" || key === "E") && isManualField && !isDisabled && isReadOnly;
  }

  function shouldDeactivateManualFieldOnKey({ key, isManualField, isReadOnly }) {
    return (key === "Escape" || key === "Enter") && isManualField && !isReadOnly;
  }

  function shouldDiscardEditPageOnEscape({ key, isTypingTarget, isManualFieldEditing, isCurrentPriceTarget }) {
    return key === "Escape" && !isManualFieldEditing && (!isTypingTarget || isCurrentPriceTarget);
  }

  function focusEditField(assetID, fieldKey) {
    const selector = `[data-asset-id="${assetID}"][data-nav-field="${fieldKey}"]`;
    const element = root.document.querySelector(selector);
    if (!element || element.disabled) {
      return;
    }

    element.focus();
    if (typeof element.select === "function" && fieldKey !== "notes") {
      element.select();
    }
  }

  function findEditRow(assetID) {
    for (const group of state.editDraft?.Groups || []) {
      const row = group.Rows.find((candidate) => candidate.AssetID === assetID);
      if (row) {
        return row;
      }
    }
    return null;
  }

  function refreshRowMetrics(assetID) {
    const row = findEditRow(assetID);
    if (!row) {
      return;
    }

    const profitCell = root.document.getElementById(`profit-value-${assetID}`);
    const percentCell = root.document.getElementById(`profit-percent-${assetID}`);
    if (!profitCell || !percentCell) {
      return;
    }

    if (row.IsCash) {
      profitCell.className = "numeric muted";
      percentCell.className = "numeric muted";
      profitCell.innerHTML = '<span class="muted">N/A</span>';
      percentCell.innerHTML = '<span class="muted">N/A</span>';
      return;
    }

    const profit = calculateProfitFromRow(row);
    const profitPercent = calculateProfitPercentFromRow(row);
    const tone = profit > 0 ? "positive" : profit < 0 ? "negative" : "neutral";

    profitCell.className = `numeric ${tone}`;
    percentCell.className = `numeric ${tone}`;
    profitCell.textContent = formatTHB(profit);
    percentCell.textContent = formatPercent(profitPercent);
  }

  function calculateProfitFromRow(row) {
    if (row.IsCash) {
      return 0;
    }
    return Number(row.CurrentPrice) - Number(row.BoughtPrice);
  }

  function calculateProfitPercentFromRow(row) {
    if (row.IsCash) {
      return 0;
    }
    const boughtPrice = Number(row.BoughtPrice);
    if (boughtPrice === 0) {
      return 0;
    }
    return calculateProfitFromRow(row) / boughtPrice;
  }

  function bindPriceInput(input, config) {
    if (input.classList.contains("manual-field")) {
      bindManualField(input);
    }

    input.addEventListener("focus", () => {
      input.select();
    });

    input.addEventListener("keydown", (event) => {
      if (state.currentView === "edit" && isEditShortcutKey(event.key)) {
        return;
      }

      if (isAllowedControlKey(event)) {
        return;
      }
      if (event.ctrlKey || event.metaKey || event.altKey) {
        return;
      }
      if (event.key.length !== 1) {
        return;
      }

      const nextValue = previewNumericValue(input, event.key);
      if (!isValidPartialDecimal(nextValue)) {
        event.preventDefault();
      }
    });

    input.addEventListener("paste", (event) => {
      const clipboardText = event.clipboardData?.getData("text") || "";
      const nextValue = previewNumericValue(input, clipboardText);
      if (!isValidPartialDecimal(nextValue)) {
        event.preventDefault();
      }
    });

    input.addEventListener("input", (event) => {
      const sanitized = sanitizePartialDecimal(event.target.value);
      if (event.target.value !== sanitized) {
        event.target.value = sanitized;
      }

      const assetID = Number(event.target.dataset.assetId);
      const row = findEditRow(assetID);
      if (!row) {
        return;
      }

      config.onChange(row, parseEditableNumber(sanitized));
      config.onDirty();
      refreshRowMetrics(assetID);
    });

    input.addEventListener("blur", (event) => {
      const assetID = Number(event.target.dataset.assetId);
      const row = findEditRow(assetID);
      if (!row) {
        return;
      }

      const value = input.classList.contains("edit-current-input") ? row.CurrentPrice : row.BoughtPrice;
      event.target.value = formatEditableNumber(value);
    });
  }

  function bindManualField(field) {
    if (field.dataset.manualFieldBound === "true") {
      return;
    }

    field.dataset.manualFieldBound = "true";
    field.addEventListener("click", () => {
      activateManualField(field);
    });
    field.addEventListener("blur", () => {
      deactivateManualField(field);
    });
  }

  function activateManualField(field) {
    if (field.disabled) {
      return;
    }
    field.readOnly = false;
    field.classList.remove("manual-field-locked");
    field.focus();
    if (typeof field.select === "function") {
      field.select();
    }
  }

  function deactivateManualField(field, options = {}) {
    if (field.disabled) {
      return;
    }
    field.readOnly = true;
    field.classList.add("manual-field-locked");
    if (options.keepFocus && typeof field.focus === "function") {
      field.focus();
    }
  }

  function isManualFieldEditing(target) {
    return Boolean(
      target instanceof root.Element &&
      target.classList.contains("manual-field") &&
      "readOnly" in target &&
      target.readOnly === false,
    );
  }

  return {
    activateManualField,
    bindManualField,
    bindPriceInput,
    calculateProfitFromRow,
    calculateProfitPercentFromRow,
    deactivateManualField,
    findEditRow,
    findVerticalTargetAssetID,
    focusEditField,
    isEditGridShortcutKey,
    isEditHorizontalShortcutKey,
    isEditShortcutKey,
    isManualFieldEditing,
    resolveEditHorizontalMove,
    resolveEditVerticalDirection,
    shouldActivateManualFieldOnEnter,
    shouldDeactivateManualFieldOnKey,
    shouldDiscardEditPageOnEscape,
  };
}));
