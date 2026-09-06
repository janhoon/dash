package datasource

import (
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Timestamp and metric helpers used by ClickHouse query-result conversion tests.
// The ClickHouse query Client lives in
// github.com/aceobservability/ace-datasource-clickhouse.

var clickHouseTimestampColumns = []string{"timestamp", "time", "ts", "datetime", "date", "_time", "event_time"}
var clickHouseMetricValueColumns = []string{"value", "val", "metric_value", "sum", "count", "avg", "min", "max"}

type clickHouseField struct {
	Key   string
	Value interface{}
}

type clickHouseMetricSeries struct {
	Metric map[string]string
	Values [][]interface{}
}

func NormaliseToMetrics(rows []map[string]interface{}) []MetricResult {
	seriesBySignature := map[string]*clickHouseMetricSeries{}

	for _, row := range rows {
		timestampField, hasTimestamp := pickClickHouseField(row, clickHouseTimestampColumns)
		valueField, hasValue := pickClickHouseField(row, clickHouseMetricValueColumns)
		if !hasTimestamp || !hasValue {
			continue
		}

		timestampSeconds, ok := parseClickHouseTimestampSeconds(timestampField.Value)
		if !ok {
			continue
		}

		value, ok := parseClickHouseFloat(valueField.Value)
		if !ok {
			continue
		}

		excludedColumns := append(append([]string{}, clickHouseTimestampColumns...), clickHouseMetricValueColumns...)
		metric := collectClickHouseLabels(row, excludedColumns)
		if metricName := pickClickHouseMetricName(row); metricName != "" {
			metric["__name__"] = metricName
		}
		if _, ok := metric["__name__"]; !ok {
			metric["__name__"] = "value"
		}

		signature := clickHouseMetricSignature(metric)
		series, exists := seriesBySignature[signature]
		if !exists {
			series = &clickHouseMetricSeries{
				Metric: metric,
				Values: make([][]interface{}, 0, 32),
			}
			seriesBySignature[signature] = series
		}

		series.Values = append(series.Values, []interface{}{
			timestampSeconds,
			strconv.FormatFloat(value, 'f', -1, 64),
		})
	}

	if len(seriesBySignature) == 0 {
		return []MetricResult{}
	}

	signatures := make([]string, 0, len(seriesBySignature))
	for signature := range seriesBySignature {
		signatures = append(signatures, signature)
	}
	sort.Strings(signatures)

	results := make([]MetricResult, 0, len(signatures))
	for _, signature := range signatures {
		series := seriesBySignature[signature]
		sort.Slice(series.Values, func(i, j int) bool {
			return clickHouseValueTimestamp(series.Values[i]) < clickHouseValueTimestamp(series.Values[j])
		})

		results = append(results, MetricResult{
			Metric: series.Metric,
			Values: series.Values,
		})
	}

	return results
}

func pickClickHouseMetricName(row map[string]interface{}) string {
	if field, ok := pickClickHouseField(row, []string{"__name__", "metric_name", "metric", "name", "series"}); ok {
		return strings.TrimSpace(anyToString(field.Value))
	}

	return ""
}

func pickClickHouseField(row map[string]interface{}, candidates []string) (clickHouseField, bool) {
	if len(row) == 0 {
		return clickHouseField{}, false
	}

	fieldsByNormalizedName := make(map[string]clickHouseField, len(row))
	for key, value := range row {
		normalized := normalizeClickHouseColumnName(key)
		if normalized == "" {
			continue
		}

		if _, exists := fieldsByNormalizedName[normalized]; !exists {
			fieldsByNormalizedName[normalized] = clickHouseField{Key: key, Value: value}
		}
	}

	for _, candidate := range candidates {
		if field, ok := fieldsByNormalizedName[normalizeClickHouseColumnName(candidate)]; ok {
			return field, true
		}
	}

	return clickHouseField{}, false
}

func collectClickHouseLabels(row map[string]interface{}, excludedColumns []string) map[string]string {
	excluded := make(map[string]struct{}, len(excludedColumns))
	for _, column := range excludedColumns {
		excluded[normalizeClickHouseColumnName(column)] = struct{}{}
	}

	keys := make([]string, 0, len(row))
	for key := range row {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	labels := map[string]string{}
	for _, key := range keys {
		if _, skip := excluded[normalizeClickHouseColumnName(key)]; skip {
			continue
		}

		value := strings.TrimSpace(anyToString(row[key]))
		if value == "" {
			continue
		}

		labels[key] = value
	}

	return labels
}

func clickHouseMetricSignature(metric map[string]string) string {
	if len(metric) == 0 {
		return ""
	}

	keys := make([]string, 0, len(metric))
	for key := range metric {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+metric[key])
	}

	return strings.Join(parts, "|")
}

func clickHouseValueTimestamp(value []interface{}) float64 {
	if len(value) == 0 {
		return 0
	}

	timestamp, ok := parseClickHouseFloat(value[0])
	if !ok {
		return 0
	}

	return timestamp
}

func clickHouseSecondsToRFC3339(seconds float64) string {
	nanos := int64(math.Round(seconds * float64(time.Second)))
	return time.Unix(0, nanos).UTC().Format(time.RFC3339Nano)
}

func parseClickHouseTimestampSeconds(value interface{}) (float64, bool) {
	if value == nil {
		return 0, false
	}

	if typed, ok := value.(time.Time); ok {
		return float64(typed.UnixNano()) / float64(time.Second), true
	}

	if typed, ok := value.(string); ok {
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return 0, false
		}

		if parsed, ok := parseClickHouseTimeString(trimmed); ok {
			return float64(parsed.UnixNano()) / float64(time.Second), true
		}

		numeric, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return 0, false
		}

		return normalizeClickHouseEpochSeconds(numeric), true
	}

	numeric, ok := anyToFloat64(value)
	if !ok {
		return 0, false
	}

	return normalizeClickHouseEpochSeconds(numeric), true
}

func parseClickHouseTimeString(value string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, true
		}

		parsed, err = time.ParseInLocation(layout, value, time.UTC)
		if err == nil {
			return parsed, true
		}
	}

	return time.Time{}, false
}

func normalizeClickHouseEpochSeconds(value float64) float64 {
	absValue := math.Abs(value)
	switch {
	case absValue >= 1e18:
		return value / 1e9
	case absValue >= 1e15:
		return value / 1e6
	case absValue >= 1e12:
		return value / 1e3
	default:
		return value
	}
}

func parseClickHouseFloat(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case nil:
		return 0, false
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return 0, false
		}
		parsed, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func normalizeClickHouseColumnName(name string) string {
	trimmed := strings.ToLower(strings.TrimSpace(name))
	if trimmed == "" {
		return ""
	}

	b := strings.Builder{}
	b.Grow(len(trimmed))
	for _, char := range trimmed {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			b.WriteRune(char)
		}
	}

	return b.String()
}
