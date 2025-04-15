package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	hpaNamespacedName = types.NamespacedName{Name: hpaName, Namespace: namespace}
	defaultHpax       = &autoscalingxv1.HorizontalPodAutoscalerX{
		ObjectMeta: metav1.ObjectMeta{Name: hpaxName, Namespace: namespace},
		Spec: autoscalingxv1.HorizontalPodAutoscalerXSpec{
			HPATargetName: hpaName,
			Fallback:      &autoscalingxv1.Fallback{MinReplicas: fallbackMinReplicas, Duration: metav1.Duration{Duration: fallbackDuration}},
			MinReplicas:   minReplicas,
		},
	}
	defaultHpa = &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: hpaName, Namespace: namespace},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MaxReplicas: fallbackMinReplicas + 100,
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
			hpax := &autoscalingxv1.HorizontalPodAutoscalerX{ObjectMeta: metav1.ObjectMeta{Name: hpaxName, Namespace: namespace}}
			Expect(k8sClient.Delete(ctx, hpax)).To(Succeed())

			By("deleting the associated HPA")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: hpaName, Namespace: namespace}}
			Expect(k8sClient.Delete(ctx, hpa)).To(Succeed())

			By("deleting any HPAOverride")
			hpaOverrideList := &autoscalingxv1.HPAOverrideList{}
			Expect(k8sClient.List(ctx, hpaOverrideList)).To(Succeed())
			for _, hpaOverride := range hpaOverrideList.Items {
				Expect(k8sClient.Delete(ctx, &hpaOverride)).To(Succeed())
			}
		})

		It("should not update minReplicas if scaling active condition is true for short time", func() {
			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

			By("updating the hpa status to have scaling active condition as true more recently than fallback duration")
			origHpa := hpa.DeepCopy()
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
			Expect(k8sClient.Status().Patch(ctx, hpa, client.MergeFrom(origHpa))).Should(Succeed())

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
			origHpa := hpa.DeepCopy()
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
			Expect(k8sClient.Status().Patch(ctx, hpa, client.MergeFrom(origHpa))).Should(Succeed())

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
			origHpa := hpa.DeepCopy()
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
			Expect(k8sClient.Status().Patch(ctx, hpa, client.MergeFrom(origHpa))).Should(Succeed())

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
			origHpa := hpa.DeepCopy()
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
			Expect(k8sClient.Status().Patch(ctx, hpa, client.MergeFrom(origHpa))).Should(Succeed())

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
			Expect(k8sClient.Update(ctx, hpa)).Should(Succeed())

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
			origHpa := hpa.DeepCopy()
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
			Expect(k8sClient.Status().Patch(ctx, hpa, client.MergeFrom(origHpa))).Should(Succeed())

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
			origHpa := hpa.DeepCopy()
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
			Expect(k8sClient.Status().Patch(ctx, hpa, client.MergeFrom(origHpa))).Should(Succeed())

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
			origHpa := hpa.DeepCopy()
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
			Expect(k8sClient.Status().Patch(ctx, hpa, client.MergeFrom(origHpa))).Should(Succeed())

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

		It("should not update minReplicas if override is found but too early", func() {
			By("creating an override that is too early")
			hpaOverride := &autoscalingxv1.HPAOverride{
				ObjectMeta: metav1.ObjectMeta{Name: "some-override", Namespace: namespace},
				Spec: autoscalingxv1.HPAOverrideSpec{
					MinReplicas:   fallbackMinReplicas + 10,
					Duration:      metav1.Duration{Duration: 1 * time.Hour},
					Time:          metav1.Time{Time: fakeclock.Now().Add(-2 * time.Hour)},
					HPATargetName: hpaName,
				},
			}
			Expect(k8sClient.Create(ctx, hpaOverride)).To(Succeed())

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

		It("should not update minReplicas if override is found but late", func() {
			By("creating an override that is late")
			hpaOverride := &autoscalingxv1.HPAOverride{
				ObjectMeta: metav1.ObjectMeta{Name: "some-override", Namespace: namespace},
				Spec: autoscalingxv1.HPAOverrideSpec{
					MinReplicas:   fallbackMinReplicas + 10,
					Duration:      metav1.Duration{Duration: 1 * time.Hour},
					Time:          metav1.Time{Time: fakeclock.Now().Add(2 * time.Hour)},
					HPATargetName: hpaName,
				},
			}
			Expect(k8sClient.Create(ctx, hpaOverride)).To(Succeed())

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

		It("should update minReplicas if override is found but fallback is not active", func() {
			By("creating an override that is active")
			hpaOverride := &autoscalingxv1.HPAOverride{
				ObjectMeta: metav1.ObjectMeta{Name: "some-override", Namespace: namespace},
				Spec: autoscalingxv1.HPAOverrideSpec{
					MinReplicas:   fallbackMinReplicas + 10,
					Duration:      metav1.Duration{Duration: 2 * time.Hour},
					Time:          metav1.Time{Time: fakeclock.Now().Add(-1 * time.Hour)},
					HPATargetName: hpaName,
				},
			}
			Expect(k8sClient.Create(ctx, hpaOverride)).To(Succeed())

			By("getting the hpa to check if minReplicas is updated")
			Eventually(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())
				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, eventuallyTimeout, interval).Should(Equal(fallbackMinReplicas + 10))
		})

		It("should update minReplicas to max of multiple overrides", func() {
			By("creating an override that is active")
			hpaOverride1 := &autoscalingxv1.HPAOverride{
				ObjectMeta: metav1.ObjectMeta{Name: "some-override-1", Namespace: namespace},
				Spec: autoscalingxv1.HPAOverrideSpec{
					MinReplicas:   fallbackMinReplicas + 10,
					Duration:      metav1.Duration{Duration: 2 * time.Hour},
					Time:          metav1.Time{Time: fakeclock.Now().Add(-1 * time.Hour)},
					HPATargetName: hpaName,
				},
			}
			Expect(k8sClient.Create(ctx, hpaOverride1)).To(Succeed())

			By("creating another override that is active")
			hpaOverride2 := &autoscalingxv1.HPAOverride{
				ObjectMeta: metav1.ObjectMeta{Name: "some-override-2", Namespace: namespace},
				Spec: autoscalingxv1.HPAOverrideSpec{
					MinReplicas:   fallbackMinReplicas + 20,
					Duration:      metav1.Duration{Duration: 2 * time.Hour},
					Time:          metav1.Time{Time: fakeclock.Now().Add(-1 * time.Hour)},
					HPATargetName: hpaName,
				},
			}
			Expect(k8sClient.Create(ctx, hpaOverride2)).To(Succeed())

			By("getting the hpa to check if minReplicas is updated")
			Eventually(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())
				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, eventuallyTimeout, interval).Should(Equal(fallbackMinReplicas + 20))
		})

		It("should update minReplicas to override if override is found and fallback is active if override is bigger", func() {
			By("creating an override that is active")
			hpaOverride := &autoscalingxv1.HPAOverride{
				ObjectMeta: metav1.ObjectMeta{Name: "some-override", Namespace: namespace},
				Spec: autoscalingxv1.HPAOverrideSpec{
					MinReplicas:   fallbackMinReplicas + 10,
					Duration:      metav1.Duration{Duration: 2 * time.Hour},
					Time:          metav1.Time{Time: fakeclock.Now().Add(-1 * time.Hour)},
					HPATargetName: hpaName,
				},
			}
			Expect(k8sClient.Create(ctx, hpaOverride)).To(Succeed())

			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

			By("updating the hpa status to have scaling active condition as false")
			origHpa := hpa.DeepCopy()
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:               autoscalingv2.ScalingActive,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration)},
				},
			}
			Expect(k8sClient.Status().Patch(ctx, hpa, client.MergeFrom(origHpa))).Should(Succeed())

			By("getting the hpa to check if minReplicas is updated")
			Eventually(func() int32 {
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())
				if hpa.Spec.MinReplicas != nil {
					return *hpa.Spec.MinReplicas
				}
				return -1
			}, eventuallyTimeout, interval).Should(Equal(fallbackMinReplicas + 10))
		})

		It("should update minReplicas to fallback if override is found and fallback is active if fallback is bigger", func() {
			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, hpaNamespacedName, hpa)).To(Succeed())

			By("updating the hpa status to have scaling active condition as false")
			origHpa := hpa.DeepCopy()
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:               autoscalingv2.ScalingActive,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: fakeclock.Now().Add(-fallbackDuration)},
				},
			}
			Expect(k8sClient.Status().Patch(ctx, hpa, client.MergeFrom(origHpa))).Should(Succeed())

			By("creating an override that is active")
			hpaOverride := &autoscalingxv1.HPAOverride{
				ObjectMeta: metav1.ObjectMeta{Name: "some-override", Namespace: namespace},
				Spec: autoscalingxv1.HPAOverrideSpec{
					MinReplicas:   fallbackMinReplicas - 1,
					Duration:      metav1.Duration{Duration: 2 * time.Hour},
					Time:          metav1.Time{Time: fakeclock.Now().Add(-1 * time.Hour)},
					HPATargetName: hpaName,
				},
			}
			Expect(k8sClient.Create(ctx, hpaOverride)).To(Succeed())

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

		It("should not update minReplicas if an override is active but for a different hpa", func() {
			By("creating an override that is active")
			hpaOverride := &autoscalingxv1.HPAOverride{
				ObjectMeta: metav1.ObjectMeta{Name: "some-override", Namespace: namespace},
				Spec: autoscalingxv1.HPAOverrideSpec{
					MinReplicas:   fallbackMinReplicas + 10,
					Duration:      metav1.Duration{Duration: 2 * time.Hour},
					Time:          metav1.Time{Time: fakeclock.Now().Add(-1 * time.Hour)},
					HPATargetName: "some-other-hpa",
				},
			}
			Expect(k8sClient.Create(ctx, hpaOverride)).To(Succeed())

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
	})
})
