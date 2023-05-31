package main

import (
	"os"

	fql "github.com/lennon-guan/filterql"
)

func main() {
	cond, err := fql.Parse(os.Args[1], &fql.ParseConfig{
		IntMethods: map[string]func(any, int) (any, error){
			"intFunc": func(any, int) (any, error) {
				return "", nil
			},
		},
		StrMethods: map[string]func(any, string) (any, error){
			"strFunc": func(any, string) (any, error) {
				return "", nil
			},
		},
	})
	if err != nil {
		panic(err)
	}
	cond.PrintTo(0, os.Stdout)
}
