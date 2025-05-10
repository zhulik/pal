package inspect

import (
	"context"
	"fmt"
	"github.com/chzyer/readline"
	"io"
	"net/http"
	"strings"
)

type RemoteConsole struct {
	Logger *Logger
}

func (r *RemoteConsole) Run(ctx context.Context) error {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "remote-js> ",
		HistoryFile:     "/tmp/pal-RemoteConsole.tmp",
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

		// TODO: url from ENV or from config
		resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/pal/eval", inspectPort), "application/text", strings.NewReader(line))
		if err != nil {
			r.Logger.Warn("Input sending error", "error", err)
			continue
		}

		result, err := io.ReadAll(resp.Body)
		if err != nil {
			r.Logger.Warn("Response body reading error", "error", err)
			continue
		}

		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			r.Logger.Warn("Non-200 response:", "body", result)
		}

		r.Logger.Info(string(result))
	}

	return nil
}
