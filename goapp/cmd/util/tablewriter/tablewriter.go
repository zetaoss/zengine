package tablewriter

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type TableWriter struct {
	w       *tabwriter.Writer
	headers []string
}

func New(out io.Writer, headers ...string) *TableWriter {
	tw := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	return &TableWriter{w: tw, headers: headers}
}

func (t *TableWriter) Header() error {
	if len(t.headers) == 0 {
		return nil
	}
	cols := make([]string, 0, len(t.headers))
	for _, h := range t.headers {
		cols = append(cols, strings.ToUpper(strings.TrimSpace(h)))
	}
	_, err := fmt.Fprintln(t.w, strings.Join(cols, "\t"))
	return err
}

func (t *TableWriter) Row(cols ...any) error {
	values := make([]string, 0, len(cols))
	for _, c := range cols {
		values = append(values, fmt.Sprint(c))
	}
	return t.writeRow(values...)
}

func (t *TableWriter) Flush() error {
	return t.w.Flush()
}

func (t *TableWriter) writeRow(cols ...string) error {
	for i := range cols {
		cols[i] = strings.TrimSpace(cols[i])
	}
	_, err := fmt.Fprintln(t.w, strings.Join(cols, "\t"))
	return err
}
