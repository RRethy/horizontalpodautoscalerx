package controller

import (
	"context"
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
	eventuallyTimeout   = 2 * time.Second
	consistentlyTimeout = 4 * time.Second
	interval            = 250 * time.Millisecond
	namespace           = "default"
	hpaxName            = "myhpax"
	hpaName             = "myhpa"
	fallbackDuration    = 5 * time.Second
	minReplicas         = int32(1)
	fallbackMinReplicas = int32(10)
)

var (
	hpaxNamespacedName = types.NamespacedName{Name: hpaxName, Namespace: namespace}
	hpaNamespacedName  = types.NamespacedName{Name: hpaName, Namespace: namespace}
	defaultHpax        = &autoscalingxv1.HorizontalPodAutoscalerX{
		ObjectMeta: metav1.ObjectMeta{Name: hpaxName, Namespace: namespace},
		Spec: autoscalingxv1.HorizontalPodAutoscalerXSpec{
			HPATargetName: hpaName,
			Fallback: &autoscalingxv1.Fallback{
				MinReplicas: fallbackMinReplicas,
				Duration:    metav1.Duration{Duration: fallbackDuration},
			},
			MinReplicas: minReplicas,
		},
	}
	defaultHpa = &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: hpaName, Namespace: namespace},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MaxReplicas: fallbackMinReplicas + 1,
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "mydeployment",
				APIVersion: "apps/v1",
			},
			Metrics: []autoscalingv2.MetricSpec{},
		},
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

			By("waiting for the HPA to get minReplicas updated")
			Eventually(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())
				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, eventuallyTimeout, interval).Should(Equal(minReplicas))
		})

		AfterEach(func() {
			By("deleting the HorizontalPodAutoscalerX")
			Expect(k8sClient.Delete(ctx, defaultHpax)).To(Succeed())

			By("deleting the associated HPA")
			Expect(k8sClient.Delete(ctx, defaultHpa)).To(Succeed())
		})

		It("should not update minReplicas if scaling active condition is true for short time", func() {
			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

			By("updating the hpa status to have scaling active condition as true more recently than fallback duration")
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:               autoscalingv2.ScalingActive,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(1 * time.Second)},
				},
				{
					Type:               autoscalingv2.AbleToScale,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(-1 * time.Second)},
				},
			}
			Expect(k8sClient.Status().Update(ctx, hpa)).To(Succeed())

			By("getting the hpa to check if minReplicas is not updated")
			Consistently(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, consistentlyTimeout, interval).Should(Equal(minReplicas))
		})

		It("should not update minReplicas if scaling active condition is true for long time", func() {
			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

			By("updating the hpa status to have scaling active condition as true more recently than fallback duration")
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:               autoscalingv2.ScalingActive,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(-1 * time.Second)},
				},
				{
					Type:               autoscalingv2.AbleToScale,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(-1 * time.Second)},
				},
			}
			Expect(k8sClient.Status().Update(ctx, hpa)).To(Succeed())

			By("getting the hpa to check if minReplicas is not updated")
			Consistently(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, consistentlyTimeout, interval).Should(Equal(minReplicas))
		})

		It("should not update minReplicas if scaling active condition is false for short time", func() {
			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

			By("updating the hpa status to have scaling active condition as false more recently than fallback duration")
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:               autoscalingv2.ScalingActive,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(1 * time.Second)},
				},
				{
					Type:               autoscalingv2.AbleToScale,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(-1 * time.Second)},
				},
			}
			Expect(k8sClient.Status().Update(ctx, hpa)).To(Succeed())

			By("getting the hpa to check if minReplicas is not updated")
			Consistently(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, consistentlyTimeout, interval).Should(Equal(minReplicas))
		})

		It("should not update minReplicas if scaling active condition is unknown for long time", func() {
			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

			By("updating the hpa status to have scaling active condition as unknown more recently than fallback duration")
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:               autoscalingv2.ScalingActive,
					Status:             corev1.ConditionUnknown,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(-1 * time.Second)},
				},
				{
					Type:               autoscalingv2.AbleToScale,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(-1 * time.Second)},
				},
			}
			Expect(k8sClient.Status().Update(ctx, hpa)).To(Succeed())

			By("getting the hpa to check if minReplicas is not updated")
			Consistently(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, consistentlyTimeout, interval).Should(Equal(minReplicas))
		})

		It("should scale down minReplicas if scaling active condition is true for short time but previously scaled up", func() {
			By("patching the minReplicas on the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())
			hpa.Spec.MinReplicas = ptr.To(fallbackMinReplicas)
			Expect(k8sClient.Update(ctx, hpa)).To(Succeed())

			By("getting the hpa to check if minReplicas is updated")
			Eventually(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())
				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, eventuallyTimeout, interval).Should(Equal(fallbackMinReplicas))

			By("updating the hpa status to have scaling active condition as true more recently than fallback duration")
			hpa = &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:               autoscalingv2.ScalingActive,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(1 * time.Second)},
				},
				{
					Type:               autoscalingv2.AbleToScale,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(-1 * time.Second)},
				},
			}
			Expect(k8sClient.Status().Update(ctx, hpa)).To(Succeed())

			By("getting the hpa to check if minReplicas is updated")
			Eventually(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())
				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, eventuallyTimeout, interval).Should(Equal(minReplicas))
		})

		It("should update minReplicas if scaling active condition is false for longer time than fallback duration", func() {
			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

			By("updating the hpa status to have scaling active condition as false more recently than fallback duration")
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:               autoscalingv2.ScalingActive,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(-1 * time.Second)},
				},
				{
					Type:               autoscalingv2.AbleToScale,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(1 * time.Second)},
				},
			}
			Expect(k8sClient.Status().Update(ctx, hpa)).To(Succeed())

			By("getting the hpa to check if minReplicas is updated")
			Eventually(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())
				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, eventuallyTimeout, interval).Should(Equal(fallbackMinReplicas))
		})

		It("should update minReplicas if scaling active condition is false for equal time than fallback duration", func() {
			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

			By("updating the hpa status to have scaling active condition as false more recently than fallback duration")
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:               autoscalingv2.ScalingActive,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration)},
				},
				{
					Type:               autoscalingv2.AbleToScale,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration).Add(1 * time.Second)},
				},
			}
			Expect(k8sClient.Status().Update(ctx, hpa)).To(Succeed())

			By("getting the hpa to check if minReplicas is updated")
			Eventually(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())
				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, eventuallyTimeout, interval).Should(Equal(fallbackMinReplicas))
		})
	})
})
