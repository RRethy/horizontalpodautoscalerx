package controller

import (
	"context"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	autoscalingxv1 "rrethy.io/horizontalpodautoscalerx/api/v1"
)

// HorizontalPodAutoscalerXReconciler reconciles a HorizontalPodAutoscalerX object
type HorizontalPodAutoscalerXReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=autoscalingx.rrethy.io,resources=horizontalpodautoscalerxes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscalingx.rrethy.io,resources=horizontalpodautoscalerxes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=autoscalingx.rrethy.io,resources=horizontalpodautoscalerxes/finalizers,verbs=update
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
// This Reconcile method uses the ObjectReconciler interface from https://github.com/kubernetes-sigs/controller-runtime/pull/2592.
func (r *HorizontalPodAutoscalerXReconciler) Reconcile(ctx context.Context, hpax *autoscalingxv1.HorizontalPodAutoscalerX) (ctrl.Result, error) {
	if !hpax.DeletionTimestamp.IsZero() {
		// The object is being deleted, don't do anything.
		return ctrl.Result{}, nil
	}

	log := ctrl.LoggerFrom(ctx)
	origStatus := hpax.Status.DeepCopy()
	// TODO: initialize status conditions

	defer func() {
		if *origStatus != hpax.Status {
			if err := r.Status().Update(ctx, hpax); err != nil {
				log.Error(err, "updating status")
			}
		}
	}()

	err := r.reconcileHPAX(ctx, hpax)
	if err != nil {
		log.Error(err, "reconciling HorizontalPodAutoscalerX")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorizontalPodAutoscalerXReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalingxv1.HorizontalPodAutoscalerX{}).
		Watches(
			&autoscalingv2.HorizontalPodAutoscaler{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				// TODO: Will this trigger for status changes?
				// We want to reconcile the HorizontalPodAutoscalerX that owns the HorizontalPodAutoscaler.
				// This can be done a variety of ways, but the simplest is to check the OwnerReferences.
				// This works because the HorizontalPodAutoscalerX controller sets the OwnerReferences for minReplicas.
				hpa := obj.(*autoscalingv2.HorizontalPodAutoscaler)
				for _, ref := range hpa.OwnerReferences {
					if ref.Kind == "HorizontalPodAutoscalerX" && ref.APIVersion == autoscalingxv1.GroupVersion.String() {
						return []reconcile.Request{
							{NamespacedName: types.NamespacedName{Name: ref.Name, Namespace: hpa.Namespace}},
						}
					}
				}

				return nil
			}),
		).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{})).
		Named("horizontalpodautoscalerx").
		Complete(reconcile.AsReconciler(mgr.GetClient(), r))
}

func (r *HorizontalPodAutoscalerXReconciler) reconcileHPAX(ctx context.Context, hpax *autoscalingxv1.HorizontalPodAutoscalerX) error {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, client.ObjectKey{Name: hpax.Spec.HPATargetName, Namespace: hpax.Namespace}, hpa)
	if err != nil {
		return fmt.Errorf("getting HorizontalPodAutoscaler %s: %w", hpax.Spec.HPATargetName, err)
	}

	return nil
}
