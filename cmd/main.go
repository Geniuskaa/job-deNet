package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"net"
	"os"
	"os/signal"
	"syscall"
	"test_task/internal/server"
)

const (
	defaultPort = "8080"
	defaultHost = "0.0.0.0"
)

func main() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP, syscall.SIGINT,
		syscall.SIGQUIT, syscall.SIGTERM)
	defer stop()

	port, ok := os.LookupEnv("APP_PORT")
	if !ok {
		port = defaultPort
	}

	host, ok := os.LookupEnv("APP_HOST")
	if !ok {
		host = defaultHost
	}

	httpSrv := applicationStart(mainCtx, net.JoinHostPort(host, port))

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		fmt.Println("HTTP server started!")
		return httpSrv.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		fmt.Println("HTTP server is shut down.")
		return httpSrv.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}
	fmt.Println("Servers were gracefully shut down.")
}

func applicationStart(ctx context.Context, addr string) *server.Server {
	logger := loggerInit()
	restServer := server.NewRestServer(ctx, addr, logger)

	return restServer
}

func loggerInit() *zap.SugaredLogger {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	core := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.ErrorLevel)

	sugarLogger := zap.New(core).Sugar()

	return sugarLogger
}
