package prompts

import (
	_ "embed"
)

// ============
// Chat Prompts
// ============

//go:embed chat.prompt
var Instructions string

//go:embed identity.prompt
var Identity string

//go:embed summary.included.prompt
var SummaryIncluded string

//go:embed chat.previous.summary.prompt
var PreviousSummary string

// ============
// Summary Prompts
// ============

//go:embed summary.prompt
var Summary string

// //go:embed existing.summary.prompt
var ExistingSummary string

// ============
// Knowledge Prompts
// ============

//go:embed knowledge.prompt
var Knowledge string
