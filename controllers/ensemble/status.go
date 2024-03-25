package controller

import (
	"context"
	"fmt"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// initJobsMatrix sets up the jobs matrix in status (for updating) to work from
func (r *EnsembleReconciler) initJobsMatrix(
	ctx context.Context,
	ensemble *api.Ensemble,
) (ctrl.Result, error) {

	fmt.Println("      Initializing Jobs Matrix")
	patch := kclient.MergeFrom(ensemble.DeepCopy())
	ensemble.Status.Jobs = map[string][]api.Job{}
	for i, member := range ensemble.Spec.Members {
		idx := fmt.Sprintf("%d", i)
		ensemble.Status.Jobs[idx] = member.Jobs
	}
	err := r.Status().Update(ctx, ensemble)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.Patch(ctx, ensemble, patch)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, err
}

// updateJobsMatrix to include jobs
func (r *EnsembleReconciler) updateJobsMatrix(
	ctx context.Context,
	ensemble *api.Ensemble,
	jobs []api.Job,
	i int,
) (ctrl.Result, error) {

	fmt.Printf("➕️ Jobs Matrix Update")
	patch := kclient.MergeFrom(ensemble.DeepCopy())
	idx := fmt.Sprintf("%d", i)

	ensemble.Status.Jobs[idx] = jobs
	err := r.Status().Update(ctx, ensemble)
	if err != nil {
		fmt.Printf("      Error with updating jobs matrix %s\n", err)
		return ctrl.Result{}, err
	}
	err = r.Patch(ctx, ensemble, patch)
	if err != nil {
		fmt.Printf("      Error with patching jobs matrix %s\n", err)
		return ctrl.Result{}, err
	}
	fmt.Printf("      Jobs matrix is updated %v\n", ensemble.Status.Jobs)
	return ctrl.Result{Requeue: true}, nil
}
