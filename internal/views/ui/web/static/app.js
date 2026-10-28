// Draws OpenFief's two charts (D-30) with the vendored Chart.js. Each canvas
// carries its data in data-chart: exact decimal strings from the read API,
// turned into numbers here only to place marks. Tooltips show the exact
// strings. Runs on page load and on every htmx swap; nothing inline (CSP).
'use strict';

(function () {
  function token(name) {
    return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  }

  function config(data) {
    const series = token('--series-1');
    const ink = token('--text-secondary');
    const grid = token('--grid');
    const exact = function (ctx) {
      return data.values[ctx.dataIndex] + ' ' + data.currency;
    };
    const line = data.kind === 'line';
    return {
      type: line ? 'line' : 'bar',
      data: {
        labels: data.labels,
        datasets: [{
          label: data.label,
          data: data.values,
          borderColor: series,
          backgroundColor: series,
          borderWidth: 2,
          borderRadius: line ? 0 : 4,
          maxBarThickness: 28,
          pointRadius: 0,
          pointHoverRadius: 4,
          tension: 0,
        }],
      },
      options: {
        indexAxis: line ? 'x' : 'y',
        maintainAspectRatio: false,
        animation: false,
        interaction: { mode: 'index', intersect: false },
        plugins: {
          legend: { display: false },
          tooltip: { callbacks: { label: exact } },
        },
        scales: {
          x: { ticks: { color: ink, maxTicksLimit: 8 }, grid: { color: grid } },
          y: { ticks: { color: ink }, grid: { color: grid } },
        },
      },
    };
  }

  function draw(root) {
    if (typeof window.Chart !== 'function') {
      return;
    }
    root.querySelectorAll('canvas[data-chart]').forEach(function (canvas) {
      if (canvas.dataset.drawn) {
        return;
      }
      canvas.dataset.drawn = 'true';
      new window.Chart(canvas, config(JSON.parse(canvas.dataset.chart)));
    });
  }

  document.addEventListener('DOMContentLoaded', function () {
    draw(document);
  });
  document.addEventListener('htmx:load', function (event) {
    draw(event.target);
  });
})();
