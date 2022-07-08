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
	tables []*table
}

func New(config *Config) (*Maker, error) {
	return &Maker{}, nil
}

func (m *Maker) AddStructs(structs ...any) {
	for _, s := range structs {
		m.tables = append(m.tables, newTable(s))
	}
}

func (m *Maker) Generate() error {
	return nil
}
