package filterql_test

import (
	"errors"
	"fmt"
	"os"
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
	cfg = &fql.ParseConfig{
		StrMethods: map[string]func(any, string) (any, error){
			"rec": func(env any, field string) (any, error) {
				rec := env.(*Record)
				switch field {
				case "ID":
					return rec.ID, nil
				case "Name":
					return rec.Name, nil
				case "Source":
					return rec.Source, nil
				case "Level":
					return rec.Level, nil
				}
				return reflect.ValueOf(env).Elem().FieldByName(field).Interface(), nil
			},
			"arg": func(env any, field string) (any, error) {
				switch field {
				case "uid":
					return 5, nil
				case "sources":
					return []int{1, 3}, nil
				default:
					return nil, errors.New("unknown arg " + field)
				}
			},
			"env": func(env any, field string) (any, error) {
				switch field {
				case "one_or_three":
					rec := env.(*Record)
					return rec.Source == 1 || rec.Source == 3, nil
				default:
					return nil, errors.New("unknown env key " + field)
				}
			},
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
	testAstAndFilter(t, query, false, expectedIds...)
}

func testAstAndFilter(t *testing.T, query string, showAst bool, expectedIds ...int) {
	t.Logf("query: %s", query)
	cond, err := fql.Parse(query, cfg)
	if err != nil {
		if pe, is := err.(*fql.ParseError); is {
			q := []rune(query)
			if pe.Pos < len(q) {
				query = string(q[:pe.Pos]) + fmt.Sprintf("\033[1;37;41m%c\033[0m", q[pe.Pos]) + string(q[pe.Pos+1:])
			}
		}
		t.Errorf("parse query [%s] error %+v", query, err)
		return
	}
	if showAst {
		cond.PrintTo(0, os.Stdout)
	}
	ids := []int{}
	ctx := fql.NewContext(nil)
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

func TestCompareWithCall(t *testing.T) {
	testFilter(t, "rec('ID') = arg('uid')", 5)
}

func TestInWithCall(t *testing.T) {
	testFilter(t, "rec('Source') in arg('sources')", 1, 2, 3, 6)
}

func TestCallResultCheck(t *testing.T) {
	testFilter(t, "env('one_or_three')", 1, 2, 3, 6)
}

func TestNotCallResultCheck(t *testing.T) {
	testFilter(t, "not env('one_or_three')", 4, 5, 7)
}

func BenchmarkFilterGetFieldBySwitch(b *testing.B) {
	cond, _ := fql.Parse("rec('Source') = 1 and not (rec('ID') = 3 or rec('ID') = 5)", cfg)
	ctx := fql.NewContext(nil)
	n := len(records)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			ctx.Env = &records[j%n]
			cond.IsTrue(ctx)
		}
	}
}

func BenchmarkFilterGetFieldByGoCode(b *testing.B) {
	filterFunc := func(ctx *fql.Context) bool {
		rec := ctx.Env.(*Record)
		return rec.Source == 1 && !(rec.ID == 3 || rec.ID == 5)
	}
	ctx := fql.NewContext(nil)
	n := len(records)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			ctx.Env = &records[j%n]
			filterFunc(ctx)
		}
	}
}

func BenchmarkParseWithoutCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := fql.Parse("rec('Source') = 1 and not (rec('ID') = 3 or rec('ID') = 5)", cfg); err != nil {
			panic(err)
		}
	}
}

func BenchmarkParseWithMapCache(b *testing.B) {
	conf := *cfg
	conf.Cache = fql.NewMapCache()
	for i := 0; i < b.N; i++ {
		if _, err := fql.Parse("rec('Source') = 1 and not (rec('ID') = 3 or rec('ID') = 5)", &conf); err != nil {
			panic(err)
		}
	}
}

func BenchmarkParseWithLRUCache(b *testing.B) {
	conf := *cfg
	conf.Cache = fql.NewLRUCache(10)
	for i := 0; i < b.N; i++ {
		if _, err := fql.Parse("rec('Source') = 1 and not (rec('ID') = 3 or rec('ID') = 5)", &conf); err != nil {
			panic(err)
		}
	}
}
