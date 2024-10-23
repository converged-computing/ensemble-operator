package controller

import (
	"context"
	"fmt"
	"strconv"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// getDeploymentAddress gets the address of the deployment
func (r *EnsembleReconciler) getDeploymentAddress(
	ctx context.Context,
	ensemble *api.Ensemble,
) (string, error) {

	// The MiniCluster service is being provided by the index 0 pod, so we can find it here.
	clientset, err := kubernetes.NewForConfig(r.RESTConfig)
	if err != nil {
		return "", err
	}

	matchLabels := getDeploymentLabels(ensemble)

	// A selector just for the lead broker pod of the ensemble MiniCluster
	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}
	pods, err := clientset.CoreV1().Pods(ensemble.Namespace).List(ctx, listOptions)
	if err != nil {
		fmt.Printf("      Error with listing pods %s\n", err)
		return "", err
	}

	// Get the ip address of the first (only for now) pod
	var ipAddress string
	for i, pod := range pods.Items {
		pod, err := clientset.CoreV1().Pods(ensemble.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("      Error with getting pods %s\n", err)
			return "", err
		}
		fmt.Printf("      Pod IP Address %s\n", pod.Status.PodIP)
		ipAddress = pod.Status.PodIP

		// We only need the first pod, if there is more than one
		// I don't think we will need to scale this, but can
		// figure out how to handle that if it comes to that.
		if i > 0 {
			break
		}
	}

	// If we don't have an ip address yet, try again later
	if ipAddress == "" {
		fmt.Println("      No pods found")
		return "", fmt.Errorf("no pods found, not ready yet")
	}
	return ipAddress, nil
}

// getServiceAddress gets the service ClusterIP serving the grpc endpoint
func (r *EnsembleReconciler) getServiceAddress(
	ctx context.Context,
	ensemble *api.Ensemble,
) (string, error) {

	// The MiniCluster service is being provided by the index 0 pod, so we can find it here.
	clientset, err := kubernetes.NewForConfig(r.RESTConfig)
	if err != nil {
		return "", err
	}

	// List all services with this name (just the one!)
	services, err := clientset.CoreV1().Services(ensemble.Namespace).List(
		ctx,
		metav1.ListOptions{
			FieldSelector: "metadata.name=" + ensemble.ServiceName(),
		},
	)
	if err != nil {
		return "", err
	}

	// Get the ip address of the first (only for now) pod
	var ipAddress string
	for _, svc := range services.Items {
		ipAddress = svc.Spec.ClusterIP
		break
	}

	// If we don't have an ip address yet, try again later
	if ipAddress == "" {
		fmt.Println("      No grpc services found")
		return "", fmt.Errorf("no grpc services found, not ready yet")
	}
	return ipAddress, nil
}

func (r *EnsembleReconciler) createServiceAccount(
	ctx context.Context,
	ensemble *api.Ensemble,
) (ctrl.Result, error) {

	// First see if we already have it!
	sa := &corev1.ServiceAccount{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      ensemble.Name,
			Namespace: ensemble.Namespace,
		},
		sa,
	)

	// If we haven't found it, create it
	if err != nil {
		if errors.IsNotFound(err) {
			sa = &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ensemble.Name,
					Namespace: ensemble.Namespace,
				},
			}
			ctrl.SetControllerReference(ensemble, sa, r.Scheme)

			// We should not have an error with creation
			err = r.Create(ctx, sa)
			if err != nil {
				return ctrl.Result{}, err
			}

			// Otherwise, requeue - we'll make another object
			// the next time around.
			return ctrl.Result{Requeue: true}, nil
		}
		// This means an error that isn't covered
		return ctrl.Result{}, err
	}
	// We already have the service account, no error
	// and continue to next thing.
	return ctrl.Result{}, nil
}

// createRole creates the RBAC role that will allow the ensemble service
// to control the ensemble object (and update it, etc.). See function
// above for comments about logic of this function.
func (r *EnsembleReconciler) createRole(
	ctx context.Context,
	ensemble *api.Ensemble,
) (ctrl.Result, error) {

	role := &rbacv1.Role{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      ensemble.Name,
			Namespace: ensemble.Namespace,
		},
		role,
	)
	if err != nil {
		if errors.IsNotFound(err) {
			role := &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ensemble.Name,
					Namespace: ensemble.Namespace,
				},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{"flux-framework.org"},
						Resources: []string{"miniclusters"},
						Verbs:     []string{"get", "list", "create", "update", "delete", "patch"},
					},
				},
			}

			ctrl.SetControllerReference(ensemble, role, r.Scheme)
			err = r.Create(ctx, role)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil

}

