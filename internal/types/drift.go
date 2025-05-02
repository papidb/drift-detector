package types

type Drift struct {
	Name     string
	Type     ResourceType
	OldValue interface{}
	NewValue interface{}
}
