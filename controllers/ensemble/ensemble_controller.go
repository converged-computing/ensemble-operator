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

	// Do we need to init the jobs matrix?
	if len(ensemble.Status.Jobs) == 0 {
		return r.initJobsMatrix(ctx, &ensemble)
	}

	// Ensure we have the MiniCluster (get or create!)
	// We only have MiniCluster now, but this design can be extended to others
	for i, member := range ensemble.Spec.Members {

		// This indicates the ensemble member is a MiniCluster
		if !reflect.DeepEqual(member.MiniCluster, minicluster.MiniClusterSpec{}) {

			// Name is the index + ensemble name
			name := fmt.Sprintf("%s-%d", ensemble.Name, i)
			result, err := r.ensureMiniClusterEnsemble(ctx, name, &ensemble, &member)
			if err != nil {
				return result, err
			}
		}
		// TODO do we want to reconcile earlier here, between ensemble components?
	}

	// When we have all ensembles, get updates
	// I'm assuming here these queries might also be member-specific
	for i, member := range ensemble.Spec.Members {

		// This indicates the ensemble member is a MiniCluster
		if !reflect.DeepEqual(member.MiniCluster, minicluster.MiniClusterSpec{}) {
			name := fmt.Sprintf("%s-%d", ensemble.Name, i)
			fmt.Printf("      Checking member %s\n", name)
			result, err := r.updateMiniClusterEnsemble(ctx, name, &ensemble, &member, i)

			// An error could indicates a requeue (without the break) since something was off
			// We likely need to tweak this a bit to account for potential updates to specific
			// ensemble members, but this is fine for a first shot.
			if err != nil {
				return result, err
			}
		}
	}
	fmt.Println("      Ensemble is Ready!")

	// If we've run updates across them, should requeue per preference of ensemble check frequency
	return ctrl.Result{RequeueAfter: ensemble.RequeueAfter()}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnsembleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Ensemble{}).
		Owns(&minicluster.MiniCluster{}).
		Complete(r)
}
