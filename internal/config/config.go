package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type DatabaseType string

var (
	DatabaseTypeMySQL    DatabaseType = "mysql"
	DatabaseTypePostgres DatabaseType = "postgres"
)

const (
	// Name of the application
	AppName = "espressoapi-go"

	configNameJSON = "config.json"
	configNameYAML = "config.yaml"
	configNameEnv  = "config.env"
)

var (
	defaultServerAddr              = ":8080"
	defaultServerReadTimeout       = 30 * time.Second
	defaultServerReadHeaderTimeout = 10 * time.Second
	defaultServerWriteTimeout      = 30 * time.Second
	defaultServerMaxRequestSize    = int64(10 * 1024 * 1024) // 10MiB

	defaultLoggerLogLevel          = "info"
	defaultLoggerDurationFieldUnit = "ms"
	defaultLoggerFormat            = "json"

	defaultDatabaseType           = DatabaseTypeMySQL
	defaultDatabaseDatasourceName = "root:root@tcp(127.0.0.1:3306)/espresso-api?parseTime=true"
)

type App struct {
	// Address for the server to listen on
	ServerAddr string `json:"server_addr" yaml:"server_addr" mapstructure:"SERVER_ADDR"`

	// Maximum duration for the http server to read the entire request, including the body
	ServerReadTimeout time.Duration `json:"server_read_timeout" yaml:"server_read_timeout" mapstructure:"SERVER_READ_TIMEOUT"`

	// Amount of time the http server allow to read request headers
	ServerReadHeaderTimeout time.Duration `json:"server_read_header_timeout" yaml:"server_read_header_timeout" mapstructure:"SERVER_READ_HEADER_TIMEOUT"`

	// Maximum duration before the http server times out writes of the response
	ServerWriteTimeout time.Duration `json:"server_write_timeout" yaml:"server_write_timeout" mapstructure:"SERVER_WRITE_TIMEOUT"`

	// Maximum size of a client request, including headers and body
	ServerMaxRequestSize int64 `json:"server_max_request_size" yaml:"server_max_request_size" mapstructure:"SERVER_MAX_REQUEST_SIZE"`

	// Logger log level
	// Available: "trace", "debug", "info", "warn", "error", "fatal", "panic"
	// ref: https://pkg.go.dev/github.com/rs/zerolog@v1.26.1#pkg-variables
	LoggerLogLevel string `json:"logger_log_level" yaml:"logger_log_level" mapstructure:"LOGGER_LOG_LEVEL"`

	// Defines the unit for `time.Duration` type fields in the logger
	// Available: "ms", "millisecond", "s", "second"
	LoggerDurationFieldUnit string `json:"logger_duration_field_unit" yaml:"logger_duration_field_unit" mapstructure:"LOGGER_DURATION_FIELD_UNIT"`

	// Format of the logs
	LoggerFormat string `json:"logger_format" yaml:"logger_format" mapstructure:"LOGGER_FORMAT"`

	// Name of the database (driver) type to use
	// Available: "mysql", "postgres"
	DatabaseType DatabaseType `json:"database_type" yaml:"database_type" mapstructure:"DATABASE_TYPE"`

	// Address of the database
	// See:
	// - https://github.com/go-sql-driver/mysql/#usage for mysql syntax
	// - https://github.com/jackc/pgx for postgres syntax
	DatabaseDatasourceName string `json:"database_datasource_name" yaml:"database_datasource_name" mapstructure:"DATABASE_DATASOURCE_NAME"`
}

// New will retrieve the runtime configuration from either
// files or environment variables.
//
// Available configuration files are:
//
// * json
//
// * yaml
//
// * dotenv
func New() (*App, error) {
	app := App{}

	// Set default configurations
	app.setDefaults()

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	if _, fileErr := os.Stat(filepath.Join(".", configNameJSON)); fileErr == nil {
		viper.SetConfigType("json")
	} else if _, fileErr := os.Stat(filepath.Join(".", configNameYAML)); fileErr == nil {
		viper.SetConfigType("yaml")
	} else if _, fileErr := os.Stat(filepath.Join(".", configNameEnv)); fileErr == nil {
		viper.SetConfigType("env")
	}

	// When the error is viper.ConfigFileNotFoundError, we try to read from
	// environment variables
	err := viper.ReadInConfig()
	if err != nil {
		switch err.(type) {
		default:
			return nil, fmt.Errorf("error while loading config file: %s", err)
		case viper.ConfigFileNotFoundError:
			readConfigFromEnvVars(app)
		}
	}

	if err := viper.Unmarshal(&app); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the config into struct, %w", err)
	}

	if err := app.validateConfig(); err != nil {
		return nil, fmt.Errorf("failed to validate the config, %w", err)
	}

	return &app, nil
}

// validateConfig will make sure the provided configuration is valid
// by looking if the values are present when they are expected to be present
func (app *App) validateConfig() error {
	return nil
}

func readConfigFromEnvVars(c App) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindEnvs(c)
}

// ref: https://github.com/spf13/viper/issues/188#issuecomment-399884438
func bindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		switch v.Kind() {
		case reflect.Struct:
			bindEnvs(v.Interface(), append(parts, tv)...)
		default:
			viper.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}

func (config *App) setDefaults() {
	config.ServerAddr = defaultServerAddr
	config.ServerReadTimeout = defaultServerReadTimeout
	config.ServerReadHeaderTimeout = defaultServerReadHeaderTimeout
	config.ServerWriteTimeout = defaultServerWriteTimeout
	config.ServerMaxRequestSize = defaultServerMaxRequestSize

	config.LoggerLogLevel = defaultLoggerLogLevel
	config.LoggerDurationFieldUnit = defaultLoggerDurationFieldUnit
	config.LoggerFormat = defaultLoggerFormat

	config.DatabaseType = defaultDatabaseType
	config.DatabaseDatasourceName = defaultDatabaseDatasourceName
}
