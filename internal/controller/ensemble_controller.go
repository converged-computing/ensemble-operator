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
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=ensemble.flux-framework.org,resources=ensembles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ensemble.flux-framework.org,resources=ensembless,verbs=status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ensemble.flux-framework.org,resources=ensembles,verbs=finalizers,verbs=update

//+kubebuilder:rbac:groups=flux-framework.org,resources=miniclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flux-framework.org,resources=miniclusters/status,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flux-framework.org,resources=miniclusters/finalizers,verbs=get;list;watch;create;update;patch;delete

// Reconcile until the cluster matches the state of the desired Ensemble
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *EnsembleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// set the log if not done yet
	r.Log = log.FromContext(ctx)

	// Create a new ensemble
	var ensemble api.Ensemble

	// Keep developer informed what is going on.
	r.Log.Info("🥞️ Ensemble! Like pancakes")
	r.Log.Info("Request: ", "req", req)

	// Does the Ensemble exist yet (based on name and namespace)
	err := r.Get(ctx, req.NamespacedName, &ensemble)
	if err != nil {

		// Create it, doesn't exist yet
		if errors.IsNotFound(err) {
			r.Log.Info("🥞️ Ensemble not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		r.Log.Info("🥞️ Failed to get Ensemble. Re-running reconcile.")
		return ctrl.Result{Requeue: true}, err
	}

	// Show parameters provided and validate one flux runner
	err = ensemble.Validate()
	if err != nil {
		r.Log.Error(err, "🥞️ Your ensemble did not validate")
		return ctrl.Result{}, nil
	}
	r.Log.Info("🥞️ Reconciling Ensemble", "Members: ", len(ensemble.Spec.Members))

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
	}

	// By the time we get here we have a Job + pods + config maps!
	// What else do we want to do?
	r.Log.Info("🥞️ Ensemble is Ready!")
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnsembleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Ensemble{}).
		Owns(&minicluster.MiniCluster{}).
		Complete(r)
}
