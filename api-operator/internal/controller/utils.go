package controller

import (
	"context"
	"sort"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// sortVersions sorts the version strings (assuming formats like "v21", "v22") // so with 2 works and keeps them by timestamp

func sortDeploymentsByTimestamp(deployments []appsv1.Deployment) []appsv1.Deployment {
	// Sort deployments by "lastDeployedAt" annotation (or fallback to creation timestamp not tested yet).
	sort.SliceStable(deployments, func(i, j int) bool {
		annI, okI := deployments[i].Annotations["lastDeployedAt"]
		annJ, okJ := deployments[j].Annotations["lastDeployedAt"]

		if okI && okJ {
			ti, err1 := strconv.ParseInt(annI, 10, 64)
			tj, err2 := strconv.ParseInt(annJ, 10, 64)

			// Both annotations are parsable as integers
			if err1 == nil && err2 == nil {
				if ti == tj {
					// If timestamps are identical, use CreationTimestamp as a tie-breaker
					return deployments[i].CreationTimestamp.Before(
						&deployments[j].CreationTimestamp,
					)
				}
				return ti < tj // Sort by "lastDeployedAt" timestamp (ascending: older first)
			}
		}

		// Fallback to creation timestamp if annotations are missing,
		// TODO to test this
		return deployments[i].CreationTimestamp.Before(&deployments[j].CreationTimestamp)
	})

	return deployments
}

func extractLatestVersions(deployments []appsv1.Deployment) []string {
	latestVersions := []string{}

	// Ensure we only extract up to two versions
	if len(deployments) > 2 {
		deployments = deployments[len(deployments)-2:]
	}

	for _, dep := range deployments {
		if version, exists := dep.Labels["version"]; exists {
			latestVersions = append(latestVersions, version)
		}
	}

	return latestVersions
}

// cleanupOldDeployments removes older deployments beyond the latest two.
func (r *SimpleapiReconciler) cleanupOldDeployments(
	ctx context.Context,
	deployments []appsv1.Deployment,
) {
	logger := log.FromContext(ctx)
	if len(deployments) <= 2 {
		return
	}
	oldDeployments := deployments[:len(deployments)-2]
	for _, oldDep := range oldDeployments {
		depToDelete := oldDep // Use a new variable for the loop to avoid issues with pointers in loops if not careful
		logger.Info(
			"Deleting old deployment",
			"deployment",
			depToDelete.Name,
			"namespace",
			depToDelete.Namespace,
		)

		if err := r.Delete(ctx, &depToDelete); err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to delete old deployment", "deployment", depToDelete.Name)
			// maybe to see how to return error as improvement
		}

		oldServiceName := serviceNameFromDeploymentName(depToDelete.Name)
		oldSvc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      oldServiceName,
				Namespace: depToDelete.Namespace,
			},
		}
		logger.Info(
			"Attempting to delete old service",
			"service",
			oldServiceName,
			"namespace",
			depToDelete.Namespace,
		)

		if err := r.Delete(ctx, oldSvc); err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to delete old service", "service", oldServiceName)
		}
	}
}

func serviceNameFromDeploymentName(deploymentName string) string {
	return deploymentName + "-svc"
}
