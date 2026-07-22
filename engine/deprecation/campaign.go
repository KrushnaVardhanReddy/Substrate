package deprecation

type DeprecationConfig struct {
	Endpoint   string `yaml:"endpoint"`
	SunsetDate string `yaml:"sunset_date"`
}
