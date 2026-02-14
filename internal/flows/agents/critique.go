package agents

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// NewCritiqueAgent creates the Critique Agent.
// The draft report is injected directly into the user message — no tools needed.
func NewCritiqueAgent(ctx context.Context, cm model.ToolCallingChatModel, jobID string) (adk.Agent, error) {
	instruction := fmt.Sprintf(`You are the Critique Agent for Reefline. Job ID: %s

Review the security report provided in the user message. Be brief.

## APPROVE if all of these are true:
- All 8 sections present: Overview, Summary, Vulnerability Analysis, CIS Benchmark Findings, Layer Efficiency Analysis, Key Findings & Risk Assessment, Score Card, Recommended Dockerfile Improvements
- Score Card filled in with real numbers (not placeholders)
- No empty tables or "N/A" without explanation

## REVISE if any section is missing or Score Card has placeholder values.

## Output (keep it short):
**Verdict:** APPROVE or REVISE
**Issues:** bullet list of specific problems (omit if none)
**Fix:** numbered list of exact corrections for Supervisor (omit if APPROVE)

Do NOT rewrite the report. Be concise — 10 lines max.`, jobID)

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "CritiqueAgent",
		Description: "Reviews the Supervisor's draft report for structural completeness, scoring, and citation quality. Issues APPROVE or REVISE.",
		Instruction: instruction,
		Model:       cm,
	})
}
