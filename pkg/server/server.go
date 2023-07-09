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

const (
	metricsEndpoint = "/metrics"
)

type Server struct {
	grpcPort, httpPort string
	GrpcServer         *grpc.Server
	HttpServer         *http.Server
	MetricsRegistry    *prometheus.Registry
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
			return srv.GrpcServer.Serve(l)
		}, func(err error) {
			srv.GrpcServer.GracefulStop()
			srv.GrpcServer.Stop()
		}
}

func (srv *Server) serveHttp() (execute func() error, interrupt func(error)) {
	httpSrv := &http.Server{Addr: srv.httpPort}
	return func() error {
			m := http.NewServeMux()
			if srv.MetricsRegistry != nil {
				m.Handle(metricsEndpoint, promhttp.HandlerFor(
					srv.MetricsRegistry,
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
