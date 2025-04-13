package controller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	autoscalingxv1 "rrethy.io/horizontalpodautoscalerx/api/v1"
)

const (
	eventuallyTimeout = 2 * time.Second
	interval          = 250 * time.Millisecond
	namespace         = "default"
	hpaxName          = "myhpax"
	hpaName           = "myhpa"
)

var (
	hpaxNamespacedName = types.NamespacedName{Name: hpaxName, Namespace: namespace}
	hpaNamespacedName  = types.NamespacedName{Name: hpaName, Namespace: namespace}
	defaultHpax        = &autoscalingxv1.HorizontalPodAutoscalerX{
		ObjectMeta: metav1.ObjectMeta{Name: hpaxName, Namespace: namespace},
		Spec:       autoscalingxv1.HorizontalPodAutoscalerXSpec{HPATargetName: hpaName},
		Status:     autoscalingxv1.HorizontalPodAutoscalerXStatus{},
	}
	defaultHpa = &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: hpaName, Namespace: namespace},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MinReplicas: ptr.To(int32(1)),
			MaxReplicas: 10,
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "mydeployment",
				APIVersion: "apps/v1",
			},
			Metrics: []autoscalingv2.MetricSpec{},
		},
		Status: autoscalingv2.HorizontalPodAutoscalerStatus{},
	}
)

var _ = Describe("HorizontalPodAutoscalerX Controller", func() {
	Context("When reconciling a resource", func() {
		ctx := context.Background()

		BeforeEach(func() {
			By("creating the associated HPA")
			Expect(k8sClient.Create(ctx, defaultHpa.DeepCopy())).To(Succeed())

			By("creating the custom resource for the Kind HorizontalPodAutoscalerX")
			Expect(k8sClient.Create(ctx, defaultHpax.DeepCopy())).To(Succeed())
		})

		AfterEach(func() {
			By("deleting the HorizontalPodAutoscalerX")
			Expect(k8sClient.Delete(ctx, defaultHpax)).To(Succeed())

			By("deleting the associated HPA")
			Expect(k8sClient.Delete(ctx, defaultHpa)).To(Succeed())
		})

		It("should store scaling active condition if false", func() {
			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

			By("updating the hpa status to have scaling active condition as true")
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{Type: autoscalingv2.ScalingActive, Status: corev1.ConditionTrue},
				{Type: autoscalingv2.AbleToScale, Status: corev1.ConditionFalse},
			}
			Expect(k8sClient.Status().Update(ctx, hpa)).To(Succeed())

			By("getting the hpax to check the status eventually has been updated")
			Eventually(func() error {
				hpax := &autoscalingxv1.HorizontalPodAutoscalerX{}
				err := k8sClient.Get(ctx, hpaxNamespacedName, hpax)
				if err != nil {
					return err
				}
				if hpax.Status.ScalingActiveCondition != corev1.ConditionTrue {
					return fmt.Errorf("expected scaling active condition to be true but got %s", hpax.Status.ScalingActiveCondition)
				}
				conditionSince := hpax.Status.ScalingActiveConditionSince
				if !conditionSince.Equal(&metav1.Time{Time: fakeclock.Now()}) {
					return fmt.Errorf("expected condition since to be %s but got %s", fakeclock.Now(), conditionSince)
				}
				return nil
			}, eventuallyTimeout, interval).Should(BeNil())
		})

		It("should store scaling active condition if true", func() {
			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

			By("updating the hpa status to have scaling active condition as true")
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{Type: autoscalingv2.ScalingActive, Status: corev1.ConditionTrue},
				{Type: autoscalingv2.AbleToScale, Status: corev1.ConditionFalse},
			}
			Expect(k8sClient.Status().Update(ctx, hpa)).To(Succeed())

			By("getting the hpax to check the status eventually has been updated")
			Eventually(func() error {
				hpax := &autoscalingxv1.HorizontalPodAutoscalerX{}
				err := k8sClient.Get(ctx, hpaxNamespacedName, hpax)
				if err != nil {
					return err
				}
				if hpax.Status.ScalingActiveCondition != corev1.ConditionTrue {
					return fmt.Errorf("expected scaling active condition to be true but got %s", hpax.Status.ScalingActiveCondition)
				}
				conditionSince := hpax.Status.ScalingActiveConditionSince
				if !conditionSince.Equal(&metav1.Time{Time: fakeclock.Now()}) {
					return fmt.Errorf("expected condition since to be %s but got %s", fakeclock.Now(), conditionSince)
				}
				return nil
			}, eventuallyTimeout, interval).Should(BeNil())
		})
	})
})
