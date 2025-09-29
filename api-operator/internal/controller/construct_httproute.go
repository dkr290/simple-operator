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
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func (r *SimpleapiReconciler) reconcileHTTPRoute(
	ctx context.Context,
	versions []string, namespace string,
	SimpleAPIApp *appsv1alpha1.Simpleapi,
) error {
	httproute := &gatewayv1.HTTPRoute{}
	// check if httproute already exists
	err := r.Get(
		ctx,
		client.ObjectKey{Namespace: namespace, Name: getHTTPRouteName(SimpleAPIApp)},
		httproute,
	)

	if errors.IsNotFound(err) {
		// ingress does not exists and creating new one

		httproute = r.constructHTTPRoute(versions, namespace, SimpleAPIApp)
		if err := controllerutil.SetControllerReference(SimpleAPIApp, httproute, r.Scheme); err != nil {
			return err
		}
		return r.Create(ctx, httproute)

	} else if err != nil {
		return err
	}
	newHttproute := r.constructHTTPRoute(versions, namespace, SimpleAPIApp)
	httproute.Spec = newHttproute.Spec
	// this is kind of ensure ownership because the ingress not gets deleted but all the others does
	if err := controllerutil.SetControllerReference(SimpleAPIApp, httproute, r.Scheme); err != nil {
		return err
	}

	return r.Update(ctx, httproute)
}

func (r *SimpleapiReconciler) constructHTTPRoute(
	versions []string,
	namespace string,
	SimpleAPIApp *appsv1alpha1.Simpleapi,
) *gatewayv1.HTTPRoute {
	rules := make([]gatewayv1.HTTPRouteRule, len(versions))
	for i, ver := range versions {
		rules[i] = gatewayv1.HTTPRouteRule{
			Matches: []gatewayv1.HTTPRouteMatch{
				{
					Path: &gatewayv1.HTTPPathMatch{
						Type:  ptr.To(gatewayv1.PathMatchPathPrefix),
						Value: ptr.To("/api/" + ver),
					},
				},
			},
			BackendRefs: []gatewayv1.HTTPBackendRef{
				{
					BackendRef: gatewayv1.BackendRef{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name: gatewayv1.ObjectName(serviceName(ver, SimpleAPIApp.Name)),
							Port: ptr.To(gatewayv1.PortNumber(SimpleAPIApp.Spec.Port)),
						},
						Weight: ptr.To[int32](1),
					},
				},
			},
		}
	}
	var httproute *gatewayv1.HTTPRoute
	if SimpleAPIApp.Spec.IngressHostName == "" {
		httproute = &gatewayv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getHTTPRouteName(SimpleAPIApp),
				Namespace: namespace,
			},
			Spec: gatewayv1.HTTPRouteSpec{
				CommonRouteSpec: gatewayv1.CommonRouteSpec{
					ParentRefs: []gatewayv1.ParentReference{
						{
							Name: gatewayv1.ObjectName(SimpleAPIApp.Spec.EnvoyGateway),
							Namespace: ptr.To(
								gatewayv1.Namespace(SimpleAPIApp.Spec.EnvoyGatewayNamespace),
							),
						},
					},
				},
				Rules: rules,
			},
		}
	} else {
		httproute = &gatewayv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getHTTPRouteName(SimpleAPIApp),
				Namespace: namespace,
			},
			Spec: gatewayv1.HTTPRouteSpec{
				Hostnames: []gatewayv1.Hostname{
					gatewayv1.Hostname(SimpleAPIApp.Spec.IngressHostName),
				},
				CommonRouteSpec: gatewayv1.CommonRouteSpec{
					ParentRefs: []gatewayv1.ParentReference{
						{
							Name: gatewayv1.ObjectName(SimpleAPIApp.Spec.EnvoyGateway),
							Namespace: ptr.To(
								gatewayv1.Namespace(SimpleAPIApp.Spec.EnvoyGatewayNamespace),
							),
						},
					},
				},
				Rules: rules,
			},
		}
	}
	return httproute
}

func getHTTPRouteName(SimpleAPIApp *appsv1alpha1.Simpleapi) string {
	return fmt.Sprintf("%s-httproute", SimpleAPIApp.Name)
}
