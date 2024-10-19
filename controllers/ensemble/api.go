package controller

import (
	"context"
	"fmt"
	"strconv"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ensureEnsembleService creates the deployment to run the ensemble service
// It is agnostic to members, but interacts with ensembles in its namespace
// and name, and then can issue requests for scale / change back to
// the operator.
func (r *EnsembleReconciler) ensureEnsembleService(
	ctx context.Context,
	ensemble *api.Ensemble,
) (ctrl.Result, error) {

	// This is the Minicluster that we found
	fmt.Println("✨ Ensuring Ensemble Deployment Service")

	// Look for an existing minicluster
	_, err := r.getExistingDeployment(ctx, ensemble)

	// Create a new job if it does not exist
	if err != nil {
		if errors.IsNotFound(err) {
			mc, err := r.newEnsembleDeployment(ensemble)
			if err != nil {
				fmt.Printf("      Failed to create Deployment object: %s\n", err)
				return ctrl.Result{}, err
			}
			fmt.Println("      Creating a new Ensemble Service Deployment")
			err = r.Create(ctx, mc)
			if err != nil {
				fmt.Printf("      Failed to create Ensemble Service Deployment: %s\n", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		// This means an error that isn't covered
		return ctrl.Result{}, err
	}
	// We need to requeue since we check the status with reconcile
	return ctrl.Result{Requeue: true}, err
}

// getExistingDeployment gets an existing deployment service
func (r *EnsembleReconciler) getExistingDeployment(
	ctx context.Context,
	ensemble *api.Ensemble,
) (*appsv1.Deployment, error) {

	existing := &appsv1.Deployment{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      ensemble.Name,
			Namespace: ensemble.Namespace,
		},
		existing,
	)
	return existing, err
}

// newMiniCluster creates a new ensemble minicluster
func (r *EnsembleReconciler) newEnsembleDeployment(ensemble *api.Ensemble) (*appsv1.Deployment, error) {

	imagePullPolicy := "IfNotPresent"
	if ensemble.Spec.Sidecar.PullAlways {
		imagePullPolicy = "Always"
	}

	// Put shared labels
	appLabels := map[string]string{
		"app":      ensemble.Name,
		"operator": "ensemble-operator",
	}

	port, err := strconv.ParseInt(ensemble.Spec.Sidecar.Port, 10, 32)
	if err != nil {
		return nil, err
	}

	// Custom command with number of workers
	workers := strconv.Itoa(int(ensemble.Spec.Sidecar.Workers))
	command := []string{
		"ensemble-server",
		"start",
		"--port", ensemble.Spec.Sidecar.Port,
		"--workers", workers,
	}

	// Assume 1 replica for now, we can always expose this
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ensemble.Name,
			Namespace: ensemble.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: appLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: appLabels,
				},
				Spec: corev1.PodSpec{
					Subdomain: ensemble.Name,
					Containers: []corev1.Container{
						{
							// matches the service
							Name:            "ensemble-service",
							Image:           ensemble.Spec.Sidecar.Image,
							ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
							Command:         command,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: int32(port),
								},
							},
						},
					},
				},
			},
		},
	}
	ctrl.SetControllerReference(ensemble, deployment, r.Scheme)
	return deployment, nil
}
