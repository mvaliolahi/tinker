package api

import (
        "encoding/json"
        "fmt"
        "os"
        "path/filepath"
        "sort"
        "strings"

        "go.yaml.in/yaml/v3"
)

// SpecEndpoint represents a single API endpoint parsed from an OpenAPI spec.
type SpecEndpoint struct {
        Method      string   `json:"method"`
        Path        string   `json:"path"`
        Summary     string   `json:"summary,omitempty"`
        OperationID string   `json:"operation_id,omitempty"`
        Tags        []string `json:"tags,omitempty"`
}

// Spec represents a parsed OpenAPI specification.
type Spec struct {
        Title     string         `json:"title"`
        Version   string         `json:"version"`
        Endpoints []SpecEndpoint `json:"endpoints"`
}

// ParseSpec reads and parses an OpenAPI/Swagger spec file (YAML or JSON).
// Returns a structured Spec with all endpoints extracted.
func ParseSpec(path string) (*Spec, error) {
        data, err := os.ReadFile(path)
        if err != nil {
                return nil, fmt.Errorf("reading spec file: %w", err)
        }

        var raw map[string]interface{}

        ext := strings.ToLower(filepath.Ext(path))
        switch ext {
        case ".yaml", ".yml":
                if err := yaml.Unmarshal(data, &raw); err != nil {
                        return nil, fmt.Errorf("parsing YAML spec: %w", err)
                }
        case ".json":
                if err := json.Unmarshal(data, &raw); err != nil {
                        return nil, fmt.Errorf("parsing JSON spec: %w", err)
                }
        default:
                // Try YAML first, then JSON
                if err := yaml.Unmarshal(data, &raw); err != nil {
                        if err2 := json.Unmarshal(data, &raw); err2 != nil {
                                return nil, fmt.Errorf("parsing spec: not valid YAML or JSON")
                        }
                }
        }

        return extractSpec(raw)
}

// extractSpec processes the raw map into a structured Spec.
func extractSpec(raw map[string]interface{}) (*Spec, error) {
        spec := &Spec{}

        // Extract info
        if info, ok := raw["info"].(map[string]interface{}); ok {
                if t, ok := info["title"].(string); ok {
                        spec.Title = t
                }
                if v, ok := info["version"].(string); ok {
                        spec.Version = v
                }
        }

        // Extract paths
        paths, ok := raw["paths"].(map[string]interface{})
        if !ok {
                return spec, nil
        }

        // Sort paths for deterministic output
        sortedPaths := make([]string, 0, len(paths))
        for p := range paths {
                sortedPaths = append(sortedPaths, p)
        }
        sort.Strings(sortedPaths)

        for _, path := range sortedPaths {
                pathItem, ok := paths[path].(map[string]interface{})
                if !ok {
                        continue
                }

                for _, method := range []string{"get", "post", "put", "patch", "delete", "options", "head"} {
                        op, ok := pathItem[method].(map[string]interface{})
                        if !ok {
                                continue
                        }

                        endpoint := SpecEndpoint{
                                Method: strings.ToUpper(method),
                                Path:   path,
                        }

                        if summary, ok := op["summary"].(string); ok {
                                endpoint.Summary = summary
                        }
                        if opID, ok := op["operationId"].(string); ok {
                                endpoint.OperationID = opID
                        }
                        if tags, ok := op["tags"].([]interface{}); ok {
                                for _, t := range tags {
                                        if tag, ok := t.(string); ok {
                                                endpoint.Tags = append(endpoint.Tags, tag)
                                        }
                                }
                        }

                        spec.Endpoints = append(spec.Endpoints, endpoint)
                }
        }

        return spec, nil
}

// FindEndpoint searches for an endpoint matching the given method and path.
// Path matching supports both exact match and parameterized paths like /users/{id}.
func (s *Spec) FindEndpoint(method, path string) *SpecEndpoint {
        method = strings.ToUpper(method)

        // Exact match first
        for i := range s.Endpoints {
                if s.Endpoints[i].Method == method && s.Endpoints[i].Path == path {
                        return &s.Endpoints[i]
                }
        }

        // Parameterized path match: /users/123 matches /users/{id}
        for i := range s.Endpoints {
                if s.Endpoints[i].Method != method {
                        continue
                }
                if pathMatches(s.Endpoints[i].Path, path) {
                        return &s.Endpoints[i]
                }
        }

        return nil
}

// EndpointsByTag groups endpoints by their first tag.
func (s *Spec) EndpointsByTag() map[string][]SpecEndpoint {
        result := make(map[string][]SpecEndpoint)
        for _, ep := range s.Endpoints {
                tag := "default"
                if len(ep.Tags) > 0 {
                        tag = ep.Tags[0]
                }
                result[tag] = append(result[tag], ep)
        }
        return result
}

// pathMatches checks if a concrete path matches a parameterized spec path.
// e.g., pathMatches("/users/{id}", "/users/123") returns true.
func pathMatches(specPath, concretePath string) bool {
        specParts := strings.Split(specPath, "/")
        concreteParts := strings.Split(concretePath, "/")

        if len(specParts) != len(concreteParts) {
                return false
        }

        for i := range specParts {
                if strings.HasPrefix(specParts[i], "{") && strings.HasSuffix(specParts[i], "}") {
                        continue // parameter segment — matches anything
                }
                if specParts[i] != concreteParts[i] {
                        return false
                }
        }
        return true
}
