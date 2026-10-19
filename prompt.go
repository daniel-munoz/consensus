package main

import _ "embed"

// masterPrompt holds the developer instructions used to optimize user requests into prompts
//
//go:embed prompts/master.md
var masterPrompt string
