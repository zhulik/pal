package pal

import (
	"context"
	"errors"
	"fmt"
)

// store is responsible for storing factories, instances and the dependency graph
type store struct {
	factories map[string]ServiceFactory
	graph     *dag

	log loggerFn
}

// newStore creates a new store instance
func newStore(factories map[string]ServiceFactory, log loggerFn) *store {
	return &store{
		factories: factories,
		log:       log,
	}
}

func (s *store) setLogger(log loggerFn) {
	s.log = log
}

func (s *store) validate(ctx context.Context) error {
	var errs []error

	for _, factory := range s.factories {
		errs = append(errs, factory.Validate(ctx))
	}

	return errors.Join(errs...)
}

func (s *store) init(ctx context.Context) error {
	var err error
	s.graph, err = newDag(s.factories)
	if err != nil {
		return err
	}

	// file, _ := os.Initialize("./mygraph.gv")
	// _ = draw.DOT(s.graph, file)

	err = s.graph.InReverseTopologicalOrder(func(factory ServiceFactory) error {
		if factory.IsSingleton() {
			s.log("initializing %s", factory.Name())

			if err := factory.Initialize(ctx); err != nil {
				return err
			}

			s.log("%s initialized", factory.Name())
		}

		return nil
	})

	s.log("Pal initialized. Services: %s", s.services())

	return err
}

func (s *store) invoke(ctx context.Context, name string) (any, error) {
	s.log("invoking %s", name)

	factory, err := s.graph.Vertex(name)
	if err != nil {
		return nil, fmt.Errorf("%w: '%s', known services: %s. %w", ErrServiceNotFound, name, s.services(), err)
	}

	instance, err := factory.Instance(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: '%s'", ErrServiceInitFailed, name)
	}

	return instance, nil
}

func (s *store) shutdown(ctx context.Context) error {
	var errs []error
	s.graph.InTopologicalOrder(func(factory ServiceFactory) error { // nolint:errcheck
		if factory.IsSingleton() {
			service, _ := factory.Instance(ctx)

			if shutdowner, ok := service.(Shutdowner); ok {
				s.log("shutting down %s", factory.Name())

				err := shutdowner.Shutdown(ctx)
				if err != nil {
					s.log("%s shot down with error=%+v", factory.Name(), err)
					errs = append(errs, err)
					return nil
				}

				s.log("%s shot down successfully", factory.Name())
			}
		}
		return nil
	})
	return errors.Join(errs...)
}

func (s *store) healthCheck(ctx context.Context) error {
	return s.graph.ForEachVertex(func(factory ServiceFactory) error { // nolint:errcheck
		if factory.IsSingleton() {
			service, _ := factory.Instance(ctx)

			if healthChecker, ok := service.(HealthChecker); ok {
				s.log("health checking %s", service)

				err := healthChecker.HealthCheck(ctx)
				if err != nil {
					s.log("%s failed health check error=%+v", service, err)
					return err
				}

				s.log("%s passed health check successfully", service)
			}
		}

		return nil
	})
}

func (s *store) services() []ServiceFactory {
	return s.graph.Vertices()
}

func (s *store) runners(ctx context.Context) map[string]Runner {
	runners := map[string]Runner{}

	s.graph.ForEachVertex(func(factory ServiceFactory) error { // nolint:errcheck
		if factory.IsRunner() {
			if runner, err := factory.Instance(ctx); err == nil {
				runners[factory.Name()] = runner.(Runner)
			}
		}
		return nil
	})

	return runners
}
