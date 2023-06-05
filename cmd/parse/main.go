package main

import (
	"os"

	fql "github.com/lennon-guan/filterql"
)

func main() {
	cond, err := fql.Parse(os.Args[1], &fql.ParseConfig{
		DefaultIntMethod: func(string, any, int) (any, error) {
			return 0, nil
		},
		DefaultStrMethod: func(string, any, string) (any, error) {
			return "", nil
		},
	})
	if err != nil {
		panic(err)
	}
	cond.PrintTo(0, os.Stdout)
}
