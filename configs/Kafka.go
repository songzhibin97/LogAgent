package configs

type Kafka struct {
	Address []string `mapstructure:"Address" json:"Address" yaml:"Address"`
}
