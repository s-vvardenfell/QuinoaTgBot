package config

type Config struct {
	ServerHost string `mapstructure:"server_addr"`
	ServerPort string `mapstructure:"server_port"`
	Token      string `mapstructure:"token"`
	Group      string `mapstructure:"group"`
	Debug      bool   `mapstructure:"debug"`
	Timeout    int    `mapstructure:"timeout"`
	Logrus     Logrus `mapstructure:"logrus"`
}

type Logrus struct {
	LogLvl int    `mapstructure:"log_level"`
	ToFile bool   `mapstructure:"to_file"`
	ToJson bool   `mapstructure:"to_json"`
	LogDir string `mapstructure:"log_dir"`
}
