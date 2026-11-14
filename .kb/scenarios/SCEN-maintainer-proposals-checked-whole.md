---
title: A proposal that breaks one rule is dropped whole and the note stays
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:31.045Z'
id: SCEN-maintainer-proposals-checked-whole
type: scenario
---
Given a model answer with a note and changes, when one change names an unknown target, repeats a target, differs from the model's current weight, has more than two decimal places, leaves the weights off 100% or has a reason of more than a line, then no change is kept, the drop is audited and the note is stored; an answer that is no note at all fails the run.