package inspect

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/dominikbraun/graph/draw"
	"github.com/goccy/go-graphviz"
	"github.com/zhulik/pal"
)

const (
	inspectPort = 24242

	gvContentType = "text/vnd.graphviz"
	// svg mime type
	svgContentType = "image/svg+xml"
)

type Inspect struct {
	P  *pal.Pal
	VM *VM
	GV *graphviz.Graphviz

	server *http.Server
}

func Provide() []pal.ServiceDef {
	return []pal.ServiceDef{
		pal.Provide[*Inspect, Inspect](),
		pal.ProvideFactory[*VM, VM](),

		pal.ProvideFn(graphviz.New).
			BeforeShutdown(func(_ context.Context, g *graphviz.Graphviz) error {
				return g.Close()
			}),
	}
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
	i.P.Logger().Info("Starting Inspect HTTP server", "address", i.server.Addr)

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

	err := i.P.HealthCheck(r.Context())

	if err != nil {
		i.P.Logger().Warn("Health check failed", "err", err)
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

	res, err := i.VM.RunString(string(bytes))

	if err != nil {
		w.WriteHeader(422)

		w.Write([]byte(err.Error())) //nolint:errcheck
		return
	}

	w.Write([]byte(res.String())) //nolint:errcheck
}

func (i *Inspect) httpGraph(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") == gvContentType {
		w.Header().Set("Content-Type", gvContentType)
		_ = draw.DOT(i.P.Container().Graph().Graph, w)
	}

	buf := &bytes.Buffer{}

	if err := draw.DOT(i.P.Container().Graph().Graph, buf); err != nil {
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

		err = i.GV.Render(r.Context(), graph, graphviz.SVG, w)
		if err != nil {
			w.WriteHeader(500)
			return
		}
	}
}
