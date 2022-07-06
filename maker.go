package myddlmaker

type Config struct {
	DB          *DBConfig
	OutFilePath string
}

type DBConfig struct {
	Driver  string
	Engine  string
	Charset string
}

type Maker struct {
}

func New(config *Config) (*Maker, error) {
	return &Maker{}, nil
}

func (m *Maker) AddStructs(s ...any) {}

func (m *Maker) Generate() error {
	return nil
}
