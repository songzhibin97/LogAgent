package configs

type SystemConfig struct {
	MaxKafkaBuffer int `mapstructure:"MaxKafkaBuffer" json:"MaxKafkaBuffer" yaml:"MaxKafkaBuffer"`
	MaxWatchBuffer int `mapstructure:"MaxWatchBuffer" json:"MaxWatchBuffer" yaml:"MaxWatchBuffer"`
	MaxSentinel    int `mapstructure:"MaxSentinel" json:"MaxSentinel" yaml:"MaxSentinel"`
}
