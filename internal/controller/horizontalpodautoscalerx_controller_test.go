package controller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
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
)

var _ = Describe("HorizontalPodAutoscalerX Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "myresource"
		const namespace = "default"
		const hpaName = "myhpa"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace,
		}
		horizontalpodautoscalerx := &autoscalingxv1.HorizontalPodAutoscalerX{}

		BeforeEach(func() {
			By("creating the associated HPA")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      hpaName,
					Namespace: namespace,
				},
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
			}
			Expect(k8sClient.Create(ctx, hpa)).To(Succeed())

			By("creating the custom resource for the Kind HorizontalPodAutoscalerX")
			err := k8sClient.Get(ctx, typeNamespacedName, horizontalpodautoscalerx)
			if err != nil && errors.IsNotFound(err) {
				resource := &autoscalingxv1.HorizontalPodAutoscalerX{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace,
					},
					Spec: autoscalingxv1.HorizontalPodAutoscalerXSpec{
						HPATargetName: hpaName,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &autoscalingxv1.HorizontalPodAutoscalerX{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance HorizontalPodAutoscalerX")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: hpaName, Namespace: namespace}, hpa)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance HorizontalPodAutoscaler")
			Expect(k8sClient.Delete(ctx, hpa)).To(Succeed())
		})

		It("should store scaling active condition if true", func() {
			By("getting the hpa")
			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: hpaName, Namespace: namespace}, hpa)
			Expect(err).NotTo(HaveOccurred())

			By("updating the hpa status to have scaling active condition as true")
			hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:   autoscalingv2.ScalingActive,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   autoscalingv2.AbleToScale,
					Status: corev1.ConditionFalse,
				},
			}
			err = k8sClient.Status().Update(ctx, hpa)
			Expect(err).NotTo(HaveOccurred())

			By("getting the hpax to check the status eventually has been updated")
			Eventually(func() error {
				err := k8sClient.Get(ctx, typeNamespacedName, horizontalpodautoscalerx)
				if err != nil {
					return err
				}
				if horizontalpodautoscalerx.Status.ScalingActiveCondition != corev1.ConditionTrue {
					return fmt.Errorf("expected scaling active condition to be true")
				}
				conditionSince := horizontalpodautoscalerx.Status.ScalingActiveConditionSince
				if conditionSince.Equal(&metav1.Time{Time: fakeclock.Now()}) {
				}
				return nil
			}).Should(BeNil())
		})
	})
})
