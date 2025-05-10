package inspect

import (
	"context"
	"github.com/dop251/goja"
	"github.com/zhulik/pal"
)

type VM struct {
	*goja.Runtime

	Logger *Logger

	cancel context.CancelFunc
}

func (vm *VM) Init(ctx context.Context) error {
	vm.Runtime = goja.New()

	ctx, vm.cancel = context.WithCancel(ctx)

	vars := map[string]goja.Value{
		"pal": vm.ToValue(pal.FromContext(ctx)),
		"console": vm.ToValue(map[string]any{
			"log": vm.Logger.With("ECMAScript", true).Info,
		}),
		"ctx": vm.ToValue(ctx),
	}

	for k, v := range vars {
		if err := vm.Set(k, v); err != nil {
			return err
		}
	}

	return nil
}

func (v *VM) Shutdown(ctx context.Context) error {
	v.cancel()
	v.Interrupt("shutdown")
	return nil
}
