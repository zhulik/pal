package inspect

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/zhulik/pal"
)

const (
	jsonContentType = "application/json"
	htmlContentType = "text/html"
)

type Inspect struct {
	P *pal.Pal

	port int

	server *http.Server
}

func (i *Inspect) Init(_ context.Context) error {
	i.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", i.port),
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		ReadTimeout:       time.Second,
		IdleTimeout:       time.Second,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/pal/tree.json", i.httpTreeJSON)
	mux.HandleFunc("/pal/tree", i.httpTree)

	staticServer := http.FileServerFS(StaticFS)

	mux.Handle("/pal/", http.StripPrefix("/pal/", staticServer))

	i.server.Handler = mux

	return nil
}

func (i *Inspect) RunConfig() *pal.RunConfig {
	return &pal.RunConfig{
		Wait: false,
	}
}

func (i *Inspect) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", i.server.Addr)
	if err != nil {
		return err
	}
	i.P.Logger().Info("Starting Inspect HTTP server", "address", i.server.Addr)

	go func() {
		<-ctx.Done()

		// create a new context as the one passed to Run is already canceled
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

func (i *Inspect) httpTreeJSON(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", jsonContentType)
	json, err := DAGToJSON(i.P.Container().Graph())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(json)
	if err != nil {
		panic(err)
	}
}

func (i *Inspect) httpTree(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", htmlContentType)

	treeHTML, err := StaticFS.ReadFile("static/tree.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(treeHTML)
	if err != nil {
		panic(err)
	}
}
