package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/mattn/go-sqlite3"

	"github.com/javibauza/final-project/grpc-service/endpoints"
	"github.com/javibauza/final-project/grpc-service/proto"
	"github.com/javibauza/final-project/grpc-service/repository"
	"github.com/javibauza/final-project/grpc-service/service"
	"github.com/javibauza/final-project/grpc-service/transport"
	"google.golang.org/grpc"
)

var (
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:9090", "gRPC server endpoint")
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger,
			"service", "grpcUserService",
			"time:", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}
	level.Info(logger).Log("msg", "grpcUserService started")
	defer level.Info(logger).Log("msg", "grpcUserService ended")

	var db *sql.DB
	{
		var err error

		db, err = sql.Open("sqlite3", "./../users.db")
		if err != nil {
			level.Error(logger).Log("exit", err)
			os.Exit(-1)
		}
	}

	flag.Parse()

	var srv endpoints.Service
	{
		repository := repository.NewRepo(db, logger)
		srv = service.NewService(repository, logger)
	}

	endpoints := endpoints.MakeEndpoints(srv, logger)
	grpcServer := transport.NewGRPCServer(endpoints, logger)

	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	grpcListener, err := net.Listen("tcp", ":9090")
	if err != nil {
		level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}

	baseServer := grpc.NewServer()
	proto.RegisterUserServiceServer(baseServer, grpcServer)
	level.Info(logger).Log("msg", "Server started successfully")

	go baseServer.Serve(grpcListener)

	responseModifier := transport.MakeHTTPResponseModifier(logger)
	mux := runtime.NewServeMux(
		runtime.WithForwardResponseOption(responseModifier),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = proto.RegisterUserServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}

	http.Handle("/", mux)
	fs := http.FileServer(http.Dir("./../swagger"))
	http.Handle("/swagger/", http.StripPrefix("/swagger/", fs))
	http.ListenAndServe(":8080", nil)

	level.Error(logger).Log("exit", <-errs)

}
