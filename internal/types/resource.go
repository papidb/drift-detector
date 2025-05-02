package types

type ResourceType string

const (
	EC2Instance ResourceType = "aws_instance"
)

// Resource represents a resource in the cloud provider
// eg AWS EC2 instance, GCP compute instance
type Resource struct {
	Name string
	Type ResourceType
	Data interface{}
}

func NewResource(name string, ResourceType ResourceType, data interface{}) Resource {
	return Resource{
		Name: name,
		Type: ResourceType,
		Data: data,
	}
}
