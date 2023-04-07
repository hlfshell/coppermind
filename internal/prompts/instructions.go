package prompts

import (
	_ "embed"
)

// ============
// Chat Prompts
// ============

//go:embed instructions/chat.prompt
var Instructions string

//go:embed instructions/chat.previous.summary.prompt
var PreviousSummary string

//go:embed instructions/conversation.continuance.prompt
var ConversationContinuance string

// ============
// Summary Prompts
// ============

//go:embed instructions/summary.prompt
var Summary string

// //go:embed instructions/existing.summary.prompt
var ExistingSummary string

// ============
// Knowledge Prompts
// ============

//go:embed instructions/knowledge.prompt
var Knowledge string

//go:embed instructions/knowledge.compression.prompt
var KnoweldgeCompression string
