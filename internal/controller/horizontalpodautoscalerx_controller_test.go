package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &HorizontalPodAutoscalerXReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
