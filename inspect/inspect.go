package inspect

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/goccy/go-graphviz"
	"github.com/zhulik/pal"
	"github.com/zhulik/pal/pkg/draw"
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
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := i.VM.RunString(string(bytes))

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)

		w.Write([]byte(err.Error())) //nolint:errcheck
		return
	}

	w.Write([]byte(res.String())) //nolint:errcheck
}

func (i *Inspect) httpGraph(w http.ResponseWriter, r *http.Request) {
	dot := draw.RenderDOT(i.P.Container().Graph())

	if r.Header.Get("Accept") == gvContentType {
		w.Header().Set("Content-Type", gvContentType)
		w.Write([]byte(dot)) //nolint:errcheck
		return
	}

	graph, err := graphviz.ParseBytes([]byte(dot))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error())) //nolint:errcheck
		return
	}

	w.Header().Set("Content-Type", svgContentType)

	err = i.GV.Render(r.Context(), graph, graphviz.SVG, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error())) //nolint:errcheck
		return
	}
	return
}
