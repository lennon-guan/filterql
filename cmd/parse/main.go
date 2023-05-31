package main

import (
	"os"

	fql "github.com/lennon-guan/filterql"
)

func main() {
	cond, err := fql.Parse(os.Args[1])
	if err != nil {
		panic(err)
	}
	cond.PrintTo(0, os.Stdout)
}
