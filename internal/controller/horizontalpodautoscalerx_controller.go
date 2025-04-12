package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

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
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HorizontalPodAutoscalerX object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *HorizontalPodAutoscalerXReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorizontalPodAutoscalerXReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalingxv1.HorizontalPodAutoscalerX{}).
		Named("horizontalpodautoscalerx").
		Complete(r)
}
