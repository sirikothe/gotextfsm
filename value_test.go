package gotextfsm

import (
	"regexp"
	"testing"
)

type valTestCase struct {
	input   string
	name    string
	regex   string
	options map[string]bool
	err     *regexp.Regexp
}

func TestValueParse(t *testing.T) {
	for _, tc := range valTestCases {
		v := TextFSMValue{}
		err := v.Parse(tc.input, 0)
		if tc.err != nil {
			if err == nil {
				t.Errorf("'%s' failed. Expected error, but none found", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("'%s' failed. Expected no error. But found error %s", tc.input, err)
			continue
		}
		if tc.name != "" {
			if tc.name != v.Name {
				t.Errorf("'%s' failed. Names dont match ('%s', '%s')", tc.input, tc.name, v.Name)
			}
		} else if v.Name != "" {
			t.Errorf("'%s failed. Names dont match ('%s', '%s')", tc.input, tc.name, v.Name)
		}
		if tc.regex != "" {
			if tc.regex != v.Regex {
				t.Errorf("'%s failed. Regex dont match ('%s', '%s')", tc.regex, tc.regex, v.Regex)
			}
		} else if v.Regex != "" {
			t.Errorf("'%s failed. Regex dont match ('%s', '%s')", tc.regex, tc.regex, v.Regex)
		}
		if len(tc.options) != len(v.Options) {
			t.Errorf("'%s' failed. Number options dont match (%d, %d)", tc.input, len(tc.options), len(v.Options))
		} else {
			if len(tc.options) > 0 {
				for option := range tc.options {
					if idx := FindIndex(v.Options, option); idx < 0 {
						t.Errorf("'%s' failed. Expected '%s' option. But not found", tc.input, option)
					}
				}
			}
		}
	}
	t.Logf("Executed %d test cases", len(valTestCases))

}

var valTestCases = []valTestCase{
	{
		input: "Hello World",
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: "Value name regex",
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: "Value name (reg(ex",
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: "Value name (regex)",
		name:  "name",
		regex: "(regex)",
	},
	{
		input:   "Value Filldown variable (regex)",
		name:    "variable",
		regex:   "(regex)",
		options: map[string]bool{"Filldown": true},
	},
	{
		input: "Value Filldown,INVALID name (regex)",
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: "Value Key,Filldown,Required,Filldown name (regex)",
		err:   regexp.MustCompile(`.+`),
	},
	{
		input:   "Value Filldown,Key name (regex)",
		name:    "name",
		regex:   "(regex)",
		options: map[string]bool{"Filldown": true, "Key": true},
	},
	{
		input:   "Value Required name (regex)",
		name:    "name",
		regex:   "(regex)",
		options: map[string]bool{"Required": true},
	},
	{
		input:   "Value Key name (reg[(]ex)",
		name:    "name",
		regex:   "(reg[(]ex)",
		options: map[string]bool{"Key": true},
	},
	{
		input:   `Value List beer (\S+)`,
		name:    "beer",
		regex:   `(\S+)`,
		options: map[string]bool{"List": true},
	},
	{
		input:   `Value Filldown,Required beer (\S+)`,
		name:    "beer",
		regex:   `(\S+)`,
		options: map[string]bool{"Filldown": true, "Required": true},
	},
	{
		input:   `Value Fillup beer (boo(hoo))`,
		name:    "beer",
		regex:   `(boo(hoo))`,
		options: map[string]bool{"Fillup": true},
	},
	{
		input: `Value beer (boo(hoo)))boo`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: `Value beer boo(boo(hoo)))`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: `Value beer (boo)hoo)`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: `Value beer (boo[(]hoo)`,
		name:  "beer",
		regex: `(boo[(]hoo)`,
	},
	{
		input: `Value beer (boo\[)\]hoo)`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: `Value beerrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr (boo\[)\]hoo)`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: `Value Beer (beer) beer`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: `Value Filldown, Required Spirits ()`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: `Value filldown,Required Wine ((c|C)laret)`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input:   `Value Filldown,Required Wine ((c|C)laret)`,
		name:    "Wine",
		regex:   `((c|C)laret)`,
		options: map[string]bool{"Filldown": true, "Required": true},
	},
	{
		input:   `Value Filldown Beer (bee(r), (and) (M)ead$)`,
		name:    "Beer",
		regex:   `(bee(r), (and) (M)ead$)`,
		options: map[string]bool{"Filldown": true},
	},
	{
		input: `Value Spirits,and,some ()`,
		name:  "Spirits,and,some",
		regex: `()`,
	},
	{
		input: `Value beer (\\S+Δ)`,
		name:  "beer",
		regex: `(\\S+Δ)`,
	},
	{
		input: `Value para_beer (\()`,
		name:  "para_beer",
		regex: `(\()`,
	},
	{
		// Test regular expression with []
		input: `Value beer ([(\S+\s\S+)]+)`,
		name:  "beer",
		regex: `([(\S+\s\S+)]+)`,
	},
}
