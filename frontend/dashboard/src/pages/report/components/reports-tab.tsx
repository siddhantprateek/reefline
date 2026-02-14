import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { JobReport } from "@/api/jobs.api";

interface ReportsTabProps {
  report: JobReport;
}

const SAMPLE_REPORT = `# Image Analysis Report

## Summary

This report provides a comprehensive security and optimization analysis of the container image.

---

## Vulnerability Overview

| Severity | Count |
|----------|-------|
| Critical | 3     |
| High     | 12    |
| Medium   | 28    |
| Low      | 47    |

---

## Key Findings

### Security Issues

- **Runs as root** — The container runs as the root user, which increases the blast radius of any compromise.
- **Outdated base image** — The base image is over 90 days old. Consider updating to a recent release.
- **Secrets detected** — 2 potential secrets were found in image layers.

### Optimization Opportunities

1. **Use a distroless or slim base image**
   Switching from \`ubuntu:latest\` to \`gcr.io/distroless/static\` can reduce image size by ~60%.

2. **Consolidate RUN instructions**
   Multiple \`RUN apt-get install\` commands can be merged to reduce layer count and image size.

3. **Remove build dependencies from final image**
   Use multi-stage builds to exclude compilers and build tools from the production image.

---

## Recommendations

\`\`\`dockerfile
# Before
FROM ubuntu:latest
RUN apt-get update
RUN apt-get install -y curl wget git
RUN pip install -r requirements.txt

# After (multi-stage)
FROM python:3.12-slim AS builder
RUN pip install --user -r requirements.txt

FROM python:3.12-slim
COPY --from=builder /root/.local /root/.local
\`\`\`

---

## Score

| Metric | Current | Estimated After |
|--------|---------|----------------|
| Security Score | 42 / 100 | 78 / 100 |
| Image Size | 512 MB | 190 MB |
| CVE Count | 90 | 15 |

> **Note:** Estimates are based on automated analysis and may vary depending on implementation.
`;

export function ReportsTab({ report: _ }: ReportsTabProps) {
  return (
    <div className="p-6">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: ({ children }) => (
            <h1 className="text-2xl font-bold mb-4 mt-2">{children}</h1>
          ),
          h2: ({ children }) => (
            <h2 className="text-xl font-semibold mb-3 mt-6 border-b border-border pb-1">
              {children}
            </h2>
          ),
          h3: ({ children }) => (
            <h3 className="text-base font-semibold mb-2 mt-4">{children}</h3>
          ),
          p: ({ children }) => (
            <p className="text-sm text-muted-foreground mb-3 leading-relaxed">{children}</p>
          ),
          ul: ({ children }) => (
            <ul className="list-disc list-inside mb-3 space-y-1 text-sm text-muted-foreground">
              {children}
            </ul>
          ),
          ol: ({ children }) => (
            <ol className="list-decimal list-inside mb-3 space-y-1 text-sm text-muted-foreground">
              {children}
            </ol>
          ),
          li: ({ children }) => <li className="leading-relaxed">{children}</li>,
          code: ({ children, className }) => {
            const isBlock = className?.includes("language-");
            return isBlock ? (
              <code className="block bg-muted rounded-md p-4 text-xs font-mono overflow-x-auto whitespace-pre">
                {children}
              </code>
            ) : (
              <code className="bg-muted px-1.5 py-0.5 rounded text-xs font-mono">
                {children}
              </code>
            );
          },
          pre: ({ children }) => (
            <pre className="mb-4 rounded-md overflow-hidden">{children}</pre>
          ),
          blockquote: ({ children }) => (
            <blockquote className="border-l-4 border-border pl-4 italic text-sm text-muted-foreground mb-3">
              {children}
            </blockquote>
          ),
          table: ({ children }) => (
            <div className="overflow-x-auto mb-4">
              <table className="w-full text-sm border-collapse border border-border">
                {children}
              </table>
            </div>
          ),
          thead: ({ children }) => (
            <thead className="bg-muted">{children}</thead>
          ),
          th: ({ children }) => (
            <th className="border border-border px-3 py-2 text-left font-medium text-xs">
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td className="border border-border px-3 py-2 text-xs text-muted-foreground">
              {children}
            </td>
          ),
          hr: () => <hr className="border-border my-4" />,
          strong: ({ children }) => (
            <strong className="font-semibold text-foreground">{children}</strong>
          ),
          a: ({ href, children }) => (
            <a
              href={href}
              className="text-primary underline underline-offset-2 hover:opacity-80"
              target="_blank"
              rel="noopener noreferrer"
            >
              {children}
            </a>
          ),
        }}
      >
        {SAMPLE_REPORT}
      </ReactMarkdown>
    </div>
  );
}
