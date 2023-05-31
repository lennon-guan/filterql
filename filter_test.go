package filterql_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	fql "github.com/lennon-guan/filterql"
)

type Record struct {
	ID     int
	Name   string
	Source int
	Level  int
}

var (
	records = []Record{
		{ID: 1, Name: "Apple", Source: 1, Level: 10},
		{ID: 2, Name: "Banana", Source: 1, Level: 6},
		{ID: 3, Name: "Cherry", Source: 1, Level: 8},
		{ID: 4, Name: "DragonFruit", Source: 2, Level: 8},
		{ID: 5, Name: "Egg", Source: 2, Level: 20},
		{ID: 6, Name: "Fig", Source: 3, Level: 5},
		{ID: 7, Name: "Grape", Source: 4, Level: 11},
	}
	strMethods = map[string]func(any, string) (any, error){
		"rec": func(env any, field string) (any, error) {
			return reflect.ValueOf(env).Elem().FieldByName(field).Interface(), nil
		},
	}
)

func joinInts(ints []int) string {
	var b strings.Builder
	for i, v := range ints {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprint(&b, v)
	}
	return b.String()
}

func testFilter(t *testing.T, query string, expectedIds ...int) {
	t.Logf("query: %s", query)
	cond, err := fql.Parse(query)
	if err != nil {
		t.Errorf("parse query [%s] error %+v", query, err)
		return
	}
	ids := []int{}
	ctx := fql.NewContextWithMethods(nil, nil, strMethods)
	for i, rec := range records {
		ctx.Env = &rec
		if matched, err := cond.IsTrue(ctx); err != nil {
			t.Errorf("filter record %d:%#v error %+v", i, rec, err)
			continue
		} else {
			t.Logf("filter record %d:%#v matched %v", i, rec, matched)
			if matched {
				ids = append(ids, rec.ID)
			}
		}
	}
	want := joinInts(expectedIds)
	got := joinInts(ids)
	if want != got {
		t.Errorf("filter result wrong. want %s got %s", want, got)
	}
}

func TestEqual(t *testing.T) {
	testFilter(t, "rec('Source') = 1", 1, 2, 3)
}

func TestNotEqual(t *testing.T) {
	testFilter(t, "rec('Source') <> 1", 4, 5, 6, 7)
}

func TestGreater(t *testing.T) {
	testFilter(t, "rec('Level') > 10", 5, 7)
}

func TestGreaterEqual(t *testing.T) {
	testFilter(t, "rec('Level') >= 10", 1, 5, 7)
}

func TestLess(t *testing.T) {
	testFilter(t, "rec('Level') < 10", 2, 3, 4, 6)
}

func TestLessEqual(t *testing.T) {
	testFilter(t, "rec('Level') <= 10", 1, 2, 3, 4, 6)
}

func TestAnd(t *testing.T) {
	testFilter(t, "rec('Level') >= 10 and rec('Level') < 20", 1, 7)
}

func TestOr(t *testing.T) {
	testFilter(t, "rec('Level') < 10 or rec('Level') >= 20", 2, 3, 4, 5, 6)
}

func TestOrAnd(t *testing.T) {
	testFilter(t, "rec('Name') = 'Banana' or rec('ID') >= 3 and rec('ID') < 5", 2, 3, 4)
}

func TestAndOr(t *testing.T) {
	testFilter(t, "rec('Source') = 1 and (rec('ID') = 3 or rec('ID') = 5)", 3)
}

func TestAndNotOr(t *testing.T) {
	testFilter(t, "rec('Source') = 1 and not (rec('ID') = 3 or rec('ID') = 5)", 1, 2)
}

func TestIn(t *testing.T) {
	testFilter(t, "rec('Name') in ('Egg', 'Fig')", 5, 6)
}

func TestNotIn(t *testing.T) {
	testFilter(t, "not rec('Name') in ('Egg', 'Fig')", 1, 2, 3, 4, 7)
}
