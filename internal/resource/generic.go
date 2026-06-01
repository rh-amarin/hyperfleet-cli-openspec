package resource

import "fmt"

// GenericResource is a JSON object returned by config-defined HyperFleet API types.
type GenericResource map[string]any

// StringField returns a top-level string field when present.
func (g GenericResource) StringField(key string) string {
	if g == nil {
		return ""
	}
	v, ok := g[key].(string)
	if !ok {
		return ""
	}
	return v
}

// ID returns the resource id field.
func (g GenericResource) ID() string { return g.StringField("id") }

// Name returns the resource name field.
func (g GenericResource) Name() string { return g.StringField("name") }

// Kind returns the resource kind field.
func (g GenericResource) Kind() string { return g.StringField("kind") }

// DeletedTime returns the resource deleted_time field when set.
func (g GenericResource) DeletedTime() string { return g.StringField("deleted_time") }

// Generation returns generation as a decimal string, or empty when absent.
func (g GenericResource) Generation() string {
	if g == nil {
		return ""
	}
	switch v := g["generation"].(type) {
	case float64:
		return fmt.Sprintf("%d", int(v))
	case int:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	default:
		return ""
	}
}
