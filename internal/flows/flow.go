package flows

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	einoopenai "github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/minio/minio-go/v7"
	"github.com/siddhantprateek/reefline/internal/flows/agents"
	"github.com/siddhantprateek/reefline/pkg/storage"
)

// retryTransport retries on HTTP 429 with exponential backoff, honouring Retry-After.
type retryTransport struct {
	base     http.RoundTripper
	maxRetry int
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Buffer the body once so we can replay it on every retry.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
	}

	var (
		resp *http.Response
		err  error
	)
	for attempt := 0; attempt <= t.maxRetry; attempt++ {
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err = t.base.RoundTrip(req)
		if err != nil || resp.StatusCode != http.StatusTooManyRequests {
			return resp, err
		}

		wait := t.backoff(attempt, resp)
		resp.Body.Close()
		log.Printf("[Flow] 429 rate-limited — waiting %s before retry %d/%d", wait, attempt+1, t.maxRetry)
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(wait):
		}
	}
	return resp, err
}

func (t *retryTransport) backoff(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.ParseFloat(ra, 64); err == nil {
				return time.Duration(secs*1000)*time.Millisecond + 500*time.Millisecond
			}
		}
		if rrt := resp.Header.Get("x-ratelimit-reset-tokens"); rrt != "" {
			if d, err := time.ParseDuration(rrt); err == nil {
				return d + 500*time.Millisecond
			}
		}
	}
	// Exponential backoff: 5s → 10s → 20s → 40s → 60s (cap)
	secs := math.Min(float64(int(1)<<uint(attempt))*5, 60)
	return time.Duration(secs) * time.Second
}

func newRetryHTTPClient() *http.Client {
	return &http.Client{
		Transport: &retryTransport{base: http.DefaultTransport, maxRetry: 5},
	}
}

const (
	nodeSupervisor = "supervisor"
	nodeCritique   = "critique"
	nodePublish    = "publish_report"
)

// flowState is shared across all graph nodes for a single run.
type flowState struct {
	Revision         int    // how many revisions have happened
	CritiqueFeedback string // critique's REVISE message, forwarded to supervisor
}

