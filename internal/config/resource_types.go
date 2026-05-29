package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ResourceTypeDef describes a config-defined HyperFleet API resource type.
type ResourceTypeDef struct {
	Name           string
	Path           string `yaml:"path"`
	Parent         string `yaml:"parent"`
	StateKey       string `yaml:"state-key"`
	PathParam      string `yaml:"path-param"`
	CreateTemplate string `yaml:"create-template"`
}

type resourceTypeFields struct {
	Path           string `yaml:"path"`
	Parent         string `yaml:"parent"`
	StateKey       string `yaml:"state-key"`
	PathParam      string `yaml:"path-param"`
	CreateTemplate string `yaml:"create-template"`
}

var placeholderRE = regexp.MustCompile(`\{[a-zA-Z0-9_]+\}`)

// ResourceTypes returns validated resource type definitions from the active environment.
func (s *Store) ResourceTypes() ([]ResourceTypeDef, error) {
	m, err := s.resourceTypesMap()
	if err != nil {
		return nil, err
	}
	out := make([]ResourceTypeDef, 0, len(m))
	for _, def := range m {
		out = append(out, def)
	}
	return out, nil
}

// ResourceTypeNames returns configured type names in stable sorted order.
func (s *Store) ResourceTypeNames() ([]string, error) {
	types, err := s.ResourceTypes()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(types))
	for i, t := range types {
		names[i] = t.Name
	}
	sortStrings(names)
	return names, nil
}

func sortStrings(ss []string) {
	for i := 0; i < len(ss); i++ {
		for j := i + 1; j < len(ss); j++ {
			if ss[j] < ss[i] {
				ss[i], ss[j] = ss[j], ss[i]
			}
		}
	}
}

// ResourceType returns a single type definition by name.
func (s *Store) ResourceType(name string) (ResourceTypeDef, error) {
	m, err := s.resourceTypesMap()
	if err != nil {
		return ResourceTypeDef{}, err
	}
	def, ok := m[name]
	if !ok {
		return ResourceTypeDef{}, fmt.Errorf("[ERROR] unknown resource type %q", name)
	}
	return def, nil
}

// ResolveResourcePath substitutes ancestor state into the type's path template.
func (s *Store) ResolveResourcePath(typeName string) (string, error) {
	def, err := s.ResourceType(typeName)
	if err != nil {
		return "", err
	}
	path := def.Path
	ancestors, err := s.resourceAncestorChain(typeName)
	if err != nil {
		return "", err
	}
	for _, a := range ancestors {
		id := s.GetState(a.StateKey)
		if id == "" {
			return "", fmt.Errorf("[ERROR] No %s set in state. Run 'hf resource %s search <name>' first.", a.StateKey, a.Name)
		}
		param := effectivePathParam(a)
		path = strings.ReplaceAll(path, "{"+param+"}", id)
	}
	if placeholderRE.MatchString(path) {
		return "", fmt.Errorf("[ERROR] unresolved path placeholders in %q for type %q", path, typeName)
	}
	return path, nil
}

// ResourceID resolves an explicit ID or the type's state-key value.
func (s *Store) ResourceID(typeName, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	def, err := s.ResourceType(typeName)
	if err != nil {
		return "", err
	}
	id := s.GetState(def.StateKey)
	if id == "" {
		return "", fmt.Errorf("[ERROR] No %s set in state. Run 'hf resource %s search <name>' first.", def.StateKey, typeName)
	}
	return id, nil
}

func (s *Store) resourceTypesMap() (map[string]ResourceTypeDef, error) {
	active := s.ActiveEnvironment()
	if active == "" {
		return nil, fmt.Errorf("[ERROR] no active environment")
	}
	return parseResourceTypes(filepath.Join(s.dir, "environments", active+".yaml"))
}

func (s *Store) resourceAncestorChain(typeName string) ([]ResourceTypeDef, error) {
	m, err := s.resourceTypesMap()
	if err != nil {
		return nil, err
	}
	def, ok := m[typeName]
	if !ok {
		return nil, fmt.Errorf("[ERROR] unknown resource type %q", typeName)
	}
	var ancestors []ResourceTypeDef
	parentName := def.Parent
	seen := map[string]bool{typeName: true}
	for parentName != "" {
		if seen[parentName] {
			return nil, fmt.Errorf("[ERROR] cycle in resource-types involving %q", parentName)
		}
		seen[parentName] = true
		p, ok := m[parentName]
		if !ok {
			return nil, fmt.Errorf("[ERROR] resource type %q references unknown parent %q", typeName, parentName)
		}
		ancestors = append([]ResourceTypeDef{p}, ancestors...)
		parentName = p.Parent
	}
	return ancestors, nil
}

func parseResourceTypes(envPath string) (map[string]ResourceTypeDef, error) {
	raw, err := os.ReadFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]ResourceTypeDef{}, nil
		}
		return nil, fmt.Errorf("read environment file: %w", err)
	}
	var doc struct {
		ResourceTypes map[string]resourceTypeFields `yaml:"resource-types"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse resource-types: %w", err)
	}
	if len(doc.ResourceTypes) == 0 {
		return map[string]ResourceTypeDef{}, nil
	}

	out := make(map[string]ResourceTypeDef, len(doc.ResourceTypes))
	stateKeys := map[string]string{}

	for name, fields := range doc.ResourceTypes {
		def := ResourceTypeDef{
			Name:           name,
			Path:           strings.TrimSpace(fields.Path),
			Parent:         strings.TrimSpace(fields.Parent),
			StateKey:       strings.TrimSpace(fields.StateKey),
			PathParam:      strings.TrimSpace(fields.PathParam),
			CreateTemplate: strings.TrimSpace(fields.CreateTemplate),
		}
		if def.Path == "" {
			return nil, fmt.Errorf("[ERROR] resource-types.%s.path is required", name)
		}
		if def.StateKey == "" {
			return nil, fmt.Errorf("[ERROR] resource-types.%s.state-key is required", name)
		}
		if other, dup := stateKeys[def.StateKey]; dup {
			return nil, fmt.Errorf("[ERROR] duplicate state-key %q on resource-types.%s and %s", def.StateKey, name, other)
		}
		stateKeys[def.StateKey] = name
		if def.Parent == "" && placeholderRE.MatchString(def.Path) {
			return nil, fmt.Errorf("[ERROR] resource-types.%s.path contains placeholders but type has no parent", name)
		}
		out[name] = def
	}

	for name, def := range out {
		if def.Parent == "" {
			continue
		}
		if _, ok := out[def.Parent]; !ok {
			return nil, fmt.Errorf("[ERROR] resource-types.%s.parent references unknown type %q", name, def.Parent)
		}
	}

	for name := range out {
		if _, err := validateNoCycle(name, out); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func validateNoCycle(start string, defs map[string]ResourceTypeDef) ([]string, error) {
	seen := map[string]bool{}
	current := start
	for {
		def := defs[current]
		if def.Parent == "" {
			return nil, nil
		}
		if seen[current] {
			return nil, fmt.Errorf("[ERROR] cycle in resource-types involving %q", current)
		}
		seen[current] = true
		current = def.Parent
	}
}

func effectivePathParam(def ResourceTypeDef) string {
	if def.PathParam != "" {
		return def.PathParam
	}
	return derivePathParam(def.StateKey)
}

func derivePathParam(stateKey string) string {
	base := strings.TrimSuffix(stateKey, "-id")
	return strings.ReplaceAll(base, "-", "_") + "_id"
}
