package utils

import (
	"fmt"
	"strings"
)

// FilterNamePrepend is a function that removes a "{BOTNAME} | " or equivalent
// prefixes from the output string, in case the LLM tries adding it back in.
func FilterNamePrepend(name string, output string) string {
	matches := []string{
		fmt.Sprintf("%s | ", name),
		fmt.Sprintf("%s|", name),
		fmt.Sprintf("%s :", name),
		fmt.Sprintf("%s:", name),
		fmt.Sprintf("%s -", name),
		fmt.Sprintf("%s-", name),
	}

	// Check to see if the output string starts with any of the matches
	for _, match := range matches {
		if len(output) > len(match) && output[:len(match)] == match {
			output = output[len(match):]
			break
		}
	}

	// Because of the possibility of leading spaces, trim the output
	output = strings.TrimSpace(output)

	return output
}
