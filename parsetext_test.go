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
		// t.Logf("Running test case %s\n", tc.name)
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
		data: " Bob: 32 NC\n Alice: 27 NY\n Jeff: 45 CA\nJulia\n\n\n\nSiri",
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
			{"boo": "e", "hoo": "ten"},
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
		name: "Test Practical case 1",
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
}
