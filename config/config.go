package config

type Config struct {
	DebugMode   bool
	SoFilePath  string
	CaseName    string
	Cpuprofile  string
	Memprofile  string
	FileStorage string
}

var Cfg = new(Config)
