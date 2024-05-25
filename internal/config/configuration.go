package config

import (
	"html/template"
	"os"
	"strings"

	"go.mau.fi/util/dbutil"
	"go.mau.fi/util/exerrors"
	"go.mau.fi/zeroconfig"
)

type ConfigFilenames []string

func (f *ConfigFilenames) String() string {
	return strings.Join(*f, ", ")
}

func (f *ConfigFilenames) Set(value string) error {
	*f = append(*f, value)
	return nil
}

type RecaptchaConfig struct {
	SiteKey   string `yaml:"site_key"`
	SecretKey string `yaml:"secret_key"`
}

type Configuration struct {
	secretKeyBytes []byte

	Database dbutil.Config `yaml:"database"`

	DevMode bool `yaml:"dev_mode"`

	Domain              string        `yaml:"domain"`
	SendgridAPIKey      string        `yaml:"sendgrid_api_key"`
	HealthcheckURL      string        `yaml:"healthcheck_url"`
	HostedByHTML        template.HTML `yaml:"hosted_by_html"`
	RegistrationEnabled bool          `yaml:"registration_enabled"`

	JWTSecretKeyFile string `yaml:"jwt_secret_key_file"`
	JWTSecretKey     string `yaml:"jwt_secret_key"`

	Recaptcha RecaptchaConfig `yaml:"recaptcha"`

	Logging zeroconfig.Config `yaml:"logging"`
}

func (c *Configuration) ReadSecretKey() []byte {
	if len(c.JWTSecretKey) > 0 {
		return []byte(c.JWTSecretKey)
	}

	if len(c.secretKeyBytes) == 0 {
		c.secretKeyBytes = exerrors.Must(os.ReadFile(c.JWTSecretKeyFile))
	}
	return c.secretKeyBytes
}
