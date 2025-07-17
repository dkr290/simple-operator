package controller

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MemcachedReconciler) checkAndCleanDeployment(
	ctx context.Context,
	req ctrl.Request,
	logger logr.Logger,
) (ctrl.Result, error) {
	existingMemcachedDeployment := &appsv1.Deployment{}

	logger.Info("Memcached resource not found. check if a deployment must be deleted.")
	// delete deployment
	err := r.Get(
		ctx,
		types.NamespacedName{Name: req.Name, Namespace: req.Namespace},
		existingMemcachedDeployment,
	)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Nothing to do, no deployment found.")
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "❌ Failed to get Deployment")
			return ctrl.Result{}, err
		}
	}

	logger.Info("☠️ Deployment exists: delete it. ☠️")
	err = r.Delete(ctx, existingMemcachedDeployment)
	if err != nil {
		logger.Error(err, "Failed to delete the deployment")
		return ctrl.Result{}, err

	}
	return ctrl.Result{}, nil
}