// createRoleBinding will bind the service account to the role we created
// This will give the grpc service permission to issue updates to the MiniCluster
func (r *EnsembleReconciler) createRoleBinding(
	ctx context.Context,
	ensemble *api.Ensemble,
) (ctrl.Result, error) {

	rb := &rbacv1.RoleBinding{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      ensemble.Name,
			Namespace: ensemble.Namespace,
		},
		rb,
	)
	if err != nil {
		if errors.IsNotFound(err) {
			rb := &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ensemble.Name,
					Namespace: ensemble.Namespace,
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     ensemble.Name,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      ensemble.Name,
						Namespace: ensemble.Namespace,
					},
				},
			}
			ctrl.SetControllerReference(ensemble, rb, r.Scheme)
			err = r.Create(ctx, rb)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil

}

// createService creates the service for the grpc
// This is used to expose the port to the cluster
// TODO stopped here - bring up interactive and debug grpc (it worked before)
func (r *EnsembleReconciler) createService(
	ctx context.Context,
	ensemble *api.Ensemble,
) (ctrl.Result, error) {

	// First see if we already have it!
	svc := &corev1.Service{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      ensemble.ServiceName(),
			Namespace: ensemble.Namespace,
		},
		svc,
	)

	// If we haven't found it, create it
	if err != nil {
		if errors.IsNotFound(err) {

			// Deployment labels to match for service
			appLabels := getDeploymentLabels(ensemble)
			port, err := strconv.Atoi(ensemble.Spec.Sidecar.Port)
			if err != nil {
				return ctrl.Result{}, err
			}

			svc = &corev1.Service{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ensemble.ServiceName(),
					Namespace: ensemble.Namespace,
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							TargetPort: intstr.FromInt(int(port)),
							Protocol:   "TCP",
							Port:       int32(port),
						},
					},
					Selector: appLabels,
				},
			}
			ctrl.SetControllerReference(ensemble, svc, r.Scheme)

			// We should not have an error with creation
			err = r.Create(ctx, svc)
			if err != nil {
				return ctrl.Result{}, err
			}

			// Otherwise, requeue - we'll make another object
			// the next time around.
			return ctrl.Result{Requeue: true}, nil
		}
		// This means an error that isn't covered
		return ctrl.Result{}, err
	}
	// We already have the service account, no error
	// and continue to next thing.
	return ctrl.Result{Requeue: true}, nil
}

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

	// First we care about the service object itself.
	// This will be generated with rbac so the service has permission
	// to update the MiniCluster (and eventually other ensemble
	// members in the space
	result, err := r.createServiceAccount(ctx, ensemble)
	if err != nil {
		return result, err
	}

	// Create the service for the deployment
	result, err = r.createService(ctx, ensemble)
	if err != nil {
		return result, err
	}

	// Once we have a service account, create a role for it
	result, err = r.createRole(ctx, ensemble)
	if err != nil {
		return result, err
	}

	// And the role binding for the TBA grpc deployment
	result, err = r.createRoleBinding(ctx, ensemble)
	if err != nil {
		return result, err
	}

	// Next, we want to create a deployment that serves the grpc
	_, err = r.getExistingDeployment(ctx, ensemble)

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

// getDeploymentLabels returns common labels for the deployment
// We use these to find the deployment later and get the
// ip address to ping grpc at!
func getDeploymentLabels(
	ensemble *api.Ensemble,
) map[string]string {
	return map[string]string{
		"app":      ensemble.Name,
		"operator": "ensemble-operator",
	}
}

// newMiniCluster creates a new ensemble minicluster
func (r *EnsembleReconciler) newEnsembleDeployment(ensemble *api.Ensemble) (*appsv1.Deployment, error) {

	imagePullPolicy := "IfNotPresent"
	if ensemble.Spec.Sidecar.ImagePullPolicy != "" {
		imagePullPolicy = ensemble.Spec.Sidecar.ImagePullPolicy

	}
	appLabels := getDeploymentLabels(ensemble)

	port, err := strconv.ParseInt(ensemble.Spec.Sidecar.Port, 10, 32)
	if err != nil {
		return nil, err
	}

	// Custom command with number of workers
	workers := strconv.Itoa(int(ensemble.Spec.Sidecar.Workers))
	command := []string{
		"ensemble-server",
		"start",
		"--kubernetes",
		"--host", "0.0.0.0",
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

					// This needs to match the service name
					Subdomain:          ensemble.ServiceName(),
					ServiceAccountName: ensemble.Name,
					Containers: []corev1.Container{
						{
							// matches the service
							Name:            "ensemble-service",
							Image:           ensemble.Spec.Sidecar.Image,
							ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
							Command:         command,
							TTY:             true,
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
