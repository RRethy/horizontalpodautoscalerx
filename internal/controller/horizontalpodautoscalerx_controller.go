package controller

import (
	"context"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	autoscalingxv1 "rrethy.io/horizontalpodautoscalerx/api/v1"
	custompredicate "rrethy.io/horizontalpodautoscalerx/internal/predicate"
)

const (
	ControllerName = "horizontalpodautoscalerx"
)

// HorizontalPodAutoscalerXReconciler reconciles a HorizontalPodAutoscalerX object
type HorizontalPodAutoscalerXReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock  clock.Clock
}

// +kubebuilder:rbac:groups=autoscalingx.rrethy.io,resources=horizontalpodautoscalerxes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscalingx.rrethy.io,resources=horizontalpodautoscalerxes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=autoscalingx.rrethy.io,resources=horizontalpodautoscalerxes/finalizers,verbs=update
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
// This Reconcile method uses the ObjectReconciler interface from https://github.com/kubernetes-sigs/controller-runtime/pull/2592.
func (r *HorizontalPodAutoscalerXReconciler) Reconcile(ctx context.Context, hpax *autoscalingxv1.HorizontalPodAutoscalerX) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if !hpax.DeletionTimestamp.IsZero() {
		// The object is being deleted, don't do anything.
		return ctrl.Result{}, nil
	}

	hpa, err := r.getHPA(ctx, hpax)
	if err != nil {
		log.Error(err, "getting associated HPA")
		return ctrl.Result{Requeue: true}, nil
	}

	orig := hpax.DeepCopy()

	defer func() {
		if orig.Status != hpax.Status {
			err := r.Status().Update(ctx, hpax)
			if err != nil {
				log.Error(err, "updating status")
			}
		}
	}()

	scalingActiveCondition := r.getScalingActiveCondition(hpa)
	minReplicas := r.getDesiredMinReplicas(hpax, scalingActiveCondition)
	err = r.updateHpaMinReplicas(ctx, hpa, minReplicas)
	if err != nil {
		log.Error(err, "updating HPA minReplicas")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorizontalPodAutoscalerXReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Clock == nil {
		r.Clock = clock.RealClock{}
	}

	// Create an index for hpax.spec.hpaTargetName so we can find the hpax that targets a given hpa.
	err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&autoscalingxv1.HorizontalPodAutoscalerX{},
		"spec.hpaTargetName",
		func(obj client.Object) []string {
			return []string{obj.(*autoscalingxv1.HorizontalPodAutoscalerX).Spec.HPATargetName}
		})
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(ControllerName).
		For(
			&autoscalingxv1.HorizontalPodAutoscalerX{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{}),
		).
		Watches(
			&autoscalingv2.HorizontalPodAutoscaler{},
			handler.EnqueueRequestsFromMapFunc(r.findHPAXForHPA),
			builder.WithPredicates(predicate.Or(
				custompredicate.HPAScalingActiveChangedPredicate{},
				custompredicate.HPAMinReplicasChangedPredicate{},
			)),
		).
		Complete(reconcile.AsReconciler(mgr.GetClient(), r))
}

// findHPAXForHPA finds all HorizontalPodAutoscalerX objects that target the given HPA.
func (r *HorizontalPodAutoscalerXReconciler) findHPAXForHPA(ctx context.Context, o client.Object) []reconcile.Request {
	hpa, ok := o.(*autoscalingv2.HorizontalPodAutoscaler)
	if !ok {
		return nil
	}

	hpaxList := &autoscalingxv1.HorizontalPodAutoscalerXList{}
	if err := r.List(ctx, hpaxList, &client.ListOptions{
		Namespace:     hpa.GetNamespace(),
		FieldSelector: fields.OneTermEqualSelector("spec.hpaTargetName", hpa.GetName()),
	}); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, hpax := range hpaxList.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      hpax.GetName(),
				Namespace: hpax.GetNamespace(),
			},
		})
	}
	return requests
}

// getHPA retrieves the HorizontalPodAutoscaler object associated with the given HorizontalPodAutoscalerX.
func (r *HorizontalPodAutoscalerXReconciler) getHPA(ctx context.Context, hpax *autoscalingxv1.HorizontalPodAutoscalerX) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, client.ObjectKey{Name: hpax.Spec.HPATargetName, Namespace: hpax.Namespace}, hpa)
	if err != nil {
		return nil, err
	}
	return hpa, nil
}

// getScalingActiveCondition retrieves the ScalingActive condition from the HPA status.
func (r *HorizontalPodAutoscalerXReconciler) getScalingActiveCondition(hpa *autoscalingv2.HorizontalPodAutoscaler) *autoscalingv2.HorizontalPodAutoscalerCondition {
	for _, condition := range hpa.Status.Conditions {
		if condition.Type == autoscalingv2.ScalingActive {
			return &condition
		}
	}
	return nil
}

// getDesiredMinReplicas calculates the desired minReplicas based on the scaling condition and fallback settings.
func (r *HorizontalPodAutoscalerXReconciler) getDesiredMinReplicas(hpax *autoscalingxv1.HorizontalPodAutoscalerX, scalingActiveCondition *autoscalingv2.HorizontalPodAutoscalerCondition) int32 {
	if hpax.Spec.Fallback == nil ||
		scalingActiveCondition == nil ||
		scalingActiveCondition.Status == corev1.ConditionTrue ||
		scalingActiveCondition.Status == corev1.ConditionUnknown ||
		scalingActiveCondition.LastTransitionTime.Time.Add(hpax.Spec.Fallback.Duration.Duration).After(r.Clock.Now()) {
		return hpax.Spec.MinReplicas
	}

	return hpax.Spec.Fallback.MinReplicas
}

func (r *HorizontalPodAutoscalerXReconciler) updateHpaMinReplicas(ctx context.Context, hpa *autoscalingv2.HorizontalPodAutoscaler, minReplicas int32) error {
	if hpa.Spec.MinReplicas != nil && minReplicas == *hpa.Spec.MinReplicas {
		return nil
	}
	hpa.Spec.MinReplicas = &minReplicas
	return r.Update(ctx, hpa)
}
