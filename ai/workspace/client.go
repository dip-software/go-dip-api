package workspace

import (
	"github.com/dip-software/go-dip-api/ai"
	"github.com/dip-software/go-dip-api/iam"
	"github.com/go-playground/validator/v10"
)

// A Client manages communication with HSDP AI-Workspace API
type Client struct {
	*ai.Client
	Workspace *Service
}

// NewClient returns a new HSDP AI-Workspace API client. A configured IAM client
// must be provided as the underlying API requires an IAM token
func NewClient(iamClient *iam.Client, config *ai.Config) (*Client, error) {
	config.Service = "interactivesession"
	client, err := ai.NewClient(iamClient, config)
	if err != nil {
		return nil, err
	}
	workspaceClient := &Client{
		Client:    client,
		Workspace: &Service{client: client, validate: validator.New()},
	}
	return workspaceClient, nil
}
