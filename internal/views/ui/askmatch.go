package ui

// maxSuggestions is how many suggestion chips the Ask box shows.
const maxSuggestions = 5

// askState is what the Ask box's suggestion chips are drawn from: how
// many items are open, the connector instances whose sync keeps failing,
// and the accounts' and sleeves' names, in their order.
type askState struct {
	Open     int
	Failing  []string
	Accounts []string
	Sleeves  []string
}

// askSuggestions returns at most maxSuggestions questions drawn from st,
// each one the box understands: what needs the owner while anything is
// open, why each failing sync fails, the first account and sleeve, then
// the allocation by currency and this year's performance.
func askSuggestions(st *askState) []string {
	var out []string
	if st.Open > 0 {
		out = append(out, "What needs me?")
	}
	for _, inst := range st.Failing {
		out = append(out, "Why is "+inst+" failing?")
	}
	if len(st.Accounts) > 0 {
		out = append(out, "Show "+st.Accounts[0])
	}
	if len(st.Sleeves) > 0 {
		out = append(out, "Show sleeve "+st.Sleeves[0])
	}
	out = append(out, "Allocation by currency", "Performance this year", "Value at end of last month")
	if len(out) > maxSuggestions {
		out = out[:maxSuggestions]
	}
	return out
}
