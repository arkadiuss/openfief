# Proposals: tracker beyond M0

Not decided. When the author decides an item, it moves into [DESIGN.md](../DESIGN.md) as a decision and leaves this file.

This area covers what else the tracker could hold once M0's brokers and bonds work (GR-3): the retirement plans an employer runs, the equity a technology employer pays in, and a way for a person the owner trusts to find it all. None of it needs a model or an agent, so it can come before M1. Each keeps M0's rules: data enters through connectors or the owner's own entries (D-7, D-79), values come from the jurisdiction pack (D-70, D-77), and nothing leaves the cluster unless the owner sends it (GR-12).

## Employer retirement plans: PPK and PPE

- **[P-50] PPK and PPE as accounts with their own tax wrappers.** Each plan is an account at the institution that runs it, in a sleeve of the retirement kind (D-71), holding fund units. The Polish pack adds PPK and PPE as tax wrappers (D-76), each with D-77's three values (R-20, R-21):
  - **PPK.** Gross is the units at the fund's unit price. Net is the same, since from 60 the path of 25% at once and the rest in at least 120 instalments is free of tax. Liquidation is what a withdrawal today returns: the owner's own contributions in full, 70% of the employer's, none of the state's payments, less 19% on the gains.
  - **PPE.** Gross and net are equal, since a payout from 60 is tax-free. Liquidation before 60 follows the programme's own rules, which the owner enters as the account's fee rule (D-78).

  Every contribution, the employer's and the state's included, enters the account from outside it, so the ledger books it as a deposit and the time-weighted return doesn't count it as performance (D-61). The employer's contributions are also taxed as salary when paid, which the pack needs for the cost basis.
- **[P-51] Getting PPK and PPE data in.** No PPK or PPE institution is known to offer an API. A connector per institution maps the statement it sends, uploaded like a broker's (D-7, D-109). Until one exists, the owner enters each period's contributions and units by hand through an owner service (D-79). Unit prices come from a market-data connector that reads the funds' published prices, and a price check compares them with the statement's own (D-22).

## Employee equity: RSUs, ESPPs and options

- **[P-52] Grants and vesting.** An equity grant is not a holding until it vests: unvested units can't be sold or moved, and leaving the employer usually forfeits them. So a grant is a record of its own: the plan, the instrument, the units granted, the vesting schedule and any forfeits. A vest books shares into the account at the plan's broker, such as Morgan Stanley at Work, Schwab or Fidelity, with the shares sold to cover tax as a sale; from then on that broker's statements bring them in like any other holdings. Unvested units appear as a projection, their value at today's price by vesting date, and stay out of totals and allocation unless the owner turns them on.
- **[P-53] The tax treatment per plan is the owner's call.** Whether a plan is a qualifying incentive programme decides when tax arises: only at the sale, or at vesting as income and again at the sale on the gain above it (R-22). Tax offices disagree about foreign plans, so the owner records the treatment for each plan, usually after asking an advisor, and the pack derives the cost basis, the net and liquidation values (D-77) and the tax-form data (P-39) from it.
- **[P-54] Exposure to the employer.** Shares of the employer move with the owner's salary, so a bad year hits both. The allocation gains an employer view: how much of the portfolio, vested and, if the owner asks, unvested, depends on one company. A model portfolio target (D-52) can cap it.

## The red folder (czerwona teczka)

- **[P-55] A red folder export.** On the owner's request, OpenFief writes one document for the person the owner trusts: each institution and account by name, its kind and tax wrapper, what it held and was worth on that date, the contact the owner noted for it, its beneficiary designation (P-56), and where the owner keeps the will. Account numbers are masked unless the owner chooses otherwise. It never holds a password, token or any other credential (R-23). The owner prints it or saves it encrypted, for example with `age` to the trusted person's key or a passphrase. It carries its date, and the UI reminds the owner when it is older than a set number of months. It leaves the cluster only when the owner exports it, so it adds no new way in.
- **[P-56] Beneficiary designations.** For each account whose institution or wrapper allows one (a bank's payable-on-death instruction, IKE, IKZE, PPK, PPE), the owner records whether a beneficiary is named and when they last checked. OpenFief stores that fact and date; it stores who the beneficiary is only if the owner enters it. The overview flags accounts that allow a designation but have none, since that money otherwise waits for probate (R-23).

## Tensions and open questions

- **[T-13] A trusted person against a single user.** A trusted person reaching OpenFief itself, such as an emergency login with a waiting period as password managers offer, would be a second user, which D-65 rules out. It would also be a standing way in for anyone holding that person's device. And a self-hosted cluster may stop running with its owner. The export (P-55) needs none of that; an online handover would need D-65 revisited.
- **[OQ-31] The default RSU treatment.** Before the owner records a plan's treatment (P-53), which does the pack assume: tax only at the sale, or tax at vesting?
- **[OQ-32] Where PPK data can come from.** Whether any PPK or PPE institution, or the central mojePPK service, offers an export a connector could read, beyond PDF statements.
