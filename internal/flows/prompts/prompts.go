package prompts

import _ "embed"

//go:embed system_prompt.txt
var SupervisorSystemPrompt string

//go:embed critique_prompt.txt
var CritiqueSystemPrompt string
