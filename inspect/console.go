package inspect

import (
	"context"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/zhulik/pal"
	"log/slog"
)

type Console struct {
	logger *slog.Logger
	vm     *VM
	p      *pal.Pal
}

func (c *Console) Shutdown(ctx context.Context) error {
	c.vm.Shutdown(ctx)
	return nil
}

func (c *Console) Init(ctx context.Context) error {
	c.p = pal.FromContext(ctx).(*pal.Pal)
	var err error

	c.logger = slog.With("palComponent", "Console")
	c.vm, err = NewVM(ctx, c.logger)
	if err != nil {
		return err
	}

	return nil
}

func (c *Console) Run(ctx context.Context) error {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "js> ",
		HistoryFile:     "/tmp/pal-console.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return err
	}
	defer rl.Close()

	go func() {
		<-ctx.Done()
		rl.Close()
	}()

	for {
		line, err := rl.Readline()
		if err != nil {
			return err
		}

		if line == "" {
			continue
		}

		result, err := c.vm.RunString(line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println(result.String())
	}

	return nil
}
