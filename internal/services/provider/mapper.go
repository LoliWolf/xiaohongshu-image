package provider

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Mapper struct {
	requestMapping map[string]interface{}
}

func NewMapper(requestMapping map[string]interface{}) *Mapper {
	if requestMapping == nil {
		requestMapping = make(map[string]interface{})
	}
	return &Mapper{
		requestMapping: requestMapping,
	}
}

func (m *Mapper) MapRequest(req UnifiedGenRequest) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range m.requestMapping {
		mappedValue, err := m.mapValue(value, req)
		if err != nil {
			return nil, fmt.Errorf("failed to map key %s: %w", key, err)
		}
		result[key] = mappedValue
	}

	if len(result) == 0 {
		reqJSON, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		var resultMap map[string]interface{}
		if err := json.Unmarshal(reqJSON, &resultMap); err != nil {
			return nil, err
		}
		return resultMap, nil
	}

	return result, nil
}

func (m *Mapper) mapValue(value interface{}, req UnifiedGenRequest) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "$.") {
			return m.extractField(v, req)
		}
		return v, nil
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			mapped, err := m.mapValue(val, req)
			if err != nil {
				return nil, err
			}
			result[key] = mapped
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			mapped, err := m.mapValue(val, req)
			if err != nil {
				return nil, err
			}
			result[i] = mapped
		}
		return result, nil
	default:
		return v, nil
	}
}

func (m *Mapper) extractField(path string, req UnifiedGenRequest) (interface{}, error) {
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid path: %s", path)
	}

	fieldName := parts[1]

	val := reflect.ValueOf(req)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("field not found: %s", fieldName)
	}

	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil, nil
		}
		return field.Elem().Interface(), nil
	}

	return field.Interface(), nil
}

func (m *Mapper) ExtractJobID(responseBody []byte) (string, error) {
	var response map[string]interface{}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	jobID, err := m.extractByJSONPath(response, "$.data.id")
	if err != nil {
		return "", err
	}

	jobIDStr, ok := jobID.(string)
	if !ok {
		return "", fmt.Errorf("job ID is not a string")
	}

	return jobIDStr, nil
}

func (m *Mapper) ExtractStatus(responseBody []byte) (JobStatus, int, *string, error) {
	var response map[string]interface{}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", 0, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	statusStr, err := m.extractByJSONPath(response, "$.status")
	if err != nil {
		return "", 0, nil, err
	}

	status := JobStatus(strings.ToLower(fmt.Sprintf("%v", statusStr)))

	progress := 0
	if progressVal, err := m.extractByJSONPath(response, "$.progress"); err == nil {
		if p, ok := progressVal.(float64); ok {
			progress = int(p)
		}
	}

	var resultURL *string
	if urlVal, err := m.extractByJSONPath(response, "$.output.url"); err == nil {
		if url, ok := urlVal.(string); ok && url != "" {
			resultURL = &url
		}
	}

	return status, progress, resultURL, nil
}

func (m *Mapper) extractByJSONPath(data map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")

	current := data
	for i := 1; i < len(parts); i++ {
		part := parts[i]

		val, exists := current[part]
		if !exists {
			return nil, fmt.Errorf("path not found: %s", path)
		}

		if i == len(parts)-1 {
			return val, nil
		}

		nextMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("path segment %s is not an object", part)
		}

		current = nextMap
	}

	return current, nil
}

func (m *Mapper) extractByGJSONPath(data map[string]interface{}, path string) (interface{}, error) {
	path = strings.TrimPrefix(path, "$.")

	parts := regexp.MustCompile(`\[(\d+)\]`).Split(path, -1)
	indices := regexp.MustCompile(`\[(\d+)\]`).FindAllStringSubmatch(path, -1)

	current := data
	for i, part := range parts {
		if part == "" {
			continue
		}

		val, exists := current[part]
		if !exists {
			return nil, fmt.Errorf("path not found: %s", path)
		}

		if i < len(indices) && len(indices[i]) > 1 {
			index, err := strconv.Atoi(indices[i][1])
			if err != nil {
				return nil, fmt.Errorf("invalid index: %s", indices[i][1])
			}

			arr, ok := val.([]interface{})
			if !ok {
				return nil, fmt.Errorf("path segment %s is not an array", part)
			}

			if index >= len(arr) {
				return nil, fmt.Errorf("index out of bounds: %d", index)
			}

			val = arr[index]
		}

		if i == len(parts)-1 {
			return val, nil
		}

		nextMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("path segment %s is not an object", part)
		}

		current = nextMap
	}

	return current, nil
}
