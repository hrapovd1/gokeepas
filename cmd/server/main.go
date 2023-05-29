package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hrapovd1/gokeepas/internal/config"
	"github.com/hrapovd1/gokeepas/internal/crypto"
	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/hrapovd1/gokeepas/internal/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// create server config
	srvConfig, err := config.NewServerConf()
	if err != nil {
		log.Fatalf("error create server configuration: %v", err)
	}

	// create logger
	logConfig := zap.NewProductionConfig()
	logConfig.EncoderConfig.TimeKey = "timestamp"
	logConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	logConfig.Level.SetLevel(srvConfig.LogLevel)
	logger, err := logConfig.Build()
	if err != nil {
		log.Fatalf("when create zap logger got error: %v", err)
	}

	// generate in-memory tls certificate
	cert, err := crypto.GenX509KeyPair()
	if err != nil {
		logger.Fatal(err.Error())
	}

	// prepare grpc server
	// tls based credentials
	creds := credentials.NewServerTLSFromCert(&cert)
	// gokeepas app
	gkp, err := server.NewKeepPasSrv(logger, *srvConfig)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer func(l *zap.Logger) {
		if err := gkp.Stor.Close(); err != nil {
			l.Fatal(err.Error())
		}
	}(logger)
	// grpc server
	srv := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(gkp.AuthInterceptor))
	// register app on the server
	pb.RegisterKeepPasServer(srv, gkp)

	// prepare server shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	// check db connection and server master key
	if err := gkp.Stor.Ping(ctx, srvConfig.ServerKey); err != nil {
		logger.Fatal(err.Error())
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(c context.Context, w *sync.WaitGroup, s *grpc.Server, l *zap.Logger) {
		defer w.Done()
		<-c.Done()
		l.Info("got signal to stop")
		s.GracefulStop()

	}(ctx, &wg, srv, logger)

	// run server
	listen, err := net.Listen("tcp", srvConfig.ServerAddr)
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Sugar().Info("server started")
	if err := srv.Serve(listen); err != http.ErrServerClosed && err != nil {
		logger.Fatal("failed to listen: " + err.Error())
	}

	wg.Wait()
	logger.Info("server stoped gracefully")
}
