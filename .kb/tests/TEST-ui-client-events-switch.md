---
title: 'The UI follows the owner''s client events setting in the core: page, endpoint, Settings card, owner-only, read once per interval, last answer kept'
status: active
origin:
  kind: agent
  ref: 'D-227 (docs/DESIGN.md): the owner can turn client events off in Settings; the author approved the deferral of the end-to-end test of /client-events (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
sourceFile: internal/views/ui/clientswitch_test.go
verification_scope: integration
id: TEST-ui-client-events-switch
type: test
---
Through the UI over its fake core, with a clock the test moves: on by default the login page loads events.js and a report is logged; the owner's post to /settings/client-events turns it off and the page loses the script while the endpoint answers 204 and logs nothing; on again it works; a read session, a viewer and no session change nothing; Settings shows the pill, the one line about what is collected and the button, and the install's config shown as forcing it off; five pages read the core once, a change in the core shows after 10 seconds and not before; with the core down the last answer is kept, and with none the UI says off. A false pass would be a page that loads the script with the setting off.