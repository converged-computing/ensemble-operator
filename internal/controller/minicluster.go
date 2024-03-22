package controller

import (
	"context"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	minicluster "github.com/flux-framework/flux-operator/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ensureEnsemble ensures that the ensemle is created!
func (r *EnsembleReconciler) ensureMiniClusterEnsemble(
	ctx context.Context,
	name string,
	ensemble *api.Ensemble,
	idx int,
	spec *minicluster.MiniCluster,
) (ctrl.Result, error) {

	// Look for an existing minicluster
	existing, err := r.getExistingMiniCluster(ctx, name, ensemble, spec)

	// Create a new job if it does not exist
	if err != nil {

		if errors.IsNotFound(err) {
			mc := r.newMiniCluster(name, ensemble, spec, idx)
			r.Log.Info(
				"✨ Creating a new Ensemble MiniCluster ✨",
				"Namespace:", mc.Namespace,
				"Name:", mc.Name,
			)
			err = r.Create(ctx, mc)
			if err != nil {
				r.Log.Error(
					err,
					"Failed to create new Ensemble MiniCluster",
					"Namespace:", mc.Namespace,
					"Name:", mc.Name,
				)
				return ctrl.Result{}, err
			}
			// Successful - return and requeue
			return ctrl.Result{Requeue: true}, nil

		} else if err != nil {
			r.Log.Error(err, "Failed to get Ensemble MiniCluster")
			return ctrl.Result{}, err
		}

	} else {
		r.Log.Info(
			"🎉 Found existing MiniCluster Service Pod 🎉",
			"Namespace:", existing.Namespace,
			"Name:", existing.Name,
		)
	}
	return ctrl.Result{}, err
}

// getExistingPod gets an existing pod service
func (r *EnsembleReconciler) getExistingMiniCluster(
	ctx context.Context,
	name string,
	ensemble *api.Ensemble,
	spec *minicluster.MiniCluster,
) (*minicluster.MiniCluster, error) {

	existing := &minicluster.MiniCluster{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      name,
			Namespace: ensemble.Namespace,
		},
		existing,
	)
	return existing, err
}

// newMiniCluster creates a new ensemble minicluster
func (r *EnsembleReconciler) newMiniCluster(
	name string,
	ensemble *api.Ensemble,
	spec *minicluster.MiniCluster,
	idx int,
) *minicluster.MiniCluster {

	// The size should be set to the desired size
	member := ensemble.Spec.Members[idx]
	spec.Spec.Size = member.DesiredSize
	spec.Spec.MaxSize = member.MaxSize
	spec.Spec.MinSize = member.MinSize
	spec.ObjectMeta = metav1.ObjectMeta{Name: name, Namespace: ensemble.Namespace}
	ctrl.SetControllerReference(ensemble, spec, r.Scheme)
	return spec
}
