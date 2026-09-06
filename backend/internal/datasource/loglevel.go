package datasource

import (
	"regexp"
	"strings"
)

func detectLogLevel(labels map[string]string, line string) string {
	for _, key := range []string{"level", "lvl", "severity", "severity_text"} {
		if level, ok := labels[key]; ok {
			if normalized := normalizeLogLevel(level); normalized != "" {
				return normalized
			}
		}
	}

	if extracted := extractStructuredLogLevel(line); extracted != "" {
		return extracted
	}

	// Simple detection from line content
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "error") || strings.Contains(lower, "err="):
		return "error"
	case strings.Contains(lower, "warn"):
		return "warning"
	case strings.Contains(lower, "info"):
		return "info"
	case strings.Contains(lower, "debug"):
		return "debug"
	default:
		return ""
	}
}

var structuredLevelPattern = regexp.MustCompile(`(?i)(?:^|[\s>\[(,])(?:level|lvl|severity|severity_text)=(?:"|')?(trace|debug|info|warn|warning|error|fatal|panic|critical)(?:\d+)?(?:"|')?(?:$|[\s,\])])`)

func extractStructuredLogLevel(line string) string {
	match := structuredLevelPattern.FindStringSubmatch(line)
	if len(match) < 2 {
		return ""
	}

	return normalizeLogLevel(match[1])
}

func normalizeLogLevel(level string) string {
	normalized := strings.ToLower(strings.TrimSpace(strings.Trim(level, `"'`)))
	if normalized == "" {
		return ""
	}

	switch {
	case strings.HasPrefix(normalized, "trace"):
		return "debug"
	case strings.HasPrefix(normalized, "debug") || normalized == "dbg":
		return "debug"
	case strings.HasPrefix(normalized, "info") || normalized == "information" || normalized == "inf":
		return "info"
	case strings.HasPrefix(normalized, "warn") || normalized == "wrn":
		return "warning"
	case strings.HasPrefix(normalized, "error") || normalized == "err":
		return "error"
	case strings.HasPrefix(normalized, "fatal") || strings.HasPrefix(normalized, "panic") || strings.HasPrefix(normalized, "critical") || normalized == "crit":
		return "error"
	case normalized == "unspecified" || normalized == "unknown" || normalized == "default":
		return ""
	default:
		return ""
	}
}
