package controller

import (
	"context"
	"fmt"

	appsv1alpha1 "github.com/dkr290/simple-operator/api-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	networkingv1 "k8s.io/api/networking/v1"
	// for gateway api networkingv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	ingressClassName string = "nginx"
)

func (r *SimpleapiReconciler) reconcileIngress(
	ctx context.Context,
	versions []string,
	namespace string,
	SimpleAPIApp *appsv1alpha1.Simpleapi,
) error {
	ingress := &networkingv1.Ingress{}

	// Check if the Ingress already exists
	err := r.Get(
		ctx,
		client.ObjectKey{Namespace: namespace, Name: getIngressName(SimpleAPIApp)},
		ingress,
	)
	// Check if the Ingress already existso
	if errors.IsNotFound(err) {
		// ingress does not exists and creating new one
		ingress = r.constructIngress(versions, namespace, SimpleAPIApp)
		if err := controllerutil.SetControllerReference(SimpleAPIApp, ingress, r.Scheme); err != nil {
			return err
		}
		return r.Create(ctx, ingress)
	} else if err != nil {
		return err
	}
	newIngress := r.constructIngress(versions, namespace, SimpleAPIApp)
	ingress.Spec = newIngress.Spec
	// this is kind of ensure ownership because the ingress not gets deleted but all the others does
	if err := controllerutil.SetControllerReference(SimpleAPIApp, ingress, r.Scheme); err != nil {
		return err
	}
	return r.Update(ctx, ingress)
}

func (r *SimpleapiReconciler) constructIngress(
	versions []string,
	namespace string,
	SimpleAPIApp *appsv1alpha1.Simpleapi,
) *networkingv1.Ingress {
	paths := make([]networkingv1.HTTPIngressPath, len(versions))

	for i, ver := range versions {

		path := networkingv1.HTTPIngressPath{
			Path: "/api/" + ver,
			PathType: func() *networkingv1.PathType {
				pt := networkingv1.PathTypeImplementationSpecific
				return &pt
			}(),
			Backend: networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: serviceName(ver, SimpleAPIApp.Name),
					Port: networkingv1.ServiceBackendPort{
						Number: SimpleAPIApp.Spec.Port,
					},
				},
			},
		}
		paths[i] = path
	}
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        getIngressName(SimpleAPIApp),
			Namespace:   namespace,
			Annotations: map[string]string{
				//	"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: ptr.To(ingressClassName),
			Rules: []networkingv1.IngressRule{
				{
					// Host: "my-api.bankingcircle.net",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: paths,
						},
					},
				},
			},
			// TLS: []networkingv1.IngressTLS{
			// 	{
			// 		Hosts:      []string{"my-api.bankingcircle.net"},
			// 		SecretName: "my-api-tls-secret",
			// 	},
			// },
		},
	}
	return ingress
}

func getIngressName(SimpleAPIApp *appsv1alpha1.Simpleapi) string {
	return fmt.Sprintf("%s-ingress", SimpleAPIApp.Name)
}
