package gotextfsm

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

type parseTestCase struct {
	name        string
	template    string
	data        string
	data1       string
	dict        []map[string]interface{}
	eof         *bool
	reset       *bool
	run_err     *regexp.Regexp
	compile_err *regexp.Regexp
}

func TestParseText(t *testing.T) {
	tc_count := 0
	for _, tc := range parseTestCases {
		t.Logf("Running test case %s\n", tc.name)
		tc_count++
		fsm := TextFSM{}
		err := fsm.ParseString(tc.template)
		if err != nil {
			if tc.compile_err == nil {
				t.Errorf("'%s' failed. TextFSM should be valid. But got error '%s'", tc.name, err.Error())
			}
			continue
		} else {
			if tc.compile_err != nil {
				t.Errorf("'%s' failed. TextFSM should invalid. But it is not", tc.name)
				continue
			}
		}
		out := ParserOutput{}
		eof := true
		if tc.eof != nil {
			eof = *tc.eof
		}
		err = out.ParseTextString(tc.data, fsm, eof)
		if tc.reset != nil && *tc.reset {
			out.Reset(fsm)
			err = out.ParseTextString(tc.data, fsm, eof)
		}
		if tc.data1 != "" {
			err = out.ParseTextString(tc.data1, fsm, eof)
		}
		if tc.run_err != nil {
			if err == nil {
				t.Errorf("'%s' failed. Expected error, but none found", tc.name)
			}
			continue
		}
		if err != nil {
			t.Errorf("'%s' failed. Expected no error. But found error %s", tc.name, err)
			continue
		}
		expected := tc.dict
		if expected == nil {
			expected = make([]map[string]interface{}, 0)
		}
		// Dict in 'out' is always initialized to empty slice. Guaranteed not to be nil
		got := out.Dict

		if len(expected) != len(got) {
			t.Errorf("'%s' failed. Expected %d records. Got %d records. %v", tc.name, len(expected), len(got), got)
			continue
		}
		for idx, exprec := range expected {
			gotrec := got[idx]
			err := comparedicts(tc, exprec, gotrec, idx)
			if err != "" {
				t.Error(err)
				break
			}
		}
	}
	t.Logf("Executed %d test cases", tc_count)
}

func comparedicts(tc parseTestCase, exprec map[string]interface{}, gotrec map[string]interface{}, idx int) string {
	if len(exprec) != len(gotrec) {
		return fmt.Sprintf("'%s' failed. Row[%d] Expected %d values, Got %d", tc.name, idx, len(exprec), len(gotrec))
	}
	for name, expval := range exprec {
		gotval, exists := gotrec[name]
		if !exists {
			return fmt.Sprintf("'%s' failed. Row[%d] Var[%s] not found in output", tc.name, idx, name)
		}
		if !valsEquals(gotval, expval) {
			return fmt.Sprintf("'%s' failed. Row[%d] Var[%s] Expected '%v'. Found '%v'", tc.name, idx, name, expval, gotval)
		}
	}
	return ""
}

func valsEquals(gotval interface{}, expval interface{}) bool {
	if reflect.TypeOf(gotval) != reflect.TypeOf(expval) {
		return false
	}
	switch gotval.(type) {
	case string:
		return gotval.(string) == expval.(string)
	case []string:
		return stringListEquals(gotval.([]string), expval.([]string))
	case map[string]string:
		return mapEquals(gotval.(map[string]string), expval.(map[string]string))
	case []map[string]string:
		return mapListEquals(gotval.([]map[string]string), expval.([]map[string]string))
	default:
		panic(fmt.Sprintf("Unknown data type %s", reflect.TypeOf(gotval)))
	}
}
func stringListEquals(t1 []string, t2 []string) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i, v := range t1 {
		if !valsEquals(v, t2[i]) {
			return false
		}
	}
	return true
}
func mapListEquals(t1 []map[string]string, t2 []map[string]string) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i, v := range t1 {
		if !valsEquals(v, t2[i]) {
			return false
		}
	}
	return true
}

var tmp_false bool = false
var tmp_true bool = true

func mapEquals(t1 map[string]string, t2 map[string]string) bool {
	if len(t1) != len(t2) {
		return false
	}
	for k, v1 := range t1 {
		v2, exists := t2[k]
		if !exists {
			return false
		}
		if !valsEquals(v1, v2) {
			return false
		}
	}
	return true
}

