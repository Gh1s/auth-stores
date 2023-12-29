module csb.nc/auth/stores/grpc/ldap

go 1.15

replace csb.nc/auth/stores => ../../

require (
	csb.nc/auth/stores v0.0.0-00010101000000-000000000000
	github.com/go-ldap/ldap v3.0.3+incompatible
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.5.1 // indirect
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.33.2
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
)
