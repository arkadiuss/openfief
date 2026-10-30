// The layout's behavior on every page of OpenFief's built-in UI: the menu
// that folds the navigation away on a phone, table rows that turn into
// cards at phone width, and the service worker that makes the UI an
// installable app (docs/concepts/ui.md). Loaded in <head> without defer, so
// the page is marked as scripted before it first paints; nothing inline
// (CSP). Without this script the navigation stays open and wide tables
// scroll inside their box.
'use strict';

(function () {
  document.documentElement.classList.add('js');

  // The menu button opens and closes the navigation on a narrow screen;
  // Escape closes it and gives the focus back to the button.
  function menu() {
    const button = document.querySelector('.menu-toggle');
    const panel = button && document.getElementById(button.getAttribute('aria-controls'));
    if (!panel) {
      return;
    }
    function set(open) {
      button.setAttribute('aria-expanded', open ? 'true' : 'false');
      panel.classList.toggle('open', open);
    }
    button.addEventListener('click', function () {
      set(button.getAttribute('aria-expanded') !== 'true');
    });
    document.addEventListener('keydown', function (event) {
      if (event.key === 'Escape' && button.getAttribute('aria-expanded') === 'true') {
        set(false);
        button.focus();
      }
    });
  }

  // Labels every cell of a table with its column's header, so at phone
  // width (style.css) each row reads as a card of "header: value" lines.
  // The roles keep the table's meaning for screen readers once the cells
  // no longer lay out as a table. Rows already labelled are skipped, so a
  // table whose rows an htmx swap adds to (more transactions) is labelled
  // again as a whole.
  function labelTable(table) {
    const head = table.tHead && table.tHead.rows[table.tHead.rows.length - 1];
    if (!head) {
      return;
    }
    const headers = [];
    Array.prototype.forEach.call(head.cells, function (th) {
      for (let i = 0; i < th.colSpan; i++) {
        headers.push(th.textContent.trim());
      }
    });
    table.setAttribute('role', 'table');
    Array.prototype.forEach.call(table.querySelectorAll('thead, tbody, tfoot'), function (group) {
      group.setAttribute('role', 'rowgroup');
    });
    Array.prototype.forEach.call(table.rows, function (row) {
      if (row.dataset.labelled) {
        return;
      }
      row.dataset.labelled = 'true';
      row.setAttribute('role', 'row');
      let column = 0;
      Array.prototype.forEach.call(row.cells, function (cell) {
        const header = cell.tagName === 'TH';
        if (header) {
          cell.setAttribute('role', row.parentNode === table.tHead ? 'columnheader' : 'rowheader');
        } else {
          cell.setAttribute('role', 'cell');
        }
        if (!header && cell.colSpan === 1 && headers[column]) {
          cell.dataset.label = headers[column];
        }
        column += cell.colSpan;
      });
    });
    table.parentNode.classList.add('stack');
  }

  function labelTables(root) {
    const within = root.closest && root.closest('.table > table');
    if (within) {
      labelTable(within);
    }
    if (root.querySelectorAll) {
      Array.prototype.forEach.call(root.querySelectorAll('.table > table'), labelTable);
    }
  }

  document.addEventListener('DOMContentLoaded', function () {
    menu();
    labelTables(document);
  });
  document.addEventListener('htmx:load', function (event) {
    labelTables(event.target);
  });

  // The service worker caches only static files; see /sw.js.
  if ('serviceWorker' in navigator && window.isSecureContext) {
    window.addEventListener('load', function () {
      navigator.serviceWorker.register('/sw.js', { scope: '/' }).catch(function () {
        // Without it the UI works the same, only not offline.
      });
    });
  }
})();
