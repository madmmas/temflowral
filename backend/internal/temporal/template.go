package temporal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/madmmas/temflowral/backend/internal/api"
)

const (
	maxTemplateExpansions = 32
	maxTemplateWalkDepth  = 16
	maxTemplateStringLen  = 8192
)

// Matches {{ nodes.<id>.output.<path...> }} with optional inner whitespace.
var templatePattern = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)

var (
	templateIdentPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	pathSegmentPattern   = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

// nodeWithResolvedConfig deep-copies node.Config and resolves
// {{ nodes.<id>.output.<path> }} using active predecessor results.
func nodeWithResolvedConfig(node api.Node, inputs []NodeResult) (api.Node, error) {
	if node.Config == nil {
		return node, nil
	}
	if !configContainsTemplate(*node.Config) {
		return node, nil
	}
	if node.Type == WaitNodeType {
		return api.Node{}, fmt.Errorf("templates are not allowed in wait node config")
	}

	root := templateDataRoot(inputs)
	resolved, err := resolveConfigValue(*node.Config, root, 0, newTemplateBudget(), skipGraphKey(node.Type))
	if err != nil {
		return api.Node{}, err
	}
	resolvedMap, ok := resolved.(map[string]interface{})
	if !ok {
		return api.Node{}, fmt.Errorf("resolved config must be an object")
	}
	node.Config = &resolvedMap
	return node, nil
}

func skipGraphKey(nodeType string) bool {
	return nodeType == ChildWorkflowNodeType
}

type templateBudget struct {
	remaining int
}

func newTemplateBudget() *templateBudget {
	return &templateBudget{remaining: maxTemplateExpansions}
}

func (budget *templateBudget) consume() error {
	if budget.remaining <= 0 {
		return fmt.Errorf("too many template expansions (max %d)", maxTemplateExpansions)
	}
	budget.remaining--
	return nil
}

func configContainsTemplate(config map[string]interface{}) bool {
	return valueContainsTemplate(config, 0)
}

func valueContainsTemplate(value interface{}, depth int) bool {
	if depth > maxTemplateWalkDepth {
		return false
	}
	switch typed := value.(type) {
	case string:
		return strings.Contains(typed, "{{")
	case map[string]interface{}:
		for _, item := range typed {
			if valueContainsTemplate(item, depth+1) {
				return true
			}
		}
	case []interface{}:
		for _, item := range typed {
			if valueContainsTemplate(item, depth+1) {
				return true
			}
		}
	}
	return false
}

func templateDataRoot(inputs []NodeResult) map[string]interface{} {
	nodes := make(map[string]interface{}, len(inputs))
	for _, input := range inputs {
		output := input.Value
		if output == nil {
			output = map[string]interface{}{}
		}
		nodes[input.NodeID] = map[string]interface{}{"output": output}
	}
	return map[string]interface{}{"nodes": nodes}
}

func resolveConfigValue(
	value interface{},
	root map[string]interface{},
	depth int,
	budget *templateBudget,
	skipGraph bool,
) (interface{}, error) {
	if depth > maxTemplateWalkDepth {
		return nil, fmt.Errorf("config nesting exceeds %d levels", maxTemplateWalkDepth)
	}
	switch typed := value.(type) {
	case string:
		return resolveTemplateString(typed, root, budget)
	case map[string]interface{}:
		out := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			if skipGraph && key == "graph" {
				out[key] = deepCopyJSON(item)
				continue
			}
			resolved, err := resolveConfigValue(item, root, depth+1, budget, false)
			if err != nil {
				return nil, err
			}
			out[key] = resolved
		}
		return out, nil
	case []interface{}:
		out := make([]interface{}, len(typed))
		for i, item := range typed {
			resolved, err := resolveConfigValue(item, root, depth+1, budget, false)
			if err != nil {
				return nil, err
			}
			out[i] = resolved
		}
		return out, nil
	default:
		return typed, nil
	}
}

