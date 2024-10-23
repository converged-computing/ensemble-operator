/*
Copyright 2024.

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
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	minicluster "github.com/flux-framework/flux-operator/api/v1alpha2"
	"github.com/go-logr/logr"
)

// EnsembleReconciler reconciles a Ensemble object
type EnsembleReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Log        logr.Logger
	RESTClient rest.Interface
	RESTConfig *rest.Config
}

//+kubebuilder:rbac:groups=ensemble.flux-framework.org,resources=ensembles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ensemble.flux-framework.org,resources=ensembles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ensemble.flux-framework.org,resources=ensembles,verbs=finalizers,verbs=get;update;patch

//+kubebuilder:rbac:groups=flux-framework.org,resources=miniclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flux-framework.org,resources=miniclusters/status,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flux-framework.org,resources=miniclusters/finalizers,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=core,resources=pods/log,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/exec,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile until the cluster matches the state of the desired Ensemble
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *EnsembleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// set the log if not done yet
	r.Log = log.FromContext(ctx)

	// Create a new ensemble
	var ensemble api.Ensemble

	// Keep developer informed what is going on.
	fmt.Println("🥞️ Ensemble!")
	fmt.Printf("   => Request: %s\n", req)

	// Does the Ensemble exist yet (based on name and namespace)
	err := r.Get(ctx, req.NamespacedName, &ensemble)
	if err != nil {

		// Create it, doesn't exist yet
		if errors.IsNotFound(err) {
			fmt.Println("      Ensemble not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		fmt.Println("      Failed to get Ensemble. Re-running reconcile.")
		return ctrl.Result{Requeue: true}, err
	}

	// Show parameters provided and validate one flux runner
	err = ensemble.Validate()
	if err != nil {
		r.Log.Error(err, "      Your ensemble did not validate")
		return ctrl.Result{}, err
	}
	fmt.Printf("      Members %d\n", len(ensemble.Spec.Members))

	// First create the grpc service that will coordinate with all ensembles
	// This takes stress off of the operator to do the individual updaters,
	// and we only need to change here to request changes to the elements
	// themselves (e.g., scale up/down).

	// This is on the same headless service as the MiniCluster (or ensemble members)
	// It needs to be running first, in case there are requests to it from members!
	result, err := r.ensureEnsembleService(ctx, &ensemble)
	if err != nil {
		return result, err
	}

	// Ensure we have the MiniCluster (get or create!)
	// We only have MiniCluster now, but this design can be extended to others
	for i, member := range ensemble.Spec.Members {

		// This indicates the ensemble member is a MiniCluster
		if !reflect.DeepEqual(member.MiniCluster, minicluster.MiniClusterSpec{}) {

			// Name is the index + ensemble name
			name := fmt.Sprintf("%s-%d", ensemble.Name, i)

			// Create the config map volume (the ensemble.yaml)
			// for the MiniCluster to run as the entrypoint
			result, err := r.ensureEnsembleConfig(ctx, name, &ensemble, &member)
			if err != nil {
				return result, err
			}

			result, err = r.ensureMiniClusterEnsemble(ctx, name, &ensemble, &member)
			if err != nil {
				return result, err
			}
		}
	}
	fmt.Println("      Ensemble is Ready!")

	// If we've run updates across them, should requeue per preference of ensemble check frequency
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnsembleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Ensemble{}).
		Owns(&minicluster.MiniCluster{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&rbacv1.Role{}).
		Complete(r)
}
