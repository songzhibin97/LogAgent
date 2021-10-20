package configs

type Server struct {
	Zap    Zap          `mapstructure:"Zap" json:"Zap" yaml:"Zap"`
	Kafka  Kafka        `mapstructure:"Kafka" json:"Kafka" yaml:"Kafka"`
	Etcd   Etcd         `mapstructure:"Etcd" json:"Etcd" yaml:"Etcd"`
	Es     EsConfig     `mapstructure:"EsConfig" json:"EsConfig" yaml:"EsConfig"`
	System SystemConfig `mapstructure:"SystemConfig" json:"SystemConfig" yaml:"SystemConfig"`
}
