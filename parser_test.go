package conditions

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalid(t *testing.T) {
	var testCases = []string{
		"",
		// "[] AND true",
		"A",
		"[var0] == DEMO",
		"[var0] == 'DEMO'",
		"![var0]",
		"[var0] <> `DEMO`",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			t.Log("condition:", tc)

			p := NewParser(strings.NewReader(tc))
			expr, err := p.Parse()
			assert.Error(t, err, "parse must return not-nil error")
			assert.Nil(t, expr, "parse must return nil expression")
		})
	}
}

func TestValid(t *testing.T) {
	var testCases = []struct {
		cond   string
		args   map[string]interface{}
		result bool
		isErr  bool
	}{
		{"true", nil, true, false},
		{"false", nil, false, false},
		{"false OR true OR false OR false OR true", nil, true, false},
		{"((false OR true) AND false) OR (false OR true)", nil, true, false},
		{"[var0]", map[string]interface{}{"var0": true}, true, false},
		{"[var0]", map[string]interface{}{"var0": false}, false, false},
		{"[var0] > true", nil, false, true},
		{"[var0] > true", map[string]interface{}{"var0": 43}, false, true},
		{"[var0] > true", map[string]interface{}{"var0": false}, false, true},
		{"[var0] and [var1]", map[string]interface{}{"var0": true, "var1": true}, true, false},
		{"[var0] AND [var1]", map[string]interface{}{"var0": true, "var1": false}, false, false},
		{"[var0] AND [var1]", map[string]interface{}{"var0": false, "var1": true}, false, false},
		{"[var0] AND [var1]", map[string]interface{}{"var0": false, "var1": false}, false, false},
		{"[var0] AND false", map[string]interface{}{"var0": true}, false, false},
		{"56.43", nil, false, true},
		{"[var5]", nil, false, true},
		{"[var0] > -100 AND [var0] < -50", map[string]interface{}{"var0": -75.4}, true, false},
		{"[var0]", map[string]interface{}{"var0": true}, true, false},
		{"[var0]", map[string]interface{}{"var0": false}, false, false},
		{"\"OFF\"", nil, false, true},
		//{"`ON`", nil, false, true},
		{"[var0] == \"OFF\"", map[string]interface{}{"var0": "OFF"}, true, false},
		{"[var0] > 10 AND [var1] == \"OFF\"", map[string]interface{}{"var0": 14, "var1": "OFF"}, true, false},
		{"([var0] > 10) AND ([var1] == \"OFF\")", map[string]interface{}{"var0": 14, "var1": "OFF"}, true, false},
		{"([var0] > 10) AND ([var1] == \"OFF\") OR true", map[string]interface{}{"var0": 1, "var1": "ON"}, true, false},
		{"[foo][dfs] == true and [bar] == true", map[string]interface{}{"foo.dfs": true, "bar": true}, true, false},
		{"[foo][dfs][a] == true and [bar] == true", map[string]interface{}{"foo.dfs.a": true, "bar": true}, true, false},
		{"[@foo][a] == true and [bar] == true", map[string]interface{}{"@foo.a": true, "bar": true}, true, false},
		{"[foo][unknow] == true and [bar] == true", map[string]interface{}{"foo.dfs": true, "bar": true}, false, true},
		//XOR
		{"false XOR false", nil, false, false},
		{"false xor true", nil, true, false},
		{"true XOR false", nil, true, false},
		{"true xor true", nil, false, false},

		//NAND
		{"false NAND false", nil, true, false},
		{"false nand true", nil, true, false},
		{"true nand false", nil, true, false},
		{"true NAND true", nil, false, false},

		// IN
		{"[foo] in [foobar]", map[string]interface{}{"foo": "findme", "foobar": []string{"notme", "may", "findme", "lol"}}, true, false},

		// NOT IN
		{"[foo] not in [foobar]", map[string]interface{}{"foo": "dontfindme", "foobar": []string{"notme", "may", "findme", "lol"}}, true, false},

		// IN with array of string
		{`[foo] in ["bonjour", "le monde", "oui"]`, map[string]interface{}{"foo": "le monde"}, true, false},
		{`[foo] in ["bonjour", "le monde", "oui"]`, map[string]interface{}{"foo": "world"}, false, false},

		// NOT IN with array of string
		{`[foo] not in ["bonjour", "le monde", "oui"]`, map[string]interface{}{"foo": "le monde"}, false, false},
		{`[foo] not in ["bonjour", "le monde", "oui"]`, map[string]interface{}{"foo": "world"}, true, false},

		// IN with array of numbers
		{`[foo] in [2,3,4]`, map[string]interface{}{"foo": 4}, true, false},
		{`[foo] in [2,3,4]`, map[string]interface{}{"foo": 5}, false, false},

		// NOT IN with array of numbers
		{`[foo] not in [2,3,4]`, map[string]interface{}{"foo": 4}, false, false},
		{`[foo] not in [2,3,4]`, map[string]interface{}{"foo": 5}, true, false},

		// =~
		{"[status] =~ /^5\\d\\d/", map[string]interface{}{"status": "500"}, true, false},
		{"[status] =~ /^4\\d\\d/", map[string]interface{}{"status": "500"}, false, false},

		// !~
		{"[status] !~ /^5\\d\\d/", map[string]interface{}{"status": "500"}, false, false},
		{"[status] !~ /^4\\d\\d/", map[string]interface{}{"status": "500"}, true, false},

		{`[foo] HAS "5"`, map[string]interface{}{"foo": []string{"5", "3"}}, true, false},
		{`[foo] HAS "4"`, map[string]interface{}{"foo": []string{"5", "3"}}, false, false},
		{`[foo] HAS ["4"]`, map[string]interface{}{"foo": []string{"5", "3"}}, false, true},
		{`[foo] HAS 3`, map[string]interface{}{"foo": []string{"5", "3"}}, false, true},

		{`[foo] INTERSECTS ["5", "7"]`, map[string]interface{}{"foo": []string{"5", "3"}}, true, false},
		{`[foo] INTERSECTS ["4", "8"]`, map[string]interface{}{"foo": []string{"5", "3"}}, false, false},
		{`[foo] INTERSECTS [5, 3]`, map[string]interface{}{"foo": []string{"5", "3"}}, false, true},
		{`[foo] INTERSECTS "4"`, map[string]interface{}{"foo": []string{"5", "3"}}, false, true},
		{`[foo] INTERSECTS ["5", "7"]`, map[string]interface{}{"foo": "4"}, false, true},
		{`[foo] INTERSECTS ["5", "7"]`, map[string]interface{}{"foo": []int{5, 7}}, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.cond, func(t *testing.T) {
			t.Log("condition:", tc.cond)

			p := NewParser(strings.NewReader(tc.cond))

			expr, err := p.Parse()
			assert.NoError(t, err, "unexpected parsing error")

			r, err := Evaluate(expr, tc.args)

			if tc.isErr {
				assert.Error(t, err, "evaluate must return not-nil error")
			} else {
				assert.NoError(t, err, "unexpected evaluation error")
			}
			assert.Equal(t, tc.result, r)
		})
	}
}

func TestExpressionsVariableNames(t *testing.T) {
	cond := "[@foo][a] == true and [bar] == true or [var9] > 10"
	p := NewParser(strings.NewReader(cond))
	expr, err := p.Parse()
	assert.Nil(t, err)

	args := Variables(expr)
	assert.Contains(t, args, "@foo.a", "...")
	assert.Contains(t, args, "bar", "...")
	assert.Contains(t, args, "var9", "...")
	assert.NotContains(t, args, "foo", "...")
	assert.NotContains(t, args, "@foo", "...")
}

func TestEvaluate_ShortCircuit(t *testing.T) {
	testCases := map[string]struct {
		cond     string
		expected bool
	}{
		"AND only left": {"false AND [nonExistent]", false},
		"AND both 1":    {"true AND true", true},
		"AND both 2":    {"true AND false", false},

		"OR only left": {"true OR [nonExistent]", true},
		"OR both 1":    {"false OR true", true},
		"OR both 2":    {"false OR false", false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			args := make(map[string]interface{})

			p := NewParser(strings.NewReader(tc.cond))
			expr, err := p.Parse()
			assert.Nil(t, err)

			result, err := Evaluate(expr, args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
