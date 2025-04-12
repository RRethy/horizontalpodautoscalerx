package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	autoscalingxv1 "rrethy.io/horizontalpodautoscalerx/api/v1"
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

		It("should successfully reconcile the resource", func() {
			hpax := &autoscalingxv1.HorizontalPodAutoscalerX{}
			err := k8sClient.Get(ctx, typeNamespacedName, hpax)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
