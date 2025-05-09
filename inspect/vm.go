package inspect

import (
	"context"
	"github.com/dop251/goja"
	"github.com/zhulik/pal"
	"log/slog"
)

type VM struct {
	*goja.Runtime
}

func NewVM(ctx context.Context, logger *slog.Logger) (*VM, error) {
	vm := goja.New()

	logger = logger.With("ECMAScript", true)

	vars := map[string]goja.Value{
		"pal": vm.ToValue(pal.FromContext(ctx)),
		"console": vm.ToValue(map[string]any{
			"log": logger.Info,
		}),
		"ctx": vm.ToValue(context.Background()), // TODO: cancel when shutting down?
	}

	for k, v := range vars {
		if err := vm.Set(k, v); err != nil {
			return nil, err
		}
	}

	return &VM{vm}, nil
}
