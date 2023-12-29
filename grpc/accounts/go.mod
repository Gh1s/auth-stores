module csb.nc/auth/stores/grpc/accounts

go 1.15

replace csb.nc/auth/stores => ../../

require (
	csb.nc/auth/stores v0.0.0-00010101000000-000000000000
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.33.2
)
