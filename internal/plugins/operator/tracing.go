package operator

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

// tracerScope names the operator's spans (telemetry.Run).
const tracerScope = "github.com/arkadiuss/openfief/internal/plugins/operator"

// traced returns r with each reconcile in a span operator.reconcile that
// names the controller, and nothing about the object: a failed reconcile
// marks the span as an error, and its error is wrapped with the
// controller's name.
func traced(controller string, r reconcile.Reconciler) reconcile.Func {
	return reconcile.Func(func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
		var (
			res reconcile.Result
			err error
		)
		telemetry.Run(ctx, tracerScope, "operator.reconcile", func(ctx context.Context) error {
			got, reconcileErr := r.Reconcile(ctx, req)
			res = got
			if reconcileErr != nil {
				err = fmt.Errorf("%s controller: %w", controller, reconcileErr)
			}
			return err
		}, telemetry.KeyController.String(controller))
		return res, err
	})
}
