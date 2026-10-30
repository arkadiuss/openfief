// Draws OpenFief's two charts (D-30) with the vendored Chart.js. Each canvas
// carries its data in data-chart: exact decimal strings from the read API,
// turned into numbers here only to place marks. Tooltips show the exact
// strings. Axes and tooltips write numbers in the owner's number format
// (D-113), whose separators each chart's data carries. Colors and fonts come from the design system's tokens, read from
// the page's CSS custom properties, so the charts follow light and dark like
// the rest of the page and redraw when the scheme changes. Runs on page load
// and on every htmx swap; nothing inline (CSP).
'use strict';

(function () {
  function token(name) {
    return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  }

  // The series in the design system's order: vert, or, kasztan, a grey, the
  // vert tint (style.css maps them onto the semantic tokens).
  function palette() {
    return ['--series-1', '--series-2', '--series-3', '--series-4', '--series-5'].map(token);
  }

  // Writes a plain decimal string, such as "-1234567.8", with the number
  // format's separators: only they change, never the digits.
  function separate(text, data) {
    const sign = text.charAt(0) === '-' ? '-' : '';
    const parts = text.replace(/^[-+]/, '').split('.');
    const whole = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, data.group || '');
    return sign + whole + (parts.length > 1 ? (data.decimal || '.') + parts[1] : '');
  }

  function config(data) {
    const ink = token('--chart-ink');
    const grid = token('--chart-grid');
    const colors = palette();
    const font = { family: token('--font-sans') };
    const exact = function (ctx) {
      const values = ctx.datasetIndex === 0 ? data.values : data.series[ctx.datasetIndex - 1].values;
      return ctx.dataset.label + ': ' + separate(values[ctx.dataIndex], data) + ' ' + data.currency;
    };
    const line = data.kind === 'line';
    const lines = [{ label: data.label, values: data.values }].concat(data.series || []);
    const datasets = lines.map(function (l, i) {
      // A line takes its series' color; the allocation's bars take one
      // color each, in the same order.
      const color = line ? colors[i % colors.length] : l.values.map(function (_, j) { return colors[j % colors.length]; });
      return {
        label: l.label,
        // An empty value is a day the line has none.
        data: l.values.map(function (v) { return v === '' ? null : v; }),
        borderColor: color,
        backgroundColor: color,
        borderWidth: line ? 2 : 0,
        borderRadius: line ? 0 : 2,
        maxBarThickness: 28,
        pointRadius: 0,
        pointHoverRadius: 4,
        tension: 0,
        spanGaps: false,
      };
    });
    // The value axis writes its numbers in the format, rounded to 12
    // significant digits against floating-point noise; the other axis
    // keeps its labels (dates, or the allocation's groups).
    const numeric = line ? 'y' : 'x';
    const ticks = function (axis) {
      const callback = axis === numeric ?
        function (value) { return separate(String(parseFloat(Number(value).toPrecision(12))), data); } :
        function (value) { return this.getLabelForValue(value); };
      return { color: ink, font: font, maxTicksLimit: axis === 'x' && line ? 8 : undefined, callback: callback };
    };
    return {
      type: line ? 'line' : 'bar',
      data: {
        labels: data.labels,
        datasets: datasets,
      },
      options: {
        indexAxis: line ? 'x' : 'y',
        maintainAspectRatio: false,
        animation: false,
        interaction: { mode: 'index', intersect: false },
        plugins: {
          legend: { display: lines.length > 1, labels: { color: ink, font: font, boxWidth: 14, boxHeight: 2 } },
          tooltip: { callbacks: { label: exact }, titleFont: font, bodyFont: { family: token('--font-mono') } },
        },
        scales: {
          x: { ticks: ticks('x'), grid: { color: grid }, border: { color: grid } },
          y: { ticks: ticks('y'), grid: { color: grid }, border: { color: grid } },
        },
      },
    };
  }

  const charts = [];

  function draw(root) {
    if (typeof window.Chart !== 'function') {
      return;
    }
    root.querySelectorAll('canvas[data-chart]').forEach(function (canvas) {
      if (canvas.dataset.drawn) {
        return;
      }
      canvas.dataset.drawn = 'true';
      charts.push({ canvas: canvas, chart: new window.Chart(canvas, config(JSON.parse(canvas.dataset.chart))) });
    });
  }

  // Redraws every chart still on the page with the tokens' current values,
  // after the reader's color scheme changes.
  function redraw() {
    for (let i = charts.length - 1; i >= 0; i--) {
      const c = charts[i];
      c.chart.destroy();
      if (!c.canvas.isConnected) {
        charts.splice(i, 1);
        continue;
      }
      c.chart = new window.Chart(c.canvas, config(JSON.parse(c.canvas.dataset.chart)));
    }
  }

  document.addEventListener('DOMContentLoaded', function () {
    draw(document);
    if (window.matchMedia) {
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', redraw);
    }
  });
  document.addEventListener('htmx:load', function (event) {
    draw(event.target);
  });
})();
