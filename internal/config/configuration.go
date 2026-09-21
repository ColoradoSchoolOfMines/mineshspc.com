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

type HomepageConfig struct {
	H2Text                 string        `yaml:"h2_text"`
	HeroText               template.HTML `yaml:"hero_text"`
	ScheduleDate           string        `yaml:"schedule_date"`
	RegistrationDeadline   string        `yaml:"registration_deadline"`
	LateRegistration       bool          `yaml:"late_registration"`
	OpenDivisionComingSoon bool          `yaml:"open_division_coming_soon"`
	OpenDivisionURL        string        `yaml:"open_division_url"`
}

type ArchiveLink struct {
	URL   string `yaml:"url"`
	Title string `yaml:"title"`
}

type ArchiveWinningTeam struct {
	Place    string `yaml:"place"`
	Name     string `yaml:"name"`
	School   string `yaml:"school"`
	Location string `yaml:"location"`
}

type ArchiveCompetitionResult struct {
	Name      string               `yaml:"name"`
	Shortname string               `yaml:"shortname"`
	Teams     []ArchiveWinningTeam `yaml:"teams"`
}

type ArchiveYearInfo struct {
	Year            int                        `yaml:"year"`
	RecapParagraphs []string                   `yaml:"recap_paragraphs"`
	Links           []ArchiveLink              `yaml:"links"`
	Results         []ArchiveCompetitionResult `yaml:"results"`
}

type Configuration struct {
	secretKeyBytes []byte

	Database dbutil.Config `yaml:"database"`

	DevMode bool `yaml:"dev_mode"`

	Domain              string         `yaml:"domain"`
	SendgridAPIKey      string         `yaml:"sendgrid_api_key"`
	HealthcheckURL      string         `yaml:"healthcheck_url"`
	HostedByHTML        template.HTML  `yaml:"hosted_by_html"`
	RegistrationEnabled bool           `yaml:"registration_enabled"`
	Homepage            HomepageConfig `yaml:"homepage"`

	Archive []ArchiveYearInfo `yaml:"archive"`

	AdminEmails []string `yaml:"admin_emails"`

	JWTSecretKeyFile string `yaml:"jwt_secret_key_file"`
	JWTSecretKey     string `yaml:"jwt_secret_key"`

	Recaptcha RecaptchaConfig `yaml:"recaptcha"`

	Logging zeroconfig.Config `yaml:"logging"`
}

func (c *Configuration) IsAdminEmail(email string) bool {
	for _, e := range c.AdminEmails {
		if e == email {
			return true
		}
	}
	return false
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
