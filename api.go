package pal

import (
	"context"
	"fmt"
	"reflect"

	typetostring "github.com/samber/go-type-to-string"
)

// Provide registers a const as a service. `T` is used to generating service name.
// Typically, `T` would be one of:
// - An interface, in this case passed value must implement it. Used when T may have multiple implementations like mocks for tests.
// - A pointer to an instance of `T`. For instance,`Provide[*Foo](&Foo{})`. Used when mocking is not required.
// If the passed value implements [Initer] or [PalIniter], the matching init method is called after dependency injection,
// unless a ToInit hook is set on the returned [ServiceConst] (see [ServiceConst.ToInit]).
func Provide[T any](value T) *ServiceConst[T] {
	validateNonNilPointer(value)

	return ProvideNamed(typetostring.GetType[T](), value)
}

// ProvideNamed registers a const as a service with a given name. Acts like Provide but allows to specify a name.
func ProvideNamed[T any](name string, value T) *ServiceConst[T] {
	validateNonNilPointer(value)

	return &ServiceConst[T]{instance: value, ServiceTyped: ServiceTyped[T]{name: name}}
}

// ProvideFn registers a singleton built with a given function.
func ProvideFn[I any, T any](fn func(ctx context.Context) (T, error)) *ServiceFnSingleton[I, T] {
	return ProvideNamedFn[I](typetostring.GetType[I](), fn)
}

// ProvideFn registers a singleton built with a given function.
func ProvideNamedFn[I any, T any](name string, fn func(ctx context.Context) (T, error)) *ServiceFnSingleton[I, T] {
	validateFactoryFunction[I, T](fn)

	return &ServiceFnSingleton[I, T]{
		fn:             fn,
		ServiceFactory: ServiceFactory[I, T]{ServiceTyped: ServiceTyped[I]{name: name}},
	}
}

// ProvideRunner turns the given function into an anounumous runner. It will run in the background, and the passed context will
// be canceled on app shutdown.
func ProvideRunner(fn func(ctx context.Context) error) *ServiceRunner {
	return &ServiceRunner{
		fn: fn,
	}
}

// ProvideList registers a list of given services.
func ProvideList(services ...ServiceDef) *ServiceList {
	return &ServiceList{Services: services}
}

// ProvideFactory0 registers a factory service that is build with a given function with no arguments.
func ProvideFactory0[I any, T any](fn func(ctx context.Context) (T, error)) *ServiceFactory0[I, T] {
	return ProvideNamedFactory0[I](typetostring.GetType[I](), fn)
}

// ProvideFactory1 registers a factory service that is built in runtime with a given function that takes one argument.
func ProvideFactory1[I any, T any, P1 any](fn func(ctx context.Context, p1 P1) (T, error)) *ServiceFactory1[I, T, P1] {
	validateFactoryFunction[I, T](fn)
	return ProvideNamedFactory1[I](typetostring.GetType[I](), fn)
}

// ProvideFactory2 registers a factory service that is built in runtime with a given function that takes two arguments.
func ProvideFactory2[I any, T any, P1 any, P2 any](fn func(ctx context.Context, p1 P1, p2 P2) (T, error)) *ServiceFactory2[I, T, P1, P2] {
	return ProvideNamedFactory2[I](typetostring.GetType[I](), fn)
}

// ProvideFactory3 registers a factory service that is built in runtime with a given function that takes three arguments.
func ProvideFactory3[I any, T any, P1 any, P2 any, P3 any](fn func(ctx context.Context, p1 P1, p2 P2, p3 P3) (T, error)) *ServiceFactory3[I, T, P1, P2, P3] {
	return ProvideNamedFactory3[I](typetostring.GetType[I](), fn)
}

// ProvideFactory4 registers a factory service that is built in runtime with a given function that takes four arguments.
func ProvideFactory4[I any, T any, P1 any, P2 any, P3 any, P4 any](fn func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4) (T, error)) *ServiceFactory4[I, T, P1, P2, P3, P4] {
	return ProvideNamedFactory4[I](typetostring.GetType[I](), fn)
}

// ProvideFactory5 registers a factory service that is built in runtime with a given function that takes five arguments.
func ProvideFactory5[I any, T any, P1 any, P2 any, P3 any, P4 any, P5 any](fn func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) (T, error)) *ServiceFactory5[I, T, P1, P2, P3, P4, P5] {
	return ProvideNamedFactory5[I](typetostring.GetType[I](), fn)
}

// ProvideNamedFactory0 is like ProvideFactory0 but allows to specify a name.
func ProvideNamedFactory0[I any, T any](name string, fn func(ctx context.Context) (T, error)) *ServiceFactory0[I, T] {
	validateFactoryFunction[I, T](fn)
	return &ServiceFactory0[I, T]{
		fn:             fn,
		ServiceFactory: ServiceFactory[I, T]{ServiceTyped: ServiceTyped[I]{name: name}},
	}
}

