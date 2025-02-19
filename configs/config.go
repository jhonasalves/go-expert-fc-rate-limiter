package configs

import "github.com/spf13/viper"

type Conf struct {
	RateLimiterMaxIPRequests int    `mapstructure:"RATE_LIMITER_MAX_IP_REQUESTS"`
	RateLimiterBurst         int    `mapstructure:"RATE_LIMITER_BURST"`
	RedisHost                string `mapstructure:"REDIS_HOST"`
	RedisPort                int    `mapstructure:"REDIS_PORT"`
	RedisPassword            string `mapstructure:"REDIS_PASSWORD"`
	RedisDB                  int    `mapstructure:"REDIS_DB"`
}

func LoadConfig(path string) (*Conf, error) {
	var cfg *Conf
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}
	return cfg, err
}
