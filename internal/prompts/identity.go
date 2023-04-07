package prompts

import (
	_ "embed"
)

//go:embed identities/rose.prompt
var Rose string

//go:embed identities/winston.prompt
var Winston string

//go:embed identities/syl.prompt
var Syl string

//go:embed identities/marcus.prompt
var Marcus string
