package sanitize

import (
	"fmt"
	"strings"

	"github.com/derailed/popeye/internal/issues"
)

func dump(msg string, o issues.Outcome) {
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(">>> " + msg)
	if len(o) == 0 {
		fmt.Println("  >>> No Outcome! <<<")
		return
	}

	for k, v := range o {
		if len(v) == 0 {
			fmt.Printf("%s ok\n", k)
			continue
		}
		fmt.Println(k)
		for _, i := range v {
			fmt.Printf("  %s %s: %s\n", i.Group, hLevel(i.Level), i.Message)
		}
	}
	fmt.Println(strings.Repeat("-", 80))
}

func hLevel(l issues.Level) string {
	switch l {
	case issues.OkLevel:
		return "Ok"
	case issues.InfoLevel:
		return "Info"
	case issues.WarnLevel:
		return "Warn"
	case issues.ErrorLevel:
		return "Error"
	default:
		return "Dope!"
	}
}
