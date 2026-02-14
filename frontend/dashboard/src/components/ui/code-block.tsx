import { useState, useCallback } from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import {
  oneDark,
  oneLight,
} from "react-syntax-highlighter/dist/esm/styles/prism";
import { Check, Clipboard } from "lucide-react";
import { useTheme } from "next-themes";

interface CodeBlockProps {
  language?: string;
  children: string;
}

const LANGUAGE_LABELS: Record<string, string> = {
  dockerfile: "Dockerfile",
  docker: "Dockerfile",
  bash: "Bash",
  sh: "Shell",
  shell: "Shell",
  yaml: "YAML",
  yml: "YAML",
  json: "JSON",
  javascript: "JavaScript",
  js: "JavaScript",
  typescript: "TypeScript",
  ts: "TypeScript",
  python: "Python",
  py: "Python",
  go: "Go",
  rust: "Rust",
  sql: "SQL",
  toml: "TOML",
  nginx: "Nginx",
  plaintext: "Text",
};

export function CodeBlock({ language = "plaintext", children }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);
  const { resolvedTheme } = useTheme();
  const isDark = resolvedTheme === "dark";

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(children).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }, [children]);

  const label = LANGUAGE_LABELS[language.toLowerCase()] ?? language;

  return (
    <div className="group relative rounded-lg border border-border overflow-hidden mb-4">
      {/* Header bar */}
      <div className="flex items-center justify-between bg-muted/60 px-3 py-1.5 border-b border-border">
        <span className="text-[11px] font-medium tracking-wide text-muted-foreground uppercase">
          {label}
        </span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground transition-colors"
          title="Copy code"
        >
          {copied ? (
            <>
              <Check className="h-3 w-3" />
              Copied
            </>
          ) : (
            <>
              <Clipboard className="h-3 w-3" />
              Copy
            </>
          )}
        </button>
      </div>

      {/* Code area */}
      <div className="bg-muted/40">
        <SyntaxHighlighter
          language={language}
          style={isDark ? oneDark : oneLight}
          customStyle={{
            margin: 0,
            padding: "1rem",
            fontSize: "0.75rem",
            lineHeight: "1.6",
            background: "none",
            backgroundColor: "transparent",
            border: "none",
          }}
          codeTagProps={{
            style: {
              fontFamily:
                'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
            },
          }}
          showLineNumbers={children.split("\n").length > 3}
          lineNumberStyle={{
            minWidth: "2em",
            paddingRight: "1em",
            color: isDark ? "rgba(255,255,255,0.25)" : "rgba(0,0,0,0.2)",
            userSelect: "none",
          }}
          wrapLongLines
        >
          {children}
        </SyntaxHighlighter>
      </div>
    </div>
  );
}
