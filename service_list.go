package pal

import (
	"context"
	"fmt"
)

// ServiceList is a proxy service to a list of services.
type ServiceList struct {
	ServiceTyped[any]
	Services []ServiceDef
}

func (s *ServiceList) Dependencies() []ServiceDef {
	return s.Services
}

func (s *ServiceList) Instance(_ context.Context, _ ...any) (any, error) {
	return nil, nil
}

func (s *ServiceList) Name() string {
	return fmt.Sprintf("$service-list-%s", randomID())
}
