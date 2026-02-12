# AI Agents Module - Architecture Overview

## 📂 Directory Structure

```
frontend/dashboard/src/api/
│
├── agents/                              # 🤖 AI Agents Module
│   ├── index.ts                        # Main entry point, exports all public APIs
│   ├── types.ts                        # TypeScript interfaces & types
│   ├── runner.ts                       # Agent execution orchestrator
│   ├── agents.ts                       # Pre-configured agent instances
│   ├── tools.ts                        # Tool definitions for API integration
│   ├── examples.ts                     # Usage examples & patterns
│   └── README.md                       # Complete documentation
│
└── prompts/                             # 📝 System Prompts
    ├── index.ts                        # Prompt loader with Vite glob import
    ├── system.txt                      # General assistant prompt
    ├── vulnerability-analysis.txt      # Security analysis prompt
    └── optimization.txt                # Image optimization prompt
```

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Your Application                        │
│  (React Components, Pages, Hooks)                           │
└────────────────────┬────────────────────────────────────────┘
                     │
                     │ import { agents, Runner, run }
                     │
┌────────────────────▼────────────────────────────────────────┐
│                  Agents Module (Frontend)                    │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Security   │  │ Optimization │  │   General    │     │
│  │    Agent     │  │    Agent     │  │    Agent     │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
│         │                  │                  │              │
│         └──────────────────┴──────────────────┘              │
│                            │                                 │
│                     ┌──────▼──────┐                         │
│                     │   Runner    │                         │
│                     │ (Orchestrator)                        │
│                     └──────┬──────┘                         │
│                            │                                 │
│                     ┌──────▼──────┐                         │
│                     │    Tools    │                         │
│                     │  (API Calls) │                        │
│                     └──────┬──────┘                         │
└────────────────────────────┼────────────────────────────────┘
                             │
                    HTTP POST /api/v1/ai/chat
                             │
┌────────────────────────────▼────────────────────────────────┐
│                    Backend API Server                        │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  AI Service (OpenAI, Anthropic, etc.)                │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Grype       │  │  Skopeo      │  │  Job Queue   │     │
│  │  Scanner     │  │  Inspector   │  │  Worker      │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

## 🔄 Data Flow

### 1. Basic Agent Run
```
User Input
    ↓
    ├─→ Runner.run(agent, input)
    │       ├─→ Load system prompt
    │       ├─→ Build message context
    │       └─→ POST /api/v1/ai/chat
    │               ↓
    │           Backend AI Service
    │               ↓
    │           LLM Response
    │               ↓
    └─→ AgentRunResult { finalOutput, steps, tokens, latency }
    ↓
User sees result
```

### 2. Agent with Tools
```
User Input: "Scan nginx:latest"
    ↓
    ├─→ Runner.run(securityAgent, input)
    │       ├─→ LLM decides to call tool: scan_vulnerabilities
    │       ├─→ Tool executes: POST /api/v1/jobs/scan
    │       ├─→ Backend runs Grype scanner
    │       ├─→ Results returned to agent
    │       └─→ LLM formats final response
    └─→ Security analysis report
```

### 3. Multi-Agent Handoff
```
User Input: "Analyze my-app:v1"
    ↓
    ├─→ Runner.run(generalAgent, input)
    │       ├─→ General agent analyzes request
    │       ├─→ Decides security focus needed
    │       ├─→ Hands off to securityAgent
    │       │       ├─→ Security scan runs
    │       │       └─→ Returns detailed findings
    │       └─→ General agent summarizes
    └─→ Comprehensive analysis
```

## 🎯 Key Components

### 1. **Agents** (`agents.ts`)
Pre-configured AI agents with specific purposes:
- **Security Agent**: CVE analysis, vulnerability assessment
- **Optimization Agent**: Size reduction, build improvements
- **General Agent**: All-purpose analysis with handoffs

### 2. **Runner** (`runner.ts`)
Orchestrates agent execution:
- Message handling
- Tool calling
- Streaming support
- Error handling
- Tracing/metadata

### 3. **Tools** (`tools.ts`)
Functions agents can call:
- `scan_vulnerabilities`: Run Grype scans
- `inspect_image`: Get image metadata
- `generate_optimizations`: Get suggestions
- `get_job_status`: Check async jobs

### 4. **Prompts** (`prompts/`)
System instructions stored as text files:
- Easy to edit without code changes
- Version controlled
- Loaded at build time via Vite

## 🚀 Integration Points

### Backend API Endpoints (To Implement)

```typescript
// AI Chat endpoint
POST /api/v1/ai/chat
{
  "agent": { "name": "...", "model": "...", "temperature": 0.7 },
  "messages": [...],
  "group_id": "trace-id",
  "metadata": { ... }
}
→ { "content": "...", "tokens_used": 150 }

// Streaming chat
POST /api/v1/ai/chat/stream
→ SSE stream: data: {"content": "..."}\n\n
```

### Frontend Usage

```typescript
// In a React component
import { agents, run } from '@/api/agents';

const handleAnalyze = async (imageRef: string) => {
  const result = await run(
    agents.security, 
    `Analyze ${imageRef}`
  );
  setAnalysis(result.finalOutput);
};
```

## 📋 Implementation Checklist

### Frontend ✅
- [x] Agent types and interfaces
- [x] Runner implementation
- [x] Pre-configured agents
- [x] Tool definitions
- [x] Prompt loader
- [x] System prompts
- [x] Documentation
- [x] Usage examples

### Backend (To Do)
- [ ] `/api/v1/ai/chat` endpoint
- [ ] `/api/v1/ai/chat/stream` endpoint
- [ ] OpenAI/Anthropic integration
- [ ] Tool execution handlers
- [ ] Tracing/logging infrastructure

## 🎨 Design Patterns

### 1. **Separation of Concerns**
- Prompts: `.txt` files (content)
- Agents: `.ts` files (logic)
- Tools: API integration layer

### 2. **Composition over Inheritance**
- Agents compose tools
- Tools compose API calls
- Runner composes execution flow

### 3. **Provider Agnostic**
- Can swap OpenAI for Anthropic
- Can add local models
- Tools are backend-agnostic

### 4. **Type Safety**
- Full TypeScript coverage
- Branded types for safety
- Generic tool definitions

## 🔧 Configuration

### Environment Variables (Backend)
```bash
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-...
AI_MODEL=gpt-4-turbo-preview
AI_TEMPERATURE=0.7
AI_MAX_TOKENS=4096
```

### Vite Config (Frontend)
Already supports glob imports for prompts!

## 📚 References

- [OpenAI Agents SDK](https://github.com/openai/openai-agents-js)
- [Vite Glob Import](https://vitejs.dev/guide/features.html#glob-import)
- [TypeScript Generics](https://www.typescriptlang.org/docs/handbook/2/generics.html)
