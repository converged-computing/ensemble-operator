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
	RequestStatus(ctx context.Context, in *pb.StatusRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error)
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

// SubmitJob submits a job to a named cluster.
// The token specific to the cluster is required
func (c *EnsembleClient) RequestStatus(ctx context.Context, in *pb.StatusRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error) {

	response := &pb.StatusResponse{}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}

	// Now contact the rainbow server with clusters...
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// Validate that the cluster exists, and we have the right token.
	// The response is the same either way - not found does not reveal
	// additional information to the client trying to find it
	response, err := c.service.RequestStatus(ctx, &pb.StatusRequest{})
	fmt.Println(response)
	return response, err
}
