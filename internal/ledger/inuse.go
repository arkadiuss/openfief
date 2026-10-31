package ledger

import "errors"

// ErrAccountHasTransactions reports an account that holds transactions or
// balance assertions, which therefore can't be deleted (D-120): nothing the
// ledger holds is ever deleted (D-18).
var ErrAccountHasTransactions = errors.New("the account holds transactions")
