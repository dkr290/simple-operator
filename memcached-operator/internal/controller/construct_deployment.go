// Package controller same part of the controller package
package controller

import (
	"context"

	cachev1alpha1 "github.com/dkr290/memcached-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *MemcachedReconciler) constructDeployment(
	memcached *cachev1alpha1.Memcached,
) appsv1.Deployment {
	labels := map[string]string{
		"app": memcached.Labels["app"],
	}

	metadata := metav1.ObjectMeta{
		Name:      "memcached",
		Namespace: memcached.Namespace,
		Labels:    labels,
	}

	replicas := memcached.Spec.Size
	spec := appsv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "memcached",
						Image: memcached.Spec.Image,
						Ports: []v1.ContainerPort{
							{
								ContainerPort: memcached.Spec.ContainerPort,
							},
						},
						Env: []v1.EnvVar{
							{
								Name:  "env",
								Value: "dev",
							},
						},
						ImagePullPolicy: "IfNotPresent",
					},
				},
				ImagePullSecrets: []v1.LocalObjectReference{
					{Name: memcached.Spec.ImagePullSecret},
				},
			},
		},
	}

	dep := appsv1.Deployment{
		ObjectMeta: metadata,
		Spec:       spec,
	}
	return dep
}

func (r *MemcachedReconciler) createDeploument(ctx context.Context, depl *appsv1.Deployment) error {
	logger := logf.FromContext(ctx)

	// Check if deployment already exists before creating
	if err := r.Create(ctx, depl); err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Info(
				"Deployment already exists, skipping creation",
				"Deployment",
				depl.Name,
			)
		} else {
			logger.Error(err, "Failed to create Deployment", "Deployment", depl.Name)
			return err
		}
	} else {
		logger.Info("Successfully created new deployment", "Deployment", depl.Name)
	}
	return nil
}
