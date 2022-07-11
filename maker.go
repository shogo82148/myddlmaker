package myddlmaker

import (
	"fmt"
	"io"
	"os"
)

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
	config  *Config
	structs []any
	tables  []*table
}

func New(config *Config) (*Maker, error) {
	return &Maker{
		config: config,
	}, nil
}

func (m *Maker) AddStructs(structs ...any) {
	m.structs = append(m.structs, structs...)
}

// GenerateFile opens
func (m *Maker) GenerateFile() error {
	f, err := os.Create(m.config.OutFilePath)
	if err != nil {
		return fmt.Errorf("myddlmaker: failed to open %q: %w", m.config.OutFilePath, err)
	}
	defer f.Close()

	if err := m.Generate(f); err != nil {
		return fmt.Errorf("myddlmaker: failed to generate ddl: %w", err)
	}

	return f.Close()
}

func (m *Maker) Generate(w io.Writer) error {
	if err := m.parse(); err != nil {
		return err
	}
	return nil
}

func (m *Maker) parse() error {
	m.tables = make([]*table, len(m.structs))
	for i, s := range m.structs {
		tbl, err := newTable(s)
		if err != nil {
			return fmt.Errorf("myddlmaker: failed to parse: %w", err)
		}
		m.tables[i] = tbl
	}
	return nil
}
