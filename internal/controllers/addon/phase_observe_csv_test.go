package addon

import (
	"context"
	"testing"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
	"github.com/openshift/addon-operator/internal/testutil"
)

func TestObserveCurrentCSV(t *testing.T) {

	addon := &addonsv1alpha1.Addon{
		ObjectMeta: metav1.ObjectMeta{
			Name: "addon-mock",
		},
		Spec: addonsv1alpha1.AddonSpec{
			Install: addonsv1alpha1.AddonInstallSpec{
				Type: addonsv1alpha1.OLMOwnNamespace,
				OLMOwnNamespace: &addonsv1alpha1.AddonInstallOLMOwnNamespace{
					AddonInstallOLMCommon: addonsv1alpha1.AddonInstallOLMCommon{
						CatalogSourceImage: "test",
						Namespace:          "test",
					},
				},
			},
			SecretPropagation: &addonsv1alpha1.AddonSecretPropagation{
				Secrets: []addonsv1alpha1.AddonSecretPropagationReference{
					{
						SourceSecret: corev1.LocalObjectReference{
							Name: "test",
						},
						DestinationSecret: corev1.LocalObjectReference{
							Name: "test",
						},
					},
				},
			},
		},
	}

	csv := &operatorsv1alpha1.ClusterServiceVersion{}

	c := testutil.NewClient()
	c.
		On("Get",
			mock.Anything,
			mock.IsType(client.ObjectKey{}),
			mock.IsType(csv),
		).Return(nil)

	r := &olmReconciler{
		client: c,
		scheme: testutil.NewTestSchemeWithAddonsv1alpha1(),
	}

	secretKey := client.ObjectKey{Name: "test", Namespace: "test"}

	ctx := context.Background()
	requeueResult, err := r.observeCurrentCSV(ctx, addon, secretKey)

	c.AssertExpectations(t)
	require.NoError(t, err)
	assert.NotNil(t, requeueResult)
}
