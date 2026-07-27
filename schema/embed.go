package schema

import "embed"

// FS embeds the JSON schemas for flagd configuration validation.
//
//go:embed v0/*
var FS embed.FS
