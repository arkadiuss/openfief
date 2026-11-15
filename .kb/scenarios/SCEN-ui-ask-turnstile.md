---
title: The Ask box shows Turnstile and loads only its script where it is configured
status: active
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T22:47:04.006Z'
id: SCEN-ui-ask-turnstile
type: scenario
---
Given a UI with Turnstile configured and a model on the Ask box, when the overview is shown then the Ask form holds the widget with the site key and the page loads Cloudflare's script, the page's policy names that origin for scripts and frames and no other outside host, a question goes with the widget's token and the widget renews it, a command needs none, and the console holds no policy violation; without the configuration the page shows no widget, names no outside host and loads nothing from Cloudflare.