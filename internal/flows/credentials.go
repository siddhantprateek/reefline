package flows

import (
	"encoding/json"
	"fmt"

	"github.com/siddhantprateek/reefline/pkg/crypto"
	"github.com/siddhantprateek/reefline/pkg/database"
	"github.com/siddhantprateek/reefline/pkg/models"
)

// aiProviderPriority is the order in which we pick a connected AI integration.
var aiProviderPriority = []string{"openai", "anthropic", "google", "openrouter"}

// resolvedCredentials holds everything needed to build the chat model.
type resolvedCredentials struct {
	ProviderID string
	APIKey     string
	ModelID    string // may be empty — RunFlow falls back to the provider default
}

// resolveCredentials looks up the job owner's connected AI integration and
// returns decrypted credentials. This keeps all DB + crypto logic out of the handler.
func resolveCredentials(jobID string) (*resolvedCredentials, error) {
	// 1. Find the job to get user_id
	var job models.Job
	if err := database.DB.Where("job_id = ?", jobID).First(&job).Error; err != nil {
		return nil, fmt.Errorf("fetching job %s: %w", jobID, err)
	}
	if job.UserID == "" {
		return nil, fmt.Errorf("job %s has no user_id", jobID)
	}

	// 2. Find the first connected AI provider for this user
	var integration models.Integration
	found := false
	for _, providerID := range aiProviderPriority {
		err := database.DB.
			Where("user_id = ? AND integration_id = ? AND status = ?", job.UserID, providerID, "connected").
			First(&integration).Error
		if err == nil {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("no connected AI provider found for user %s", job.UserID)
	}

	// 3. Decrypt credentials → {"apiKey": "...", "model": "..."}
	raw, err := crypto.Decrypt(integration.Credentials)
	if err != nil {
		return nil, fmt.Errorf("decrypting credentials for %s: %w", integration.IntegrationID, err)
	}
	var creds map[string]string
	if err := json.Unmarshal(raw, &creds); err != nil {
		return nil, fmt.Errorf("parsing credentials for %s: %w", integration.IntegrationID, err)
	}

	apiKey := creds["apiKey"]
	if apiKey == "" {
		return nil, fmt.Errorf("no apiKey in credentials for %s", integration.IntegrationID)
	}

	return &resolvedCredentials{
		ProviderID: integration.IntegrationID,
		APIKey:     apiKey,
		ModelID:    creds["model"],
	}, nil
}