func resolveTemplateString(raw string, root map[string]interface{}, budget *templateBudget) (interface{}, error) {
	if len(raw) > maxTemplateStringLen {
		return nil, fmt.Errorf("template string exceeds %d bytes", maxTemplateStringLen)
	}
	if !strings.Contains(raw, "{{") {
		return raw, nil
	}

	trimmed := strings.TrimSpace(raw)
	if matches := templatePattern.FindStringSubmatch(trimmed); len(matches) == 2 && matches[0] == trimmed {
		if err := budget.consume(); err != nil {
			return nil, err
		}
		return lookupTemplatePath(matches[1], root)
	}

	var builder strings.Builder
	last := 0
	matches := templatePattern.FindAllStringSubmatchIndex(raw, -1)
	if matches == nil {
		if strings.Contains(raw, "{{") || strings.Contains(raw, "}}") {
			return nil, fmt.Errorf("invalid template syntax in %q", raw)
		}
		return raw, nil
	}
	for _, match := range matches {
		if err := budget.consume(); err != nil {
			return nil, err
		}
		builder.WriteString(raw[last:match[0]])
		expr := raw[match[2]:match[3]]
		value, err := lookupTemplatePath(expr, root)
		if err != nil {
			return nil, err
		}
		text, err := stringifyTemplateValue(value)
		if err != nil {
			return nil, err
		}
		builder.WriteString(text)
		last = match[1]
	}
	builder.WriteString(raw[last:])
	resolved := builder.String()
	if strings.Contains(resolved, "{{") || strings.Contains(resolved, "}}") {
		return nil, fmt.Errorf("unresolved template syntax in %q", raw)
	}
	return resolved, nil
}

func lookupTemplatePath(expr string, root map[string]interface{}) (interface{}, error) {
	expr = strings.TrimSpace(expr)
	parts := strings.Split(expr, ".")
	if len(parts) < 4 || parts[0] != "nodes" || parts[2] != "output" {
		return nil, fmt.Errorf(
			"template path %q must be nodes.<nodeId>.output.<field>[.<field>...]",
			expr,
		)
	}
	if !templateIdentPattern.MatchString(parts[1]) {
		return nil, fmt.Errorf("template node id %q is invalid", parts[1])
	}
	for _, part := range parts[3:] {
		if !pathSegmentPattern.MatchString(part) {
			return nil, fmt.Errorf("template path segment %q is invalid", part)
		}
	}

	current, ok := root["nodes"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("template data root is missing nodes")
	}
	nodeObj, ok := current[parts[1]].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("template references unknown or inactive predecessor %q", parts[1])
	}
	value, ok := nodeObj["output"]
	if !ok {
		return nil, fmt.Errorf("template predecessor %q has no output", parts[1])
	}
	for _, part := range parts[3:] {
		asMap, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("template path %q is not an object at %q", expr, part)
		}
		next, exists := asMap[part]
		if !exists {
			return nil, fmt.Errorf("template path %q not found", expr)
		}
		value = next
	}
	return value, nil
}

func stringifyTemplateValue(value interface{}) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "", fmt.Errorf("template value is null")
	case string:
		return typed, nil
	case bool:
		return strconv.FormatBool(typed), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(typed), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	default:
		return "", fmt.Errorf("template value of type %T cannot be interpolated into a string", value)
	}
}

func deepCopyJSON(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			out[key] = deepCopyJSON(item)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(typed))
		for i, item := range typed {
			out[i] = deepCopyJSON(item)
		}
		return out
	default:
		return typed
	}
}

func containsTemplate(s string) bool {
	return strings.Contains(s, "{{")
}

// validateTemplateSyntaxInString ensures every {{ ... }} island is a valid
// nodes.<id>.output.<path> expression without resolving it.
func validateTemplateSyntaxInString(raw string) error {
	if !strings.Contains(raw, "{{") {
		return nil
	}
	matches := templatePattern.FindAllStringSubmatch(raw, -1)
	if matches == nil {
		return fmt.Errorf("invalid template syntax")
	}
	for _, match := range matches {
		if err := validateTemplatePathShape(match[1]); err != nil {
			return err
		}
	}
	stripped := templatePattern.ReplaceAllString(raw, "")
	if strings.Contains(stripped, "{{") || strings.Contains(stripped, "}}") {
		return fmt.Errorf("invalid template syntax")
	}
	return nil
}

func validateTemplatePathShape(expr string) error {
	expr = strings.TrimSpace(expr)
	parts := strings.Split(expr, ".")
	if len(parts) < 4 || parts[0] != "nodes" || parts[2] != "output" {
		return fmt.Errorf(
			"template path %q must be nodes.<nodeId>.output.<field>[.<field>...]",
			expr,
		)
	}
	if !templateIdentPattern.MatchString(parts[1]) {
		return fmt.Errorf("template node id %q is invalid", parts[1])
	}
	for _, part := range parts[3:] {
		if !pathSegmentPattern.MatchString(part) {
			return fmt.Errorf("template path segment %q is invalid", part)
		}
	}
	return nil
}
