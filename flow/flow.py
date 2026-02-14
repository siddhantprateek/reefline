import logging
from typing import Literal
from agents import Agent, Runner, function_tool, handoff
from agents.models.openai_chatcompletions import OpenAIChatCompletionsModel

from provider import ProviderConfig
from integration.minio import read_artifact, write_artifact

log = logging.getLogger(__name__)

ALLOWED_READ  = {"grype.json", "dockle.json", "dive.json", "draft.md", "report.md"}
ALLOWED_WRITE = {"report.md", "draft.md"}


# ── Tool factories (per job_id) ───────────────────────────────────────────────

def make_scan_tools(job_id: str):
    """Return read_file and write_file tools bound to the given job_id."""

    @function_tool
    def read_file(filename: Literal["grype.json", "dockle.json", "dive.json", "draft.md", "report.md"]) -> str:
        """Read a scan artifact or report file for the current job.
        Use 'grype.json' for vulnerability data, 'dockle.json' for CIS benchmark,
        'dive.json' for layer efficiency, 'draft.md' or 'report.md' to re-read a report.
        """
        if filename not in ALLOWED_READ:
            return f"Error: '{filename}' not allowed. Choose from: {', '.join(sorted(ALLOWED_READ))}"
        try:
            return read_artifact(job_id, filename).decode()
        except Exception as e:
            return f"{filename} not available: {e}"

    @function_tool
    def write_file(filename: Literal["report.md", "draft.md"], content: str) -> str:
        """Write report content to report.md or draft.md for the current job."""
        if filename not in ALLOWED_WRITE:
            return f"Error: '{filename}' not allowed. Choose from: {', '.join(sorted(ALLOWED_WRITE))}"
        try:
            write_artifact(job_id, filename, content.encode(), "text/markdown")
            return f"OK: {filename} saved ({len(content)} bytes)"
        except Exception as e:
            return f"Error writing {filename}: {e}"

    return [read_file, write_file]


# ── Flow ──────────────────────────────────────────────────────────────────────

