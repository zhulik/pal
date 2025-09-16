package inspect

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/zhulik/pal"
)

const (
	inspectPort = 24242

	jsonContentType = "application/json"
	htmlContentType = "text/html"
)

type Inspect struct {
	P  *pal.Pal
	VM *VM

	server *http.Server
}

func (i *Inspect) Shutdown(ctx context.Context) error {
	err := i.VM.Shutdown(ctx)
	if err != nil {
		return nil
	}
	return i.server.Shutdown(ctx)
}

func (i *Inspect) Init(_ context.Context) error {
	i.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", inspectPort),
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		ReadTimeout:       time.Second,
		IdleTimeout:       time.Second,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/pal/eval", i.httpEval)
	mux.HandleFunc("/pal/tree.json", i.httpTreeJSON)
	mux.HandleFunc("/pal/tree", i.httpTree)

	staticServer := http.FileServerFS(StaticFS)

	mux.Handle("/pal/", http.StripPrefix("/pal/", staticServer))

	i.server.Handler = mux

	return nil
}

func (i *Inspect) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", i.server.Addr)
	if err != nil {
		return err
	}
	i.P.Logger().Info("Starting Inspect HTTP server", "address", i.server.Addr)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), i.P.Config().ShutdownTimeout)
		defer cancel()

		i.server.Shutdown(ctx) //nolint:errcheck
	}()

	err = i.server.Serve(ln)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (i *Inspect) httpEval(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	res, err := i.VM.RunString(string(bytes))

	if err != nil {
		w.WriteHeader(422)

		w.Write([]byte(err.Error())) //nolint:errcheck
		return
	}

	w.Write([]byte(res.String())) //nolint:errcheck
}

func (i *Inspect) httpTreeJSON(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", jsonContentType)
	json, err := DAGToJSON(i.P.Container().Graph())
	if err != nil {
		w.WriteHeader(500)
		return
	}

	_, err = w.Write(json)
	if err != nil {
		panic(err)
	}
}

func (i *Inspect) httpTree(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", htmlContentType)

	treeHtml, err := StaticFS.ReadFile("static/tree.html")
	if err != nil {
		w.WriteHeader(500)
		return
	}

	_, err = w.Write(treeHtml)
	if err != nil {
		panic(err)
	}
}
