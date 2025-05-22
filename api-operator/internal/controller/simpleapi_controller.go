/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	appsv1alpha1 "github.com/dkr290/simple-operator/api-operator/api/v1alpha1"
)

const (
	appLabel        = "my-api"
	versionLabel    = "version"
	ingressName     = "my-api-ingress"
	imagePullsecret = "regcred"
)

// SimpleapiReconciler reconciles a Simpleapi object
type SimpleapiReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.api.test,resources=simpleapis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.api.test,resources=simpleapis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.api.test,resources=simpleapis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Simpleapi object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *SimpleapiReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// trying to fetch appversion CR instance, which is like custom CRD called Simpleapi
	// that is defined in v1alpha1 simpleapi_types
	var SimpleapiApp appsv1alpha1.Simpleapi
	if err := r.Get(ctx, req.NamespacedName, &SimpleapiApp); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("AppVersion resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get AppVersion")
		return ctrl.Result{}, err
	}
	// List existing Deployments for the API using the label "app=my-api from the contants it is subject to change"
	var deploymentList appsv1.DeploymentList
	if err := r.List(ctx, &deploymentList, client.MatchingLabels{"app": appLabel}); err != nil {
		logger.Error(err, "Failed to list Deployments")
		return ctrl.Result{}, err
	}

	// get the active current version from the Deployment byt version= key
	existingVersions := map[string]*appsv1.Deployment{}
	for k, v := range deploymentList.Items {
		if ver, exists := v.Labels[versionLabel]; exists {
			existingVersions[ver] = &deploymentList.Items[k]
		}
	}

	desiredVersion := SimpleapiApp.Spec.Version
	// If the desired version is not deployed, create new Deployment and respective Service. We just do ingress mapping after
	if _, exists := existingVersions[desiredVersion]; !exists {
		logger.Info("Creating new deployment and service for version", "version", desiredVersion)
		// adding new deployment, this is only construct
		deployment := r.constructDeployment(SimpleapiApp)
		// setting appVersion as owner for garbage collection best practices
		if err := controllerutil.SetControllerReference(&SimpleapiApp, deployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, deployment); err != nil {
			logger.Error(err, "Failed to create Deployment", "Deployment", deployment)
			return ctrl.Result{}, err
		}
		// creating the connecting service
		// creating the service --------------------------------

		service := r.constructService(SimpleapiApp)
		if err := controllerutil.SetControllerReference(&SimpleapiApp, service, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		// actual creation of the service was missing so adding
		if err := r.Create(ctx, service); err != nil {
			logger.Error(err, "Failed to create Service", "Service", service)
			return ctrl.Result{}, err
		}
		// adding desired version ,adding to our collection of existing versions
		existingVersions[desiredVersion] = deployment

	}

	// Sort the versions (assumes versions in this application are only in the form of like "v21", "v22", etc.)
	sortedVersions := sortVersions(existingVersions)
	if len(sortedVersions) > 2 {
		toDelete := sortedVersions[:len(sortedVersions)-2]
		for _, ver := range toDelete {
			if ver == desiredVersion {
				// trying to fix here the rollback case (rollback case)
				logger.Info("Existing versions found:", "versions", sortedVersions)
				continue
			}
			logger.Info("Deleting old deployment and service for version", "version", ver)
			// Delete the old deployment
			if dep, ok := existingVersions[ver]; ok {
				if err := r.Delete(ctx, dep); err != nil {
					logger.Error(err, "Failed to delete Deployment", "version", ver)
				}
			}

			// Delete the old service which should be associated with the old deployment

			svcName := serviceName(ver)
			var svc corev1.Service
			if err := r.Get(ctx, client.ObjectKey{Namespace: SimpleapiApp.Namespace, Name: svcName}, &svc); err == nil {
				if err := r.Delete(ctx, &svc); err != nil {
					logger.Error(
						err,
						"Failed to delete Service",
						"Service",
						svcName,
						"version",
						ver,
					)
					return ctrl.Result{}, err
				}
			}
		}

		// Keep only the newest two versions.
		sortedVersions = sortedVersions[len(sortedVersions)-2:]
	}
	if err := r.reconcileIngress(ctx, sortedVersions, req.Namespace, &SimpleapiApp); err != nil {
		logger.Error(err, "Failed to reconcile Ingress")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleapiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Simpleapi{}).
		Complete(r)
}
