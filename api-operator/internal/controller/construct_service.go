package controller

import (
	"fmt"
	"strings"

	appsv1alpha1 "github.com/dkr290/simple-operator/api-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *SimpleapiReconciler) constructService(
	SimpleAPIApp appsv1alpha1.Simpleapi, timestamp int64,
) *corev1.Service {
	metadata := metav1.ObjectMeta{
		Name:      serviceName(SimpleAPIApp.Spec.Version, SimpleAPIApp.Name),
		Namespace: SimpleAPIApp.Namespace,
		Labels: map[string]string{
			"app":     SimpleAPIApp.Labels["app"],
			"version": SimpleAPIApp.Spec.Version,
		},
		Annotations: map[string]string{
			"lastDeployedAt": fmt.Sprintf("%d", timestamp),
		},
	}
	spec := corev1.ServiceSpec{
		Selector: map[string]string{
			"app":     SimpleAPIApp.Labels["app"],
			"version": SimpleAPIApp.Spec.Version,
		},
		Ports: []corev1.ServicePort{
			{
				Port:       SimpleAPIApp.Spec.Port,
				TargetPort: intstr.FromInt(int(SimpleAPIApp.Spec.Port)),
				Protocol:   corev1.ProtocolTCP,
			},
		},
	}

	svc := &corev1.Service{
		ObjectMeta: metadata,
		Spec:       spec,
	}

	return svc
}

// returns a standardized Service name based on the version, important for replacor in utils is th make version at the end.
// TODO maybe for improvement
func serviceName(version string, serviceName string) string {
	return fmt.Sprintf(serviceName+"-%s-svc", strings.ToLower(version))
}
