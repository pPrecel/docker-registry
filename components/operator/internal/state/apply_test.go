package state

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Test_buildSFnApplyResources(t *testing.T) {
	t.Run("switch state and add condition when condition is missing", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{},
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testInstalledDockerRegistry.GetName(),
					Namespace: testInstalledDockerRegistry.GetNamespace(),
				},
				Release: chart.Release{
					Name:      testInstalledDockerRegistry.GetName(),
					Namespace: testInstalledDockerRegistry.GetNamespace(),
				},
			},
			flagsBuilder: chart.NewFlagsBuilder(),
		}

		next, result, err := sFnApplyResources(context.Background(), nil, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnVerifyResources, next)

		expectedFlags := map[string]interface{}{
			"commonLabels": map[string]interface{}{
				"app.kubernetes.io/managed-by": "dockerregistry-operator",
			},
		}

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)

		status := s.instance.Status
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionUnknown,
			v1alpha1.ConditionReasonInstallation,
			"Installing for configuration",
		)
	})

	t.Run("apply resources", func(t *testing.T) {
		s := &systemState{
			instance: *testInstalledDockerRegistry.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testInstalledDockerRegistry.GetName(),
					Namespace: testInstalledDockerRegistry.GetNamespace(),
				},
				Release: chart.Release{
					Name:      testInstalledDockerRegistry.GetName(),
					Namespace: testInstalledDockerRegistry.GetNamespace(),
				},
			},
			flagsBuilder: chart.NewFlagsBuilder(),
		}
		r := &reconciler{}

		// run installation process and return verificating state
		next, result, err := sFnApplyResources(context.Background(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnVerifyResources, next)
	})

	t.Run("install chart error", func(t *testing.T) {
		s := &systemState{
			instance: *testInstalledDockerRegistry.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixManifestCache("\t"),
				CacheKey: types.NamespacedName{
					Name:      testInstalledDockerRegistry.GetName(),
					Namespace: testInstalledDockerRegistry.GetNamespace(),
				},
			},
			flagsBuilder: chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		// handle error and return update condition state
		next, result, err := sFnApplyResources(context.Background(), r, s)
		require.EqualError(t, err, "could not parse chart manifest: yaml: found character that cannot start any token")
		require.Nil(t, result)
		require.Nil(t, next)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateError, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonInstallationErr,
			"could not parse chart manifest: yaml: found character that cannot start any token",
		)
	})
}
