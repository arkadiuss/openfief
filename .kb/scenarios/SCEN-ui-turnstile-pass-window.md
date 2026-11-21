---
title: Within the window a second model question is not checked, after it or past 2 hours from the check it is
status: active
origin:
  kind: agent
  ref: 'The design session''s amendment of D-208 (docs/DESIGN.md): a visitor who passes Turnstile isn''t checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
id: SCEN-ui-turnstile-pass-window
type: scenario
---
Given the Ask box with Turnstile in front of a model and a visitor whose first model question passed the check, when they ask again within 30 minutes of their last model question the question goes with no token and no check runs, and each such question renews the 30 minutes up to 2 hours from the first check; when 31 minutes pass with no question, or 2 hours have passed from the check although the last question was 21 minutes ago, the next question is told a check is needed and runs it.