package controller

import (
	"context"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	autoscalingxv1 "rrethy.io/horizontalpodautoscalerx/api/v1"
)

const (
	ControllerName = "horizontalpodautoscalerx"
	requeueAfter   = 30
)

// HorizontalPodAutoscalerXReconciler reconciles a HorizontalPodAutoscalerX object
type HorizontalPodAutoscalerXReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

	_, err := r.getHPA(ctx, hpax)
	if err != nil {
		log.Error(err, "getting associated HPA")
		return ctrl.Result{
			RequeueAfter: requeueAfter,
		}, nil
	}

	defer func() {
		err := r.Status().Update(ctx, hpax)
		if err != nil {
			log.Error(err, "updating status")
		}
	}()

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorizontalPodAutoscalerXReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalingxv1.HorizontalPodAutoscalerX{}).
		Named(ControllerName).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{})).
		Complete(reconcile.AsReconciler(mgr.GetClient(), r))
}

func (r *HorizontalPodAutoscalerXReconciler) getHPA(ctx context.Context, hpax *autoscalingxv1.HorizontalPodAutoscalerX) (*autoscalingv1.HorizontalPodAutoscaler, error) {
	hpa := &autoscalingv1.HorizontalPodAutoscaler{}
	err := r.Get(ctx, client.ObjectKey{Name: hpax.Spec.HPATargetName, Namespace: hpax.Namespace}, hpa)
	if err != nil {
		return nil, err
	}
	return hpa, nil
}
