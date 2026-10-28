# Proposals: hosting

Not decided. When the author decides an item, it moves into [DESIGN.md](../DESIGN.md) as a decision and leaves this file.

This area covers GR-1's "it can eventually also be hosted for a subscription".

- **[P-42] Hosted offering scope.** A hosted subscription covers tracking, valuation, analytics and tax only. Agent execution stays self-hosted, on the owner's own account, and the hosted service gives no personalised buy or sell recommendations. Running agents on other people's money, or advising them, would need a KNF permit (see R-7).
- **[P-43] Tenant controls.** Hosted tenants get isolation, per-tenant encryption keys, full export and deletion, and model access to their holdings only when they opt in.

## Security model: hosted, multi-user

- **Data seen by the operator:** every tenant's holdings, transactions and tax data.
- **Models:** "local only" for whole-portfolio data would mean operator-hosted models, which moves the trust boundary from the owner to the operator; any model access is opt-in per tenant (GR-11, P-43).
- **Threats:** cross-tenant leaks; one breach exposing many portfolios; drifting into portfolio management or investment advice.
- **Controls:** P-42, P-43, and a legal opinion before launch (OQ-7).

## Tensions

- **[T-4] GR-1 vs the permit rules.** A hosted subscription is allowed by GR-1, but executing agents for other people, or giving them personalised recommendations, needs a KNF permit (R-7). Proposed: P-42, plus a legal opinion before launch (OQ-7).

## Open questions

- **[OQ-7]** A legal opinion before any hosted offering.
- **[OQ-17]** Does publishing a default agent (GR-10) that other people run on their own accounts raise investment-advice questions? Should the legal opinion in OQ-7 cover it?
