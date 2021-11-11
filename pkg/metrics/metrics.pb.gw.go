// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: metrics/metrics.proto

/*
Package metrics is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package metrics

import (
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Suppress "imported and not used" errors
var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = metadata.Join

func request_Metrics_ProjectByID_0(ctx context.Context, marshaler runtime.Marshaler, client MetricsClient, req *http.Request, pathParams map[string]string) (Metrics_ProjectByIDClient, runtime.ServerMetadata, error) {
	var protoReq ProjectByIDRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["namespace"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "namespace")
	}

	protoReq.Namespace, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "namespace", err)
	}

	val, ok = pathParams["pod"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "pod")
	}

	protoReq.Pod, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "pod", err)
	}

	stream, err := client.ProjectByID(ctx, &protoReq)
	if err != nil {
		return nil, metadata, err
	}
	header, err := stream.Header()
	if err != nil {
		return nil, metadata, err
	}
	metadata.HeaderMD = header
	return stream, metadata, nil

}

// RegisterMetricsHandlerServer registers the http handlers for service Metrics to "mux".
// UnaryRPC     :call MetricsServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterMetricsHandlerFromEndpoint instead.
func RegisterMetricsHandlerServer(ctx context.Context, mux *runtime.ServeMux, server MetricsServer) error {

	mux.Handle("GET", pattern_Metrics_ProjectByID_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		err := status.Error(codes.Unimplemented, "streaming calls are not yet supported in the in-process transport")
		_, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
		return
	})

	return nil
}

// RegisterMetricsHandlerFromEndpoint is same as RegisterMetricsHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterMetricsHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterMetricsHandler(ctx, mux, conn)
}

// RegisterMetricsHandler registers the http handlers for service Metrics to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterMetricsHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterMetricsHandlerClient(ctx, mux, NewMetricsClient(conn))
}

// RegisterMetricsHandlerClient registers the http handlers for service Metrics
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "MetricsClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "MetricsClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "MetricsClient" to call the correct interceptors.
func RegisterMetricsHandlerClient(ctx context.Context, mux *runtime.ServeMux, client MetricsClient) error {

	mux.Handle("GET", pattern_Metrics_ProjectByID_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/.Metrics/ProjectByID", runtime.WithHTTPPathPattern("/api/metrics/namespace/{namespace}/pods/{pod}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_Metrics_ProjectByID_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_Metrics_ProjectByID_0(ctx, mux, outboundMarshaler, w, req, func() (proto.Message, error) { return resp.Recv() }, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_Metrics_ProjectByID_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 1, 0, 4, 1, 5, 2, 2, 3, 1, 0, 4, 1, 5, 4}, []string{"api", "metrics", "namespace", "pods", "pod"}, ""))
)

var (
	forward_Metrics_ProjectByID_0 = runtime.ForwardResponseStream
)