package server

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

type Server struct {
	grpcPort, httpPort string
	Grpc               *grpc.Server
	Http               *http.Server
	Metrics            *prometheus.Registry
}

func (srv *Server) Serve() {
	g := &run.Group{}
	g.Add(srv.serveGrpc())
	g.Add(srv.serveHttp())

	if err := g.Run(); err != nil {
		os.Exit(1)
	}
}

func (srv *Server) serveGrpc() (execute func() error, interrupt func(error)) {
	return func() error {
			l, err := net.Listen(tcp, srv.grpcPort)
			if err != nil {
				return err
			}
			return srv.Grpc.Serve(l)
		}, func(err error) {
			srv.Grpc.GracefulStop()
			srv.Grpc.Stop()
		}
}

func (srv *Server) serveHttp() (execute func() error, interrupt func(error)) {
	httpSrv := &http.Server{Addr: srv.httpPort}
	return func() error {
			m := http.NewServeMux()
			if srv.Metrics != nil {
				m.Handle(metricsEndpoint, promhttp.HandlerFor(
					srv.Metrics,
					promhttp.HandlerOpts{
						EnableOpenMetrics: true,
					},
				))
			}

			httpSrv.Handler = m
			log.Println("starting http playground at " + httpSrv.Addr)
			return httpSrv.ListenAndServe()
		}, func(error) {
			if err := httpSrv.Close(); err != nil {
				log.Fatalf("failed to close http playground: %v", err)
			}
		}
}
