(function initAssetReorder(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const assetReorder = factory(shared, root);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = assetReorder;
  }
  root.WorthlyAssetReorder = assetReorder;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildAssetReorder(shared, root) {
  const {
    escapeHTML,
    runTransition,
    state,
  } = shared;

  function ensureAssetReorderState(page) {
    const signature = buildAssetManagementSignature(page);
    if (state.assetReorder && state.assetReorder.signature === signature) {
      return state.assetReorder;
    }

    state.assetReorder = {
      signature,
      typeFilterActiveOnly: true,
      assetFilterActiveOnly: true,
      typeRows: cloneRows(page.AssetTypes || []),
      originalTypeRows: cloneRows(page.AssetTypes || []),
      assetGroups: buildAssetGroups(page),
      originalAssetGroups: buildAssetGroups(page),
      typeError: "",
      assetError: "",
      drag: null,
    };

    return state.assetReorder;
  }

  function buildAssetManagementSignature(page) {
    const typePart = (page?.AssetTypes || []).map((item) => `${item.ID}:${item.Ordering}:${item.IsActive}`).join("|");
    const assetPart = (page?.Assets || []).map((item) => `${item.ID}:${item.AssetTypeID}:${item.Ordering}:${item.IsActive}`).join("|");
    return `${typePart}#${assetPart}`;
  }

  function cloneRows(rows) {
    return rows.map((row) => ({ ...row }));
  }

  function buildAssetGroups(page) {
    const rows = cloneRows(page?.Assets || []);
    const groups = [];
    for (const row of rows) {
      let group = groups.find((item) => item.AssetTypeID === row.AssetTypeID);
      if (!group) {
        group = {
          AssetTypeID: row.AssetTypeID,
          AssetTypeName: row.AssetTypeName,
          Rows: [],
        };
        groups.push(group);
      }
      group.Rows.push(row);
    }
    return groups;
  }

  function renderAssetTypeReorderPage() {
    const reorder = ensureAssetReorderState(state.assetManagementPage);
    const visibleRows = getVisibleTypeRows(reorder);
    const dirty = isTypeReorderDirty(reorder);

    return `
      <section class="panel asset-reorder-panel">
        <div class="table-header asset-reorder-header">
          <div>
            <h2>Reorder Asset Types</h2>
            <span>Drag and drop to change display order.</span>
          </div>
          <div class="asset-reorder-toolbar">
            ${renderReorderFilterToggle("reorder-asset-type-filter", reorder.typeFilterActiveOnly)}
            <div class="actions">
              <button id="reorder-asset-type-reset" class="button" type="button" ${dirty ? "" : "disabled"}>Reset</button>
              <button id="reorder-asset-type-save" class="button button-primary" type="button" ${dirty ? "" : "disabled"}>Save</button>
            </div>
          </div>
        </div>
        ${reorder.typeError ? `<p class="error-copy error-banner">${escapeHTML(reorder.typeError)}</p>` : ""}
        <div class="asset-reorder-list" data-reorder-list="asset_type">
          ${visibleRows.map((row) => renderReorderItem({
            id: row.ID,
            title: row.Name,
            meta: `${row.AssetCount} asset(s)`,
            active: row.IsActive,
          })).join("")}
        </div>
      </section>
    `;
  }

  function renderAssetReorderPage() {
    const reorder = ensureAssetReorderState(state.assetManagementPage);
    const groups = getVisibleAssetGroups(reorder);
    const dirty = isAssetReorderDirty(reorder);

    return `
      <section class="panel asset-reorder-panel">
        <div class="table-header asset-reorder-header">
          <div>
            <h2>Reorder Assets</h2>
            <span>Drag assets within the same asset type.</span>
          </div>
          <div class="asset-reorder-toolbar">
            ${renderReorderFilterToggle("reorder-asset-filter", reorder.assetFilterActiveOnly)}
            <div class="actions">
              <button id="reorder-asset-reset" class="button" type="button" ${dirty ? "" : "disabled"}>Reset</button>
              <button id="reorder-asset-save" class="button button-primary" type="button" ${dirty ? "" : "disabled"}>Save</button>
            </div>
          </div>
        </div>
        ${reorder.assetError ? `<p class="error-copy error-banner">${escapeHTML(reorder.assetError)}</p>` : ""}
        <div class="asset-reorder-groups">
          ${groups.map((group) => `
            <section class="asset-reorder-group">
              <div class="table-header">
                <h3>${escapeHTML(group.AssetTypeName)}</h3>
                <span>${group.Rows.length} visible asset(s)</span>
              </div>
              <div class="asset-reorder-list" data-reorder-list="asset" data-asset-type-id="${group.AssetTypeID}">
                ${group.Rows.map((row) => renderReorderItem({
                  id: row.ID,
                  title: row.Name,
                  meta: row.Broker || row.AssetTypeName,
                  active: row.IsActive,
                })).join("")}
              </div>
            </section>
          `).join("")}
        </div>
      </section>
    `;
  }

  function renderReorderItem(config) {
    return `
      <article
        class="asset-reorder-item"
        draggable="true"
        data-reorder-id="${config.id}"
      >
        <span class="asset-reorder-handle" aria-hidden="true">::</span>
        <div class="asset-reorder-copy">
          <strong>${escapeHTML(config.title)}</strong>
          <span>${escapeHTML(config.meta)}</span>
        </div>
        <span class="metric-delta ${config.active ? "delta-positive" : "delta-negative"}">${config.active ? "Yes" : "No"}</span>
      </article>
    `;
  }

  function renderReorderFilterToggle(id, isOn) {
    return `
      <label class="field-inline field-inline-toggle asset-reorder-filter-toggle">
        <span class="field-label">Active only</span>
        <label class="toggle-field" for="${id}">
          <input id="${id}" class="toggle-field-input" type="checkbox" ${isOn ? "checked" : ""} />
          <span class="mock-toggle" aria-hidden="true">
            <span class="mock-toggle-thumb"></span>
          </span>
        </label>
      </label>
    `;
  }

  function bindAssetTypeReorderPage(app) {
    const reorder = ensureAssetReorderState(state.assetManagementPage);

    const filter = root.document.getElementById("reorder-asset-type-filter");
    if (filter) {
      filter.addEventListener("change", () => {
        reorder.typeFilterActiveOnly = filter.checked;
        reorder.typeError = "";
        renderCurrentReorderPage(app);
      });
    }

    const resetButton = root.document.getElementById("reorder-asset-type-reset");
    if (resetButton) {
      resetButton.addEventListener("click", () => {
        reorder.typeRows = cloneRows(reorder.originalTypeRows);
        reorder.typeError = "";
        renderCurrentReorderPage(app);
      });
    }

    const saveButton = root.document.getElementById("reorder-asset-type-save");
    if (saveButton) {
      saveButton.addEventListener("click", async () => {
        await runTransition(async () => {
          reorder.typeError = "";
          try {
            await shared.resolveBackend().ReorderAssetTypes(buildReorderAssetTypesPayload(reorder));
            await app.loadAssetManagementPage({ view: "reorder_asset_type" });
          } catch (error) {
            reorder.typeError = error?.message || String(error);
            renderCurrentReorderPage(app);
          }
        });
      });
    }

    bindReorderDragList(root.document.querySelector('[data-reorder-list="asset_type"]'), {
      onMove: (draggedID, targetID) => {
        reorder.typeRows = reorderVisibleRows(reorder.typeRows, draggedID, targetID, reorder.typeFilterActiveOnly);
      },
      rerender: () => renderCurrentReorderPage(app),
      getDragScope: () => "asset_type",
    });
  }

  function bindAssetReorderPage(app) {
    const reorder = ensureAssetReorderState(state.assetManagementPage);

    const filter = root.document.getElementById("reorder-asset-filter");
    if (filter) {
      filter.addEventListener("change", () => {
        reorder.assetFilterActiveOnly = filter.checked;
        reorder.assetError = "";
        renderCurrentReorderPage(app);
      });
    }

    const resetButton = root.document.getElementById("reorder-asset-reset");
    if (resetButton) {
      resetButton.addEventListener("click", () => {
        reorder.assetGroups = buildAssetGroups({ Assets: flattenAssetGroups(reorder.originalAssetGroups) });
        reorder.assetError = "";
        renderCurrentReorderPage(app);
      });
    }

    const saveButton = root.document.getElementById("reorder-asset-save");
    if (saveButton) {
      saveButton.addEventListener("click", async () => {
        await runTransition(async () => {
          reorder.assetError = "";
          try {
            const backend = shared.resolveBackend();
            for (const payload of buildReorderAssetPayloads(reorder)) {
              await backend.ReorderAssets(payload);
            }
            await app.loadAssetManagementPage({ view: "reorder_asset" });
          } catch (error) {
            reorder.assetError = error?.message || String(error);
            renderCurrentReorderPage(app);
          }
        });
      });
    }

    for (const list of root.document.querySelectorAll('[data-reorder-list="asset"]')) {
      const assetTypeID = Number(list.dataset.assetTypeId);
      bindReorderDragList(list, {
        onMove: (draggedID, targetID) => {
          reorder.assetGroups = reorder.assetGroups.map((group) => {
            if (group.AssetTypeID !== assetTypeID) {
              return group;
            }
            return {
              ...group,
              Rows: reorderVisibleRows(group.Rows, draggedID, targetID, reorder.assetFilterActiveOnly),
            };
          });
        },
        rerender: () => renderCurrentReorderPage(app),
        getDragScope: () => `asset:${assetTypeID}`,
      });
    }
  }

  function bindReorderDragList(list, config) {
    if (!list) {
      return;
    }

    for (const item of list.querySelectorAll("[data-reorder-id]")) {
      item.addEventListener("dragstart", (event) => {
        startReorderDrag(event, {
          id: Number(item.dataset.reorderId),
          scope: config.getDragScope(),
        });
        item.classList.add("asset-reorder-item-dragging");
      });

      item.addEventListener("dragend", () => {
        state.assetReorder.drag = null;
        item.classList.remove("asset-reorder-item-dragging");
      });

      item.addEventListener("dragover", (event) => {
        const drag = state.assetReorder?.drag;
        if (!shouldAcceptReorderDrag(drag, config.getDragScope())) {
          return;
        }
        event.preventDefault();
        if (event.dataTransfer) {
          event.dataTransfer.dropEffect = "move";
        }
      });

      item.addEventListener("drop", (event) => {
        const drag = state.assetReorder?.drag;
        if (!shouldAcceptReorderDrag(drag, config.getDragScope())) {
          return;
        }
        event.preventDefault();
        const targetID = Number(item.dataset.reorderId);
        if (drag.id === targetID) {
          return;
        }
        config.onMove(drag.id, targetID);
        state.assetReorder.drag = null;
        config.rerender();
      });
    }
  }

  function startReorderDrag(event, payload) {
    state.assetReorder.drag = payload;
    if (event?.dataTransfer) {
      event.dataTransfer.effectAllowed = "move";
      event.dataTransfer.setData("text/plain", String(payload.id));
    }
  }

  function shouldAcceptReorderDrag(drag, scope) {
    return Boolean(drag && drag.scope === scope);
  }

  function renderCurrentReorderPage(app) {
    const page = root.WorthlyAssetManagement;
    if (page && typeof page.renderAssetManagementPage === "function") {
      page.renderAssetManagementPage(app);
    }
  }

  function getVisibleTypeRows(reorder) {
    if (!reorder.typeFilterActiveOnly) {
      return reorder.typeRows;
    }
    return reorder.typeRows.filter((row) => row.IsActive);
  }

  function getVisibleAssetGroups(reorder) {
    return reorder.assetGroups
      .map((group) => ({
        ...group,
        Rows: reorder.assetFilterActiveOnly ? group.Rows.filter((row) => row.IsActive) : group.Rows,
      }))
      .filter((group) => group.Rows.length > 0);
  }

  function reorderVisibleRows(rows, draggedID, targetID, activeOnly) {
    const visibleRows = activeOnly ? rows.filter((row) => row.IsActive) : rows.slice();
    const draggedIndex = visibleRows.findIndex((row) => row.ID === draggedID);
    const targetIndex = visibleRows.findIndex((row) => row.ID === targetID);
    if (draggedIndex < 0 || targetIndex < 0) {
      return rows;
    }

    const reorderedVisible = visibleRows.slice();
    const [draggedRow] = reorderedVisible.splice(draggedIndex, 1);
    reorderedVisible.splice(targetIndex, 0, draggedRow);

    if (!activeOnly) {
      return reorderedVisible;
    }

    const visibleIDs = new Set(visibleRows.map((row) => row.ID));
    let visibleCursor = 0;
    return rows.map((row) => {
      if (!visibleIDs.has(row.ID)) {
        return row;
      }
      const nextRow = reorderedVisible[visibleCursor];
      visibleCursor += 1;
      return nextRow;
    });
  }

  function isTypeReorderDirty(reorder) {
    return buildRowSignature(reorder.typeRows) !== buildRowSignature(reorder.originalTypeRows);
  }

  function isAssetReorderDirty(reorder) {
    return buildGroupSignature(reorder.assetGroups) !== buildGroupSignature(reorder.originalAssetGroups);
  }

  function buildRowSignature(rows) {
    return rows.map((row) => row.ID).join("|");
  }

  function buildGroupSignature(groups) {
    return groups.map((group) => `${group.AssetTypeID}:${buildRowSignature(group.Rows)}`).join("#");
  }

  function flattenAssetGroups(groups) {
    return groups.flatMap((group) => group.Rows);
  }

  function buildReorderAssetTypesPayload(reorder) {
    return {
      OrderedIDs: getVisibleTypeRows(reorder).map((row) => row.ID),
      ActiveOnly: reorder.typeFilterActiveOnly,
    };
  }

  function buildReorderAssetPayloads(reorder) {
    return getVisibleAssetGroups(reorder).map((group) => ({
      AssetTypeID: group.AssetTypeID,
      OrderedIDs: group.Rows.map((row) => row.ID),
      ActiveOnly: reorder.assetFilterActiveOnly,
    }));
  }

  return {
    bindAssetReorderPage,
    bindAssetTypeReorderPage,
    buildAssetGroups,
    buildReorderAssetPayloads,
    buildReorderAssetTypesPayload,
    ensureAssetReorderState,
    renderAssetReorderPage,
    renderReorderFilterToggle,
    renderAssetTypeReorderPage,
    reorderVisibleRows,
    shouldAcceptReorderDrag,
    startReorderDrag,
  };
}));
