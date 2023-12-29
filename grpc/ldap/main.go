package main

import (
	"context"

	"csb.nc/auth/stores/grpc/ldap/svc"
	"csb.nc/auth/stores/tools"
	"csb.nc/auth/stores/users"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	cfgName = "config"
	cfgType = "yaml"
	cfgPath = "."
)

type server struct {
	users.UnimplementedUserServer
}

func (s *server) Authenticate(ctx context.Context, req *users.AuthRequest) (*users.AuthResponse, error) {
	return svc.Authenticate(req), nil
}

func (s server) FindClaims(ctx context.Context, req *users.ClaimsRequest) (*users.ClaimsResponse, error) {
	return svc.FindClaims(req), nil
}

func (s server) SearchClaims(ctx context.Context, req *users.SearchRequest) (*users.SearchResponse, error) {
	return svc.SearchClaims(req), nil
}

func init() {
	tools.InitConfig(cfgName, cfgType, cfgPath)
}

func main() {
	defer zap.L().Sync()

	lis := tools.CreateNetworkListner()
	defer lis.Close()

	tlsConfig := tools.CreateTLSConfig()
	tlsCreds := credentials.NewTLS(tlsConfig)

	zap.L().Info("Buidling the gRPC server.")
	srv := grpc.NewServer(
		grpc.Creds(tlsCreds),
		grpc.StreamInterceptor(
			grpcmiddleware.ChainStreamServer(
				grpczap.StreamServerInterceptor(zap.L()),
			),
		),
		grpc.UnaryInterceptor(
			grpcmiddleware.ChainUnaryServer(
				grpczap.UnaryServerInterceptor(zap.L()),
			),
		),
	)
	defer srv.Stop()
	users.RegisterUserServer(srv, &server{})

	zap.L().Info("Starting the gRPC server.")
	if err := srv.Serve(lis); err != nil {
		zap.L().Fatal(
			"Could not start the gRPC server",
			zap.Error(err),
		)
	}
}
