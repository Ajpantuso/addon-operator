package addon

import (
	"context"
	"fmt"
	"testing"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
	"github.com/openshift/addon-operator/internal/testutil"
)

func TestObserveCurrentCSV(t *testing.T) {
	type expected struct {
		Conditions []metav1.Condition
		Result     requeueResult
	}

	for name, tc := range map[string]struct {
		CSV      *operatorsv1alpha1.ClusterServiceVersion
		Expected expected
	}{
		"No CSV present": {
			CSV: new(operatorsv1alpha1.ClusterServiceVersion),
			Expected: expected{
				Conditions: []metav1.Condition{unreadyCSVCondition("unkown/pending")},
				Result:     resultRetry,
			},
		},
	} {
		tc := tc

		t.Run(name, func(t *testing.T) {
			c := testutil.NewClient()
			c.
				On("Get",
					mock.Anything,
					mock.IsType(client.ObjectKey{}),
					testutil.IsOperatorsV1Alpha1ClusterServiceVersionPtr,
				).
				Run(func(args mock.Arguments) {
					tc.CSV.DeepCopyInto(args.Get(2).(*operatorsv1alpha1.ClusterServiceVersion))
				}).
				Return(nil)

			r := &olmReconciler{
				client: c,
				scheme: testutil.NewTestSchemeWithAddonsv1alpha1(),
			}

			var addon addonsv1alpha1.Addon

			res, err := r.observeCurrentCSV(context.Background(), &addon, client.ObjectKey{})
			require.NoError(t, err)

			c.AssertExpectations(t)

			assert.Equal(t, tc.Expected.Result, res)
			assertEqualConditions(t, tc.Expected.Conditions, addon.Status.Conditions)
		})
	}
}

func unreadyCSVCondition(msg string) metav1.Condition {
	return metav1.Condition{
		Type:    addonsv1alpha1.Available,
		Status:  metav1.ConditionFalse,
		Reason:  addonsv1alpha1.AddonReasonUnreadyCSV,
		Message: fmt.Sprintf("ClusterServiceVersion is not ready: %s", msg),
	}
}

func assertEqualConditions(t *testing.T, expected []metav1.Condition, actual []metav1.Condition) {
	t.Helper()

	assert.ElementsMatch(t, dropConditionTransients(expected...), dropConditionTransients(actual...))
}

func dropConditionTransients(conds ...metav1.Condition) []nonTransientCondition {
	res := make([]nonTransientCondition, 0, len(conds))

	for _, c := range conds {
		res = append(res, nonTransientCondition{
			Type:    c.Type,
			Status:  c.Status,
			Reason:  c.Reason,
			Message: c.Message,
		})
	}

	return res
}

type nonTransientCondition struct {
	Type    string
	Status  metav1.ConditionStatus
	Reason  string
	Message string
}
