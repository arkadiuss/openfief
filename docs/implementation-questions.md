---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Taxes

75. **What an OKI account without a form counts its PLN cash as.**
   - *Problem:* D-147 has the owner set an OKI account's form when creating it and lets the form decide whether its PLN cash is a savings or an investment asset, but an OKI account created before forms were asked for has none, and nothing says how its cash counts until it gets one.
   - *Today:* its PLN cash counts as other assets, which no exemption covers, so the estimate errs high as D-132 does for an instrument; the owner can give such an account its form, and an account moved into OKI must take one then; once set, a form never changes ([Valuation](concepts/valuation.md#okis-yearly-tax), [Ingest](concepts/ingest.md)).
   - *Options:* keep; or count it as an investment asset, the most common form; or raise an attention item asking for the form.

76. **An IKE seeded from a starting balance before its earlier payments are entered.**
   - *Problem:* D-149 has the owner enter the total paid into an IKE whose earlier history is unknown, which its early return is taxed above, but doesn't say what its liquidation value is until they are entered.
   - *Today:* its liquidation total is missing, as an IKZE's is without the marginal rate (D-78), and a `setting-missing` attention item asks for the payments, linking the accounts page; the owner may enter them for any IKE, and they count only once it is seeded. Before D-149 the cost of its opening lots and its opening cash counted as paid in.
   - *Options:* keep; or count the opening lots' cost and opening cash until the owner enters a total, and show that it is a guess.

77. **Which holdings the instrument-to-classify check looks at.**
   - *Problem:* D-148 makes the unclassified instrument's item a check the attention store runs, and D-132 counts the instrument for OKI's exemptions in any year it is held, but neither says over which days the check looks for it.
   - *Today:* the check reads each OKI account's stored valuations in the tax year of the day it is valued through, so an instrument sold in an earlier year no longer raises an item, while that year's tax still counts it as other assets until the owner classifies it.
   - *Options:* keep; or look at every year the act is in force for; or at the current and the previous tax year until the tax is due.
