/*
Copyright 2022 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reconcile

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"

	v2 "github.com/fluxcd/helm-controller/api/v2beta2"
	"github.com/fluxcd/helm-controller/internal/action"
)

const (
	mockReleaseName      = "mock-release"
	mockReleaseNamespace = "mock-ns"
)

func Test_summarize(t *testing.T) {
	tests := []struct {
		name       string
		generation int64
		spec       *v2.HelmReleaseSpec
		conditions []metav1.Condition
		expect     []metav1.Condition
	}{
		{
			name:       "summarize conditions",
			generation: 1,
			conditions: []metav1.Condition{
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.InstallSucceededReason,
					Message:            "Install complete",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.TestSuccessCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.TestFailedReason,
					Message:            "test hook(s) failure",
					ObservedGeneration: 1,
				},
			},
			expect: []metav1.Condition{
				{
					Type:               meta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.InstallSucceededReason,
					Message:            "Install complete",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.InstallSucceededReason,
					Message:            "Install complete",
					ObservedGeneration: 1,
				},
			},
		},
		{
			name:       "with tests enabled",
			generation: 1,
			conditions: []metav1.Condition{
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.InstallSucceededReason,
					Message:            "Install complete",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.TestSuccessCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.TestSucceededReason,
					Message:            "test hook(s) succeeded",
					ObservedGeneration: 1,
				},
			},
			spec: &v2.HelmReleaseSpec{
				Test: &v2.Test{
					Enable: true,
				},
			},
			expect: []metav1.Condition{
				{
					Type:               meta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.TestSucceededReason,
					Message:            "test hook(s) succeeded",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.InstallSucceededReason,
					Message:            "Install complete",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.TestSuccessCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.TestSucceededReason,
					Message:            "test hook(s) succeeded",
					ObservedGeneration: 1,
				},
			},
		},
		{
			name:       "with tests enabled and failure tests",
			generation: 1,
			conditions: []metav1.Condition{
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.InstallSucceededReason,
					Message:            "Install complete",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.TestSuccessCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.TestFailedReason,
					Message:            "test hook(s) failure",
					ObservedGeneration: 1,
				},
			},
			spec: &v2.HelmReleaseSpec{
				Test: &v2.Test{
					Enable: true,
				},
			},
			expect: []metav1.Condition{
				{
					Type:               meta.ReadyCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.TestFailedReason,
					Message:            "test hook(s) failure",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.InstallSucceededReason,
					Message:            "Install complete",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.TestSuccessCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.TestFailedReason,
					Message:            "test hook(s) failure",
					ObservedGeneration: 1,
				},
			},
		},
		{
			name: "with test hooks enabled and pending tests",
			conditions: []metav1.Condition{
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.InstallSucceededReason,
					Message:            "Install complete",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.TestSuccessCondition,
					Status:             metav1.ConditionUnknown,
					Reason:             "AwaitingTests",
					Message:            "Release is awaiting tests",
					ObservedGeneration: 1,
				},
			},
			spec: &v2.HelmReleaseSpec{
				Test: &v2.Test{
					Enable: true,
				},
			},
			expect: []metav1.Condition{
				{
					Type:               meta.ReadyCondition,
					Status:             metav1.ConditionUnknown,
					Reason:             "AwaitingTests",
					Message:            "Release is awaiting tests",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.InstallSucceededReason,
					Message:            "Install complete",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.TestSuccessCondition,
					Status:             metav1.ConditionUnknown,
					Reason:             "AwaitingTests",
					Message:            "Release is awaiting tests",
					ObservedGeneration: 1,
				},
			},
		},
		{
			name:       "with remediation failure",
			generation: 1,
			conditions: []metav1.Condition{
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.InstallSucceededReason,
					Message:            "Install complete",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.TestSuccessCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.TestFailedReason,
					Message:            "test hook(s) failure",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.RemediatedCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.UninstallFailedReason,
					Message:            "Uninstall failure",
					ObservedGeneration: 1,
				},
			},
			spec: &v2.HelmReleaseSpec{
				Test: &v2.Test{
					Enable: true,
				},
			},
			expect: []metav1.Condition{
				{
					Type:               meta.ReadyCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.UninstallFailedReason,
					Message:            "Uninstall failure",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.InstallSucceededReason,
					Message:            "Install complete",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.TestSuccessCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.TestFailedReason,
					Message:            "test hook(s) failure",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.RemediatedCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.UninstallFailedReason,
					Message:            "Uninstall failure",
					ObservedGeneration: 1,
				},
			},
		},
		{
			name:       "with remediation success",
			generation: 1,
			conditions: []metav1.Condition{
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.UpgradeFailedReason,
					Message:            "Upgrade failure",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.RemediatedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.RollbackSucceededReason,
					Message:            "Uninstall complete",
					ObservedGeneration: 1,
				},
			},
			expect: []metav1.Condition{
				{
					Type:               meta.ReadyCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.RollbackSucceededReason,
					Message:            "Uninstall complete",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.UpgradeFailedReason,
					Message:            "Upgrade failure",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.RemediatedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.RollbackSucceededReason,
					Message:            "Uninstall complete",
					ObservedGeneration: 1,
				},
			},
		},
		{
			name:       "with stale ready",
			generation: 1,
			conditions: []metav1.Condition{
				{
					Type:    meta.ReadyCondition,
					Status:  metav1.ConditionFalse,
					Reason:  "ChartNotFound",
					Message: "chart not found",
				},
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.UpgradeSucceededReason,
					Message:            "Upgrade finished",
					ObservedGeneration: 1,
				},
			},
			expect: []metav1.Condition{
				{
					Type:               meta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.UpgradeSucceededReason,
					Message:            "Upgrade finished",
					ObservedGeneration: 1,
				},
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.UpgradeSucceededReason,
					Message:            "Upgrade finished",
					ObservedGeneration: 1,
				},
			},
		},
		{
			name:       "with stale observed generation",
			generation: 5,
			spec: &v2.HelmReleaseSpec{
				Test: &v2.Test{
					Enable: true,
				},
			},
			conditions: []metav1.Condition{
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.UpgradeSucceededReason,
					Message:            "Upgrade finished",
					ObservedGeneration: 4,
				},
				{
					Type:               v2.RemediatedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.RollbackSucceededReason,
					Message:            "Rollback finished",
					ObservedGeneration: 3,
				},
				{
					Type:               v2.TestSuccessCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.TestFailedReason,
					Message:            "test hook(s) failure",
					ObservedGeneration: 2,
				},
			},
			expect: []metav1.Condition{
				{
					Type:               meta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.UpgradeSucceededReason,
					Message:            "Upgrade finished",
					ObservedGeneration: 5,
				},
				{
					Type:               v2.ReleasedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.UpgradeSucceededReason,
					Message:            "Upgrade finished",
					ObservedGeneration: 4,
				},
				{
					Type:               v2.RemediatedCondition,
					Status:             metav1.ConditionTrue,
					Reason:             v2.RollbackSucceededReason,
					Message:            "Rollback finished",
					ObservedGeneration: 3,
				},
				{
					Type:               v2.TestSuccessCondition,
					Status:             metav1.ConditionFalse,
					Reason:             v2.TestFailedReason,
					Message:            "test hook(s) failure",
					ObservedGeneration: 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			obj := &v2.HelmRelease{
				ObjectMeta: metav1.ObjectMeta{
					Generation: tt.generation,
				},
				Status: v2.HelmReleaseStatus{
					Conditions: tt.conditions,
				},
			}
			if tt.spec != nil {
				obj.Spec = *tt.spec.DeepCopy()
			}
			summarize(&Request{Object: obj})

			g.Expect(obj.Status.Conditions).To(conditions.MatchConditions(tt.expect))
		})
	}
}

func mockLogBuffer(size int, lines int) *action.LogBuffer {
	log := action.NewLogBuffer(action.NewDebugLog(logr.Discard()), size)
	for i := 0; i < lines; i++ {
		log.Log("line %d", i+1)
	}
	return log
}
