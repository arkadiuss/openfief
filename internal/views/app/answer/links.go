package answer

// Links says where a card's links go. The pages are the UI's, so it supplies
// the routes; a caller that shows no page, such as the MCP server's tools,
// supplies none and every Href is empty (NoLinks).
type Links interface {
	// Inbox is the Inbox, and InboxItem one of its items.
	Inbox() string
	InboxItem(id int64) string
	// Plugins is the installed plugins' page.
	Plugins() string
	// Account is an account's page, and Sleeve a sleeve on the overview.
	Account(id string) string
	Sleeve(id string) string
	// Instrument is an instrument's page, by ISIN.
	Instrument(isin string) string
	// Allocation is the overview's allocation by the dimension's key
	// (Dimension.Key).
	Allocation(dimension string) string
	// Overview is the overview on the day, YYYY-MM-DD.
	Overview(day string) string
	// PerformanceChart is the performance chart from the day, YYYY-MM-DD.
	PerformanceChart(from string) string
}

// NoLinks is the Links of a caller that draws no page: every link is empty.
type NoLinks struct{}

var _ Links = NoLinks{}

// Inbox returns "".
func (NoLinks) Inbox() string { return "" }

// InboxItem returns "".
func (NoLinks) InboxItem(int64) string { return "" }

// Plugins returns "".
func (NoLinks) Plugins() string { return "" }

// Account returns "".
func (NoLinks) Account(string) string { return "" }

// Sleeve returns "".
func (NoLinks) Sleeve(string) string { return "" }

// Instrument returns "".
func (NoLinks) Instrument(string) string { return "" }

// Allocation returns "".
func (NoLinks) Allocation(string) string { return "" }

// Overview returns "".
func (NoLinks) Overview(string) string { return "" }

// PerformanceChart returns "".
func (NoLinks) PerformanceChart(string) string { return "" }
