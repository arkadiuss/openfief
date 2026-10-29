// Package app holds the portfolio context's use cases for the model
// portfolio: the owner reads and replaces it (D-52), and the read API gets
// its backtest and max drawdowns while it is on (D-53, D-54). The ports it
// needs are declared here: the model's store, the stored end-of-day
// history, and a check that a session is the owner's.
package app
