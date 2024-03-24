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
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
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

	// The MiniCluster service is being provided by the index 0 pod, so we can find it here.
	clientset, err := kubernetes.NewForConfig(r.RESTConfig)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
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
		return ctrl.Result{Requeue: true}, err
	}

	// Get the ip address of the lead broker pod - requeue if not ready yet
	var ipAddress string
	for _, pod := range pods.Items {
		pod, err := clientset.CoreV1().Pods(ensemble.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("      Error with getting pods %s\n", err)
			return ctrl.Result{Requeue: true}, err
		}
		fmt.Printf("      Pod IP Address %s\n", pod.Status.PodIP)
		ipAddress = pod.Status.PodIP
	}

	// If we don't have an ip address yet, try again later
	if ipAddress == "" {
		fmt.Println("      No pods found")
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
	in := pb.StatusRequest{Member: member.Type()}
	response, err := c.RequestStatus(ctx, &in)
	if err != nil {
		fmt.Printf("      Error with status request %s\n", err)
		return ctrl.Result{Requeue: true}, err
	}
	fmt.Println(response.Status)

	// Make a decision based on the queue (and changing jobs matrix)
	idx := fmt.Sprintf("%d", i)
	jobs := ensemble.Status.Jobs[idx]
	decision, err := algo.MakeDecision(response.Payload, jobs)
	if err != nil {
		fmt.Printf("      Decision error %s\n", err)
		return ctrl.Result{}, err
	}

	// If we are requesting an action to the queue (sidecar gRPC) do it
	// This second request should be OK because I think it will be infrequent.
	// Most algorithms should do submission in bulk (infrequently) and then monitor
	if decision.Action != "" {
		in := pb.ActionRequest{
			Member:    member.Type(),
			Algorithm: algo.Name(),
			Payload:   decision.Payload,
			Action:    decision.Action,
		}
		response, err := c.RequestAction(ctx, &in)
		if err != nil {
			fmt.Printf("      Error with action request %s\n", err)
			return ctrl.Result{Requeue: true}, err
		}
		fmt.Println(response.Status)
	}

	// Since we requeue anyway, we don't check error. But probably should.
	if decision.Updated {
		return r.updateJobsMatrix(ctx, ensemble, decision.Jobs, i)
	}
	// This is the last return, this says to check every N seconds
	return ctrl.Result{Requeue: true}, nil
}

// updateJobsMatrix to include jobs
func (r *EnsembleReconciler) updateJobsMatrix(
	ctx context.Context,
	ensemble *api.Ensemble,
	jobs []api.Job,
	i int,
) (ctrl.Result, error) {
	patch := kclient.MergeFrom(ensemble.DeepCopy())

	idx := fmt.Sprintf("%d", i)

	ensemble.Status.Jobs[idx] = jobs
	// Should probably check error here
	err := r.Status().Update(ctx, ensemble)
	if err != nil {
		return ctrl.Result{}, nil
	}
	err = r.Patch(ctx, ensemble, patch)
	if err != nil {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{Requeue: true}, err
}
