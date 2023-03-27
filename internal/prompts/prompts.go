package prompts

import (
	_ "embed"
)

//go:embed chat.prompt
var Instructions string

//go:embed knowledge.prompt
var Knowledge string

//go:embed summary.prompt
var Summary string

//go:embed identity.prompt
var Identity string

//go:embed existing.summary.prompt
var ExistingSummary string
