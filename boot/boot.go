package boot

func Initialize(path ...string) {
	Viper.Initialize(path...)
	Zap.Initialize()
}
