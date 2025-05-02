package pal

import (
	"context"
	"errors"
	"fmt"
)

// store is responsible for storing services, instances and the dependency graph
type store struct {
	services map[string]Service
	graph    *dag

	log loggerFn
}

// newStore creates a new store instance
func newStore(services map[string]Service, log loggerFn) *store {
	return &store{
		services: services,
		log:      log,
	}
}

func (s *store) setLogger(log loggerFn) {
	s.log = log
}

func (s *store) validate(ctx context.Context) error {
	var errs []error

	for _, service := range s.services {
		errs = append(errs, service.Validate(ctx))
	}

	return errors.Join(errs...)
}

func (s *store) init(ctx context.Context) error {
	var err error
	s.graph, err = newDag(s.services)
	if err != nil {
		return err
	}

	// file, _ := os.Initialize("./mygraph.gv")
	// _ = draw.DOT(s.graph, file)

	err = s.graph.InReverseTopologicalOrder(func(service Service) error {
		if service.IsSingleton() {
			s.log("initializing %s", service.Name())

			if err := service.Initialize(ctx); err != nil {
				return err
			}

			s.log("%s initialized", service.Name())
		}

		return nil
	})

	s.log("Pal initialized. Services: %s", s.Services())

	return err
}

func (s *store) invoke(ctx context.Context, name string) (any, error) {
	s.log("invoking %s", name)

	service, err := s.graph.Vertex(name)
	if err != nil {
		return nil, fmt.Errorf("%w: '%s', known services: %s. %w", ErrServiceNotFound, name, s.Services(), err)
	}

	instance, err := service.Instance(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: '%s'", ErrServiceInitFailed, name)
	}

	return instance, nil
}

func (s *store) shutdown(ctx context.Context) error {
	var errs []error
	s.graph.InTopologicalOrder(func(service Service) error { // nolint:errcheck
		if service.IsSingleton() {
			instance, _ := service.Instance(ctx)

			if shutdowner, ok := instance.(Shutdowner); ok {
				s.log("shutting down %s", service.Name())

				err := shutdowner.Shutdown(ctx)
				if err != nil {
					s.log("%s shot down with error=%+v", service.Name(), err)
					errs = append(errs, err)
					return nil
				}

				s.log("%s shot down successfully", service.Name())
			}
		}
		return nil
	})
	return errors.Join(errs...)
}

func (s *store) healthCheck(ctx context.Context) error {
	return s.graph.ForEachVertex(func(service Service) error { // nolint:errcheck
		if service.IsSingleton() {
			instance, _ := service.Instance(ctx)

			if healthChecker, ok := instance.(HealthChecker); ok {
				s.log("health checking %s", instance)

				err := healthChecker.HealthCheck(ctx)
				if err != nil {
					s.log("%s failed health check error=%+v", instance, err)
					return err
				}

				s.log("%s passed health check successfully", instance)
			}
		}

		return nil
	})
}

func (s *store) Services() []Service {
	return s.graph.Vertices()
}

func (s *store) runners(ctx context.Context) map[string]Runner {
	runners := map[string]Runner{}

	s.graph.ForEachVertex(func(service Service) error { // nolint:errcheck
		if service.IsRunner() {
			if runner, err := service.Instance(ctx); err == nil {
				runners[service.Name()] = runner.(Runner)
			}
		}
		return nil
	})

	return runners
}
