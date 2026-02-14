package agents

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/siddhantprateek/reefline/internal/flows/prompts"
)

// NewSupervisorAgent creates the Supervisor Agent.
// It receives raw scan data (grype / dockle / dive JSON) and produces a
// structured, evidence-cited Markdown security report (draft.md).
func NewSupervisorAgent(ctx context.Context, cm model.ToolCallingChatModel) (adk.Agent, error) {
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name: "SupervisorAgent",
		Description: `Synthesizes raw container image scan outputs (Grype, Dockle, Dive)
into a structured, evidence-backed Markdown security report. Produces draft.md with
sections covering vulnerability analysis, CIS benchmark findings, layer efficiency,
key findings, score card, and Dockerfile improvement recommendations.`,
		Instruction: prompts.SupervisorSystemPrompt,
		Model:       cm,
	})
}
