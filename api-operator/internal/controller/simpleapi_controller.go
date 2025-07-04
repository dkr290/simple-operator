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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"

	appsv1alpha1 "github.com/dkr290/simple-operator/api-operator/api/v1alpha1"
)

const (
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
	// List existing Deployments for the API using the label "app=my-api from the constants it is subject to change"
	var deploymentList appsv1.DeploymentList
	if err := r.List(ctx, &deploymentList, client.MatchingLabels{"app": SimpleapiApp.Labels["app"]}); err != nil {
		logger.Error(err, "Failed to list Deployments")
		return ctrl.Result{}, err
	}
	// Always create new deployment with unique timestamp
	timestamp := time.Now().Unix()

	// adding new deployment, this is only construct
	newDeployment := r.constructDeployment(SimpleapiApp, timestamp)
	// setting appVersion as owner for garbage collection best practices
	if err := controllerutil.SetControllerReference(&SimpleapiApp, newDeployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if deployment already exists before creating
	if err := r.Create(ctx, newDeployment); err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Info(
				"Deployment already exists, skipping creation",
				"Deployment",
				newDeployment.Name,
			)
		} else {
			logger.Error(err, "Failed to create Deployment", "Deployment", newDeployment.Name)
			return ctrl.Result{}, err
		}
	} else {
		logger.Info("Successfully created new deployment", "Deployment", newDeployment.Name)
	}

	// Create corresponding Service.
	newService := r.constructService(SimpleapiApp, timestamp)
	if err := controllerutil.SetControllerReference(&SimpleapiApp, newService, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if service already exists before creating
	if err := r.Create(ctx, newService); err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Info("Service already exists, skipping creation", "Service", newService.Name)
		} else {
			logger.Error(err, "Failed to create Service", "Service", newService.Name)
			return ctrl.Result{}, err
		}
	} else {
		logger.Info("Successfully created new service", "Service", newService.Name)
	}

	// Re-list deployments to capture the new state.
	if err := r.List(ctx, &deploymentList, client.MatchingLabels{"app": SimpleapiApp.Labels["app"]}); err != nil {
		return ctrl.Result{}, err
	}
	// Extract last two versions based on timestamps.
	sortedDeployments := sortDeploymentsByTimestamp(deploymentList.Items)
	latestVersions := extractLatestVersions(sortedDeployments)

	// Delete older versions beyond the latest two.
	r.cleanupOldDeployments(ctx, sortedDeployments)

	// Reconcile Ingress paths to reflect the latest two versions.
	if err := r.reconcileIngress(ctx, latestVersions, SimpleapiApp.Namespace, &SimpleapiApp); err != nil {
		logger.Error(err, "Failed to reconcile Ingress")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleapiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Simpleapi{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
