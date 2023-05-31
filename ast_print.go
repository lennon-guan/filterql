package filterql

import (
	"fmt"
	"io"
	"strings"
)

func (a *ANDs) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	fmt.Fprintf(out, "%sAND (\n", indent)
	for _, child := range a.Children {
		child.PrintTo(level+1, out)
	}
	fmt.Fprintf(out, "%s)\n", indent)
}

func (a *ORs) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	fmt.Fprintf(out, "%sOR (\n", indent)
	for _, child := range a.Children {
		child.PrintTo(level+1, out)
	}
	fmt.Fprintf(out, "%s)\n", indent)
}

func (a *NOT) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	fmt.Fprintf(out, "%sNOT (\n", indent)
	a.Child.PrintTo(level+1, out)
	fmt.Fprintf(out, "%s)\n", indent)
}

func (a *Call) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	if a.ParamType == TOKEN_INT {
		fmt.Fprintf(out, "%s%s(%d)\n", indent, a.Name, a.IntParam)
	} else {
		fmt.Fprintf(out, "%s%s(%#v)\n", indent, a.Name, a.StrParam)
	}
}

func (a *Compare[T]) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	fmt.Fprintf(out, "%sCompare(%s) (\n", indent, tokenName(a.Op))
	a.Call.PrintTo(level+1, out)
	fmt.Fprintf(out, "%s  %#v\n", indent, a.Target)
	fmt.Fprintf(out, "%s)\n", indent)
}

func (a *In[T]) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	fmt.Fprintf(out, "%sIn (\n", indent)
	a.Call.PrintTo(level+1, out)
	for _, choice := range a.Choices {
		fmt.Fprintf(out, "%s  %#v\n", indent, choice)
	}
	fmt.Fprintf(out, "%s)\n", indent)
}
