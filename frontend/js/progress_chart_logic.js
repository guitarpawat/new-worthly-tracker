(function initProgressChartLogic(root, factory) {
  const shared = root.WorthlyShared || (typeof require !== "undefined" ? require("./shared.js") : null);
  const progressChartLogic = factory(shared);
  if (typeof module !== "undefined" && module.exports) {
    module.exports = progressChartLogic;
  }
  root.WorthlyProgressChartLogic = progressChartLogic;
}(typeof globalThis !== "undefined" ? globalThis : this, function buildProgressChartLogic(shared) {
  const {
    formatDateLabel,
    formatPercent,
    formatTHB,
    state,
  } = shared;

  const PROGRESS_VIEWS = [
    { id: "trend", label: "Trend Chart" },
    { id: "summary", label: "Summary Table" },
  ];

  const CHART_MODES = [
    { id: "net_worth", label: "Total Net Worth" },
    { id: "profit_rate", label: "% Profit" },
    { id: "cash_ratio", label: "% Cash / Net Worth" },
    { id: "category_breakdown", label: "Cash / Non-Cash / Liabilities" },
  ];

  const ALLOCATION_MODES = [
    { id: "asset_type", label: "By Asset Type" },
    { id: "asset", label: "By Asset" },
    { id: "category", label: "Cash / Non-Cash / Liabilities" },
  ];

  const PROJECTION_OPTIONS = [6, 12, 18, 24, 36];

  const PROGRESS_CHART_COLORS = [
    "#7b530c",
    "#d0b16e",
    "#8d8578",
    "#26231d",
    "#c79f3d",
    "#877d68",
    "#b89a5d",
    "#5f5a52",
    "#dfd7cb",
    "#8b5d12",
  ];

  function buildTrendChartConfig(page, chartMode, projectionMonths) {
    if (chartMode === "category_breakdown") {
      return buildCategoryBreakdownTrendChartConfig(page, projectionMonths);
    }

    const visibleProjectionPoints = chartMode === "net_worth"
      ? sliceProjectionPoints(page.ProjectionPoints || [], projectionMonths)
      : [];
    const labels = page.TrendPoints.map((point) => formatDateLabel(point.SnapshotDate));
    const actualValues = page.TrendPoints.map((point) => resolveTrendValue(point, chartMode));
    const projectionLabelTail = chartMode === "net_worth" && visibleProjectionPoints.length > 1
      ? visibleProjectionPoints.slice(1).map((point) => formatProjectionMonthLabel(point.SnapshotDate))
      : [];
    const allLabels = labels.concat(projectionLabelTail);
    const datasets = [{
      label: resolveChartModeLabel(chartMode),
      data: padSeries(actualValues, allLabels.length),
      borderColor: "#7b530c",
      backgroundColor: "rgba(123, 83, 12, 0.12)",
      borderWidth: 3,
      pointRadius: 4,
      pointHoverRadius: 6,
      pointHitRadius: 24,
      fill: false,
      tension: 0.24,
    }];

    if (chartMode === "net_worth" && visibleProjectionPoints.length) {
      const projectionValues = new Array(allLabels.length).fill(null);
      for (const [index, point] of visibleProjectionPoints.entries()) {
        const targetIndex = Math.max(page.TrendPoints.length - 1, 0) + index;
        if (targetIndex < projectionValues.length) {
          projectionValues[targetIndex] = Number(point.TotalCurrent || 0);
        }
      }

      datasets.push({
        label: "Projection",
        data: projectionValues,
        borderColor: "#8d8578",
        backgroundColor: "rgba(141, 133, 120, 0.10)",
        borderDash: [8, 6],
        borderWidth: 3,
        pointRadius: 2,
        pointHoverRadius: 4,
        pointHitRadius: 22,
        fill: false,
        tension: 0.18,
      });
    }

    if (chartMode === "net_worth") {
      const visibleGoalLines = buildVisibleGoalDatasets(page, allLabels, actualValues, visibleProjectionPoints);
      datasets.push(...visibleGoalLines);
    }

    return {
      type: "line",
      data: {
        labels: allLabels,
        datasets,
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          mode: "index",
          intersect: false,
          axis: "x",
        },
        plugins: {
          legend: {
            display: datasets.length > 1,
          },
          tooltip: {
            callbacks: {
              label(context) {
                const value = Number(context.parsed.y ?? context.parsed);
                return `${context.dataset.label}: ${formatChartTooltipValue(value, "category_breakdown")}`;
              },
            },
          },
        },
        scales: {
          y: {
            ticks: {
              maxTicksLimit: 6,
              precision: 2,
              callback(value) {
                if (chartMode === "net_worth") {
                  return formatNumberWithTwoDecimals(value);
                }
                return `${(Number(value) * 100).toFixed(2)}%`;
              },
            },
            grid: {
              color: "rgba(181, 172, 156, 0.28)",
            },
          },
          x: {
            ticks: {
              autoSkip: true,
              maxRotation: 0,
              maxTicksLimit: 8,
            },
            grid: {
              display: false,
            },
          },
        },
      },
    };
  }

  function buildCategoryBreakdownTrendChartConfig(page, projectionMonths) {
    const labels = page.TrendPoints.map((point) => formatDateLabel(point.SnapshotDate));
    const visibleProjectionPoints = sliceProjectionPoints(page.ProjectionPoints || [], projectionMonths);
    const projectionLabelTail = visibleProjectionPoints.length > 1
      ? visibleProjectionPoints.slice(1).map((point) => formatProjectionMonthLabel(point.SnapshotDate))
      : [];
    const allLabels = labels.concat(projectionLabelTail);
    const datasets = [
      buildCategoryTrendDataset(page, "Cash", "#7b530c", allLabels.length),
      buildCategoryTrendDataset(page, "Non Cash Asset", "#d0b16e", allLabels.length),
      buildCategoryTrendDataset(page, "Liabilities", "#26231d", allLabels.length),
    ];
    datasets.push(...buildCategoryProjectionDatasets(visibleProjectionPoints, page.TrendPoints.length, allLabels.length));

    return {
      type: "line",
      data: {
        labels: allLabels,
        datasets,
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          mode: "index",
          intersect: false,
          axis: "x",
        },
        plugins: {
          legend: {
            display: true,
          },
          tooltip: {
            callbacks: {
              label(context) {
                const value = Number(context.parsed.y ?? context.parsed);
                return `${context.dataset.label}: ${formatChartTooltipValue(value, "category_breakdown")}`;
              },
            },
          },
        },
        scales: {
          y: {
            ticks: {
              maxTicksLimit: 6,
              precision: 2,
              callback(value) {
                return formatNumberWithTwoDecimals(value);
              },
            },
            grid: {
              color: "rgba(181, 172, 156, 0.28)",
            },
          },
          x: {
            ticks: {
              autoSkip: true,
              maxRotation: 0,
              maxTicksLimit: 8,
            },
            grid: {
              display: false,
            },
          },
        },
      },
    };
  }

  function buildCategoryTrendDataset(page, categoryName, borderColor, totalLabelCount) {
    const values = (page.TrendPoints || []).map((point) => resolveCategoryTrendValue(page, point.SnapshotDate, categoryName));
    return {
      label: categoryName,
      data: padSeries(values, totalLabelCount),
      borderColor,
      backgroundColor: "transparent",
      borderWidth: 3,
      pointRadius: 4,
      pointHoverRadius: 6,
      pointHitRadius: 24,
      fill: false,
      tension: 0.24,
    };
  }

  function buildCategoryProjectionDatasets(projectionPoints, historicalPointCount, totalLabelCount) {
    if (!projectionPoints.length) {
      return [];
    }

    return [
      buildCategoryProjectionDataset(projectionPoints, historicalPointCount, totalLabelCount, "Cash", "#7b530c"),
      buildCategoryProjectionDataset(projectionPoints, historicalPointCount, totalLabelCount, "Non Cash Asset", "#d0b16e"),
      buildCategoryProjectionDataset(projectionPoints, historicalPointCount, totalLabelCount, "Liabilities", "#26231d"),
    ];
  }

  function buildCategoryProjectionDataset(
    projectionPoints,
    historicalPointCount,
    totalLabelCount,
    categoryName,
    borderColor,
  ) {
    const values = new Array(totalLabelCount).fill(null);
    for (const [index, point] of projectionPoints.entries()) {
      const targetIndex = Math.max(historicalPointCount - 1, 0) + index;
      if (targetIndex >= totalLabelCount) {
        continue;
      }
      values[targetIndex] = resolveCategoryProjectionValue(point, categoryName);
    }

    return {
      label: `${categoryName} Projection`,
      data: values,
      borderColor,
      backgroundColor: "transparent",
      borderDash: [8, 6],
      borderWidth: 3,
      pointRadius: 2,
      pointHoverRadius: 4,
      pointHitRadius: 22,
      fill: false,
      tension: 0.18,
    };
  }

  function resolveCategoryTrendValue(page, snapshotDate, categoryName) {
    const snapshot = (page.AllocationSnapshots || []).find((item) => item.SnapshotDate === snapshotDate);
    if (!snapshot) {
      return 0;
    }
    const row = (snapshot.ByCategory || []).find((item) => item.Name === categoryName);
    return Number(row?.Value || 0);
  }

  function resolveCategoryProjectionValue(point, categoryName) {
    if (categoryName === "Cash") {
      return Number(point.TotalCash || 0);
    }
    if (categoryName === "Liabilities") {
      return Number(point.Liabilities || 0);
    }
    return Number(point.TotalNonCash || 0);
  }

  function buildAllocationChartConfig(page, allocationDate, allocationMode) {
    const rows = buildAllocationRows(page, allocationDate, allocationMode);
    const backgroundColor = allocationMode === "category"
      ? rows.map((row) => resolveCategoryChartColor(row.name))
      : rows.map((_, index) => PROGRESS_CHART_COLORS[index % PROGRESS_CHART_COLORS.length]);
    return {
      type: "pie",
      data: {
        labels: rows.map((row) => row.name),
        datasets: [{
          data: rows.map((row) => row.value),
          backgroundColor,
          borderColor: "#f9f5ec",
          borderWidth: 3,
          hoverOffset: 10,
        }],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            display: false,
          },
          tooltip: {
            callbacks: {
              title() {
                return "";
              },
              label(context) {
                const label = String(context.label || rows[context.dataIndex]?.name || "");
                const value = Number(context.raw ?? rows[context.dataIndex]?.value ?? 0);
                return `${label}: ${formatTHB(value)}`;
              },
            },
          },
        },
      },
    };
  }

  function buildVisibleGoalDatasets(page, allLabels, actualValues, projectionPoints) {
    const values = actualValues.slice();
    values.push(...projectionPoints.map((point) => Number(point.TotalCurrent || 0)));
    if (!values.length) {
      return [];
    }

    const minValue = Math.min(...values);
    const maxValue = Math.max(...values);

    return (page.Goals || [])
      .filter((goal) => Number(goal.TargetAmount) >= minValue && Number(goal.TargetAmount) <= maxValue)
      .map((goal, index) => ({
        label: `${goal.Name} Goal`,
        data: new Array(allLabels.length).fill(Number(goal.TargetAmount)),
        borderColor: goal.TargetDate ? "#1b6b7f" : "#275a86",
        backgroundColor: "rgba(27, 107, 127, 0.10)",
        borderDash: [12, 4],
        borderWidth: 2,
        pointRadius: 0,
        pointHoverRadius: 0,
        pointHitRadius: 20,
        fill: false,
        tension: 0,
        order: 10 + index,
      }));
  }

  function buildAllocationRows(page, snapshotDate, allocationMode) {
    const snapshot = (page.AllocationSnapshots || []).find((item) => item.SnapshotDate === snapshotDate);
    if (!snapshot) {
      return [];
    }

    let rows = snapshot.ByAssetType || [];
    if (allocationMode === "asset") {
      rows = snapshot.ByAsset || [];
    }
    if (allocationMode === "category") {
      rows = snapshot.ByCategory || [];
    }

    return rows.map((row) => ({
      name: row.Name,
      value: Number(row.Value || 0),
    }));
  }

  function normalizeProgressDateSelection(snapshots, selectedDate) {
    if (!snapshots || snapshots.length === 0) {
      return "";
    }

    const availableDates = snapshots.map((snapshot) => snapshot.SnapshotDate);
    if (availableDates.includes(selectedDate)) {
      return selectedDate;
    }

    return availableDates[availableDates.length - 1];
  }

  function resolveProgressView(options = {}) {
    if (options.view) {
      return options.view;
    }
    return state.progressView || "trend";
  }

  function sliceProjectionPoints(points, projectionMonths) {
    if (!points.length) {
      return [];
    }
    const months = Number(projectionMonths || 6);
    return points.slice(0, Math.min(points.length, months + 1));
  }

  function goalStatusToneClass(status) {
    if (status === "Reached" || status === "On track") {
      return "delta-positive";
    }
    if (status === "Projected") {
      return "delta-caution";
    }
    if (status === "Behind target" || status === "Needs positive trend") {
      return "delta-negative";
    }
    return "delta-neutral";
  }

  function resolveTrendValue(point, chartMode) {
    if (chartMode === "profit_rate") {
      return Number(point.ProfitRate || 0);
    }
    if (chartMode === "cash_ratio") {
      return Number(point.CashRatio || 0);
    }
    return Number(point.TotalCurrent || 0);
  }

  function formatChartAxisValue(value, chartMode) {
    if (chartMode === "net_worth" || chartMode === "category_breakdown") {
      return formatNumberWithTwoDecimals(value);
    }
    return formatPercent(value);
  }

  function formatChartTooltipValue(value, chartMode) {
    if (chartMode === "net_worth" || chartMode === "category_breakdown") {
      return formatTHB(value);
    }
    return formatPercent(value);
  }

  function resolveChartModeLabel(chartMode) {
    return CHART_MODES.find((mode) => mode.id === chartMode)?.label || "Total Net Worth";
  }

  function resolveCategoryChartColor(name) {
    const normalized = String(name || "").trim().toLowerCase();
    if (normalized === "cash") {
      return "#7b530c";
    }
    if (normalized === "liabilities") {
      return "#26231d";
    }
    return "#d0b16e";
  }

  function formatProjectionMonthLabel(value) {
    if (!value) {
      return "";
    }

    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) {
      return value;
    }

    return parsed.toLocaleDateString("en-GB", {
      month: "short",
      year: "numeric",
      timeZone: "UTC",
    });
  }

  function formatNumberWithTwoDecimals(value) {
    return Number(value || 0).toLocaleString("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
  }

  function padSeries(values, length) {
    const padded = new Array(length).fill(null);
    values.forEach((value, index) => {
      padded[index] = value;
    });
    return padded;
  }

  return {
    ALLOCATION_MODES,
    CHART_MODES,
    PROGRESS_CHART_COLORS,
    PROGRESS_VIEWS,
    PROJECTION_OPTIONS,
    buildAllocationChartConfig,
    buildAllocationRows,
    buildTrendChartConfig,
    formatChartAxisValue,
    formatChartTooltipValue,
    formatProjectionMonthLabel,
    goalStatusToneClass,
    normalizeProgressDateSelection,
    resolveChartModeLabel,
    resolveCategoryChartColor,
    resolveProgressView,
    sliceProjectionPoints,
  };
}));
