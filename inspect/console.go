package inspect

import (
	"context"
	"fmt"

	"github.com/zhulik/pal"

	"github.com/chzyer/readline"
)

type Console struct {
	P  *pal.Pal
	VM *VM
}

func (c *Console) Shutdown(ctx context.Context) error {
	return c.VM.Shutdown(ctx)
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

		result, err := c.VM.RunString(line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println(result.String())
	}
}
