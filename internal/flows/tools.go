package flows

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/minio/minio-go/v7"
	"github.com/siddhantprateek/reefline/pkg/storage"
)

// ─── Tool input types ─────────────────────────────────────────────────────────

const readMaxBytes = 40000 // ~40 KB per chunk to stay within context limits

type readScanFileArgs struct {
	JobID    string `json:"job_id"    jsonschema:"description=The job ID whose scan artifact to read"`
	Filename string `json:"filename"  jsonschema:"description=Artifact to read: grype.json | dockle.json | dive.json | draft.md | report.md"`
	Offset   int    `json:"offset"    jsonschema:"description=Byte offset to start reading from (0 for the beginning). Use this to paginate large files — if the response contains TRUNCATED, call again with the returned next_offset value."`
}

type listScanFilesArgs struct {
	JobID string `json:"job_id" jsonschema:"description=The job ID whose artifacts to list"`
}

type writeDraftArgs struct {
	JobID   string `json:"job_id"  jsonschema:"description=The job ID under which report.md will be written"`
	Content string `json:"content" jsonschema:"description=Full Markdown content to write into report.md"`
}

// ─── Tool constructors ────────────────────────────────────────────────────────

// NewReadScanFileTool reads a specific scan artifact from MinIO for the given job.
// Object path pattern: {job_id}/artifacts/{filename}
func NewReadScanFileTool(bucket string) (tool.BaseTool, error) {
	return utils.InferTool(
		"read_scan_file",
		"Read a scan artifact file (grype.json, dockle.json, dive.json, draft.md, or report.md) from object storage for the given job.",
		func(ctx context.Context, args readScanFileArgs) (string, error) {
			allowed := map[string]bool{
				"grype.json":  true,
				"dockle.json": true,
				"dive.json":   true,
				"draft.md":    true,
				"report.md":   true,
			}
			if !allowed[args.Filename] {
				return "", fmt.Errorf("filename %q not allowed; choose: grype.json, dockle.json, dive.json, draft.md, report.md", args.Filename)
			}

			objectName := fmt.Sprintf("%s/artifacts/%s", args.JobID, args.Filename)

			obj, err := storage.Client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
			if err != nil {
				return "", fmt.Errorf("getting object %s: %w", objectName, err)
			}
			defer obj.Close()

			// Skip to offset in Go (avoids HTTP Range header issues with some MinIO versions)
			if args.Offset > 0 {
				if _, err := io.CopyN(io.Discard, obj, int64(args.Offset)); err != nil {
					if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "does not exist") {
						return fmt.Sprintf("artifact %q not found for job %q", args.Filename, args.JobID), nil
					}
					return "", fmt.Errorf("seeking to offset %d in %s: %w", args.Offset, objectName, err)
				}
			}

			buf := make([]byte, readMaxBytes+1)
			n, err := io.ReadFull(obj, buf)
			if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
				if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "does not exist") {
					return fmt.Sprintf("artifact %q not found for job %q", args.Filename, args.JobID), nil
				}
				return "", fmt.Errorf("reading object %s: %w", objectName, err)
			}

			chunk := buf[:n]
			if len(chunk) > readMaxBytes {
				nextOffset := args.Offset + readMaxBytes
				return fmt.Sprintf("%s\n\n[TRUNCATED — file continues. Call read_scan_file again with offset=%d to get the next chunk.]", string(chunk[:readMaxBytes]), nextOffset), nil
			}
			return string(chunk), nil
		},
	)
}

// NewListScanFilesTool lists available scan artifacts in MinIO for the given job.
func NewListScanFilesTool(bucket string) (tool.BaseTool, error) {
	return utils.InferTool(
		"list_scan_files",
		"List the scan artifact files available in object storage for the given job ID.",
		func(ctx context.Context, args listScanFilesArgs) (string, error) {
			prefix := fmt.Sprintf("%s/artifacts/", args.JobID)
			objects, err := storage.ListFiles(ctx, bucket, prefix)
			if err != nil {
				return "", fmt.Errorf("listing artifacts for job %q: %w", args.JobID, err)
			}

			if len(objects) == 0 {
				return fmt.Sprintf("no artifacts found for job %q", args.JobID), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Artifacts for job %q:\n", args.JobID))
			for _, obj := range objects {
				// Strip the prefix so the LLM sees just the filename
				name := strings.TrimPrefix(obj.Key, prefix)
				sb.WriteString(fmt.Sprintf("  - %s (%d bytes)\n", name, obj.Size))
			}
			return sb.String(), nil
		},
	)
}

// NewWriteDraftTool writes (or overwrites) report.md in MinIO for the given job.
func NewWriteDraftTool(bucket string) (tool.BaseTool, error) {
	return utils.InferTool(
		"write_draft",
		"Write or overwrite report.md in object storage for the given job with the provided Markdown content.",
		func(ctx context.Context, args writeDraftArgs) (string, error) {
			objectName := fmt.Sprintf("%s/artifacts/report.md", args.JobID)
			reader := strings.NewReader(args.Content)
			_, err := storage.Client.PutObject(ctx, bucket, objectName, reader, int64(len(args.Content)), minio.PutObjectOptions{
				ContentType: "text/markdown",
			})
			if err != nil {
				return "", fmt.Errorf("writing report.md for job %q: %w", args.JobID, err)
			}
			return fmt.Sprintf("report.md written for job %q (%d bytes)", args.JobID, len(args.Content)), nil
		},
	)
}
