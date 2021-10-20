package configs

type Etcd struct {
	Title   string `mapstructure:"Title" json:"Title" yaml:"Title"`
	Address string `mapstructure:"Address" json:"Address" yaml:"Address"`
}
