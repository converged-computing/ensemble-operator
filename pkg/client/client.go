package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/converged-computing/ensemble-operator/protos"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// EnsembleClient interacts with client endpoints
type EnsembleClient struct {
	host       string
	connection *grpc.ClientConn
	service    pb.EnsembleOperatorClient
}

var _ Client = (*EnsembleClient)(nil)

// Client interface defines functions required for a valid client
type Client interface {

	// Ensemble interactions
	RequestStatus(ctx context.Context, in *pb.StatusRequest, opts ...grpc.CallOption) (*pb.Response, error)
	RequestAction(ctx context.Context, in *pb.ActionRequest, opts ...grpc.CallOption) (*pb.Response, error)
}

// NewClient creates a new EnsembleClient
func NewClient(host string) (Client, error) {
	if host == "" {
		return nil, errors.New("host is required")
	}

	log.Printf("🥞️ starting client (%s)...", host)
	c := &EnsembleClient{host: host}

	// Set up a connection to the server.
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(c.GetHost(), creds, grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to %s", host)
	}

	c.connection = conn
	c.service = pb.NewEnsembleOperatorClient(conn)

	return c, nil
}

// Close closes the created resources (e.g. connection).
func (c *EnsembleClient) Close() error {
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}

// Connected returns  true if we are connected and the connection is ready
func (c *EnsembleClient) Connected() bool {
	return c.service != nil && c.connection != nil && c.connection.GetState() == connectivity.Ready
}

// GetHost returns the private hostn name
func (c *EnsembleClient) GetHost() string {
	return c.host
}

// RequestStatus gets the queue and jobs status.
// This is primarily for scaling/termination
func (c *EnsembleClient) RequestStatus(
	ctx context.Context,
	in *pb.StatusRequest,
	opts ...grpc.CallOption,
) (*pb.Response, error) {

	response := &pb.Response{}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	response, err := c.service.RequestStatus(ctx, in)
	fmt.Println(response)
	return response, err
}

func (c *EnsembleClient) RequestAction(
	ctx context.Context,
	in *pb.ActionRequest,
	opts ...grpc.CallOption,
) (*pb.Response, error) {

	response := &pb.Response{}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	response, err := c.service.RequestAction(ctx, in)
	fmt.Println(response)
	return response, err
}
