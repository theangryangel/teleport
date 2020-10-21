package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/lib/services"

	"github.com/gravitational/teleport/lib/auth"
)

// clusterLabelCRUD performs each cluster label crud function as an example.
func clusterLabelCRUD(ctx context.Context, client *auth.Client) error {
	// Create a cluster join token with labels. remote clusters added with
	// this token will inherit the token's labels.
	tokenString, err := client.GenerateToken(ctx, auth.GenerateTokenRequest{
		// You can provide 'Token' for a static token name
		Roles: teleport.Roles{teleport.RoleTrustedCluster},
		TTL:   time.Hour,
		Labels: map[string]string{
			"env": "staging",
		},
	})
	if err != nil {
		return fmt.Errorf("Failed to generate rc token: %v", err)
	}

	log.Printf("Generated token: %v", tokenString)

	// update the remote cluster with new cluster labels
	rc, err := services.NewRemoteCluster("remote")
	if err != nil {
		return fmt.Errorf("Failed to make new remote cluster: %v", err)
	}

	md := rc.GetMetadata()
	md.Labels = map[string]string{"env": "prod"}
	rc.SetMetadata(md)

	if err = client.UpdateRemoteCluster(ctx, rc); err != nil {
		return fmt.Errorf("Failed to update remote cluster: %v", err)
	}

	log.Println("Updated remote cluster")

	return nil
}