async def run_flow(job_id: str, cfg: ProviderConfig) -> str:
    """
    Supervisor writes draft.md → hands off to Critique → Critique hands off back
    to Supervisor (REVISE) or finishes (APPROVE). Max 3 revisions.
    """
    client = cfg.openai_client()
    model = OpenAIChatCompletionsModel(model=cfg.model_id, openai_client=client)

    scan_tools = make_scan_tools(job_id)

    # Forward-declare so agents can reference each other
    supervisor: Agent
    critique: Agent

    supervisor = Agent(
        name="SupervisorAgent",
        handoff_description="Writes the security report draft based on scan data.",
        instructions="""You are a container image security analyst writing a professional report.

## WORKFLOW:
1. Call read_file(filename="grype.json") for vulnerability data.
2. Call read_file(filename="dockle.json") for CIS benchmark data.
3. Call read_file(filename="dive.json") for layer efficiency data.
4. If you received critique feedback, call read_file(filename="draft.md") to read the previous draft.
5. Write your complete Markdown report using write_file(filename="draft.md", content=...).
6. Hand off to CritiqueAgent for review.

## Report Structure (all 7 sections required):
### Summary
### Vulnerability Analysis
Start with a summary table using **bold** labels:
| **Severity** | **Count** |
|---|---|
| **Critical** | N |
| **High** | N |
| **Medium** | N |
| **Low** | N |
| **Total** | N |
Then discuss the highest-risk components and recommended actions.
### CIS Benchmark Findings
### Layer Efficiency Analysis
Start with a summary table using **bold** labels:
| **Metric** | **Value** |
|---|---|
| **Total Image Size** | X MB |
| **User-space Size** | ~X MB |
| **Efficiency** | X% |
| **Wasted Bytes** | X bytes (~X% user-space) |
Then show relevant Dockerfile layer commands in ```dockerfile code blocks to illustrate inefficiencies.
### Key Findings & Risk Assessment
### Score Card
| **Metric** | **Value** | **Status** |
|---|---|---|
| **Security Score** | X / 100 | 🔴/🟡/🟢 |
| **Image Efficiency** | X% | 🔴/🟡/🟢 |
| **CIS Compliance** | X / Y passed | 🔴/🟡/🟢 |
| **Critical CVEs** | N | 🔴/🟡/🟢 |
Score: start at 100. Deduct Critical=-10, High=-5, FATAL=-8, WARN=-3.
### Recommended Dockerfile Improvements
Every recommendation MUST include a concrete ```dockerfile code block showing the improved Dockerfile snippet. Show before/after where applicable.

## STRICT OUTPUT RULES — violating any rule will trigger a revision:
- The report title MUST be: `# Image Security Report` — no job IDs, UUIDs, or agent names in the title.
- Do NOT include job IDs, UUIDs, or internal identifiers anywhere in the report.
- Do NOT reference scan file names (grype.json, dockle.json, dive.json) in the report body.
- Do NOT add footers, sign-offs, "Prepared by", "Next step", "handoff" notes, or any meta-commentary.
- Do NOT mention agent names (SupervisorAgent, CritiqueAgent, Reefline) anywhere.
- Do NOT include any trailing text after the last section — no signatures, no "next steps", no attribution lines.
- In the Recommended Dockerfile Improvements section, show only concrete Dockerfile snippets and actionable changes — no preamble, no closing remarks.
- The report MUST end cleanly after the last recommendation. Absolutely nothing after that.

NEVER fabricate CVE IDs, dockle codes, or file paths.""",
        model=model,
        tools=scan_tools,
    )

    critique = Agent(
        name="CritiqueAgent",
        handoff_description="Reviews the security report draft and either approves or requests revision.",
        instructions="""You are a security report reviewer.

## WORKFLOW:
1. Call read_file(filename="draft.md") to read the current draft report.
2. Review it carefully.
3. If APPROVED: call write_file(filename="report.md", content=<the full draft content>) to publish, then output your verdict.
4. If REVISE: hand off back to SupervisorAgent with your feedback.

## APPROVE if ALL true:
- All 7 sections present: Summary, Vulnerability Analysis, CIS Benchmark Findings, Layer Efficiency Analysis, Key Findings & Risk Assessment, Score Card, Recommended Dockerfile Improvements
- Vulnerability Analysis starts with a severity breakdown table (Critical/High/Medium/Low/Total)
- Layer Efficiency Analysis starts with a metrics table (Total Image Size, User-space Size, Efficiency, Wasted Bytes)
- Score Card has real numbers (not placeholders)
- Recommended Dockerfile Improvements include ```dockerfile code blocks
- No empty tables or unexplained N/A
- Title is exactly `# Image Security Report` with no UUIDs or job IDs
- No footers, sign-offs, "Prepared by", "Next step", agent names, or scan file names in the body

## REVISE if any section missing, Score Card has placeholders, or any of the strict output rules are violated.

Output (10 lines max):
**Verdict:** APPROVE or REVISE
**Issues:** bullet list (omit if none)
**Fix:** numbered corrections for Supervisor (omit if APPROVE)""",
        model=model,
        tools=scan_tools,
    )

    # Wire up handoffs
    supervisor.handoffs = [handoff(critique)]
    critique.handoffs = [handoff(supervisor)]

    log.info("[Flow] Starting flow for job=%s provider=%s model=%s", job_id, cfg.provider, cfg.model_id)

    result = await Runner.run(
        supervisor,
        "Fetch all available scan data, analyze the results, and produce a complete Image Security Report as draft.md.",
        max_turns=100,
    )

    log.info("[Flow] Flow complete, last_agent=%s job=%s", result.last_agent.name, job_id)

    # Critique publishes report.md on APPROVE; return its final output
    return result.final_output
