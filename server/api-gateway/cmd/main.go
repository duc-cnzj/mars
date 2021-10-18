package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/duc-cnzj/mars/pkg/cluster"
	"github.com/duc-cnzj/mars/pkg/gitlab"
	"github.com/duc-cnzj/mars/pkg/mars"
	"github.com/duc-cnzj/mars/pkg/namespace"
	"github.com/duc-cnzj/mars/pkg/project"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

var endpoint = "localhost:9999"

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := http.NewServeMux()

	gmux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseEnumNumbers:  true,
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var err error
	err = namespace.RegisterNamespaceHandlerFromEndpoint(ctx, gmux, endpoint, opts)
	fatalError(err)
	err = cluster.RegisterClusterHandlerFromEndpoint(ctx, gmux, endpoint, opts)
	fatalError(err)
	err = gitlab.RegisterGitlabHandlerFromEndpoint(ctx, gmux, endpoint, opts)
	fatalError(err)
	err = mars.RegisterMarsHandlerFromEndpoint(ctx, gmux, endpoint, opts)
	fatalError(err)
	err = project.RegisterProjectHandlerFromEndpoint(ctx, gmux, endpoint, opts)
	fatalError(err)

	mux.Handle("/", gmux)

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	s := &http.Server{
		Addr:    ":4000",
		Handler: routeLogger(allowCORS(mux)),
	}

	go func() {
		log.Println("api-gateway start at: ", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			log.Println("error: ", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	timeout, cancelFunc := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancelFunc()
	s.Shutdown(timeout)
	log.Println("api-gateway shutdown")
}

func fatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "X-Requested-With"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	return
}

func routeLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(t time.Time) {
			log.Printf("method: %v, url: %v, use %v", r.Method, r.URL, time.Since(t))
		}(time.Now())
		h.ServeHTTP(w, r)
	})
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
