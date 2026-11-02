package ui

import (
	"net/http"
	"testing"
)

// TestOKIAccountForm checks the OKI account form at creation (D-147): the
// account forms offer its four forms, labelled, with none chosen, and the
// form posted goes to CreateAccount; the accounts list names an account's
// form next to its wrapper.
func TestOKIAccountForm(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, accountsPath, ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `<label for="new-account-form">OKI account form</label>`, `name="account_form"`,
		`<option value="" selected>Not an OKI account</option>`, `<option value="brokerage-account">Brokerage account</option>`,
		`<option value="treasury-bond-account">Treasury bond account</option>`, "set once")
	post := h.serve(t, withSession(form(accountsPath, "name=OKI&sleeve=core&institution=xtb&tax_wrapper=oki&account_form=brokerage-account&next=/accounts"), ownerSession))
	wantStatus(t, post, http.StatusSeeOther)
	if got := h.accounts.lastCreate(); got.GetAccountForm() != "brokerage-account" || got.GetTaxWrapper() != "oki" {
		t.Errorf("CreateAccount got form %q in %q, want brokerage-account in oki", got.GetAccountForm(), got.GetTaxWrapper())
	}
	if got := formLabel("fund-register"); got != "Fund register" {
		t.Errorf(`formLabel("fund-register") = %q`, got)
	}
}

// TestIKEEarlierPayments checks the earlier payments of an IKE whose history
// is unknown (D-149): its edit form shows the amount entered with its
// currency, and an amount typed with grouping spaces and a decimal comma
// goes to UpdateAccount as a decimal.
func TestIKEEarlierPayments(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.read.locked(func() { h.read.ike = true })
	rec := h.serve(t, withSession(request(http.MethodGet, accountsPath, ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `<label for="acc-`+accountID+`-paid">Paid in before tracking</label>`, `name="earlier_payments" value="30000"`, "<span>PLN</span>",
		"liquidation value stays missing until you enter it")
	post := h.serve(t, withSession(form("/accounts/"+accountID+"/edit", "name=IKE&sleeve=core&tax_wrapper=ike&earlier_payments=30+000,50&next=/accounts"), ownerSession))
	wantStatus(t, post, http.StatusSeeOther)
	if got := h.accounts.lastUpdate().GetEarlierPayments(); got != "30000.50" {
		t.Errorf("UpdateAccount got earlier payments %q, want 30000.50", got)
	}
}

// TestEarlierPaymentsItemPage checks that the item of an IKE's missing
// earlier payments sends the owner to the accounts page, where they are
// entered (D-149), and the marginal rate's to the settings page.
func TestEarlierPaymentsItemPage(t *testing.T) {
	t.Parallel()
	if got := attentionItemKind("setting-missing", "earlier-payments/acc-ike"); got.Page != accountsPath {
		t.Errorf("earlier payments' item goes to %q, want %q", got.Page, accountsPath)
	}
	if got := attentionItemKind("setting-missing", "marginal-rate"); got.Page != "/settings" {
		t.Errorf("the marginal rate's item goes to %q, want /settings", got.Page)
	}
}
