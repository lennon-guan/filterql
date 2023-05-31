package filterql_test

import (
	"fmt"
	"os"
	"testing"

	fql "github.com/lennon-guan/filterql"
)

func TestParse(t *testing.T) {
	query := `header('x-cli-pn') = 'com.sawa.ksa' And (user('uid') = 10000 or NOT tag(10000) or not user('source') in (1, 2))`
	cond, err := fql.Parse(query)
	if err != nil {
		t.Errorf("parse error %+v", err)
		return
	}
	fmt.Println("query is", query)
	fmt.Println("ast is:")
	cond.PrintTo(0, os.Stdout)
}
