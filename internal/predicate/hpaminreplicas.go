package predicate

import (
	"reflect"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// HPAMinReplicasChangedPredicate focuses only on specific HPA field changes
type HPAMinReplicasChangedPredicate struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating HPA specific changes
func (HPAMinReplicasChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	oldHPA, ok := e.ObjectOld.(*autoscalingv2.HorizontalPodAutoscaler)
	if !ok {
		return false
	}

	newHPA, ok := e.ObjectNew.(*autoscalingv2.HorizontalPodAutoscaler)
	if !ok {
		return false
	}

	return !reflect.DeepEqual(oldHPA.Spec.MinReplicas, newHPA.Spec.MinReplicas)
}
