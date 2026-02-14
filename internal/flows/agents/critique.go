package agents

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/siddhantprateek/reefline/internal/flows/prompts"
)

// NewCritiqueAgent creates the Critique Agent.
// It reviews a draft report against the raw scan data and issues an
// APPROVE or REVISE verdict with specific correction instructions.
func NewCritiqueAgent(ctx context.Context, cm model.ToolCallingChatModel) (adk.Agent, error) {
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name: "CritiqueAgent",
		Description: `Reviews the Supervisor Agent's draft.md report against raw scan data.
Cross-references all CVE IDs, dockle codes, scores, and citations for accuracy.
Issues APPROVE if the report is factually complete, or REVISE with a numbered
list of specific corrections for the Supervisor to apply.`,
		Instruction: prompts.CritiqueSystemPrompt,
		Model:       cm,
	})
}
