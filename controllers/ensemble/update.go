package controller

import (
	"context"
	"fmt"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	"github.com/converged-computing/ensemble-operator/pkg/algorithm"
	"github.com/converged-computing/ensemble-operator/pkg/client"
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
	i int,
) (ctrl.Result, error) {

	fmt.Println("🦀 MiniCluster Ensemble Update")

	// Get the ip address of our pod
	ipAddress, err := r.getLeaderAddress(ctx, ensemble, name)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// Create a client to the pod (host)
	host := fmt.Sprintf("%s:%s", ipAddress, member.Sidecar.Port)
	fmt.Printf("      Host %s\n", host)

	c, err := client.NewClient(host)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// Get the algorithm - if this fails we stop
	algo, err := algorithm.Get(member.Algorithm.Name)
	if err != nil {
		fmt.Printf("      Failed to retrieve algorithm %s\n", err)
		return ctrl.Result{Requeue: true}, err
	}
	// check that the algorithm is valid for the member type
	err = algo.Check(algorithm.AlgorithmOptions{}, member)
	if err != nil {
		fmt.Printf("      Algorithm %s does not support %s", algo.Name(), member.Type())
		return ctrl.Result{}, err
	}

	// The status request comes first to peek at the queue
	// TODO add secret here, maybe don't need Name
	in := pb.StatusRequest{Member: member.Type(), Algorithm: algo.Name()}
	response, err := c.RequestStatus(ctx, &in)
	if err != nil || response.Status == pb.Response_ERROR {
		fmt.Printf("      Error with status request %s\n", err)
		return ctrl.Result{Requeue: true}, err
	}
	fmt.Println(response.Status)

	// Make a decision based on the queue (and changing jobs matrix)
	idx := fmt.Sprintf("%d", i)
	jobs := ensemble.Status.Jobs[idx]
	decision, err := algo.MakeDecision(ensemble, member, response.Payload, jobs)
	if err != nil || response.Status == pb.Response_ERROR {
		fmt.Printf("      Decision error %s\n", err)
		return ctrl.Result{}, err
	}

	// If we are requesting an action to the queue (sidecar gRPC) do it
	// This second request should be OK because I think it will be infrequent.
	// Most algorithms should do submission in bulk (infrequently) and then monitor
	if decision.Action == algorithm.JobsMatrixUpdateAction {
		in := pb.ActionRequest{
			Member:    member.Type(),
			Algorithm: algo.Name(),
			Payload:   decision.Payload,
			Action:    algorithm.SubmitAction,
		}
		response, err := c.RequestAction(ctx, &in)
		fmt.Println(response.Status)
		if err != nil {
			fmt.Printf("      Error with action request %s\n", err)
			return ctrl.Result{}, err
		}
		// Since we requeue anyway, we don't check error. But probably should.
		return r.updateJobsMatrix(ctx, ensemble, decision.Jobs, i)
	}

	// Are we done? If we might have terminated by the user indicated
	// not to, just reconcile for a last time
	if decision.Action == algorithm.CompleteAction {
		return ctrl.Result{}, err
	}

	// Are we terminating? Note that the next check for updated
	// cannot happen at the same time as a termination request
	if decision.Action == algorithm.TerminateAction {
		err = r.Delete(ctx, ensemble)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, err
	}

	// Are we scaling?
	if decision.Action == algorithm.ScaleAction && decision.Scale > 0 {
		if member.Type() == api.MiniclusterType {
			return r.updateMiniClusterSize(ctx, ensemble, decision.Scale, name, idx)
		}
	}

	// This is the last return, this says to check every N seconds
	return ctrl.Result{RequeueAfter: ensemble.RequeueAfter()}, nil
}

// getLeaderAddress gets the ipAddress of the lead broker
// In all cases of error we requeue
func (r *EnsembleReconciler) getLeaderAddress(
	ctx context.Context,
	ensemble *api.Ensemble,
	name string,
) (string, error) {

	// The MiniCluster service is being provided by the index 0 pod, so we can find it here.
	clientset, err := kubernetes.NewForConfig(r.RESTConfig)
	if err != nil {
		return "", err
	}

	// A selector just for the lead broker pod of the ensemble MiniCluster
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{
		// job-name corresponds to the ensemble name plus index in the list
		"job-name": name,
		// job index is the lead broker (0) within
		"job-index": "0",
	}}

	// There should only be one pod!
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}
	pods, err := clientset.CoreV1().Pods(ensemble.Namespace).List(ctx, listOptions)
	if err != nil {
		fmt.Printf("      Error with listing pods %s\n", err)
		return "", err
	}

	// Get the ip address of the lead broker pod - requeue if not ready yet
	var ipAddress string
	for _, pod := range pods.Items {
		pod, err := clientset.CoreV1().Pods(ensemble.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("      Error with getting pods %s\n", err)
			return "", err
		}
		fmt.Printf("      Pod IP Address %s\n", pod.Status.PodIP)
		ipAddress = pod.Status.PodIP
	}

	// If we don't have an ip address yet, try again later
	if ipAddress == "" {
		fmt.Println("      No pods found")
		return "", fmt.Errorf("no pods found, not ready yet")
	}
	return ipAddress, nil
}
