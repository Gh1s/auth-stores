package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"csb.nc/auth/stores/tools"
	"csb.nc/auth/stores/users"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
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

type user struct {
	ID           string            `id:"id"`
	Username     string            `json:"username"`
	PasswordHash string            `json:"password_hash"`
	Claims       map[string]string `json:"claims"`
}

const (
	UserNotFound = iota + 1
	UsersMissing
	InvalidPassword

	userNotFoundError = "User not found"
)

func (s *server) Authenticate(ctx context.Context, req *users.AuthRequest) (*users.AuthResponse, error) {
	resp := &users.AuthResponse{}

	u, err := findUser(req.Username, users.IdentifierType_USER_NAME)
	if err != nil {
		resp.Error = UserNotFound
	} else {
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(req.Password)))
		if strings.EqualFold(hash, u.PasswordHash) {
			resp.Succeeded = true
			resp.Subject = u.ID
		} else {
			resp.Error = InvalidPassword
		}
	}

	return resp, nil
}

func (s server) FindClaims(ctx context.Context, req *users.ClaimsRequest) (*users.ClaimsResponse, error) {
	resp := &users.ClaimsResponse{
		Claims: make(map[string]string, len(req.Claims)),
	}

	u, err := findUser(req.Identifier, req.IdentifierType)
	if err != nil {
		resp.Error = UserNotFound
	} else {
		for _, k := range req.Claims {
			resp.Claims[k] = u.Claims[k]
		}
		resp.Succeeded = true
	}

	return resp, nil
}

func (s server) SearchClaims(ctx context.Context, req *users.SearchRequest) (*users.SearchResponse, error) {
	resp := &users.SearchResponse{}

	usrs, err := getUsers()
	if err != nil {
		resp.Error = UsersMissing
	}

	resp.Results = make([]*users.SearchResponseResult, 0)
	search := strings.ToLower(req.Search)
	for _, u := range usrs {
		if strings.Contains(strings.ToLower(u.Username), search) || strings.Contains(strings.ToLower(u.Claims["name"]), search) {
			item := &users.SearchResponseResult{
				Properties: make(map[string]string, len(req.Claims)),
			}
			for _, k := range req.Claims {
				item.Properties[k] = u.Claims[k]
			}
			resp.Results = append(resp.Results, item)
		}
	}
	resp.Succeeded = true

	return resp, nil
}

func getUsers() ([]user, error) {
	jsonData, err := ioutil.ReadFile("users.json")
	if err != nil {
		return []user{}, err
	}

	var usrs []user
	err = json.Unmarshal(jsonData, &usrs)
	if err != nil {
		return []user{}, err
	}

	return usrs, nil
}

func findUser(identifier string, identifierType users.IdentifierType) (*user, error) {
	usrs, err := getUsers()
	if err != nil {
		return nil, err
	}

	for _, u := range usrs {
		switch identifierType {
		case users.IdentifierType_SUBJECT:
			if strings.EqualFold(identifier, u.ID) {
				return &u, nil
			}
		case users.IdentifierType_USER_NAME:
			if strings.EqualFold(identifier, u.Username) {
				return &u, nil
			}
		}
	}

	return nil, fmt.Errorf(userNotFoundError)
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
			grpc_middleware.ChainStreamServer(
				grpc_zap.StreamServerInterceptor(zap.L()),
			),
		),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_zap.UnaryServerInterceptor(zap.L()),
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
