package controller

import (
	"context"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	autoscalingxv1 "rrethy.io/horizontalpodautoscalerx/api/v1"
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

	curCondition := r.getScalingActiveCondition(hpa)
	since := r.getConditionSince(hpax, curCondition)
	hpax.Status.ScalingActiveCondition = curCondition
	hpax.Status.ScalingActiveConditionSince = since

	// if curCondition == corev1.ConditionFalse && r.Clock.Now().Sub(since.Time) > hpax.Spec.Fallback.Duration.Duration {
	// 	if hpax.Spec.Fallback != nil {
	// 		hpa.Spec.MinReplicas = &hpax.Spec.Fallback.Replicas
	// 		err := r.Update(ctx, hpa)
	// 		if err != nil {
	// 			log.Error(err, "updating HPA")
	// 			return ctrl.Result{}, err
	// 		}
	// 	}
	// }

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorizontalPodAutoscalerXReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(ControllerName).
		For(&autoscalingxv1.HorizontalPodAutoscalerX{}).
		// TODO: setup watches for hpa and figure out how to watch updates to hpa status
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{})).
		Complete(reconcile.AsReconciler(mgr.GetClient(), r))
}

func (r *HorizontalPodAutoscalerXReconciler) getHPA(ctx context.Context, hpax *autoscalingxv1.HorizontalPodAutoscalerX) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, client.ObjectKey{Name: hpax.Spec.HPATargetName, Namespace: hpax.Namespace}, hpa)
	if err != nil {
		return nil, err
	}
	return hpa, nil
}

func (r *HorizontalPodAutoscalerXReconciler) getScalingActiveCondition(hpa *autoscalingv2.HorizontalPodAutoscaler) corev1.ConditionStatus {
	for _, condition := range hpa.Status.Conditions {
		if condition.Type == autoscalingv2.ScalingActive {
			return condition.Status
		}
	}
	return corev1.ConditionUnknown
}

func (r *HorizontalPodAutoscalerXReconciler) getConditionSince(hpax *autoscalingxv1.HorizontalPodAutoscalerX, curCondition corev1.ConditionStatus) *metav1.Time {
	if hpax.Status.ScalingActiveCondition == curCondition {
		return hpax.Status.ScalingActiveConditionSince
	}
	return &metav1.Time{Time: r.Clock.Now()}
}
