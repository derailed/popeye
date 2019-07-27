package report

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"

	"gopkg.in/yaml.v2"
)

var sanitizerDir = path.Join(os.TempDir(), "popeye")

// Diff represents a diff report between 2 sanitizer runs.
type Diff struct {
	io.Writer

	jurassicMode bool
}

// NewDiff returns a new sanitizer diff instance.
func NewDiff(w io.Writer, jurassic bool) *Diff {
	return &Diff{
		Writer:       w,
		jurassicMode: jurassic,
	}
}

// Jurassic detects if jurassic mode is in effect.
func (d *Diff) Jurassic() bool {
	return d.jurassicMode
}

// Run runs a diff report from the last two sanitize runs.
func (d *Diff) Run(cluster string) error {
	reports, err := lastReports(cluster)
	if err != nil {
		return err
	}

	r1, r2, err := loadReports(reports)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	report := newDiffReport(r1, r2)
	const dateFmt = "2006-01-02 15:04:05"
	open(d, fmt.Sprintf("Candidate Reports [%s]", cluster))
	{
		fmt.Fprintf(d, Colorizef(ColorAqua, "🛀 %-15s: %s\n", reports[1].ModTime().Format(dateFmt), reports[1].Name()))
		fmt.Fprintf(d, Colorizef(ColorAqua, "🛀 %-15s: %s\n", reports[0].ModTime().Format(dateFmt), reports[0].Name()))
	}
	close(d)
	report.Build()
	writer := NewDiffWriter(d, d.jurassicMode)
	writer.Dump(report)

	return nil
}

func loadReports(reports []os.FileInfo) (*Report, *Report, error) {
	b1, err := loadBuilder(reports[1].Name())
	if err != nil {
		return nil, nil, err
	}
	b2, err := loadBuilder(reports[0].Name())
	if err != nil {
		return nil, nil, err
	}

	return &b1.Report, &b2.Report, nil
}

// ----------------------------------------------------------------------------
// Helpers...

func loadBuilder(r string) (*Builder, error) {
	f, err := ioutil.ReadFile(path.Join(sanitizerDir, r))
	if err != nil {
		return nil, err
	}

	var b Builder
	if err := yaml.Unmarshal(f, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func lastReports(cluster string) ([]os.FileInfo, error) {
	rx := regexp.MustCompile(fmt.Sprintf(`\Asanitizer_%s`, cluster))
	var files []os.FileInfo
	err := filepath.Walk(sanitizerDir, func(path string, f os.FileInfo, err error) error {
		if rx.MatchString(f.Name()) {
			files = append(files, f)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(files) < 2 {
		return nil, fmt.Errorf("Must at least have 2 sanitizer reports. Found %d", len(files))
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Unix() > files[j].ModTime().Unix()
	})

	return files[:2], nil
}
