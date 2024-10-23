package controller

import (
	"context"
	"fmt"
	"path/filepath"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	minicluster "github.com/flux-framework/flux-operator/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	// Assume the user chose the wrong view, and install python3 cffi
	preCommand = `
apt-get update && apt-get install -y python3-cffi || yum update && yum install -y python3-cffi
python3 -m pip install ensemble-python || echo "please install ensemble-python"
`

	// Custom install from a branch
	branchPreCommand = `
apt-get update && apt-get install -y python3-cffi git || yum update && yum install -y python3-cffi git
git clone -b %s --depth 1 https://github.com/converged-computing/ensemble-python.git /tmp/ensemble-python
cd /tmp/ensemble-python
python3 -m pip install . || echo "please install ensemble-python"
cd -
`
)

// ensureMiniClusterEnsemble ensures that the ensemle is created!
func (r *EnsembleReconciler) ensureMiniClusterEnsemble(
	ctx context.Context,
	name string,
	ensemble *api.Ensemble,
	member *api.Member,
) (ctrl.Result, error) {

	// This is the Minicluster that we found
	spec := &member.MiniCluster
	fmt.Println("✨ Ensuring Ensemble MiniCluster")

	// Look for an existing minicluster
	_, err := r.getExistingMiniCluster(ctx, name, ensemble)

	// Create a new job if it does not exist
	if err != nil {
		if errors.IsNotFound(err) {

			// We first need the address of the grpc service
			// if this fails, we try again - it might not be ready
			ipAddress, err := r.getServiceAddress(ctx, ensemble, name)
			if err != nil {
				return ctrl.Result{Requeue: true}, err
			}

			// The address si given to the minicluster start command
			// The MiniCluster queue communicates to it for grow/shrink requests
			mc := r.newMiniCluster(name, ensemble, member, spec, ipAddress)
			fmt.Println("      Creating a new Ensemble MiniCluster")
			err = r.Create(ctx, mc)
			if err != nil {
				fmt.Println("      Failed to create Ensemble MiniCluster")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		// This means an error that isn't covered
		return ctrl.Result{}, err
	} else {
		fmt.Println("      Found existing Ensemble MiniCluster")
	}
	// We need to requeue since we check the status with reconcile
	return ctrl.Result{Requeue: true}, err
}

// getExistingPod gets an existing pod service
func (r *EnsembleReconciler) getExistingMiniCluster(
	ctx context.Context,
	name string,
	ensemble *api.Ensemble,
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

// updateMiniCluster size gets its current size from the status and updated
// if it is valid
func (r *EnsembleReconciler) updateMiniClusterSize(
	ctx context.Context,
	ensemble *api.Ensemble,
	scale int32,
	name string,
) (ctrl.Result, error) {

	mc, err := r.getExistingMiniCluster(ctx, name, ensemble)

	// Check the size against what we have
	size := mc.Spec.Size

	// We can only scale if we are left with at least one node
	// If we want to scale to 0, this should be a termination event
	newSize := size + scale
	if newSize < 1 {
		fmt.Printf("        Ignoring scaling event, new size %d is < 1\n", newSize)
		return ctrl.Result{}, err
	}
	if newSize <= mc.Spec.MaxSize {
		fmt.Printf("        Updating size from %d to %d\n", size, newSize)
		mc.Spec.Size = newSize

		// TODO: this will trigger reconcile. Can we set the time?
		err = r.Update(ctx, mc)
		if err != nil {
			return ctrl.Result{}, err
		}

	} else {
		fmt.Printf("        Ignoring scaling event %d to %d, outside allowed boundary\n", size, newSize)
	}

	// Check again in the allotted time
	return ctrl.Result{}, err
}

// newMiniCluster creates a new ensemble minicluster
func (r *EnsembleReconciler) newMiniCluster(
	name string,
	ensemble *api.Ensemble,
	member *api.Member,
	spec *minicluster.MiniCluster,
	host string,
) *minicluster.MiniCluster {

	// The size should be set to the desired size
	spec.ObjectMeta = metav1.ObjectMeta{Name: name, Namespace: ensemble.Namespace}

	// Ensure the service name is the ensemble name so the ensemble service
	// can share it too!
	spec.Spec.Network = minicluster.Network{HeadlessName: ensemble.Name}

	// Files to mount from configMap
	items := map[string]string{
		ensembleYamlName: ensembleYamlName,
	}

	// Add the config map as a volume to the main container
	container := spec.Spec.Containers[0]
	volume := minicluster.ContainerVolume{
		ConfigMapName: ensemble.Name,
		Path:          "/ensemble-entrypoint",
		Items:         items,
	}
	container.Volumes = map[string]minicluster.ContainerVolume{ensemble.Name: volume}
	container.RunFlux = true
	container.Launcher = true

	command := preCommand
	if member.Branch != "" {
		command = fmt.Sprintf(branchPreCommand, member.Branch)
	}
	// Install ensemble via python, either from pip or github
	container.Commands = minicluster.Commands{Pre: command}

	// Note that we aren't creating a headless service so that the different members are isolated.
	// Otherwise they would all be on the same service address, which might get ugly.
	ensembleYamlPath := filepath.Join(ensembleYamlDirName, ensembleYamlName)
	prefix := "ensemble run --executor minicluster --host"
	container.Command = fmt.Sprintf("%s %s --port %s --name %s %s",
		prefix, host,
		ensemble.Spec.Sidecar.Port, ensemble.Name,
		ensembleYamlPath,
	)
	spec.Spec.Containers[0] = container
	fmt.Println(spec.Spec)
	ctrl.SetControllerReference(ensemble, spec, r.Scheme)
	return spec
}
