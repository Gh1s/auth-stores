package tools

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ViperKeyEnvironment        = "environment"
	ViperKeyEnvironmentDefault = "development"

	ViperKeyLogLevel  = "logLevel"
	ViperKeyLogFormat = "logFormat"
	ViperKeyLogTz     = "logTz"
	ViperKeyLogTf     = "logTf"

	ViperKeyListenProtocol = "listen.protocol"
	ViperKeyListenPort     = "listen.port"
	ViperKeyListenTLSCert  = "listen.tls.cert"
	ViperKeyListenTLSKey   = "listen.tls.key"
)

func init() {
	// Define a default global logger.
	zap.ReplaceGlobals(zap.New(createConsoleLogger()))
}

var logLevel zapcore.Level

func createConsoleLogger() zapcore.Core {
	// Configures a console logger that outputs logs in a human readable format.
	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= logLevel
		}),
	)
}

// LogLevel returns the log level.
func LogLevel() zapcore.Level {
	return logLevel
}

// InitConfig initializes the config.
func InitConfig(cfgName string, cfgType string, cfgPath string) {
	// Define defaults.
	viper.SetDefault(ViperKeyEnvironment, ViperKeyEnvironmentDefault)

	// Setup environment variables.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Setup mandatory config file.
	viper.SetConfigName(cfgName)
	viper.SetConfigType(cfgType)
	viper.AddConfigPath(cfgPath)

	if err := viper.ReadInConfig(); err != nil {
		zap.L().Fatal(
			"Could not use the config file.",
			zap.Error(err),
			zap.String("config", viper.ConfigFileUsed()),
		)
	}

	zap.L().Sugar().Infof("Hosting environment: %s", viper.GetString(ViperKeyEnvironment))

	// Setup environment config file.
	viper.SetConfigName(fmt.Sprintf("%s.%s", cfgName, viper.GetString(ViperKeyEnvironment)))

	if err := viper.MergeInConfig(); err != nil {
		zap.L().Warn(
			"Could not use the environment config file.",
			zap.Error(err),
			zap.String("config", cfgName),
			zap.String("environment", viper.GetString(ViperKeyEnvironment)),
		)
	}

	// Setup the log level.
	if err := logLevel.UnmarshalText([]byte(viper.GetString(ViperKeyLogLevel))); err != nil {
		zap.L().Warn(
			"Could not change the logging level based on configuration.",
			zap.Error(err),
		)
	}

	// TODO: Configures a Kafka or logstash logger.
	// TODO: Reconfigure the global logger.
}

// CreateTLSConfig creates an instance of *tls.Config with the X509 certificate found in the configuration.
func CreateTLSConfig() *tls.Config {
	certFile := viper.GetString(ViperKeyListenTLSCert)
	keyFile := viper.GetString(ViperKeyListenTLSKey)
	zap.L().Info(
		"Loading the X509 key pair.",
		zap.String("cert", certFile),
		zap.String("key", keyFile),
	)
	x509, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		zap.L().Fatal(
			"Could not load the X509 key pair.",
			zap.Error(err),
			zap.String("cert", certFile),
			zap.String("key", keyFile),
		)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{x509},
		ClientAuth:   tls.NoClientCert,
	}
	return tlsConfig
}

// CreateNetworkListner creates a network listener on the default protocol and port found in the configuration.
func CreateNetworkListner() net.Listener {
	protocol := viper.GetString(ViperKeyListenProtocol)
	port := viper.GetInt(ViperKeyListenPort)
	addr := fmt.Sprintf(":%d", port)
	zap.L().Info(
		"Starting the network listener.",
		zap.String("protocol", protocol),
		zap.Int("port", port),
	)
	lis, err := net.Listen(protocol, addr)
	if err != nil {
		zap.L().Fatal(
			"Could not listen on the configured network interface.",
			zap.Error(err),
			zap.String("protocol", protocol),
			zap.String("network", addr),
		)
	}
	return lis
}

// CreateNetworkListnerWithTLS creates a network listener on the default protocol and port found in the configuration, with the provided *tls.Config.
func CreateNetworkListnerWithTLS(tlsConfig *tls.Config) net.Listener {
	protocol := viper.GetString(ViperKeyListenProtocol)
	port := viper.GetInt(ViperKeyListenPort)
	addr := fmt.Sprintf(":%d", port)
	zap.L().Info(
		"Starting the network listener.",
		zap.String("protocol", protocol),
		zap.Int("port", port),
	)
	lis, err := tls.Listen(protocol, addr, tlsConfig)
	if err != nil {
		zap.L().Fatal(
			"Could not listen on the configured network interface.",
			zap.Error(err),
			zap.String("protocol", protocol),
			zap.String("network", addr),
		)
	}
	return lis
}
