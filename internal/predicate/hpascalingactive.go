package predicate

import (
	"reflect"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// HPAScalingActiveChangedPredicate focuses only on specific HPA field changes
type HPAScalingActiveChangedPredicate struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating HPA specific changes
func (HPAScalingActiveChangedPredicate) Update(e event.UpdateEvent) bool {
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

	var oldCondition *autoscalingv2.HorizontalPodAutoscalerCondition
	for _, condition := range oldHPA.Status.Conditions {
		if condition.Type == autoscalingv2.ScalingActive {
			oldCondition = &condition
			break
		}
	}

	var newCondition *autoscalingv2.HorizontalPodAutoscalerCondition
	for _, condition := range newHPA.Status.Conditions {
		if condition.Type == autoscalingv2.ScalingActive {
			newCondition = &condition
			break
		}
	}

	return !reflect.DeepEqual(oldCondition, newCondition)
}
