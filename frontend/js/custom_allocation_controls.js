(function initCustomAllocationControls(root, factory) {
  const controls = root.WorthlyControls || (typeof require !== "undefined" ? require("./controls.js") : null);
  const api = factory(controls, root);
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  root.WorthlyCustomAllocationControls = api;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildCustomAllocationControls(controls, root) {
  function render(charts, mode, prefix) {
    if (!charts?.length) return "";
    const custom = mode.startsWith("custom:");
    return controls.renderSelectControl({
      id: `${prefix}-custom-allocation`,
      value: mode,
      ariaLabel: "Custom allocation chart",
      triggerLabel: "Custom",
      buttonClass: `button progress-chip custom-allocation-trigger ${custom ? "progress-chip-active" : ""}`,
      panelClass: "home-menu-panel custom-allocation-menu",
      options: charts.map((chart) => ({ value: `custom:${chart.ID}`, label: chart.Name })),
    });
  }

  function bind(prefix, onChange) {
    root.document.getElementById(`${prefix}-custom-allocation`)?.addEventListener("change", (event) => {
      onChange(event.target.value);
      focus(prefix);
    });
  }

  function focus(prefix) {
    root.document.getElementById(`${prefix}-custom-allocation-trigger`)?.focus({ preventScroll: true });
  }

  function normalizeMode(mode, charts) {
    if (!mode?.startsWith("custom:")) return mode || "asset_type";
    return charts?.some((chart) => `custom:${chart.ID}` === mode) ? mode : "asset_type";
  }

  return { render, bind, normalizeMode, focus };
}));
