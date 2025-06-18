package pal

import (
	"context"
)

// ServiceList is a proxy service to a list of services.
type ServiceList struct {
	Services []ServiceDef
}

func (s *ServiceList) Dependencies() []ServiceDef {
	return s.Services
}

func (s *ServiceList) Run(_ context.Context) error {
	return nil
}

func (s *ServiceList) Init(_ context.Context) error {
	return nil
}

func (s *ServiceList) HealthCheck(_ context.Context) error {
	return nil
}

func (s *ServiceList) Shutdown(_ context.Context) error {
	return nil
}

func (s *ServiceList) Make() any {
	return nil
}

func (s *ServiceList) Instance(_ context.Context) (any, error) {
	return nil, nil
}

func (s *ServiceList) Name() string {
	return "$ServiceList"
}

func (s *ServiceList) RunConfig() *RunConfig {
	return nil
}