var parseTestCases = []parseTestCase{
	{
		name:     "Trivial FSM 1, no records produced.",
		template: "Value unused (.)\n\nStart\n  ^Trivial SFM\n",
		data:     "Non-matching text\nline1\nline 2\n",
	},
	{
		name:     "Trivial FSM 2, no records produced.",
		template: "Value unused (.)\n\nStart\n  ^Trivial SFM\n",
		data:     "Matching text\nTrivial SFM\nline 2\n",
	},
	{
		name:     "Test Next & Record - 1",
		template: "Value boo (.*)\n\nStart\n  ^$boo -> Next.Record\n\nEOF\n",
		data:     "Matching text",
		dict: []map[string]interface{}{
			{"boo": "Matching text"},
		},
	},
	{
		name:     "Test Next & Record - 2",
		template: "Value boo (.*)\n\nStart\n  ^$boo -> Next.Record\n\nEOF\n",
		data:     "Matching text\nAnd again",
		dict: []map[string]interface{}{
			{"boo": "Matching text"},
			{"boo": "And again"},
		},
	},
	{
		name:     "Null data",
		template: "Value boo (.*)\n\nStart\n  ^$boo -> Next.Record\n\nEOF\n",
		data:     "",
	},
	{
		// Matching two lines. Only one records returned due to 'Required' flag.
		// Tests 'Filldown' and 'Required' options.
		name: "Two Variables: 1.",
		template: `Value Required boo (one)
Value Filldown hoo (two)

Start
  ^$boo -> Next.Record
  ^$hoo -> Next.Record

EOF
`,
		data: "two\none",
		dict: []map[string]interface{}{
			{"boo": "one", "hoo": "two"},
		},
	},
	{
		// Matching two lines. Two records returned due to 'Filldown' flag.
		name: "Two Variables: 2 ",
		template: `Value Required boo (one)
Value Filldown hoo (two)

Start
  ^$boo -> Next.Record
  ^$hoo -> Next.Record

EOF
`,
		data: "two\none\none",
		dict: []map[string]interface{}{
			{"boo": "one", "hoo": "two"},
			{"boo": "one", "hoo": "two"},
		},
	},
	{
		// Matching two lines. Two records returned due to 'Filldown' flag.
		name: "Mulitple variables and options ",
		template: `Value Required,Filldown boo (one)
Value Filldown,Required hoo (two)

Start
  ^$boo -> Next.Record
  ^$hoo -> Next.Record

EOF
`,
		data: "two\none\none",
		dict: []map[string]interface{}{
			{"boo": "one", "hoo": "two"},
			{"boo": "one", "hoo": "two"},
		},
	},
	{
		name: "Test Clear",
		template: `Value Required boo (on.)
Value Filldown,Required hoo (tw.)

Start
  ^$boo -> Next.Record
  ^$hoo -> Next.Clear
`,
		data: "one\ntwo\nonE\ntwO",
		dict: []map[string]interface{}{
			{"boo": "onE", "hoo": "two"},
		},
	},
	{
		name: "Test Clear All",
		template: `Value Filldown  boo (on.)
Value Filldown hoo (tw.)

Start
  ^$boo -> Next.Clearall
  ^$hoo
`,
		data: "one\ntwo",
		dict: []map[string]interface{}{
			{"boo": "", "hoo": "two"},
		},
	},
	{
		name: "Test Continue",
		template: `Value Required  boo (on.)
Value Filldown,Required hoo (on.)

Start
  ^$boo -> Continue
  ^$hoo -> Continue.Record
`,
		data: "one\non0",
		dict: []map[string]interface{}{
			{"boo": "one", "hoo": "one"},
			{"boo": "on0", "hoo": "on0"},
		},
	},
	{
		name: "Test Error 1",
		template: `Value Required boo (on.)
Value Filldown,Required hoo (on.)

Start
  ^$boo -> Continue
  ^$hoo -> Error
`,
		data:    "one",
		run_err: regexp.MustCompile(".+"),
	},
	{
		name: "Test Error with string",
		template: `Value Required boo (on.)
Value Filldown,Required hoo (on.)

Start
  ^$boo -> Continue
  ^$hoo -> Error "Hello World"
`,
		data:    "one",
		run_err: regexp.MustCompile(".+"),
	},
	{
		// Key really does not have any significance anyway
		name: "Test Key",
		template: `Value Required,Key boo (one)
Value Filldown hoo (two)

Start
  ^$boo -> Next.Record
  ^$hoo -> Next.Record

EOF
`,
		data: "two\none",
		dict: []map[string]interface{}{
			{"boo": "one", "hoo": "two"},
		},
	},
	{
		name: "Test List",
		template: `Value List boo (on.)
Value hoo (tw.)

Start
  ^$boo
  ^$hoo -> Next.Record

EOF
`,
		data: "one\ntwo\non0\ntw0",
		dict: []map[string]interface{}{
			{"boo": []string{"one"}, "hoo": "two"},
			{"boo": []string{"on0"}, "hoo": "tw0"},
		},
	},
	{
		name: "Test List: 2",
		template: `Value List,Filldown boo (on.)
Value hoo (on.)

Start
  ^$boo -> Continue
  ^$hoo -> Next.Record

EOF
`,
		data: "one\non0\non1",
		dict: []map[string]interface{}{
			{"boo": []string{"one"}, "hoo": "one"},
			{"boo": []string{"one", "on0"}, "hoo": "on0"},
			{"boo": []string{"one", "on0", "on1"}, "hoo": "on1"},
		},
	},
	{
		name: "Test Empty List",
		template: `Value List boo (never)
Value hoo (on.)

Start
  ^$boo -> Continue
  ^$hoo -> Next.Record

EOF
`,
		data: "one\non0\non1",
		dict: []map[string]interface{}{
			{"boo": []string{}, "hoo": "one"},
			{"boo": []string{}, "hoo": "on0"},
			{"boo": []string{}, "hoo": "on1"},
		},
	},
	{
		name: "Test Nested Scalar",
		template: `Value foo ((?P<name>\w+):\s+(?P<age>\d+)\s+(?P<state>\w{2})\s*)
Value name (^\w+$)

Start
  ^\s*${foo}
  ^\s*${name}
  ^\s*$$ -> Record
`,
		data: " Bob: 32 NC\n Alice: 27 NY\n Jeff: 45 CA\nJulia\n\n",
		dict: []map[string]interface{}{
			{
				"foo":  map[string]string{"name": "Jeff", "age": "45", "state": "CA"},
				"name": "Julia",
			},
		},
	},
	{
		name: "Test Nested Scalar, Filldown",
		template: `Value Filldown foo ((?P<name>\w+):\s+(?P<age>\d+)\s+(?P<state>\w{2})\s*)
Value name (^\w+$)

Start
  ^\s*${foo}
  ^\s*${name}
  ^\s*$$ -> Record
`,
		data: " Bob: 32 NC\n Alice: 27 NY\n Jeff: 45 CA\nJulia\n\nSiri",
		dict: []map[string]interface{}{
			{
				"foo":  map[string]string{"name": "Jeff", "age": "45", "state": "CA"},
				"name": "Julia",
			},
			{
				"foo":  map[string]string{"name": "Jeff", "age": "45", "state": "CA"},
				"name": "Siri",
			},
		},
	},
	{
		name: "Test Nested Scalar, Filldown, ClearAll",
		template: `Value Filldown foo ((?P<name>\w+):\s+(?P<age>\d+)\s+(?P<state>\w{2})\s*)
Value name (^\w+$)

Start
  ^\s*${foo}
  ^\s*${name}
  ^\s*$$ -> Record
  ^\s*Clear all$$ -> Clearall
`,
		data: " Bob: 32 NC\n Alice: 27 NY\n Jeff: 45 CA\nJulia\n\nSiri\n\nClear all\nShirley",
		dict: []map[string]interface{}{
			{
				"foo":  map[string]string{"name": "Jeff", "age": "45", "state": "CA"},
				"name": "Julia",
			},
			{
				"foo":  map[string]string{"name": "Jeff", "age": "45", "state": "CA"},
				"name": "Siri",
			},
			{
				"foo":  map[string]string{},
				"name": "Shirley",
			},
		},
	},
	{
		// Ensures that List-type values with nested regex capture groups are parsed
		// correctly as a list of dictionaries.
		// Additionaly, another value is used with the same group-name as one of the
		// nested groups to ensure that there are no conflicts when the same name is
		// used.
		name: "Test Nested List",
		template: `Value List foo ((?P<name>\w+):\s+(?P<age>\d+)\s+(?P<state>\w{2})\s*)
Value name (^\w+$)

Start
  ^\s*${foo}
  ^\s*${name}
  ^\s*$$ -> Record
`,
		data: " Bob: 32 NC\n Alice: 27 NY\n Jeff: 45 CA\nJulia\n\n",
		dict: []map[string]interface{}{
			{
				"foo": []map[string]string{
					map[string]string{"name": "Bob", "age": "32", "state": "NC"},
					map[string]string{"name": "Alice", "age": "27", "state": "NY"},
					map[string]string{"name": "Jeff", "age": "45", "state": "CA"},
				},
				"name": "Julia",
			},
		},
	},
	{
		name: "Test Nested with List,Filldown",
		template: `Value List,Filldown foo ((?P<name>\w+):\s+(?P<age>\d+)\s+(?P<state>\w{2})\s*)
Value name (^\w+$)

Start
  ^\s*${foo}
  ^\s*${name}
  ^\s*$$ -> Record
`,
		data: " Bob: 32 NC\n Alice: 27 NY\n Jeff: 45 CA\nJulia\n\nSiri\n\nDavid: 60 VA\nShirley",
		dict: []map[string]interface{}{
			{
				"foo": []map[string]string{
					map[string]string{"name": "Bob", "age": "32", "state": "NC"},
					map[string]string{"name": "Alice", "age": "27", "state": "NY"},
					map[string]string{"name": "Jeff", "age": "45", "state": "CA"},
				},
				"name": "Julia",
			},
			{
				"foo": []map[string]string{
					map[string]string{"name": "Bob", "age": "32", "state": "NC"},
					map[string]string{"name": "Alice", "age": "27", "state": "NY"},
					map[string]string{"name": "Jeff", "age": "45", "state": "CA"},
				},
				"name": "Siri",
			},
			{
				"foo": []map[string]string{
					map[string]string{"name": "Bob", "age": "32", "state": "NC"},
					map[string]string{"name": "Alice", "age": "27", "state": "NY"},
					map[string]string{"name": "Jeff", "age": "45", "state": "CA"},
					map[string]string{"name": "David", "age": "60", "state": "VA"},
				},
				"name": "Shirley",
			},
		},
	},
	{
		name: "Test Nested with List,Filldown - ClearAll",
		template: `Value List,Filldown foo ((?P<name>\w+):\s+(?P<age>\d+)\s+(?P<state>\w{2})\s*)
Value name (^\w+$)

Start
  ^\s*${foo}
  ^\s*${name}
  ^\s*$$ -> Record
  ^\s*Clear All$$ -> Clearall
`,
		data: " Bob: 32 NC\n Alice: 27 NY\n Jeff: 45 CA\nJulia\n\nSiri\n\nClear All\nDavid: 60 VA\nShirley",
		dict: []map[string]interface{}{
			{
				"foo": []map[string]string{
					map[string]string{"name": "Bob", "age": "32", "state": "NC"},
					map[string]string{"name": "Alice", "age": "27", "state": "NY"},
					map[string]string{"name": "Jeff", "age": "45", "state": "CA"},
				},
				"name": "Julia",
			},
			{
				"foo": []map[string]string{
					map[string]string{"name": "Bob", "age": "32", "state": "NC"},
					map[string]string{"name": "Alice", "age": "27", "state": "NY"},
					map[string]string{"name": "Jeff", "age": "45", "state": "CA"},
				},
				"name": "Siri",
			},
			{
				"foo": []map[string]string{
					map[string]string{"name": "David", "age": "60", "state": "VA"},
				},
				"name": "Shirley",
			},
		},
	},
	{
		name: "Simple state change, no actions",
		template: `Value boo (one)
Value hoo (two)

Start
  ^$boo -> State1

State1
  ^$hoo -> Start

EOF
`,
		data: "one",
	},
	{
		name: "Simple state change, with actions",
		template: `Value boo (one)
Value hoo (two)

Start
  ^$boo ->  Next.Record State1

State1
  ^$hoo -> Start

EOF
`,
		data: "one",
		dict: []map[string]interface{}{
			{"boo": "one", "hoo": ""},
		},
	},
	{
		name: "Test implicit EOF",
		template: `Value boo (.*)

Start
  ^$boo ->  Next
`,
		data: "Matching Text",
		dict: []map[string]interface{}{
			{"boo": "Matching Text"},
		},
	},
	{
		name: "Test explicit EOF",
		template: `Value boo (.*)

Start
  ^$boo ->  Next

EOF
`,
		data: "Matching Text",
	},
	{
		name: "Implicit EOF suppressed by argument.",
		template: `Value boo (.*)

Start
  ^$boo ->  Next

`,
		data: "Matching Text",
		eof:  &tmp_false,
	},
	{
		name: "End State, EOF is skipped.",
		template: `Value boo (.*)

Start
  ^$boo ->  End
  ^$boo ->  Record
`,
		data: "Matching text A\nMatching text B",
	},
	{
		name: "EOF state transition is followed by implicit End State.",
		template: `Value boo (.*)

Start
  ^$boo ->  EOF
  ^boo -> Record
`,
		data: "Matching text A\nMatching text B",
		dict: []map[string]interface{}{
			{"boo": "Matching text A"},
		},
	},
	{
		name: "Test Invalid Regex",
		template: `Value boo (.$**)

Start
  ^$boo ->  EOF
`,
		data:        "Matching text A\nMatching text B",
		compile_err: regexp.MustCompile(".+"),
	},
	{
		name: "Test Fillup",
		template: `Value Required Col1 ([^-]+)
Value Fillup Col2 ([^-]+)
Value Fillup Col3 ([^-]+)

Start
  ^$Col1 -- -- -> Record
  ^$Col1 $Col2 -- -> Record
  ^$Col1 -- $Col3 -> Record
  ^$Col1 $Col2 $Col3 -> Record`,
		data: `
1 -- B1
2 A2 --
3 -- B3
`,
		dict: []map[string]interface{}{
			{"Col1": "1", "Col2": "A2", "Col3": "B1"},
			{"Col1": "2", "Col2": "A2", "Col3": "B3"},
			{"Col1": "3", "Col2": "", "Col3": "B3"},
		},
	},
	{
		name:  "Reset Test Case",
		reset: &tmp_true,
		template: `Value Required Col1 ([^-]+)
Value Fillup Col2 ([^-]+)
Value Fillup Col3 ([^-]+)

Start
	^$Col1 -- -- -> Record
	^$Col1 $Col2 -- -> Record
	^$Col1 -- $Col3 -> Record
	^$Col1 $Col2 $Col3 -> Record`,
		data: `
1 -- B1
2 A2 --
3 -- B3
`,
		dict: []map[string]interface{}{
			{"Col1": "1", "Col2": "A2", "Col3": "B1"},
			{"Col1": "2", "Col2": "A2", "Col3": "B3"},
			{"Col1": "3", "Col2": "", "Col3": "B3"},
		},
	},
	{
		name: "Test Reentrant parser",
		template: `Value Required Col1 ([^-]+)
Value Fillup Col2 ([^-]+)
Value Fillup Col3 ([^-]+)

Start
	^$Col1 -- -- -> Record
	^$Col1 $Col2 -- -> Record
	^$Col1 -- $Col3 -> Record
	^$Col1 $Col2 $Col3 -> Record`,
		data: `
1 -- B1`,
		data1: `2 A2 --
3 -- B3
`,
		dict: []map[string]interface{}{
			{"Col1": "1", "Col2": "A2", "Col3": "B1"},
			{"Col1": "2", "Col2": "A2", "Col3": "B3"},
			{"Col1": "3", "Col2": "", "Col3": "B3"},
		},
	},
	{
		name: "Filldown with new value",
		template: `# Headline
Value Filldown boo (o.*)
Value hoo (t.*)

Start
  ^$boo
  ^$hoo -> Record

`,
		data: "one\ntwo\nthree\nother\nten",
		dict: []map[string]interface{}{
			{"boo": "one", "hoo": "two"},
			{"boo": "one", "hoo": "three"},
			{"boo": "other", "hoo": "ten"},
			{"boo": "other", "hoo": ""},
		},
	},
	{
		name: "Filldown with new value 2",
		template: `# Headline
Value Filldown boo (o.*)
Value Required hoo (t.*)

Start
  ^$boo
  ^$hoo -> Record

`,
		data: "one\ntwo\nthree\nother\nten",
		dict: []map[string]interface{}{
			{"boo": "one", "hoo": "two"},
			{"boo": "one", "hoo": "three"},
			{"boo": "other", "hoo": "ten"},
		},
	},
	{
		// For a Filldown variable, the value is populated down
		// In case of no match OR in case of match, but it matched to empty string
		// This test case checks the filldown works when it is a match, but output is empty
		name: "Filldown with new value",
		template: `# Headline
Value Filldown boo (.?)
Value hoo (t.*)

Start
  ^on$boo
  ^$hoo -> Record

`,
		data: "one\ntwo\nthree\non\nten",
		dict: []map[string]interface{}{
			{"boo": "e", "hoo": "two"},
			{"boo": "e", "hoo": "three"},
			{"boo": "", "hoo": "ten"},
		},
	},
	{
		name: "All types of outputs",
		template: `Value continent (.*)
Value List countries (.*)
Value state_abbr ((?P<fullstate>\w+):\s+(?P<abbr>\w{2}))
Value List persons ((?P<name>\w+):\s+(?P<age>\d+)\s+(?P<state>\w{2})\s*)

Start
  ^Continent: ${continent}
  ^Country: ${countries}
  ^State: ${state_abbr}
  ^${persons}
`,
		data: `Continent: North America
Country: USA
Country: Canada
Country: Mexico
State: California: CA
Siri: 50 CA
Raj: 22 NM
Gandhi: 150 NV
`,
		dict: []map[string]interface{}{
			{
				"continent":  "North America",
				"countries":  []string{"USA", "Canada", "Mexico"},
				"state_abbr": map[string]string{"fullstate": "California", "abbr": "CA"},
				"persons": []map[string]string{
					{"name": "Siri", "age": "50", "state": "CA"},
					{"name": "Raj", "age": "22", "state": "NM"},
					{"name": "Gandhi", "age": "150", "state": "NV"},
				},
			},
		},
	},
	{
		name: "Test Square brackets",
		template: `Value country ([a-zA-Z]+)

Start
  ^Country: ${country} -> Record
`,
		data: `Continent: North America
Country: USA
Country: Canada
Country: Mexico
`,
		dict: []map[string]interface{}{
			{"country": "USA"}, {"country": "Canada"}, {"country": "Mexico"},
		},
	},
	{
		name: "Test Fillup with explicit EOF",
		template: `Value Filldown FILE_SYSTEM (\S+)
Value PERMISSIONS (\S+)
Value SIZE (\d+)
Value DATE_TIME (\S+\s+\d+\s+((\d+)|(\d+:\d+)))
Value NAME (\S+)
Value Fillup TOTAL_SIZE (\d+)
Value Fillup TOTAL_FREE (\d+)
		
Start
  ^Directory of\s+${FILE_SYSTEM} -> DIR

DIR
  ^\s+${PERMISSIONS}\s+${SIZE}\s+${DATE_TIME}\s+${NAME} -> Record
  ^${TOTAL_SIZE}\s+\S+\s+\S+\s+\(${TOTAL_FREE}\s+\S+\s+\S+\)
  ^\s+$$
  ^$$
  ^.* -> Error "LINE NOT FOUND"

EOF
`,
		data: `Directory of flash:/

		-rwx   591941836            Aug 2  2017  EOS-4.18.3.1F.swi
		-rwx   609823300           Feb 14 02:03  EOS-4.19.5M.swi
		-rwx          29           Aug 23  2017  boot-config

3519041536 bytes total (1725112320 bytes free)
 
`,
		dict: []map[string]interface{}{
			{"FILE_SYSTEM": "flash:/", "PERMISSIONS": "-rwx", "SIZE": "591941836", "DATE_TIME": "Aug 2  2017", "NAME": "EOS-4.18.3.1F.swi", "TOTAL_SIZE": "3519041536", "TOTAL_FREE": "1725112320"},
			{"FILE_SYSTEM": "flash:/", "PERMISSIONS": "-rwx", "SIZE": "609823300", "DATE_TIME": "Feb 14 02:03", "NAME": "EOS-4.19.5M.swi", "TOTAL_SIZE": "3519041536", "TOTAL_FREE": "1725112320"},
			{"FILE_SYSTEM": "flash:/", "PERMISSIONS": "-rwx", "SIZE": "29", "DATE_TIME": "Aug 23  2017", "NAME": "boot-config", "TOTAL_SIZE": "3519041536", "TOTAL_FREE": "1725112320"},
		},
	},
	{
		name: "Show Interfaces",
		template: `Value Required INTERFACE (\S+)
Value LINK_STATUS (.*)
Value PROTOCOL_STATUS (.*)
Value HARDWARE_TYPE ([\w+-]+)
Value ADDRESS ([a-zA-Z0-9]+.[a-zA-Z0-9]+.[a-zA-Z0-9]+)
Value BIA ([a-zA-Z0-9]+.[a-zA-Z0-9]+.[a-zA-Z0-9]+)
Value DESCRIPTION (.*)
Value IP_ADDRESS (\d+\.\d+\.\d+\.\d+\/\d+)
Value MTU (\d+)
Value BANDWIDTH (\d+\s+\w+)
		
Start
  ^${INTERFACE}\s+is\s+${LINK_STATUS},\s+line\s+protocol\s+is\s+${PROTOCOL_STATUS}$$
  ^\s+Hardware\s+is\s+${HARDWARE_TYPE}(.*address\s+is\s+${ADDRESS})*(.*bia\s+${BIA})*
  ^\s+Description:\s+${DESCRIPTION}
  ^\s+Internet\s+address\s+is\s+${IP_ADDRESS}
  ^.*MTU\s+${MTU}(.*BW\s+${BANDWIDTH})* -> Record		
`,
		data: `Ethernet1 is up, line protocol is up (connected)
  Hardware is Ethernet, address is 0800.27dc.5443
  Internet address is 172.16.1.1/24
  Broadcast address is 255.255.255.255
  Address determined by manual configuration
  IP MTU 1500 bytes , BW 10000000 kbit
  Full-duplex, 10Gb/s, auto negotiation: off, uni-link: unknown
  Up 14 minutes, 2 seconds
  1 link status changes since last clear
  Last clearing of "show interface" counters never
  5 minutes input rate 0 bps (0.0% with framing overhead), 0 packets/sec
  5 minutes output rate 0 bps (0.0% with framing overhead), 0 packets/sec
    292 packets input, 31440 bytes
    Received 3 broadcasts, 0 multicast
    0 runts, 0 giants
    0 input errors, 0 CRC, 0 alignment, 0 symbol, 0 input discards
    0 PAUSE input
    203 packets output, 33221 bytes
    Sent 0 broadcasts, 32 multicast
    0 output errors, 0 collisions
    0 late collision, 0 deferred, 0 output discards
	0 PAUSE output
Ethernet49/1 is administratively down, line protocol is notpresent (disabled)
	Hardware is Ethernet, address is fcbd.67e2.b922 (bia fcbd.67e2.b922)
	Ethernet MTU 9214 bytes , BW 100000000 kbit
	Full-duplex, 100Gb/s, auto negotiation: off, uni-link: n/a
	Down 6 days, 11 hours, 16 minutes, 54 seconds
	Loopback Mode : None
	1 link status changes since last clear
	Last clearing of "show interface" counters 6 days, 11:19:37 ago
	5 minutes input rate 0 bps (0.0% with framing overhead), 0 packets/sec
	5 minutes output rate 0 bps (0.0% with framing overhead), 0 packets/sec
	   0 packets input, 0 bytes
	   Received 0 broadcasts, 0 multicast
	   0 runts, 0 giants
	   0 input errors, 0 CRC, 0 alignment, 0 symbol, 0 input discards
	   0 PAUSE input
	   0 packets output, 0 bytes
	   Sent 0 broadcasts, 0 multicast
	   0 output errors, 0 collisions
	   0 late collision, 0 deferred, 0 output discards
	   0 PAUSE output
`,
		dict: []map[string]interface{}{
			{"INTERFACE": "Ethernet1", "LINK_STATUS": "up", "PROTOCOL_STATUS": "up (connected)", "HARDWARE_TYPE": "Ethernet", "ADDRESS": "0800.27dc.5443", "BIA": "", "DESCRIPTION": "", "IP_ADDRESS": "172.16.1.1/24", "MTU": "1500", "BANDWIDTH": "10000000 kbit"},
			{"INTERFACE": "Ethernet49/1", "LINK_STATUS": "administratively down", "PROTOCOL_STATUS": "notpresent (disabled)", "HARDWARE_TYPE": "Ethernet", "ADDRESS": "fcbd.67e2.b922", "BIA": "fcbd.67e2.b922", "DESCRIPTION": "", "IP_ADDRESS": "", "MTU": "9214", "BANDWIDTH": "100000000 kbit"},
		},
	},
	{
		name: "Show ip helper",
		template: `Value Required INTERFACE (\S+)
Value List IP_HELPER (\d+\.\d+\.\d+\.\d+|\S+)

Start
  ^DHCP
  ^Interface -> Continue.Record
  ^Interface:\s+${INTERFACE}$$
  ^\s+DHCP\s+Smart
  ^\s+DHCP\s+servers:\s+${IP_HELPER}$$
  ^\s+${IP_HELPER}$$
  ^$$
  ^. -> Error	
`,
		data: `
DHCP Relay is active
DHCP Relay Option 82 is disabled
DHCPv6 Relay Link-layer Address Option (79) is disabled
DHCP Smart Relay is disabled
Interface: Vlan1
    DHCP Smart Relay is disabled
    DHCP servers: 10.1.0.0
                  10.1.0.1
                  10.1.0.2
                  10.1.0.3
                  server.domain
Interface: Vlan2
    DHCP Smart Relay is disabled
    DHCP servers: 10.1.0.4
                  server.domain
`,
		dict: []map[string]interface{}{
			{"INTERFACE": "Vlan1", "IP_HELPER": []string{"10.1.0.0", "10.1.0.1", "10.1.0.2", "10.1.0.3", "server.domain"}},
			{"INTERFACE": "Vlan2", "IP_HELPER": []string{"10.1.0.4", "server.domain"}},
		},
	},
	{
		name: "Show ip route",
		template: `Value Filldown VRF (\S+)
Value Filldown PROTOCOL (\S+\s\S+?|\w?)
Value Filldown NETWORK (\d+.\d+.\d+.\d+)
Value Filldown MASK (\d+)
Value Filldown DISTANCE (\d+)
Value Filldown METRIC (\d+)
Value DIRECT (directly)
Value Required NEXT_HOP (connected|\d+\.\d+\.\d+\.\d+)
Value INTERFACE (\S+)

Start
  ^\s+${PROTOCOL}\s+${NETWORK}/${MASK}\s+(?:\[${DISTANCE}/${METRIC}\]|is\s+${DIRECT})(?:.+?)${NEXT_HOP},\s+${INTERFACE}$$ -> Record
  ^\s+via\s+${NEXT_HOP},\s+${INTERFACE} -> Record
  ^VRF\s+name:\s+${VRF}\s*$$
  ^VRF:\s+${VRF}\s*$$
  ^WARNING
  ^kernel
  ^Codes:
  # Match for codes
  ^\s+\S+\s+-\s+\S+
  ^Gateway\s+of\s+last
  ^\s*$$
  ^. -> Error		
`,
		data: `
Codes: C - connected, S - static, K - kernel, 
    O - OSPF, IA - OSPF inter area, E1 - OSPF external type 1,
	E2 - OSPF external type 2, N1 - OSPF NSSA external type 1,
	N2 - OSPF NSSA external type2, B I - iBGP, B E - eBGP,
	R - RIP, I - ISIS, A B - BGP Aggregate, A O - OSPF Summary,
	NG - Nexthop Group Static Route 
 
Gateway of last resort is not set

 B E    10.1.31.100/32 [200/0] via 192.168.17.5, Ethernet18
 B E    10.1.31.101/32 [200/0] via 192.168.17.5, Ethernet18
 C      10.1.31.102/32 is directly connected, Loopback100
`,
		dict: []map[string]interface{}{
			{"VRF": "", "PROTOCOL": "B E", "NETWORK": "10.1.31.100", "MASK": "32", "DISTANCE": "200", "METRIC": "0", "DIRECT": "", "NEXT_HOP": "192.168.17.5", "INTERFACE": "Ethernet18"},
			{"VRF": "", "PROTOCOL": "B E", "NETWORK": "10.1.31.101", "MASK": "32", "DISTANCE": "200", "METRIC": "0", "DIRECT": "", "NEXT_HOP": "192.168.17.5", "INTERFACE": "Ethernet18"},
			{"VRF": "", "PROTOCOL": "C", "NETWORK": "10.1.31.102", "MASK": "32", "DISTANCE": "", "METRIC": "", "DIRECT": "directly", "NEXT_HOP": "connected", "INTERFACE": "Loopback100"},
		},
	},
	{
		name: "Show ip route",
		template: `Value MODULE (\S+)
Value PORTS (\d+)
Value CARD (.+?)
Value TYPE (\S+)
Value MODEL (\S+)
Value SERIAL_NUM (\S+)
Value Fillup MAC_ADDRESS_START (.+?)
Value Fillup MAC_ADDRESS_END (.+?)
Value Fillup HW_VER (\S+)
Value Fillup SW_VER (\S+|\s+)
Value Fillup STATUS (\S+)
Value Fillup UPTIME (.+)

Start
	^-.+
	^Module\s+Ports\s+Card\s+Type\s+Model\s+Serial\s+No\.\s*$$
	^${MODULE}\s+${PORTS}\s+${CARD}\s+${TYPE}\s+${MODEL}\s+${SERIAL_NUM}\s*$$ -> Record
	^Module\s+MAC\s+addresses\s+Hw\s+Sw\s*$$
	^${MODULE}\s+(?:${MAC_ADDRESS_START}\s+-\s+${MAC_ADDRESS_END})?\s+${HW_VER}(\s+${SW_VER})?\s*$$
	^Module\s+Status\s+Uptime\s*$$
	^${MODULE}\s+${STATUS}(\s+${UPTIME})?\s*$$
	^\s*$$
	^. -> Error "LINE NOT FOUND"

EOF			
`,
		data: `
Module  Ports Card Type                Model          Serial No.
------- ----- ------------------------ -------------- -----------
1       3     DCS-7500-SUP2 Supervisor DCS-7500-SUP2  XX16380393
3       144   36-port QSFP100 Linecard 7500R-36CQ-LC  XX16340219
Fabric1 0     DCS-7508R Fabric         7508R-FM       XX16472732

Module  MAC addresses                         Hw    Sw
------- ------------------------------------- ----- -------
1       44:4c:a8:e6:17:5e - 44:4c:a8:e6:17:5f 14.20 4.19.5M
3       44:4c:a8:e2:d0:28 - 44:4c:a8:e2:d0:b7 13.00
4       44:4c:a8:ee:a9:2c - 44:4c:a8:ee:a9:bb 13.00
5       44:4c:a8:ee:97:2c - 44:4c:a8:ee:97:bb 13.00
6       44:4c:a8:ee:2f:1c - 44:4c:a8:ee:2f:ab 13.00
7       28:99:3a:a4:01:58 - 28:99:3a:a4:01:e7 12.01
Fabric1                                       12.03
Fabric2                                       12.03
Fabric3                                       12.03
Fabric4                                       12.03
Fabric5                                       12.03
Fabric6                                       12.03

Module  Status Uptime
------- ------ ----------------
1       Active
3       Ok     74 days, 0:25:22
4       Ok     74 days, 0:25:22
5       Ok     74 days, 0:25:22
6       Ok     74 days, 0:25:22
7       Ok     74 days, 0:25:22
Fabric1 Ok     74 days, 0:25:22
Fabric2 Ok     74 days, 0:25:22
Fabric3 Ok     74 days, 0:25:22
Fabric4 Ok     74 days, 0:25:22
Fabric5 Ok     74 days, 0:25:22
Fabric6 Ok     74 days, 0:25:22`,
		dict: []map[string]interface{}{
			{"MODULE": "1", "PORTS": "3", "CARD": "DCS-7500-SUP2", "TYPE": "Supervisor", "MODEL": "DCS-7500-SUP2", "SERIAL_NUM": "XX16380393", "MAC_ADDRESS_START": "44:4c:a8:e6:17:5e", "MAC_ADDRESS_END": "44:4c:a8:e6:17:5f",
				"HW_VER": "14.20", "SW_VER": "4.19.5M", "STATUS": "Ok", "UPTIME": "74 days, 0:25:22"},
			{"MODULE": "3", "PORTS": "144", "CARD": "36-port QSFP100", "TYPE": "Linecard", "MODEL": "7500R-36CQ-LC", "SERIAL_NUM": "XX16340219", "MAC_ADDRESS_START": "44:4c:a8:e6:17:5e", "MAC_ADDRESS_END": "44:4c:a8:e6:17:5f",
				"HW_VER": "14.20", "SW_VER": "4.19.5M", "STATUS": "Ok", "UPTIME": "74 days, 0:25:22"},
			{"MODULE": "Fabric1", "PORTS": "0", "CARD": "DCS-7508R", "TYPE": "Fabric", "MODEL": "7508R-FM", "SERIAL_NUM": "XX16472732", "MAC_ADDRESS_START": "44:4c:a8:e6:17:5e", "MAC_ADDRESS_END": "44:4c:a8:e6:17:5f",
				"HW_VER": "14.20", "SW_VER": "4.19.5M", "STATUS": "Ok", "UPTIME": "74 days, 0:25:22"},
		},
	},
	{
		name: "Test broadcom fastiron show version",
		template: `Value List SWITCH_ID (\d+)
Value List POE (POE)

Start
	^show\s+version
	^\s+UNIT\s+${SWITCH_ID}
	^\s+\(.+?\)\s+from\s+\S+\s+.*\s*$$
	^\s*SW:\s+Version\s+.*
	^\s*Boot-Monitor.*,\s+Version:.*\s*$$
	^\s*HW:\s+.*\s*$$
	^\s+Copyright
	^\s*$$
	^=+\s*$$ -> Hardware
	^. -> Error

Hardware
	^UNIT\s+\d+:\s+SL\s+\S+:\s+\S+\s+(?:${POE}\s+|)\d+(?:-|)port\s+\S+\s+Module\s*$$
`,
		data: `
show version

  Copyright (c) 1996-2016 Brocade Communications Systems, Inc. All rights reserved.
  UNIT 1: compiled on May 19 2016 at 01:15:45 labeled as ICX64S08030h
	  (8500344 bytes) from Primary ICX64S08030h.bin
	  SW: Version 08.0.30hT311
  UNIT 2: compiled on May 19 2016 at 01:15:45 labeled as ICX64S08030h
	  (8500344 bytes) from Primary ICX64S08030h.bin
	  SW: Version 08.0.30hT311
  UNIT 3: compiled on May 19 2016 at 01:15:45 labeled as ICX64S08030h
	  (8500344 bytes) from Primary ICX64S08030h.bin
	  SW: Version 08.0.30hT311
  UNIT 4: compiled on May 19 2016 at 01:15:45 labeled as ICX64S08030h
	  (8500344 bytes) from Primary ICX64S08030h.bin
	  SW: Version 08.0.30hT311
  UNIT 5: compiled on May 19 2016 at 01:15:45 labeled as ICX64S08030h
	  (8500344 bytes) from Primary ICX64S08030h.bin
	  SW: Version 08.0.30hT311
  UNIT 6: compiled on May 19 2016 at 01:15:45 labeled as ICX64S08030h
	  (8500344 bytes) from Primary ICX64S08030h.bin
	  SW: Version 08.0.30hT311
  Boot-Monitor Image size = 786944, Version:10.1.05T310 (kxz10105)
  HW: Stackable ICX6450-48-HPOE
==========================================================================
UNIT 1: SL 1: ICX6450-48P POE 48-port Management Module
	Serial  #: BZT3217M025
	License: BASE_SOFT_PACKAGE   (LID: dbvIHGMoFHK)
	P-ENGINE  0: type DEF0, rev 01
	P-ENGINE  1: type DEF0, rev 01
==========================================================================
UNIT 1: SL 2: ICX6450-SFP-Plus 4port 40G Module
`,
		dict: []map[string]interface{}{
			{"SWITCH_ID": []string{"1", "2", "3", "4", "5", "6"}, "POE": []string{"POE", ""}},
		},
	},
	{
		name: "Test match number based regex",
		template: `Value List DHCP_SELECTION ([linksubet\-co]+)
Value List DHCP_SERVER (\d+\.\d+\.\d+\.\d+)

Start
	^\s+dhcp\-server(\s+){0,1}(${DHCP_SELECTION}{0,1})\s+${DHCP_SERVER}\s*

`,
		data: `
	dhcp-server 10.10.10.10
	dhcp-server link-selection 10.10.10.11
	dhcp-server subnet-selection 10.10.10.12
	   `,
		dict: []map[string]interface{}{
			{"DHCP_SELECTION": []string{"", "link-selection", "subnet-selection"}, "DHCP_SERVER": []string{"10.10.10.10", "10.10.10.11", "10.10.10.12"}},
		},
	},
	{
		name: "Last row empty, but has a filldown",
		template: `Value Filldown MODE (((\w+)(\s)?){1,})
Value ALIAS (\w+)
Value COMMAND ((\w+(\s|\S))+)

Start
  ^${MODE}\saliases: -> ALIAS
  # Capture time-stamp if vty line has command time-stamping turned on
  ^Load\s+for\s+
  ^Time\s+source\s+is

ALIAS
  ^${MODE}\saliases: -> ALIAS  
  ^\s+${ALIAS}\s+${COMMAND} -> Record
`,
		data: `
Exec mode aliases:
  h                     help

ATM virtual circuit configuration mode aliases:
  vbr                   vbr-nrt
`,
		dict: []map[string]interface{}{
			{"MODE": "Exec mode", "ALIAS": "h", "COMMAND": "help"},
			{"MODE": "ATM virtual circuit configuration mode", "ALIAS": "vbr", "COMMAND": "vbr-nrt"},
			{"MODE": "ATM virtual circuit configuration mode", "ALIAS": "", "COMMAND": ""},
		},
	},
}
