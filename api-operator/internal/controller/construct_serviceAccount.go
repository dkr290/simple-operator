package controller

import (
	appsv1alpha1 "github.com/dkr290/simple-operator/api-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *SimpleapiReconciler) constructServiceAccount(
	SimpleAPIApp appsv1alpha1.Simpleapi,
) *corev1.ServiceAccount {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SimpleAPIApp.Spec.ServiceAccountName,
			Namespace: SimpleAPIApp.Namespace,
			Labels: map[string]string{
				"app": SimpleAPIApp.Name,
			},
			// Optionally set owner references for garbage collection
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: SimpleAPIApp.APIVersion,
					Kind:       SimpleAPIApp.Kind,
					Name:       SimpleAPIApp.Name,
					UID:        SimpleAPIApp.UID,
				},
			},
		},
	}
	return serviceAccount
}
