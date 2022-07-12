package myddlmaker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
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
	var buf bytes.Buffer
	if err := m.parse(); err != nil {
		return err
	}

	buf.WriteString("SET foreign_key_checks=0;\n")
	for _, table := range m.tables {
		m.generateTable(&buf, table)
	}

	buf.WriteString("SET foreign_key_checks=1;\n")

	if _, err := buf.WriteTo(w); err != nil {
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

func (m *Maker) generateTable(w io.Writer, table *table) {
	fmt.Fprintf(w, "DROP TABLE IF EXISTS %s;\n\n", quote(table.Name))
	fmt.Fprintf(w, "CREATE TABLE %s (\n", quote(table.Name))
	for _, col := range table.Columns {
		m.generateColumn(w, col)
	}
	m.generateIndex(w, table)
	fmt.Fprintf(w, "    PRIMARY KEY (%s)\n", strings.Join(quoteAll(table.PrimaryKey.columns), ", "))
	fmt.Fprintf(w, ") ENGINE=InnoDB DEFAULT CHARACTER SET = 'utf8mb4';\n\n")
}

func (m *Maker) generateColumn(w io.Writer, col *column) {
	io.WriteString(w, "    ")
	io.WriteString(w, quote(col.Name))
	io.WriteString(w, " ")
	io.WriteString(w, col.Type)
	if col.Size != 0 {
		fmt.Fprintf(w, "(%d)", col.Size)
	}
	if col.Unsigned {
		io.WriteString(w, " unsigned")
	}
	if col.Null {
		io.WriteString(w, " NULL")
	} else {
		io.WriteString(w, " NOT NULL")
	}
	io.WriteString(w, ",\n")
}

func (m *Maker) generateIndex(w io.Writer, table *table) {
	for _, idx := range table.Indexes {
		switch idx := idx.(type) {
		case *index:
			io.WriteString(w, "    INDEX ")
			io.WriteString(w, quote(idx.Name))
			io.WriteString(w, " (")
			io.WriteString(w, strings.Join(quoteAll(idx.Columns), ", "))
			io.WriteString(w, "),\n")
		default:
			panic("must not reach")
		}
	}
}

func quote(s string) string {
	var buf strings.Builder
	// Strictly speaking, we need to count the number of backquotes in s.
	// However, in many cases, s doesn't include backquotes.
	buf.Grow(len(s) + len("``"))

	buf.WriteByte('`')
	for _, r := range s {
		if r == '`' {
			buf.WriteByte('`')
		}
		buf.WriteRune(r)
	}
	buf.WriteByte('`')
	return buf.String()
}

func quoteAll(strings []string) []string {
	ret := make([]string, len(strings))
	for i, s := range strings {
		ret[i] = quote(s)
	}
	return ret
}

type PrimaryKey struct {
	columns []string
}

type primaryKey interface {
	PrimaryKey() *PrimaryKey
}

func NewPrimaryKey(field ...string) *PrimaryKey {
	return &PrimaryKey{
		columns: field,
	}
}
