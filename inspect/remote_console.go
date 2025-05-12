package inspect

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/chzyer/readline"
)

type RemoteConsole struct {
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
			log.Printf("Input sending error: %+v", err)
			continue
		}

		result, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Response body reading error: %+v", err)
			continue
		}

		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Non-200 response, bod: %s", result)
		}

		log.Printf("%s", result)
	}
}
