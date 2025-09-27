package main

import (
	"fmt"
	"regexp"
	"strings"
)

type RedactionResult struct {
	Sanitized    string
	Replacements map[string]string
}

var (
	reJWT              = regexp.MustCompile(`[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}`)
	reKeyValueSecrets  = regexp.MustCompile(`(?i)(password|token|secret|apikey|api_key|passphrase)\s*[:=]\s*(?:"([^"]+)"|'([^']+)'|([^\s]+))`)
	reBearerToken      = regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._\-]{20,}`)
	reBase64Blob       = regexp.MustCompile(`(?m)(?:^|\s)([A-Za-z0-9+/]{40,}={0,2})(?:$|\s)`)
	reSecretKeyRefName = regexp.MustCompile(`(?m)(secretKeyRef:\s*(?:\{[^}]*\}|\n(?:\s{2,}.+\n?)+))`)
)

func RedactSensitive(input string) RedactionResult {
	result := RedactionResult{
		Sanitized:    input,
		Replacements: make(map[string]string),
	}
	counter := 0
	replace := func(match string) string {
		counter++
		placeholder := fmt.Sprintf("__SECRET_%d__", counter)
		result.Replacements[placeholder] = match
		return placeholder
	}

	result.Sanitized = reJWT.ReplaceAllStringFunc(result.Sanitized, replace)
	result.Sanitized = reBearerToken.ReplaceAllStringFunc(result.Sanitized, replace)
	result.Sanitized = reKeyValueSecrets.ReplaceAllStringFunc(result.Sanitized, func(match string) string {
		parts := reKeyValueSecrets.FindStringSubmatch(match)
		if len(parts) >= 5 {
			value := ""
			for i := 2; i <= 4; i++ {
				if parts[i] != "" {
					value = parts[i]
					break
				}
			}
			placeholder := replace(value)
			if strings.Contains(match, ":") {
				return fmt.Sprintf("%s: %s", parts[1], placeholder)
			}
			return fmt.Sprintf("%s=%s", parts[1], placeholder)
		}
		return replace(match)
	})
	result.Sanitized = reBase64Blob.ReplaceAllStringFunc(result.Sanitized, func(match string) string {
		trimmed := strings.TrimSpace(match)
		placeholder := replace(trimmed)
		return strings.Replace(match, trimmed, placeholder, 1)
	})

	result.Sanitized = reSecretKeyRefName.ReplaceAllStringFunc(result.Sanitized, func(match string) string {
		return replace(match)
	})

	return result
}

func RestoreSecrets(content string, replacements map[string]string) string {
	restored := content
	for placeholder, original := range replacements {
		restored = strings.ReplaceAll(restored, placeholder, original)
	}
	return restored
}

func RedactText(input string) string {
	return RedactSensitive(input).Sanitized
}
