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
	fmt.Fprintf(w, "DROP TABLE IF EXISTS %s;\n\n", quote(table.name))
	fmt.Fprintf(w, "CREATE TABLE %s (\n", quote(table.name))
	for _, col := range table.columns {
		m.generateColumn(w, col)
	}
	m.generateIndex(w, table)
	fmt.Fprintf(w, "    PRIMARY KEY (%s)\n", strings.Join(quoteAll(table.primaryKey.columns), ", "))
	fmt.Fprintf(w, ") ENGINE=InnoDB DEFAULT CHARACTER SET = 'utf8mb4';\n\n")
}

func (m *Maker) generateColumn(w io.Writer, col *column) {
	io.WriteString(w, "    ")
	io.WriteString(w, quote(col.name))
	io.WriteString(w, " ")
	io.WriteString(w, col.typ)
	if col.size != 0 {
		fmt.Fprintf(w, "(%d)", col.size)
	}
	if col.unsigned {
		io.WriteString(w, " unsigned")
	}
	if col.null {
		io.WriteString(w, " NULL")
	} else {
		io.WriteString(w, " NOT NULL")
	}
	if col.autoIncr {
		io.WriteString(w, " AUTO_INCREMENT")
	}
	io.WriteString(w, ",\n")
}

func (m *Maker) generateIndex(w io.Writer, table *table) {
	for _, idx := range table.indexes {
		io.WriteString(w, "    INDEX ")
		io.WriteString(w, quote(idx.name))
		io.WriteString(w, " (")
		io.WriteString(w, strings.Join(quoteAll(idx.columns), ", "))
		io.WriteString(w, ")")
		if idx.comment != "" {
			io.WriteString(w, " COMMENT ")
			io.WriteString(w, stringQuote(idx.comment))
		}
		io.WriteString(w, ",\n")
	}

	for _, idx := range table.uniqueIndexes {
		io.WriteString(w, "    UNIQUE ")
		io.WriteString(w, quote(idx.name))
		io.WriteString(w, " (")
		io.WriteString(w, strings.Join(quoteAll(idx.columns), ", "))
		io.WriteString(w, ")")
		if idx.comment != "" {
			io.WriteString(w, " COMMENT ")
			io.WriteString(w, stringQuote(idx.comment))
		}
		io.WriteString(w, ",\n")
	}

	for _, idx := range table.foreignKeys {
		io.WriteString(w, "    CONSTRAINT ")
		io.WriteString(w, quote(idx.name))
		io.WriteString(w, " FOREIGN KEY (")
		io.WriteString(w, strings.Join(quoteAll(idx.columns), ", "))
		io.WriteString(w, ") REFERENCES ")
		io.WriteString(w, quote(idx.table))
		io.WriteString(w, " (")
		io.WriteString(w, strings.Join(quoteAll(idx.references), ", "))
		io.WriteString(w, ")")
		if idx.onUpdate != "" {
			io.WriteString(w, " ON UPDATE ")
			io.WriteString(w, string(idx.onUpdate))
		}
		if idx.onDelete != "" {
			io.WriteString(w, " ON DELETE ")
			io.WriteString(w, string(idx.onDelete))
		}
		io.WriteString(w, ",\n")
	}
}

// quote quotes s with `s`.
func quote(s string) string {
	var buf strings.Builder
	// Strictly speaking, we need to count the number of back quotes in s.
	// However, in many cases, s doesn't include back quotes.
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

// escape sequence table
// https://dev.mysql.com/doc/refman/8.0/en/string-literals.html
var stringQuoter = strings.NewReplacer(
	"\x00", `\0`,
	"'", `\'`,
	`"`, `\"`,
	"\b", `\b`,
	"\n", `\n`,
	"\r", `\r`,
	"\t", `\t`,
	"\x1a", `\Z`,
	"\\", `\\`,
)

// stringQuote quotes s with 's'.
func stringQuote(s string) string {
	var buf strings.Builder
	// Strictly speaking, we need to count the number of quotes in s.
	// However, in many cases, s doesn't include quotes.
	buf.Grow(len(s) + len("''"))

	buf.WriteByte('\'')
	stringQuoter.WriteString(&buf, s)
	buf.WriteByte('\'')
	return buf.String()
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
