package filterql_test

import (
	"testing"

	fql "github.com/lennon-guan/filterql"
)

func assertTokens(t *testing.T, code string, expects ...int) {
	ts := fql.NewTokenStream(code)
	for _, et := range expects {
		if !ts.Next() {
			t.Errorf("expected %d but no more token", et)
			return
		}
		if ts.Current.Type != et {
			t.Errorf("expected %d but got %d", et, ts.Current.Type)
			return
		}
		t.Logf("--> %d %s", ts.Current.Type, string(ts.Current.Text))
	}
	if ts.Next() {
		t.Error("expected EOF but got some mored")
	}
}

func TestGetToken(t *testing.T) {
	assertTokens(t, "header('x-cli-pn') = 'com.sawa.ksa'",
		fql.TOKEN_ID,
		fql.TOKEN_LEFT_BRACKET,
		fql.TOKEN_STR,
		fql.TOKEN_RIGHT_BRACKET,
		fql.TOKEN_OP_EQ,
		fql.TOKEN_STR,
	)
	assertTokens(t, "header('x-cli-pn') = 'com.sawa.ksa' AnD 1 < 2",
		fql.TOKEN_ID,
		fql.TOKEN_LEFT_BRACKET,
		fql.TOKEN_STR,
		fql.TOKEN_RIGHT_BRACKET,
		fql.TOKEN_OP_EQ,
		fql.TOKEN_STR,
		fql.TOKEN_AND,
		fql.TOKEN_INT,
		fql.TOKEN_OP_LT,
		fql.TOKEN_INT,
	)
}
