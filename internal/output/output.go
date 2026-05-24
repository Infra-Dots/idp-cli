package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// Printer writes formatted data to stdout.
type Printer struct {
	Format Format
	Quiet  bool
	Out    io.Writer
}

// New creates a Printer for the given format string.
func New(format string, quiet bool) *Printer {
	f := Format(format)
	switch f {
	case FormatJSON, FormatYAML, FormatTable:
	default:
		f = FormatTable
	}
	return &Printer{Format: f, Quiet: quiet, Out: os.Stdout}
}

// Table renders rows as a table with headers.
func (p *Printer) Table(headers []string, rows [][]string) {
	t := tablewriter.NewWriter(p.Out)
	// Convert []string to []any for the new v1 API.
	hAny := make([]any, len(headers))
	for i, h := range headers {
		hAny[i] = h
	}
	t.Header(hAny...)
	for _, row := range rows {
		rAny := make([]any, len(row))
		for i, cell := range row {
			rAny[i] = cell
		}
		_ = t.Append(rAny...)
	}
	_ = t.Render()
}

// JSON prints v as indented JSON.
func (p *Printer) JSON(v any) error {
	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// YAML prints v as YAML.
func (p *Printer) YAML(v any) error {
	enc := yaml.NewEncoder(p.Out)
	enc.SetIndent(2)
	return enc.Encode(v)
}

// Print dispatches to the correct renderer based on p.Format.
// tableHeaders and tableRows are used for table format.
// v is used for json/yaml formats.
func (p *Printer) Print(v any, headers []string, rows [][]string) error {
	switch p.Format {
	case FormatJSON:
		return p.JSON(v)
	case FormatYAML:
		return p.YAML(v)
	default:
		p.Table(headers, rows)
		return nil
	}
}

// PrintID is used with --quiet to print only an identifier.
func (p *Printer) PrintID(id string) {
	fmt.Fprintln(p.Out, id)
}

// NewError returns a formatted error (for use in cobra RunE).
func NewError(msg string, args ...any) error {
	return fmt.Errorf(msg, args...)
}

// Err writes an error message to stderr.
func Err(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+msg+"\n", args...)
}

// Fatal writes an error to stderr and exits 1.
func Fatal(msg string, args ...any) {
	Err(msg, args...)
	os.Exit(1)
}