// RunFlow builds a graph: Supervisor → Critique → branch(APPROVE→publish→END | REVISE→Supervisor)
//
//	Graph:
//	  START → supervisor → critique → [APPROVE] → publish_report → END
//	                          ↑                                       |
//	                          └──────────── [REVISE] ←───────────────┘
//	                                      (max 3 revisions)
func RunFlow(ctx context.Context, jobID, bucket string) error {
	creds, err := resolveCredentials(jobID)
	if err != nil {
		return fmt.Errorf("resolving credentials: %w", err)
	}

	p := Provider(creds.ProviderID)

	modelID := creds.ModelID
	if modelID == "" {
		info, ok := defaultModels[p]
		if !ok {
			return fmt.Errorf("unknown provider %q", creds.ProviderID)
		}
		modelID = info.ID
	}

	baseURL, ok := providerBaseURLs[p]
	if !ok {
		return fmt.Errorf("unknown provider %q", creds.ProviderID)
	}

	log.Printf("[Flow] provider=%s model=%s job=%s", creds.ProviderID, modelID, jobID)

	cm, err := einoopenai.NewClient(ctx, &einoopenai.Config{
		APIKey:     creds.APIKey,
		BaseURL:    string(baseURL),
		Model:      modelID,
		HTTPClient: newRetryHTTPClient(),
	})
	if err != nil {
		return fmt.Errorf("building chat model: %w", err)
	}

	// Build MinIO tools
	listTool, err := NewListScanFilesTool(bucket)
	if err != nil {
		return fmt.Errorf("list_scan_files tool: %w", err)
	}
	readTool, err := NewReadScanFileTool(bucket)
	if err != nil {
		return fmt.Errorf("read_scan_file tool: %w", err)
	}
	writeTool, err := NewWriteDraftTool(bucket)
	if err != nil {
		return fmt.Errorf("write_draft tool: %w", err)
	}

	supervisorTools := []tool.BaseTool{listTool, readTool, writeTool}

	// Build agents
	supervisor, err := agents.NewSupervisorAgent(ctx, cm, supervisorTools, jobID)
	if err != nil {
		return fmt.Errorf("creating supervisor agent: %w", err)
	}
	// Critique has no tools — draft content is passed directly in the message
	critique, err := agents.NewCritiqueAgent(ctx, cm, jobID)
	if err != nil {
		return fmt.Errorf("creating critique agent: %w", err)
	}

	// supervisorLambda: on first run uses the initial prompt; on REVISE runs injects critique feedback.
	supervisorLambda := compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
		var feedback string
		_ = compose.ProcessState(ctx, func(_ context.Context, s *flowState) error {
			feedback = s.CritiqueFeedback
			return nil
		})

		var trigger *schema.Message
		if feedback != "" {
			trigger = schema.UserMessage(fmt.Sprintf(
				"REVISE the report for job_id=%q based on critique feedback.\n\n"+
					"Read report.md, apply all corrections, then call write_draft to save the updated report.\n\n"+
					"Critique feedback:\n%s",
				jobID, feedback,
			))
		} else {
			trigger = schema.UserMessage(fmt.Sprintf(
				"Generate a security report for job_id=%q. "+
					"List artifacts, read grype.json, dockle.json and dive.json, "+
					"then call write_draft to save the complete report.",
				jobID,
			))
		}
		input := &adk.AgentInput{Messages: []*schema.Message{trigger}}
		iter := supervisor.Run(ctx, input)
		return drainAgent(iter, "SupervisorAgent")
	})

	// critiqueLambda: reads report.md directly and passes it in the message — no tool calls needed.
	critiqueLambda := compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
		draft, err := readMinIOFile(ctx, bucket, fmt.Sprintf("%s/artifacts/report.md", jobID))
		if err != nil {
			return nil, fmt.Errorf("reading report.md for critique: %w", err)
		}
		trigger := schema.UserMessage(fmt.Sprintf("Review this security report and return APPROVE or REVISE:\n\n%s", draft))
		input := &adk.AgentInput{Messages: []*schema.Message{trigger}}
		iter := critique.Run(ctx, input)
		return drainAgent(iter, "CritiqueAgent")
	})

	// publish_report: report.md was written directly by the supervisor — just confirm it exists.
	publishLambda := compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) (*schema.Message, error) {
		content, err := readMinIOFile(ctx, bucket, fmt.Sprintf("%s/artifacts/report.md", jobID))
		if err != nil {
			return nil, fmt.Errorf("report.md not found after supervisor: %w", err)
		}
		log.Printf("[Flow] report.md confirmed for job=%s (%d bytes)", jobID, len(content))
		return schema.AssistantMessage("report.md confirmed", nil), nil
	})

	// Build graph: []*schema.Message → *schema.Message
	g := compose.NewGraph[[]*schema.Message, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *flowState {
			return &flowState{}
		}),
	)

	if err := g.AddLambdaNode(nodeSupervisor, supervisorLambda); err != nil {
		return fmt.Errorf("adding supervisor node: %w", err)
	}
	if err := g.AddLambdaNode(nodeCritique, critiqueLambda); err != nil {
		return fmt.Errorf("adding critique node: %w", err)
	}
	if err := g.AddLambdaNode(nodePublish, publishLambda); err != nil {
		return fmt.Errorf("adding publish node: %w", err)
	}

	// Edges: START → supervisor → critique
	if err := g.AddEdge(compose.START, nodeSupervisor); err != nil {
		return fmt.Errorf("edge START→supervisor: %w", err)
	}
	if err := g.AddEdge(nodeSupervisor, nodeCritique); err != nil {
		return fmt.Errorf("edge supervisor→critique: %w", err)
	}

	// Branch after critique: APPROVE → publish | REVISE → supervisor (up to 3 revisions).
	// We also inject the critique feedback into shared state so supervisor can read it.
	branch := compose.NewGraphBranch(
		func(ctx context.Context, msgs []*schema.Message) (string, error) {
			var revision int
			if err := compose.ProcessState(ctx, func(_ context.Context, s *flowState) error {
				revision = s.Revision
				return nil
			}); err != nil {
				return nodePublish, nil
			}

			// Find the last assistant message from critique
			for i := len(msgs) - 1; i >= 0; i-- {
				if msgs[i].Role == schema.Assistant && msgs[i].Content != "" {
					verdict := msgs[i].Content
					if strings.Contains(verdict, "APPROVE") || revision >= 3 {
						log.Printf("[Flow] Critique verdict=APPROVE (revision=%d) job=%s", revision, jobID)
						return nodePublish, nil
					}
					log.Printf("[Flow] Critique verdict=REVISE (revision=%d) job=%s", revision, jobID)
					// Store critique feedback in shared state so supervisor lambda can read it
					_ = compose.ProcessState(ctx, func(_ context.Context, s *flowState) error {
						s.Revision++
						s.CritiqueFeedback = verdict
						return nil
					})
					return nodeSupervisor, nil
				}
			}
			return nodePublish, nil
		},
		map[string]bool{
			nodeSupervisor: true,
			nodePublish:    true,
		},
	)
	if err := g.AddBranch(nodeCritique, branch); err != nil {
		return fmt.Errorf("adding critique branch: %w", err)
	}

	// publish → END
	if err := g.AddEdge(nodePublish, compose.END); err != nil {
		return fmt.Errorf("edge publish→END: %w", err)
	}

	// Compile and run
	runnable, err := g.Compile(ctx)
	if err != nil {
		return fmt.Errorf("compiling graph: %w", err)
	}

	initialPrompt := []*schema.Message{
		schema.UserMessage(fmt.Sprintf(
			"Generate a security report for job_id=%q. "+
				"List available artifacts, read grype.json, dockle.json and dive.json, "+
				"then write a complete report.md.",
			jobID,
		)),
	}

	result, err := runnable.Invoke(ctx, initialPrompt)
	if err != nil {
		return fmt.Errorf("running flow: %w", err)
	}

	if result != nil {
		log.Printf("[Flow] done job=%s result=%s", jobID, truncate(result.Content, 80))
	}
	return nil
}

// drainAgent consumes an adk.AsyncIterator, logs each message, and returns all messages seen.
func drainAgent(iter *adk.AsyncIterator[*adk.AgentEvent], name string) ([]*schema.Message, error) {
	var msgs []*schema.Message
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return msgs, fmt.Errorf("[%s] agent error: %w", name, event.Err)
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if msg := event.Output.MessageOutput.Message; msg != nil {
				if msg.Content != "" {
					log.Printf("[Flow][%s] %s", name, truncate(msg.Content, 120))
				}
				msgs = append(msgs, msg)
			}
		}
	}
	return msgs, nil
}

// lastUserMessage returns the last user-role message from a slice, or nil.
func lastUserMessage(msgs []*schema.Message) *schema.Message {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == schema.User {
			return msgs[i]
		}
	}
	return nil
}

// readMinIOFile reads the full content of an object from MinIO.
func readMinIOFile(ctx context.Context, bucket, objectName string) (string, error) {
	obj, err := storage.Client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}
	defer obj.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(obj); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
