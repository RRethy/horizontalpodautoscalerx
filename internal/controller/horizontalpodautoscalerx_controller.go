package controller

import (
	"context"
	"fmt"
	"slices"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"
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
// +kubebuilder:rbac:groups=autoscalingx.rrethy.io,resources=hpaoverrides,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=autoscalingx.rrethy.io,resources=hpaoverrides/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
// This Reconcile method uses the ObjectReconciler interface from https://github.com/kubernetes-sigs/controller-runtime/pull/2592.
func (r *HorizontalPodAutoscalerXReconciler) Reconcile(ctx context.Context, hpax *autoscalingxv1.HorizontalPodAutoscalerX) (res ctrl.Result, retErr error) {
	if !hpax.DeletionTimestamp.IsZero() {
		// The object is being deleted, don't do anything.
		return ctrl.Result{}, nil
	}

	log := log.FromContext(ctx)
	orig := hpax.DeepCopy()
	defer func() {
		// we don't even need to do this really, we're always updating the status
		if !apiequality.Semantic.DeepEqual(orig, hpax) {
			if err := r.Status().Update(ctx, hpax); err != nil {
				log.Error(err, "updating status")
			}
		}
	}()

	hpa, err := r.getHPA(ctx, hpax)
	if err != nil {
		log.Error(err, "getting HPA")
		return ctrl.Result{}, fmt.Errorf("getting HPA: %w", err)
	}

	err = r.updateHpaMinReplicas(ctx, hpax, hpa)
	if err != nil {
		log.Error(err, "updating HPA spec.minReplicas")
		return ctrl.Result{}, fmt.Errorf("updating HPA spec.minReplicas: %w", err)
	}

	hpax.Status.ObservedGeneration = ptr.To(hpax.Generation)
	r.setCondition(hpax, autoscalingxv1.ConditionReady, corev1.ConditionTrue, "HPAUpdated", "updated the minReplicas of the hpa")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorizontalPodAutoscalerXReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Clock == nil {
		r.Clock = clock.RealClock{}
	}

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

	err = mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&autoscalingxv1.HPAOverride{},
		"spec.hpaTargetName",
		func(obj client.Object) []string {
			return []string{obj.(*autoscalingxv1.HPAOverride).Spec.HPATargetName}
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
		Watches(
			&autoscalingxv1.HPAOverride{},
			handler.EnqueueRequestsFromMapFunc(r.findHPAXForHPAOverride),
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		Complete(reconcile.AsReconciler(mgr.GetClient(), r))
}

// setCondition sets the condition of the HorizontalPodAutoscalerX.
func (r *HorizontalPodAutoscalerXReconciler) setCondition(
	hpax *autoscalingxv1.HorizontalPodAutoscalerX,
	conditionType autoscalingxv1.HorizontalPodAutoscalerXConditionType,
	status corev1.ConditionStatus,
	reason string,
	message string,
) {
	for i, cond := range hpax.Status.Conditions {
		if cond.Type != conditionType {
			continue
		}

		lastTransitionTime := metav1.Time{Time: r.Clock.Now()}
		if cond.Status == status {
			lastTransitionTime = cond.LastTransitionTime
		}

		hpax.Status.Conditions[i] = autoscalingxv1.HorizontalPodAutoscalerXCondition{
			Type:               conditionType,
			Status:             status,
			LastTransitionTime: lastTransitionTime,
			Reason:             reason,
			Message:            message,
		}
		return
	}

	hpax.Status.Conditions = append(hpax.Status.Conditions, autoscalingxv1.HorizontalPodAutoscalerXCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Time{Time: r.Clock.Now()},
		Reason:             reason,
		Message:            message,
	})
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

	requests := make([]reconcile.Request, 0, len(hpaxList.Items))
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

// findHPAXForHPAOverride finds all HorizontalPodAutoscalerX objects that target the given HPAOverride.
func (r *HorizontalPodAutoscalerXReconciler) findHPAXForHPAOverride(ctx context.Context, o client.Object) []reconcile.Request {
	hpaOverride, ok := o.(*autoscalingxv1.HPAOverride)
	if !ok {
		return nil
	}

	hpaxList := &autoscalingxv1.HorizontalPodAutoscalerXList{}
	if err := r.List(ctx, hpaxList, &client.ListOptions{
		Namespace:     hpaOverride.GetNamespace(),
		FieldSelector: fields.OneTermEqualSelector("spec.hpaTargetName", hpaOverride.Spec.HPATargetName),
	}); err != nil {
		return nil
	}
	requests := make([]reconcile.Request, 0, len(hpaxList.Items))
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
		r.setCondition(hpax, autoscalingxv1.ConditionReady, corev1.ConditionFalse, "FailedToGetHPA", "failed getting the target hpa")
		return nil, err
	}

	if hpa.Status.ObservedGeneration != nil {
		hpax.Status.HPAObservedGeneration = ptr.To(*hpa.Status.ObservedGeneration)
	}

	return hpa, nil
}

// getFallbackSuggestion calculates the desired minReplicas for the HorizontalPodAutoscalerX based on the ScalingActive condition for the hpa.
func (r *HorizontalPodAutoscalerXReconciler) getFallbackSuggestion(hpax *autoscalingxv1.HorizontalPodAutoscalerX, hpa *autoscalingv2.HorizontalPodAutoscaler) int32 {
	var cond *autoscalingv2.HorizontalPodAutoscalerCondition
	for _, condition := range hpa.Status.Conditions {
		if condition.Type == autoscalingv2.ScalingActive {
			cond = &condition
		}
	}

	if hpax.Spec.Fallback == nil ||
		cond == nil ||
		cond.Status == corev1.ConditionTrue ||
		cond.Status == corev1.ConditionUnknown {
		r.setCondition(hpax, autoscalingxv1.ConditionFallback, corev1.ConditionFalse, "ScalingActive", "scaling active condition is not false")
		return hpax.Spec.MinReplicas
	}

	if cond.LastTransitionTime.Time.Add(hpax.Spec.Fallback.Duration.Duration).After(r.Clock.Now()) {
		r.setCondition(hpax, autoscalingxv1.ConditionFallback, corev1.ConditionTrue, "ScalingRecentlyInactive", "scaling active condition is false for not long enough")
		return hpax.Spec.MinReplicas
	}

	r.setCondition(hpax, autoscalingxv1.ConditionFallback, corev1.ConditionFalse, "ScalingInactive", "scaling active condition is false for long enough")
	return hpax.Spec.Fallback.MinReplicas
}

func (r *HorizontalPodAutoscalerXReconciler) getOverrideSuggestion(ctx context.Context, hpax *autoscalingxv1.HorizontalPodAutoscalerX) int32 {
	hpaOverrideList := &autoscalingxv1.HPAOverrideList{}
	if err := r.List(ctx, hpaOverrideList, &client.ListOptions{
		Namespace:     hpax.Namespace,
		FieldSelector: fields.OneTermEqualSelector("spec.hpaTargetName", hpax.Spec.HPATargetName),
	}); err != nil {
		r.setCondition(hpax, autoscalingxv1.ConditionReady, corev1.ConditionFalse, "FailedToGetHPAOverride", "failed getting target hpa overrides")
		return hpax.Spec.MinReplicas
	}

	replicas := []int32{}
	now := r.Clock.Now()
	for _, hpaOverride := range hpaOverrideList.Items {
		if hpaOverride.Spec.Time.After(now) {
			continue
		}
		if hpaOverride.Spec.Time.Add(hpaOverride.Spec.Duration.Duration).Before(now) {
			continue
		}
		replicas = append(replicas, hpaOverride.Spec.MinReplicas)
	}

	if len(replicas) == 0 {
		r.setCondition(hpax, autoscalingxv1.ConditionOverrideActive, corev1.ConditionFalse, "NoActiveOverride", "no active override was found")
		return hpax.Spec.MinReplicas
	}

	r.setCondition(hpax, autoscalingxv1.ConditionOverrideActive, corev1.ConditionTrue, "OverrideActive", "an override that is active was found")
	return slices.Max(replicas)
}

func (r *HorizontalPodAutoscalerXReconciler) updateHpaMinReplicas(ctx context.Context, hpax *autoscalingxv1.HorizontalPodAutoscalerX, hpa *autoscalingv2.HorizontalPodAutoscaler) error {
	minReplicas := slices.Max([]int32{hpax.Spec.MinReplicas, r.getFallbackSuggestion(hpax, hpa), r.getOverrideSuggestion(ctx, hpax)})

	hpaCopy := hpa.DeepCopy()
	hpa.Spec.MinReplicas = &minReplicas
	err := r.Patch(ctx, hpa, client.StrategicMergeFrom(hpaCopy))
	if err != nil {
		r.setCondition(hpax, autoscalingxv1.ConditionReady, corev1.ConditionFalse, "FailedToUpdateHPA", "failed updating the target hpa spec.minReplicas")
	}
	return err
}
