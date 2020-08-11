package gotextfsm

import (
	"regexp"
	"testing"
)

type fsmTestCase struct {
	name   string
	input  string
	values map[string]string
	states map[string][]string
	err    *regexp.Regexp
}

func TestFSMParse(t *testing.T) {
	tc_count := 0
	for _, tc := range fsmtestcases {
		// fmt.Printf("Test case %s\n", tc.name)
		tc_count++
		v := TextFSM{}
		err := v.ParseString(tc.input)
		if tc.err != nil {
			if err == nil {
				t.Errorf("'%s' failed. Expected error, but none found", tc.name)
			}
			continue
		}
		if err != nil {
			t.Errorf("'%s' failed. Expected no error. But found error '%s'", tc.name, err)
			continue
		}
		if tc.values != nil {
			if len(tc.values) != len(v.Values) {
				t.Errorf("'%s' failed. Expected %d values found %d", tc.name, len(tc.values), len(v.Values))
				continue
			}
			failed := false
			for name, val := range tc.values {
				if curval, exists := v.Values[name]; exists {
					if val != curval.String() {
						t.Errorf("'%s' failed. Expected %s. Found %s", tc.name, val, curval.String())
						failed = true
						break
					}
				} else {
					t.Errorf("'%s' failed. Expected %s. Found none", tc.name, val)
					failed = true
					break
				}
			}
			if failed {
				continue
			}
		} else if v.Values != nil && len(v.Values) > 0 {
			t.Errorf("'%s' failed. No values expected. But %d values found", tc.name, len(v.Values))
			continue
		}
		if tc.states != nil {
			if len(tc.states) != len(v.States) {
				t.Errorf("'%s' failed. Expected %d states found %d", tc.name, len(tc.states), len(v.States))
				continue
			}
			for name, val := range tc.states {
				if curval, exists := v.States[name]; exists {
					if len(val) != len(curval.rules) {
						t.Errorf("'%s' failed. Expected %d rules in state %s. Found %d", tc.name, len(val), name, len(curval.rules))
						break
					}
					failed := false
					for i, rule := range val {
						if rule != curval.rules[i].String() {
							failed = true
							t.Errorf("'%s' failed. State %s: Rule doesnt match ('%s', '%s')", tc.name, name, rule, curval.rules[i].String())
						}
					}
					if failed {
						break
					}
				}
			}
		}
	}
	t.Logf("Executed %d test cases", tc_count)
}

