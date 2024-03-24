package controller

import (
	"context"
	"fmt"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	"github.com/converged-computing/ensemble-operator/internal/algorithm"
	"github.com/converged-computing/ensemble-operator/internal/client"
	pb "github.com/converged-computing/ensemble-operator/protos"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ensureEnsemble ensures that the ensemle is created!
func (r *EnsembleReconciler) updateMiniClusterEnsemble(
	ctx context.Context,
	name string,
	ensemble *api.Ensemble,
	member *api.Member,
) (ctrl.Result, error) {

	// The MiniCluster service is being provided by the index 0 pod, so we can find it here.
	clientset, err := kubernetes.NewForConfig(r.RESTConfig)
	if err != nil {
		r.Log.Info("🦀 MiniCluster Ensemble Update", "Error with Creating Client", err)
		return ctrl.Result{Requeue: true}, err
	}

	// A selector just for the lead broker pod of the ensemble MiniCluster
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{
		// job-name corresponds to the ensemble name plus index in the list
		"job-name": fmt.Sprintf("%s-0", ensemble.Name),
		// job index is the lead broker (0) within
		"job-index": "0",
	}}

	// There should only be one pod!
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}
	pods, err := clientset.CoreV1().Pods(ensemble.Namespace).List(ctx, listOptions)
	if err != nil {
		r.Log.Info("🦀 MiniCluster Ensemble Update", "Error with listing pods", err)
		return ctrl.Result{Requeue: true}, err
	}

	// Get the ip address of the lead broker pod - requeue if not ready yet
	var ipAddress string
	for _, pod := range pods.Items {
		pod, err := clientset.CoreV1().Pods(ensemble.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			r.Log.Info("🦀 MiniCluster Ensemble Update", "Error with listing pods", err)
			return ctrl.Result{Requeue: true}, err
		}
		r.Log.Info("Pod", "IP Address", pod.Status.PodIP)
		ipAddress = pod.Status.PodIP
	}

	// If we don't have an ip address yet, try again later
	if ipAddress == "" {
		r.Log.Info("🦀 MiniCluster Ensemble Update", "No pods found", err)
		return ctrl.Result{Requeue: true}, err
	}

	// Create a client to the pod (host)
	host := fmt.Sprintf("%s:%s", ipAddress, member.SidecarPort)
	r.Log.Info("🦀 MiniCluster Ensemble Update", "Host", host)
	c, err := client.NewClient(host)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// Get the queue status!
	in := pb.StatusRequest{}
	response, err := c.RequestStatus(ctx, &in)
	if err != nil {
		r.Log.Info("🦀 MiniCluster Ensemble GRPC Client", "Error with status request", err)
		return ctrl.Result{Requeue: true}, err
	}
	// Get the algorithm - if this fails we stop
	a, err := algorithm.Get(member.Algorithm.Name)
	if err != nil {
		r.Log.Info("🦀 MiniCluster Ensemble GRPC Client", "Failed to retrieve algorithm", err)
		return ctrl.Result{Requeue: true}, err
	}
	// check that the algorithm is valid for the member type
	// TODO add actual options here
	err = a.Check(algorithm.AlgorithmOptions{}, member)
	if err != nil {
		r.Log.Info("🦀 MiniCluster Ensemble GRPC Client", "Algorithm %s does not support %s", a.Name(), member.Type())
		return ctrl.Result{}, err
	}

	fmt.Println(response.Status)
	fmt.Println(response.Payload)

	// Make a decision
	decision, err := a.MakeDecision(response.Payload, member)
	if err != nil {
		r.Log.Info("🦀 MiniCluster Ensemble GRPC Client", "Decision Error", err)
		return ctrl.Result{}, err
	}

	fmt.Println(decision)

	// This is the last return, this says to check every N seconds
	return ctrl.Result{Requeue: true}, nil
}
