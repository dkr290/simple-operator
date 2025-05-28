package controller

import (
	"fmt"
	"strings"

	appsv1alpha1 "github.com/dkr290/simple-operator/api-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *SimpleapiReconciler) constructDeployment(
	SimpleApiApp appsv1alpha1.Simpleapi, timestamp int64,
) *appsv1.Deployment {
	labels := map[string]string{
		"app":     appLabel,
		"version": SimpleApiApp.Spec.Version,
	}

	replicas := int32(1)
	if SimpleApiApp.Spec.Replicas != nil {
		replicas = *SimpleApiApp.Spec.Replicas
	}

	objectMetaData := metav1.ObjectMeta{
		Name:      "my-api-" + strings.ToLower(SimpleApiApp.Spec.Version),
		Namespace: SimpleApiApp.Namespace,
		Labels:    labels,
		Annotations: map[string]string{
			"lastDeployedAt": fmt.Sprintf("%d", timestamp),
		},
	}

	specData := appsv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "api",
						Image: SimpleApiApp.Spec.Image,
						Ports: []corev1.ContainerPort{
							{ContainerPort: SimpleApiApp.Spec.Port},
						},
					},
				},
				ImagePullSecrets: []corev1.LocalObjectReference{
					{
						Name: imagePullsecret,
					},
				},
			},
		},
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: objectMetaData,
		Spec:       specData,
	}

	return deploy
}