// ProvideNamedFactory1 is like ProvideFactory1 but allows to specify a name.
func ProvideNamedFactory1[I any, T any, P1 any](name string, fn func(ctx context.Context, p1 P1) (T, error)) *ServiceFactory1[I, T, P1] {
	validateFactoryFunction[I, T](fn)

	return &ServiceFactory1[I, T, P1]{
		fn:             fn,
		ServiceFactory: ServiceFactory[I, T]{ServiceTyped: ServiceTyped[I]{name: name}},
	}
}

// ProvideNamedFactory2 is like ProvideFactory2 but allows to specify a name.
func ProvideNamedFactory2[I any, T any, P1 any, P2 any](name string, fn func(ctx context.Context, p1 P1, p2 P2) (T, error)) *ServiceFactory2[I, T, P1, P2] {
	validateFactoryFunction[I, T](fn)

	return &ServiceFactory2[I, T, P1, P2]{
		fn:             fn,
		ServiceFactory: ServiceFactory[I, T]{ServiceTyped: ServiceTyped[I]{name: name}},
	}
}

// ProvideNamedFactory3 is like ProvideFactory3 but allows to specify a name.
func ProvideNamedFactory3[I any, T any, P1 any, P2 any, P3 any](name string, fn func(ctx context.Context, p1 P1, p2 P2, p3 P3) (T, error)) *ServiceFactory3[I, T, P1, P2, P3] {
	validateFactoryFunction[I, T](fn)

	return &ServiceFactory3[I, T, P1, P2, P3]{
		fn:             fn,
		ServiceFactory: ServiceFactory[I, T]{ServiceTyped: ServiceTyped[I]{name: name}},
	}
}

// ProvideNamedFactory4 is like ProvideFactory4 but allows to specify a name.
func ProvideNamedFactory4[I any, T any, P1 any, P2 any, P3 any, P4 any](name string, fn func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4) (T, error)) *ServiceFactory4[I, T, P1, P2, P3, P4] {
	validateFactoryFunction[I, T](fn)

	return &ServiceFactory4[I, T, P1, P2, P3, P4]{
		fn:             fn,
		ServiceFactory: ServiceFactory[I, T]{ServiceTyped: ServiceTyped[I]{name: name}},
	}
}

// ProvideNamedFactory5 is like ProvideFactory5 but allows to specify a name.
func ProvideNamedFactory5[I any, T any, P1 any, P2 any, P3 any, P4 any, P5 any](name string, fn func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) (T, error)) *ServiceFactory5[I, T, P1, P2, P3, P4, P5] {
	validateFactoryFunction[I, T](fn)

	return &ServiceFactory5[I, T, P1, P2, P3, P4, P5]{
		fn:             fn,
		ServiceFactory: ServiceFactory[I, T]{ServiceTyped: ServiceTyped[I]{name: name}},
	}
}

// ProvidePal registers all services for the given pal instance
func ProvidePal(pal *Pal) *ServiceList {
	services := make([]ServiceDef, 0, len(pal.Services()))
	for _, v := range pal.Services() {
		if v.Name() != "*github.com/zhulik/pal.Pal" {
			services = append(services, v)
		}
	}

	return ProvideList(services...)
}

// Invoke retrieves or creates an instance of type T from the given Pal container.
// Invoker may be nil, in this case an instance of Pal will be extracted from the context,
// if the context does not contain a Pal instance, an error will be returned.
func Invoke[T any](ctx context.Context, invoker Invoker, args ...any) (T, error) {
	name := typetostring.GetType[T]()
	return InvokeNamed[T](ctx, invoker, name, args...)
}

// MustInvoke is like Invoke but panics if an error occurs.
func MustInvoke[T any](ctx context.Context, invoker Invoker, args ...any) T {
	return must(Invoke[T](ctx, invoker, args...))
}

// InvokeNamed is like Invoke but allows to specify a name.
func InvokeNamed[T any](ctx context.Context, invoker Invoker, name string, args ...any) (T, error) {
	if invoker == nil {
		var err error
		invoker, err = FromContext(ctx)
		if err != nil {
			return empty[T](), err
		}
	}

	a, err := invoker.Invoke(ctx, name, args...)
	if err != nil {
		return empty[T](), err
	}

	casted, ok := a.(T)
	if !ok {
		return empty[T](), fmt.Errorf("%w: %s. %+v does not implement %s", ErrServiceInvalid, name, a, name)
	}

	return casted, nil
}

// MustInvokeNamed is like InvokeNamed but panics if an error occurs.
func MustInvokeNamed[T any](ctx context.Context, invoker Invoker, name string, args ...any) T {
	return must(InvokeNamed[T](ctx, invoker, name, args...))
}

// InvokeAs invokes a service and casts it to the expected type. It returns an error if the cast fails.
// May be useful when invoking a service with an interface type and you want to cast it to a concrete type.
// Invoker may be nil, in this case an instance of Pal will be extracted from the context,
// if the context does not contain a Pal instance, an error will be returned.
func InvokeAs[T any, C any](ctx context.Context, invoker Invoker, args ...any) (*C, error) {
	name := typetostring.GetType[T]()
	return InvokeNamedAs[T, C](ctx, invoker, name, args...)
}

