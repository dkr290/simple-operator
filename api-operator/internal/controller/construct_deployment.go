// Package controller main engine package
package controller

import (
	"fmt"
	"strings"

	appsv1alpha1 "github.com/dkr290/simple-operator/api-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func (r *SimpleapiReconciler) constructDeployment(
	SimpleAPIApp appsv1alpha1.Simpleapi, timestamp int64,
) *appsv1.Deployment {
	labels := map[string]string{
		"app":     SimpleAPIApp.Labels["app"],
		"version": SimpleAPIApp.Spec.Version,
	}

	var podSpec corev1.PodSpec
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

	serviceAccountName := SimpleAPIApp.Spec.ServiceAccountName
	if serviceAccountName == "" {
		serviceAccountName = "default"
	}

	imagePullPolicy := SimpleAPIApp.Spec.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = corev1.PullIfNotPresent
	}
	imagePullSecret := SimpleAPIApp.Spec.ImagePullSecret

	if imagePullSecret == "" {
		podSpec = GetPodSpec(SimpleAPIApp, serviceAccountName, false, imagePullPolicy)
	} else {
		podSpec = GetPodSpec(SimpleAPIApp, serviceAccountName, true, imagePullPolicy)
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
			Spec: podSpec,
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

func GetPodSpec(
	SimpleAPIApp appsv1alpha1.Simpleapi,
	serviceAccountName string,
	isImagePullSecret bool,
	ImagePullPolicy corev1.PullPolicy,
) corev1.PodSpec {
	if isImagePullSecret {
		return corev1.PodSpec{
			AutomountServiceAccountToken: ptr.To(false),
			ServiceAccountName:           serviceAccountName,
			SecurityContext:              SimpleAPIApp.Spec.PodSecurityContext,
			Containers: []corev1.Container{
				{
					Name: SimpleAPIApp.Name,
					Image: fmt.Sprintf(
						"%s:%s",
						SimpleAPIApp.Spec.Image,
						SimpleAPIApp.Spec.Version,
					),
					Ports: []corev1.ContainerPort{
						{ContainerPort: SimpleAPIApp.Spec.Port},
					},
					StartupProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: SimpleAPIApp.Spec.StartupProbe.HTTPGet.Path,
								Port: SimpleAPIApp.Spec.StartupProbe.HTTPGet.Port,
							},
						},
						InitialDelaySeconds: SimpleAPIApp.Spec.StartupProbe.InitialDelaySeconds,
						PeriodSeconds:       SimpleAPIApp.Spec.StartupProbe.PeriodSeconds,
						FailureThreshold:    SimpleAPIApp.Spec.StartupProbe.FailureThreshold,
					},
					ImagePullPolicy: ImagePullPolicy,
					Resources:       SimpleAPIApp.Spec.Resources,
				},
			},
			Affinity:    SimpleAPIApp.Spec.Affinity,
			Tolerations: SimpleAPIApp.Spec.Tolerations,

			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: SimpleAPIApp.Spec.ImagePullSecret,
				},
			},
		}
	} else {
		return corev1.PodSpec{
			AutomountServiceAccountToken: ptr.To(false),
			ServiceAccountName:           serviceAccountName,
			SecurityContext:              SimpleAPIApp.Spec.PodSecurityContext,
			Containers: []corev1.Container{
				{
					Name: SimpleAPIApp.Name,
					Image: fmt.Sprintf(
						"%s:%s",
						SimpleAPIApp.Spec.Image,
						SimpleAPIApp.Spec.Version,
					),
					Ports: []corev1.ContainerPort{
						{ContainerPort: SimpleAPIApp.Spec.Port},
					},
					StartupProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: SimpleAPIApp.Spec.StartupProbe.HTTPGet.Path,
								Port: SimpleAPIApp.Spec.StartupProbe.HTTPGet.Port,
							},
						},
						InitialDelaySeconds: SimpleAPIApp.Spec.StartupProbe.InitialDelaySeconds,
						PeriodSeconds:       SimpleAPIApp.Spec.StartupProbe.PeriodSeconds,
						FailureThreshold:    SimpleAPIApp.Spec.StartupProbe.FailureThreshold,
					},
					ImagePullPolicy: ImagePullPolicy,
					Resources:       SimpleAPIApp.Spec.Resources,
				},
			},
			Affinity:    SimpleAPIApp.Spec.Affinity,
			Tolerations: SimpleAPIApp.Spec.Tolerations,
		}
	}
}
