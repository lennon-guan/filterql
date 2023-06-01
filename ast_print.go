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

func (a *call[T]) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	prefix := ""
	if a.not {
		prefix = "!"
	}
	fmt.Fprintf(out, "%s%s%s(%#v)\n", indent, prefix, a.name, a.arg)
}

func (a *callThenCompare[T1, T2]) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	fmt.Fprintf(out, "%sCallThenCompare(%s) (\n", indent, tokenName(a.op))
	fmt.Fprintf(out, "%s  %s(%#v)\n", indent, a.name, a.arg)
	fmt.Fprintf(out, "%s  %#v\n", indent, a.target)
	fmt.Fprintf(out, "%s)\n", indent)
}

func (a *Compare[T]) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	fmt.Fprintf(out, "%sCompare(%s) (\n", indent, tokenName(a.Op))
	a.Call.PrintTo(level+1, out)
	fmt.Fprintf(out, "%s  %#v\n", indent, a.Target)
	fmt.Fprintf(out, "%s)\n", indent)
}

func (a *CompareWithCall) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	fmt.Fprintf(out, "%sCompare(%s) (\n", indent, tokenName(a.Op))
	a.Left.PrintTo(level+1, out)
	a.Right.PrintTo(level+1, out)
	fmt.Fprintf(out, "%s)\n", indent)
}

func (a *In[T]) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	if a.NotIn {
		fmt.Fprintf(out, "%sNotIn (\n", indent)
	} else {
		fmt.Fprintf(out, "%sIn (\n", indent)
	}
	a.Call.PrintTo(level+1, out)
	for _, choice := range a.Choices {
		fmt.Fprintf(out, "%s  %#v\n", indent, choice)
	}
	fmt.Fprintf(out, "%s)\n", indent)
}

func (a *InWithCall) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	if a.NotIn {
		fmt.Fprintf(out, "%sNotIn (\n", indent)
	} else {
		fmt.Fprintf(out, "%sIn (\n", indent)
	}
	a.Left.PrintTo(level+1, out)
	a.Right.PrintTo(level+1, out)
	fmt.Fprintf(out, "%s)\n", indent)
}

func (a *callThenIn[T1, T2]) PrintTo(level int, out io.Writer) {
	indent := strings.Repeat("  ", level)
	if a.not {
		fmt.Fprintf(out, "%sCallThenNotIn (\n", indent)
	} else {
		fmt.Fprintf(out, "%sCallThenIn (\n", indent)
	}
	fmt.Fprintf(out, "%s  %s(%#v)\n", indent, a.name, a.arg)
	for _, choice := range a.choices {
		fmt.Fprintf(out, "%s  %#v\n", indent, choice)
	}
	fmt.Fprintf(out, "%s)\n", indent)
}
