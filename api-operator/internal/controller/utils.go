package controller

import (
	"context"
	"sort"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// sortVersions sorts the version strings (assuming formats like "v21", "v22") Only thoose are accepted in this application

func sortDeploymentsByTimestamp(deployments []appsv1.Deployment) []appsv1.Deployment {
	// Sort deployments by "lastDeployedAt" annotation (or fallback to creation timestamp).
	sort.SliceStable(deployments, func(i, j int) bool {
		// Try parsing the annotation as an integer timestamp
		ti, err1 := strconv.ParseInt(deployments[i].Annotations["lastDeployedAt"], 10, 64)
		tj, err2 := strconv.ParseInt(deployments[j].Annotations["lastDeployedAt"], 10, 64)

		if err1 == nil && err2 == nil {
			return ti < tj // Sort by timestamp
		}

		// Fallback to creation timestamp if annotations are missing
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
	if len(deployments) <= 2 {
		return
	}
	oldDeployments := deployments[:len(deployments)-2]
	for _, oldDep := range oldDeployments {
		_ = r.Delete(ctx, &oldDep)
		oldServiceName := serviceNameFromDeploymentName(oldDep.Name)
		var oldSvc corev1.Service
		if err := r.Get(ctx, client.ObjectKey{Namespace: oldDep.Namespace, Name: oldServiceName}, &oldSvc); err == nil {
			_ = r.Delete(ctx, &oldSvc)
		}
	}
}

func serviceNameFromDeploymentName(deploymentName string) string {
	return strings.Replace(deploymentName, "my-api", "my-api-svc", 1)
}
