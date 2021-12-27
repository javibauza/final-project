package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	_ "github.com/mattn/go-sqlite3"

	"github.com/javibauza/final-project/grpc-service/endpoints"
	"github.com/javibauza/final-project/grpc-service/pb"
	"github.com/javibauza/final-project/grpc-service/repository"
	"github.com/javibauza/final-project/grpc-service/service"
	"github.com/javibauza/final-project/grpc-service/transport"
	"google.golang.org/grpc"
)

func main() {
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

		db, err = sql.Open("sqlite3", "./users.db")
		if err != nil {
			level.Error(logger).Log("exit", err)
			os.Exit(-1)
		}
	}

	flag.Parse()

	var srv service.Service
	{
		repository := repository.NewRepo(db, logger)
		srv = service.NewService(repository, logger)
	}

	endpoints := endpoints.MakeEndpoints(srv)
	grpcServer := transport.NewGRPCServer(endpoints, logger)

	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	grpcListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}

	go func() {
		baseServer := grpc.NewServer()
		pb.RegisterUserServiceServer(baseServer, grpcServer)
		level.Info(logger).Log("msg", "Server started successfully")
		baseServer.Serve(grpcListener)
	}()

	level.Error(logger).Log("exit", <-errs)
}
