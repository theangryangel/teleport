package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/lib/auth"
)

// tokenCRUD performs each token crud function as an example.
func tokenCRUD(ctx context.Context, client *auth.Client) error {
	// generate a randomly generated cluster join token for adding another proxy to a cluster.
	tokenString, err := client.GenerateToken(ctx, auth.GenerateTokenRequest{
		// You can provide 'Token' for a static token name
		Roles: teleport.Roles{teleport.RoleProxy},
		TTL:   time.Hour,
	})
	if err != nil {
		return fmt.Errorf("Failed to generate token: %v", err)
	}

	log.Printf("Generated token: %v", tokenString)

	// retrieve all active cluster join tokens
	tokens, err := client.GetTokens()
	if err != nil {
		return fmt.Errorf("Failed to get tokens: %v", err)
	}

	log.Println("Retrieved tokens:")
	for _, t := range tokens {
		log.Printf("  %v", t.GetName())
	}

	// update the token to be a proxy token
	token, err := client.GetToken(tokenString)
	if err != nil {
		return fmt.Errorf("Failed to retrieve token for update: %v", err)
	}

	token.SetRoles(teleport.Roles{teleport.RoleProxy})
	if err = client.UpsertToken(token); err != nil {
		return fmt.Errorf("Failed to update token: %v", err)
	}

	log.Println("Updated token")

	// delete the cluster tokens we just created
	if err = client.DeleteToken(tokenString); err != nil {
		return fmt.Errorf("Failed to delete token: %v", err)
	}

	log.Println("Deleted token")

	return nil
}
