package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"google.golang.org/grpc"

	"github.com/javibauza/final-project/rest-service/endpoints"
	"github.com/javibauza/final-project/rest-service/repository"
	"github.com/javibauza/final-project/rest-service/service"
	"github.com/javibauza/final-project/rest-service/transport"
)

func main() {
	var grpcUserServiceAddr *string

	env := os.Getenv("ENV")
	if env == "cluster" {
		grpcUserServiceAddr = flag.String("addr", "user-grpc-service-service.default.svc.cluster.local:50051", "The gprcUserServer address in the format of host:port")
	} else {
		grpcUserServiceAddr = flag.String("addr", "localhost:50051", "The gprcUserServer address in the format of host:port")
	}

	var (
		httpAddr = flag.String("http", ":8080", "http listen address")
	)
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger,
			"service", "httpService",
			"time:", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}

	level.Info(logger).Log("msg", "http service started")
	defer level.Info(logger).Log("msg", "http service ended")

	flag.Parse()

	var err error
	var grpcUserServiceConn *grpc.ClientConn
	{
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		grpcUserServiceConn, err = grpc.Dial(*grpcUserServiceAddr, opts...)
		if err != nil {
			level.Error(logger).Log("exit", err)
			os.Exit(-1)
		}
	}

	var srv service.Service
	{
		repository := repository.NewUserRepo(grpcUserServiceConn, logger)
		srv = service.NewService(repository, logger)
	}

	errChan := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	endpoints := endpoints.MakeEndpoints(srv)
	go func() {
		httpHandler := transport.NewHTTPServer(endpoints, logger)
		errChan <- http.ListenAndServe(*httpAddr, httpHandler)
	}()

	level.Error(logger).Log("exit", <-errChan)
}
