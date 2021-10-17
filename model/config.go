package model

type KafkaConfig struct {
	Address []string
}

type EsConfig struct {
	Address string
}

type SystemConfig struct {
	MaxKafkaBuffer int
	MaxWatchBuffer int
}

type Config struct {
	KafkaConfig  *KafkaConfig  `json:"kafka_config"`
	EsConfig     *EsConfig     `json:"es_config"`
	SystemConfig *SystemConfig `json:"system_config"`
}
