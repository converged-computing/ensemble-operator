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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
)

var (
	ensembleYamlName    = "ensemble.yaml"
	ensembleYamlDirName = "/ensemble-entrypoint"
)

// getConfigMap gets the entrypoint config map
func (r *EnsembleReconciler) ensureEnsembleConfig(
	ctx context.Context,
	name string,
	ensemble *api.Ensemble,
	member *api.Member,
) (ctrl.Result, error) {

	// Look for the config map by name
	r.Log.Info("👀️ Looking for Ensemble YAML 👀️")
	existing := &corev1.ConfigMap{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      name,
			Namespace: ensemble.Namespace,
		},
		existing,
	)

	if err != nil {

		// Case 1: not found yet, and hostfile is ready (recreate)
		if errors.IsNotFound(err) {

			// Finally create the config map
			cm := r.createConfigMap(ensemble, member, name)
			r.Log.Info("✨ Creating Ensemble YAML ✨")
			err = r.Create(ctx, cm)
			if err != nil {
				r.Log.Error(err, "❌ Failed to create Ensemble YAML")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil

		} else if err != nil {
			r.Log.Error(err, "Failed to get Ensemble YAML")
			return ctrl.Result{}, err
		}

	}
	return ctrl.Result{}, err
}

// createConfigMap generates a config map with some kind of data
func (r *EnsembleReconciler) createConfigMap(
	ensemble *api.Ensemble,
	member *api.Member,
	name string,
) *corev1.ConfigMap {

	data := map[string]string{
		ensembleYamlName: member.Ensemble,
	}
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ensemble.Namespace,
		},
		Data: data,
	}
	fmt.Println(cm.Data)
	ctrl.SetControllerReference(ensemble, cm, r.Scheme)
	return cm
}
