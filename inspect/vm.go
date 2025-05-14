package inspect

import (
	"context"

	"github.com/dop251/goja"
	"github.com/zhulik/pal"
)

type VM struct {
	*goja.Runtime

	P *pal.Pal

	cancel context.CancelFunc
}

func (vm *VM) Init(ctx context.Context) error {
	vm.Runtime = goja.New()

	ctx, vm.cancel = context.WithCancel(ctx)

	vars := map[string]goja.Value{
		"pal": vm.ToValue(vm.P),
		"console": vm.ToValue(map[string]any{
			"log": vm.P.Logger().With("ECMAScript", true).Info,
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

func (vm *VM) Shutdown(_ context.Context) error {
	vm.cancel()
	vm.Interrupt("shutdown")
	return nil
}
