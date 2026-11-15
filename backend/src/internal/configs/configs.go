package configs

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Firebase FirebaseConfig
}

type AppConfig struct {
	Environment            string `mapstructure:"APP_ENV"`
	Port                   string `mapstructure:"PORT"`
	ServiceName            string `mapstructure:"SERVICE_NAME"`
	Timezone               string `mapstructure:"APP_TIMEZONE"`
	CORSAllowedOrigins     string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	FrontendBaseURL        string `mapstructure:"FRONTEND_BASE_URL"`
	SchedulePollingEnabled bool   `mapstructure:"SCHEDULE_POLLING_ENABLED"`
}

type AuthConfig struct {
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `mapstructure:"GOOGLE_REDIRECT_URL"`
	JWTSecret          string `mapstructure:"JWT_SECRET"`
	UserInfoURL        string `mapstructure:"GOOGLE_USERINFO_URL"` // override-able for testing
	AdminEmails        string `mapstructure:"ADMIN_EMAILS"`
}

type DatabaseConfig struct {
	URL     string `mapstructure:"DATABASE_URL"`
	Driver  string `mapstructure:"DB_DRIVER"`
	SSLMode string `mapstructure:"DB_SSLMODE"`
	Host    string `mapstructure:"DB_HOST"`
	DbName  string `mapstructure:"DB_NAME"`
	Pwd     string `mapstructure:"DB_PASSWORD"`
	Port    string `mapstructure:"DB_PORT"`
	User    string `mapstructure:"DB_USER"`
	Backoff string `mapstructure:"DB_RETRY_BACKOFF"`
}

type FirebaseConfig struct {
	CredentialFile string `mapstructure:"FIREBASE_CREDENTIAL_FILE"`
	StorageBucket  string `mapstructure:"FIREBASE_STORAGE_BUCKET"`
}

func (c *Config) ValidateRuntime() error {
	if strings.ToLower(c.App.Environment) != "prod" {
		return nil
	}

	required := []struct {
		key   string
		value string
	}{
		{"FRONTEND_BASE_URL", c.App.FrontendBaseURL},
		{"GOOGLE_CLIENT_ID", c.Auth.GoogleClientID},
		{"GOOGLE_CLIENT_SECRET", c.Auth.GoogleClientSecret},
		{"GOOGLE_REDIRECT_URL", c.Auth.GoogleRedirectURL},
		{"JWT_SECRET", c.Auth.JWTSecret},
		{"ADMIN_EMAILS", c.Auth.AdminEmails},
		{"FIREBASE_STORAGE_BUCKET", c.Firebase.StorageBucket},
	}

	var missing []string
	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			missing = append(missing, item.key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required prod configuration: %s", strings.Join(missing, ", "))
	}

	return nil
}

type Configuration struct {
	ConfigPath string
	FileName   string
	FileType   string
}

func LoadConfig(cfg Configuration) (*Config, error) {
	v := viper.New()
	v.SetConfigName(cfg.FileName)
	v.SetConfigType(cfg.FileType)
	v.AddConfigPath(cfg.ConfigPath)

	v.SetDefault("APP_ENV", "local")
	v.SetDefault("PORT", "8080")
	v.SetDefault("SERVICE_NAME", "tex-backend")
	v.SetDefault("APP_TIMEZONE", "Asia/Bangkok")
	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:4200,http://localhost:3000")
	v.SetDefault("FRONTEND_BASE_URL", "http://localhost:4200")
	v.SetDefault("SCHEDULE_POLLING_ENABLED", false)
	v.SetDefault("DB_DRIVER", "postgres")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_NAME", "tex_db")
	v.SetDefault("DB_PASSWORD", "password")
	v.SetDefault("DB_PORT", "6632")
	v.SetDefault("DB_USER", "admin")
	v.SetDefault("DB_RETRY_BACKOFF", "5s")

	v.AutomaticEnv()
	bindStructEnvKeys(v, reflect.TypeOf(AppConfig{}))
	bindStructEnvKeys(v, reflect.TypeOf(DatabaseConfig{}))
	bindStructEnvKeys(v, reflect.TypeOf(AuthConfig{}))
	bindStructEnvKeys(v, reflect.TypeOf(FirebaseConfig{}))

	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		v.OnConfigChange(func(_ fsnotify.Event) {
			log.Println("configuration file changed")
		})
		v.WatchConfig()
	}

	conf := &Config{}
	if err := v.Unmarshal(&conf.App); err != nil {
		return nil, fmt.Errorf("failed to decode app config: %w", err)
	}
	if err := v.Unmarshal(&conf.Database); err != nil {
		return nil, fmt.Errorf("failed to decode database config: %w", err)
	}
	if err := v.Unmarshal(&conf.Auth); err != nil {
		return nil, fmt.Errorf("failed to decode auth config: %w", err)
	}
	if err := v.Unmarshal(&conf.Firebase); err != nil {
		return nil, fmt.Errorf("failed to decode firebase config: %w", err)
	}

	return conf, nil
}

func bindStructEnvKeys(v *viper.Viper, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("mapstructure")
		if tag == "" {
			continue
		}
		_ = v.BindEnv(tag)
	}
}
