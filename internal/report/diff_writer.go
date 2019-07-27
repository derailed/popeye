package report

import (
	"fmt"
	"io"
	"strings"
)

// JurassicWriter represent a writer that can print to a fancy or dumb terminal.
type JurassicWriter interface {
	io.Writer

	Jurassic() bool
}

// DiffWriter spews a sanitizer diff report to an output stream.
type DiffWriter struct {
	io.Writer

	jurassicMode bool
}

// Jurassic determines output type.
func (w *DiffWriter) Jurassic() bool {
	return w.jurassicMode
}

// NewDiffWriter returns a new diff report writer.
func NewDiffWriter(w io.Writer, jurassic bool) *DiffWriter {
	return &DiffWriter{
		Writer:       w,
		jurassicMode: jurassic,
	}
}

// Dump outputs a diff report.
func (w *DiffWriter) Dump(r *DiffReport) {
	w.renderOverall(r)
	for k, s := range r.sections {
		if len(s.tallies) == 0 && len(s.outcomes) == 0 {
			continue
		}
		open(w, resToTitle()[k])
		{
			for _, t := range s.tallies {
				ic := "😁"
				if t.worst() {
					ic = "😡"
				}
				fmt.Fprintln(w, fmt.Sprintf("%s Score %s has %s (%s)", ic, EmojiForLevel(t.level, w.Jurassic()), t.summarize(), t.delta()))
			}
			fmt.Fprintln(w)
			for k, ii := range s.outcomes {
				if len(ii) == 0 {
					continue
				}
				fmt.Fprintf(w, "  ⚙️  %s\n", k)
				for _, i := range ii {
					sign := "+"
					if !i.add {
						sign = "-"
					}
					color(w, fmt.Sprintf("    %s %s %s", sign, EmojiForLevel(i.Level, w.Jurassic()), i.Message), colorForLevel(i.Level))
				}
			}
		}
		close(w)
	}
	if len(r.errors) > 0 {
		open(w, "Delta Report Errors")
		{
			for _, e := range r.errors {
				color(w, fmt.Sprintf("  😡  %s", e), ColorRed)
			}
		}
		close(w)
	}
	fmt.Fprintln(w)
}

func (w *DiffWriter) renderOverall(r *DiffReport) {
	open(w, "Cluster Score")
	{
		emoji, paint := "😡", ColorRed
		if !r.overall.changed() {
			emoji, paint = "😟", ColorAqua
		} else if r.overall.better() {
			emoji, paint = "😃", ColorGreen
		}
		msg := fmt.Sprintf("%s Cluster quality score has %s. ", emoji, r.overall.summarize())
		if r.overall.changed() {
			msg += fmt.Sprintf("Score was %d, now %d (%s)", r.overall.s1, r.overall.s2, r.overall.delta())
		}
		color(w, msg, paint)
	}
	close(w)
}

// Helpers...

func open(w JurassicWriter, msg string) {
	color(w, strings.Title(msg), ColorLighSlate)
	color(w, strings.Repeat("┅", Width), ColorLighSlate)
	fmt.Fprintln(w)
}

// Color or not this message by inject ansi colors.
func color(w JurassicWriter, msg string, c Color) {
	if w.Jurassic() {
		fmt.Fprintf(w, "\n%s", msg)
	}
	fmt.Fprintf(w, "\n%s", Colorize(msg, c))
}

// Close a report section.
func close(w JurassicWriter) {
	fmt.Fprintln(w)
}
