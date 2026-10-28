---
title: Transaction pages follow date then key order with an opaque cursor
status: active
sourceFile: internal/views/page_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:18:02.190Z'
id: TEST-views-page
type: test
---
Five synthetic transactions given out of order are paged with views.PageTransactions: the first page of two is a and b with a cursor at b, the next page after b is c and d, after d only e with no cursor, a page of five holds all, and after the last transaction the page is empty. The input slice is checked to be unchanged. Page sizes 0, -1 and 501 fail with ErrInvalidPageSize and 500 is accepted. A cursor with a key containing URL-special characters survives Encode and ParseTransactionCursor, and empty, non-base64, non-JSON, bad-date and keyless cursors fail with ErrInvalidCursor.