// InvokeNamedAs is like InvokeAs but allows to specify a name.
func InvokeNamedAs[T any, C any](ctx context.Context, invoker Invoker, name string, args ...any) (*C, error) {
	service, err := InvokeNamed[T](ctx, invoker, name, args...)
	if err != nil {
		return nil, err
	}
	casted, ok := any(service).(*C)
	if !ok {
		var c *C
		return nil, fmt.Errorf("%w: %T cannot be cast to %T", ErrServiceInvalidCast, service, c)
	}

	return casted, nil
}

// MustInvokeNamedAs is like InvokeNamedAs but panics if an error occurs.
func MustInvokeNamedAs[T any, C any](ctx context.Context, invoker Invoker, name string, args ...any) *C {
	return must(InvokeNamedAs[T, C](ctx, invoker, name, args...))
}

// MustInvokeAs is like InvokeAs but panics if an error occurs.
func MustInvokeAs[T any, C any](ctx context.Context, invoker Invoker, args ...any) *C {
	return must(InvokeAs[T, C](ctx, invoker, args...))
}

// InvokeByInterface invokes a service by interface.
// It iterates over all services and returns the only one that implements the interface.
// If no service implements the interface, or multiple services implement the interface, or given I is not an interface
// an error will be returned.
// Invoker may be nil, in this case an instance of Pal will be extracted from the context,
// if the context does not contain a Pal instance, an error will be returned.
func InvokeByInterface[I any](ctx context.Context, invoker Invoker, args ...any) (I, error) {
	if invoker == nil {
		var err error
		invoker, err = FromContext(ctx)
		if err != nil {
			return empty[I](), err
		}
	}
	iface := reflect.TypeOf((*I)(nil)).Elem()
	if invoker == nil {
		var err error
		invoker, err = FromContext(ctx)
		if err != nil {
			return empty[I](), err
		}
	}

	instance, err := invoker.InvokeByInterface(ctx, iface, args...)
	if err != nil {
		return empty[I](), err
	}
	return instance.(I), nil
}

// MustInvokeByInterface is like InvokeByInterface but panics if an error occurs.
func MustInvokeByInterface[I any](ctx context.Context, invoker Invoker, args ...any) I {
	return must(InvokeByInterface[I](ctx, invoker, args...))
}

// Build resolves dependencies for a struct of type T using the provided context and Invoker.
// It initializes the struct's fields by injecting appropriate dependencies based on the field types.
// Returns the fully initialized struct or an error if dependency resolution fails.
// Invoker may be nil, in this case an instance of Pal will be extracted from the context,
// if the context does not contain a Pal instance, an error will be returned.
func Build[T any](ctx context.Context, invoker Invoker) (*T, error) {
	s := new(T)

	err := InjectInto(ctx, invoker, s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// MustBuild is like Build but panics if an error occurs.x
func MustBuild[T any](ctx context.Context, invoker Invoker) *T {
	return must(Build[T](ctx, invoker))
}

// InjectInto populates the fields of a struct of type T with dependencies obtained from the given Invoker.
// It only sets fields that are exported and match a resolvable dependency, skipping fields when ErrServiceNotFound occurs.
// Returns an error if dependency invocation fails or other unrecoverable errors occur during injection.
func InjectInto[T any](ctx context.Context, invoker Invoker, s *T) error {
	if invoker == nil {
		var err error
		invoker, err = FromContext(ctx)
		if err != nil {
			return err
		}
	}
	return invoker.InjectInto(ctx, s)
}

// MustInjectInto is like InjectInto but panics if an error occurs.
func MustInjectInto[T any](ctx context.Context, invoker Invoker, s *T) {
	must("", InjectInto(ctx, invoker, s))
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func validateNonNilPointer(value any) {
	val := reflect.ValueOf(value)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		panic(fmt.Sprintf("Argument must be a non-nil pointer to a struct, got %T", value))
	}
}

func validateFactoryFunction[I any, T any](fn any) {
	// Factory function must return a pointer to a struct that implements I
	// I and T must be the same pointer type.
	// This way pal can inspect the type of the returned value to build the correct dependency tree.
	if reflect.TypeOf(fn).Out(0).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Factory function must return a pointer, got %s", reflect.TypeOf(fn).Out(0).Kind()))
	}

	if typetostring.GetType[I]() == typetostring.GetType[T]() {
		return
	}

	iType := reflect.TypeOf((*I)(nil)).Elem()
	tType := reflect.TypeOf((*T)(nil)).Elem()

	if iType.Kind() != reflect.Interface {
		panic(fmt.Sprintf("I must be an interface, got %s", iType.Kind()))
	}

	if !tType.Implements(iType) {
		panic(fmt.Sprintf("T (%s) must implement interface I (%s)", tType, iType))
	}
}