var fsmtestcases = []fsmTestCase{
	{
		name:  "Null template",
		input: "",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "No state definition 1",
		input: "Value beer (.*)",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "No state definition 2",
		input: "Value beer (.*)\n\n",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Missing Start",
		input: "Value beer (.*)\n\nHello\n  ^.*",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Invalid rule 1",
		input: "Value unused (.)\n\nStart\n A simple string.",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Invalid rule 2",
		input: "Value unused (.)\n\nStart\n.^A simple string.",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Invalid rule 3",
		input: "Value unused (.)\n\nStart\n\tA simple string.",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Invalid rule 4",
		input: "Value unused (.)\n\nStart\nA simple string.",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:   "Simple Start state 1",
		input:  "Value unused (.)\n\nStart\n ^A simple string.",
		values: map[string]string{"unused": "Value unused (.)"},
		states: map[string][]string{"Start": []string{" ^A simple string."}},
	},
	{
		name:   "Simple Start state 2",
		input:  "Value unused (.)\n\nStart\n  ^A simple string.",
		values: map[string]string{"unused": "Value unused (.)"},
		states: map[string][]string{"Start": []string{" ^A simple string."}},
	},
	{
		name:   "Simple Start state 3",
		input:  "Value unused (.)\n\nStart\n\t^A simple string.",
		values: map[string]string{"unused": "Value unused (.)"},
		states: map[string][]string{"Start": []string{" ^A simple string."}},
	},
	{
		name:   "Empty Start State",
		input:  "Value unused (.)\n\nStart",
		values: map[string]string{"unused": "Value unused (.)"},
		states: map[string][]string{"Start": []string{}},
	},
	{
		name:   "Empty NON-Start State",
		input:  "Value unused (.)\n\nStart\n  ^.*\n\nEMPTY",
		values: map[string]string{"unused": "Value unused (.)"},
		states: map[string][]string{"Start": []string{" ^.*"}, "EMPTY": []string{}},
	},
	{
		name:   "Empty NON-Start State",
		input:  "Value unused (.)\n\nStart\n  ^.*\n\n#Comment",
		values: map[string]string{"unused": "Value unused (.)"},
		states: map[string][]string{"Start": []string{" ^.*"}},
	},
	{
		name:   "Empty Start with Filldown",
		input:  "Value Filldown Beer (beer)\n\nStart\n",
		values: map[string]string{"Beer": "Value Filldown Beer (beer)"},
		states: map[string][]string{"Start": []string{}},
	},
	{
		name:   "Single variable with commented header",
		input:  "# Headline\nValue Filldown Beer (beer)\n\nStart\n",
		values: map[string]string{"Beer": "Value Filldown Beer (beer)"},
		states: map[string][]string{"Start": []string{}},
	},
	{
		name: "Multiple variables 1",
		input: `# Headline
Value Filldown Beer (beer)
Value Required Spirits (whiskey)
Value Filldown Wine (claret)

Start
`,
		values: map[string]string{"Beer": "Value Filldown Beer (beer)", "Spirits": "Value Required Spirits (whiskey)", "Wine": "Value Filldown Wine (claret)"},
		states: map[string][]string{"Start": []string{}},
	},
	{
		name: "Multiple variables 2",
		input: `# Headline
Value Filldown Beer (beer)
# A Comment
Value Required Spirits ()
Value Filldown,Required Wine ((c|C)laret)

Start
`,
		values: map[string]string{"Beer": "Value Filldown Beer (beer)", "Spirits": "Value Required Spirits ()", "Wine": "Value Filldown,Required Wine ((c|C)laret)"},
		states: map[string][]string{"Start": []string{}},
	},
	{
		name: "Values that look bad but are OK",
		input: `# Headline
Value Filldown Beer (bee(r), (and) (M)ead$)
# A Comment
Value Spirits,and,some ()
Value Filldown,Required Wine ((c|C)laret)

Start
`,
		values: map[string]string{
			"Beer":             "Value Filldown Beer (bee(r), (and) (M)ead$)",
			"Spirits,and,some": "Value Spirits,and,some ()",
			"Wine":             "Value Filldown,Required Wine ((c|C)laret)"},
		states: map[string][]string{"Start": []string{}},
	},
	{
		name:   "Two simple variables",
		input:  "Value Beer (.)\nValue Wine (\\w)\n\nStart\n",
		values: map[string]string{"Beer": "Value Beer (.)", "Wine": "Value Wine (\\w)"},
		states: map[string][]string{"Start": []string{}},
	},
	{
		name:  "Duplicate Start state",
		input: "Value Beer (.)\nValue Wine (\\w)\n\nStart\nStart\n",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:   "Single State",
		input:  "# Headline\nValue Beer (.)\n\n\nStart\n ^.\n",
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{"Start": []string{" ^."}},
	},
	{
		name:   "Multiple rules",
		input:  "# Headline\nValue Beer (.)\n\nStart\n  ^Hello World\n#Comment In Rules\n  ^Last-[Cc]ha$$nge\n",
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{"Start": []string{" ^Hello World", " ^Last-[Cc]ha$$nge"}},
	},
	{
		name:  "Malformed states 1",
		input: "Value Beer (.)\n\nSt%art\n  ^.\n  ^Hello World\n",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Malformed states 2",
		input: "Value Beer (.)\n\nStart\n^.\n  ^Hello World\n",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Malformed states 3",
		input: "Value Beer (.)\n\n  Start\n  ^.\n  ^Hello World\n",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Malformed states 4",
		input: "Value Beer (.)\n\n  Start\n  ^.\n  ^Hello World\n",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name: "Multiple variables and substitution",
		input: `Value Beer (.)

Start
  ^.${Beer}${Wine}.
  ^Hello $Beer
  ^Last-[Cc]ha$$nge
`,
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{"Start": []string{
			" ^.${Beer}${Wine}.",
			" ^Hello $Beer",
			" ^Last-[Cc]ha$$nge",
		}},
	},
	{
		name:  "State name too long (>32 char)",
		input: "Value Beer (.)\n\nrnametoolong_nametoolong_nametoolong_nametoolong_nametoolo\n  ^.\n  ^Hello World\n",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Continue should not accept a destination",
		input: "Value Beer (.)\n\nStart\n  ^.* -> Continue Start\n",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name: "Error accepts a text string",
		input: `Value Beer (.)

Start
  ^.* -> Error "hi there"
`,
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{"Start": []string{` ^.* -> Error "hi there"`}},
	},
	{
		name: "Next does not accept a text string",
		input: `Value Beer (.)

Start
  ^.* -> Next "hi there"
`,
		err: regexp.MustCompile(`.+`),
	},
	{
		name:  "Invalid rule - 2 variables",
		input: `Value Beer (.)\nValue Wine (\\w)\n\nStart\n  A Simple line`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "No Values",
		input: `\nNotStart\n`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "No States",
		input: `Value unused (.)\n\n`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "No Start State",
		input: `Value unused (.)\n\nNotStart\n ^.*`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:   "Has Start state with valid destination",
		input:  "Value Beer (.)\n\nStart\n ^.* -> Start",
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{"Start": []string{` ^.* -> Start`}},
	},
	{
		name:  "Invalid destination",
		input: "Value Beer (.)\n\nStart\n ^.* -> Start\n ^.* -> bogus",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:   "valid destinations",
		input:  "Value Beer (.)\n\nStart\n ^.* -> Start\n ^.* -> bogus\n\nbogus\n ^.* -> Start",
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{
			"Start": []string{` ^.* -> Start`, ` ^.* -> bogus`},
			"bogus": []string{` ^.* -> Start`},
		},
	},
	{
		name:   "valid destination with options",
		input:  "Value Beer (.)\n\nStart\n ^.* -> Start\n ^.* -> bogus\n\nbogus\n ^.* -> Next.Record Start",
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{
			"Start": []string{` ^.* -> Start`, ` ^.* -> bogus`},
			"bogus": []string{` ^.* -> Next.Record Start`},
		},
	},
	{
		name:   "Error without messages",
		input:  "Value Beer (.)\n\nStart\n ^.* -> Start\n ^.* -> bogus\n\nbogus\n ^.* -> Error",
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{
			"Start": []string{` ^.* -> Start`, ` ^.* -> bogus`},
			"bogus": []string{` ^.* -> Error`},
		},
	},
	{
		name:   "Error with messages",
		input:  "Value Beer (.)\n\nStart\n ^.* -> Start\n ^.* -> bogus\n\nbogus\n ^.* -> Error \"Boo hoo\"",
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{
			"Start": []string{` ^.* -> Start`, ` ^.* -> bogus`},
			"bogus": []string{` ^.* -> Error "Boo hoo"`},
		},
	},
	{
		name:  "No State definition",
		input: "Value Beer (.)",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:   "No Line operation, but Record operation",
		input:  "Value Beer (.)\n\nStart\n ^.* -> Start\n ^.* -> bogus\n\nbogus\n ^.* -> Record",
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{
			"Start": []string{` ^.* -> Start`, ` ^.* -> bogus`},
			"bogus": []string{` ^.* -> Record`},
		},
	},
	{
		name:   "No Line operation, but Record operation + new state",
		input:  "Value Beer (.)\n\nStart\n ^.* -> Start\n ^.* -> bogus\n\nbogus\n ^.* -> Record Start",
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{
			"Start": []string{` ^.* -> Start`, ` ^.* -> bogus`},
			"bogus": []string{` ^.* -> Record Start`},
		},
	},
	{
		name:  "No empty line after Values",
		input: "Value Beer (.)\nStart",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "State name as a key word",
		input: "Value Beer (.)\n\nRecord\n ^.*",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Duplicate non Start state",
		input: "Value Beer (.)\n\nStart\n ^.*\n\nbogus\n ^.*\n\nbogus\n ^.*",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Non empty End state",
		input: "Value Beer (.)\n\nStart\n ^.*\n\nEnd\n ^.*\n",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:   "Remove End State",
		input:  "Value Beer (.)\n\nStart\n ^.*\n\nEnd\n",
		values: map[string]string{"Beer": "Value Beer (.)"},
		states: map[string][]string{
			"Start": []string{` ^.*`},
		},
	},
	{
		name:  "Non empty EOF state",
		input: "Value Beer (.)\n\nStart\n ^.*\n\nEOF\n ^.*\n",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name:  "Non alphanumeric char in state name",
		input: "Value Beer (.)\n\nStart\n ^.*\n\nDUMMY\n ^.* -> \"f_$f\"\n\nf_$f\n",
		err:   regexp.MustCompile(`.+`),
	},
	{
		name: "Test Nested with conflict",
		input: `Value List foo ((?P<name>\w+)\s+(?P<name>\w+):\s+(?P<age>\d+)\s+(?P<state>\w{2})\s*)

Start
  ^\s*${foo}
  ^
  ^\s*$$ -> Record
`,
		err: regexp.MustCompile(`.+`),
	},
	{
		name: "Complext template with no options",
		input: `# Header
# Header 2
Value Beer (.*)
Value Wine (\\w+)

# An explanation with a unicode character Δ
Start
  ^hi there ${Wine}. -> Next.Record State1

State1
  ^\\wΔ
  ^$Beer .. -> Start
# Some comments
  ^$$ -> Next
  ^$$ -> End

End
# Tail comment.`,
		values: map[string]string{
			"Beer": "Value Beer (.*)",
			"Wine": `Value Wine (\\w+)`,
		},
		states: map[string][]string{
			"Start":  []string{` ^hi there ${Wine}. -> Next.Record State1`},
			"State1": []string{` ^\\wΔ`, ` ^$Beer .. -> Start`, ` ^$$ -> Next`, ` ^$$ -> End`},
		},
	},
	{
		// TODO: Is this a valid template or should this be an error?
		name:   "Regex in variable name",
		input:  "Value Filldown B.*r (beer)\n\nStart\n",
		values: map[string]string{"B.*r": "Value Filldown B.*r (beer)"},
		states: map[string][]string{"Start": []string{}},
	},
	{
		name: "Template with { in regex",
		input: `Value INBOUND_SETTINGS_IN_USE (.*)

Start
	^\s+in\s+use\s+settings\s+=\{${INBOUND_SETTINGS_IN_USE},\s+\}\s*

EOF
`,
		values: map[string]string{"INBOUND_SETTINGS_IN_USE": "Value INBOUND_SETTINGS_IN_USE (.*)"},
		states: map[string][]string{"Start": []string{` ^\s+in\s+use\s+settings\s+=\{${INBOUND_SETTINGS_IN_USE},\s+\}\s*`}, "EOF": []string{}},
	},
}
