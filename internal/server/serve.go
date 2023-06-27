package server

import (
	"log"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

const (
	tcp             = "tcp"
	metricsEndpoint = "/metrics"
)

// ServeGrpc starts a gRPC server on the given address.
func ServeGrpc(srv *grpc.Server, addr string) (func() error, func(error)) {
	return func() error {
			l, err := net.Listen(tcp, addr)
			if err != nil {
				return err
			}
			return srv.Serve(l)
		}, func(err error) {
			srv.GracefulStop()
			srv.Stop()
		}
}

// ServeHttp starts an HTTP server on the given address and sets up the
// prometheus metrics endpoint.
func ServeHttp(
	httpSrv *http.Server,
	registry *prometheus.Registry,
) (func() error, func(error)) {
	return func() error {
			m := http.NewServeMux()
			m.Handle(metricsEndpoint, promhttp.HandlerFor(
				registry,
				promhttp.HandlerOpts{
					EnableOpenMetrics: true,
				},
			))
			httpSrv.Handler = m
			log.Println("starting http server at " + httpSrv.Addr)
			return httpSrv.ListenAndServe()
		}, func(error) {
			if err := httpSrv.Close(); err != nil {
				log.Fatalf("failed to close http server: %v", err)
			}
		}
}
