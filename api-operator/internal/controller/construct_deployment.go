// Package controller main engine package
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
	SimpleAPIApp appsv1alpha1.Simpleapi, timestamp int64,
) *appsv1.Deployment {
	labels := map[string]string{
		"app":     SimpleAPIApp.Labels["app"],
		"version": SimpleAPIApp.Spec.Version,
	}

	replicas := int32(1)
	if SimpleAPIApp.Spec.Replicas != nil {
		replicas = *SimpleAPIApp.Spec.Replicas
	}

	objectMetaData := metav1.ObjectMeta{
		Name: deploymentName(
			strings.ToLower(SimpleAPIApp.Spec.Version),
			SimpleAPIApp.Name,
		),
		Namespace: SimpleAPIApp.Namespace,
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
						Name:  SimpleAPIApp.Name,
						Image: SimpleAPIApp.Spec.Image,
						Ports: []corev1.ContainerPort{
							{ContainerPort: SimpleAPIApp.Spec.Port},
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

func deploymentName(version string, deploymentName string) string {
	return fmt.Sprintf(deploymentName+"-%s", strings.ToLower(version))
}
