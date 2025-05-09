package inspect

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/dominikbraun/graph/draw"
	"github.com/goccy/go-graphviz"
	"github.com/zhulik/pal"
)

const (
	gvContentType = "text/vnd.graphviz"
	// svg mime type
	svgContentType = "image/svg+xml"
)

type Inspect struct {
	logger *slog.Logger
	vm     *VM
	p      *pal.Pal
	gv     *graphviz.Graphviz

	server *http.Server
}

func (i *Inspect) Shutdown(ctx context.Context) error {
	err := i.server.Shutdown(ctx)
	if err != nil {
		return err
	}

	i.vm.Shutdown(ctx)
	return nil
}

func (i *Inspect) Init(ctx context.Context) error {
	i.p = pal.FromContext(ctx).(*pal.Pal)
	var err error

	i.logger = slog.With("palComponent", "Inspect")
	i.vm, err = NewVM(ctx, i.logger)
	if err != nil {
		return err
	}

	i.gv, err = graphviz.New(ctx)
	if err != nil {
		return err
	}

	i.server = &http.Server{
		Addr:              ":24242",
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		ReadTimeout:       time.Second,
		IdleTimeout:       time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", i.httpHealth)

	mux.HandleFunc("/pal/eval", i.httpEval)
	mux.HandleFunc("/pal/graph", i.httpGraph)

	i.server.Handler = mux

	return nil
}

func (i *Inspect) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", i.server.Addr)
	if err != nil {
		return err
	}
	i.logger.Info("Starting Inspect HTTP server", "address", i.server.Addr)

	go func() {
		<-ctx.Done()
		// TODO: figure out a good context here, Run's ctx is cancelled.
		i.server.Shutdown(context.TODO()) //nolint:errcheck
	}()

	err = i.server.Serve(ln)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (i *Inspect) httpHealth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	err := i.p.HealthCheck(r.Context())

	if err != nil {
		i.logger.Warn("Health check failed: %+v", err)
		w.WriteHeader(500)
	}
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

	res, err := i.vm.RunString(string(bytes))

	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(422)
		return
	}

	w.Write([]byte(res.String()))
}

func (i *Inspect) httpGraph(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") == gvContentType {
		w.Header().Set("Content-Type", gvContentType)
		_ = draw.DOT(i.p.Container().Graph().Graph, w)
	}

	buf := &bytes.Buffer{}

	if err := draw.DOT(i.p.Container().Graph().Graph, buf); err != nil {
		w.WriteHeader(500)
		return
	}

	graph, err := graphviz.ParseBytes(buf.Bytes())
	if err != nil {
		w.WriteHeader(500)
		return
	}

	if r.Header.Get("Accept") == svgContentType {
		w.Header().Set("Content-Type", svgContentType)

		err = i.gv.Render(r.Context(), graph, graphviz.SVG, w)
		if err != nil {
			w.WriteHeader(500)
			return
		}
	}
}
