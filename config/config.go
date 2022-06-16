package config

type Config struct {
	Token  string `mapstructure:"token"`
	Group  string `mapstructure:"group"`
	Debug  bool   `mapstructure:"debug"`
	Logrus Logrus `mapstructure:"logrus"`
}

type Logrus struct {
	LogLvl int    `mapstructure:"log_level"`
	ToFile bool   `mapstructure:"to_file"`
	ToJson bool   `mapstructure:"to_json"`
	LogDir string `mapstructure:"log_dir"`
}
