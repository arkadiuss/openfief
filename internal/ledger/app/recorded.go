package app

import (
	"time"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// RecordedVersion is a stored version of a transaction with when it was
// recorded and its row ID, which orders versions recorded at the same time.
// The read API's activity lists coupons and corrections from these (D-193).
type RecordedVersion struct {
	ledger.VersionedTransaction

	ID       int64
	Recorded time.Time
}

// IsCorrection reports whether v corrects an earlier version of its
// transaction: any version after the first, a deletion included (D-18,
// D-100).
func (v *RecordedVersion) IsCorrection() bool { return v.Version > 1 }

// IsCoupon reports whether v books a coupon: the first version of an
// interest transaction, such as a treasury bond's coupon (D-101).
func (v *RecordedVersion) IsCoupon() bool {
	return v.Version == 1 && v.Transaction.Kind == ledger.KindInterest
}
