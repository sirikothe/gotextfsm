package gotextfsm

import (
	"regexp"
	"testing"
)

type ruleTestCase struct {
	input     string
	match     string
	line_op   string
	record_op string
	new_state string
	err       *regexp.Regexp
}

func TestRuleParse(t *testing.T) {
	for _, tc := range ruleTestCases {
		v := TextFSMRule{}
		err := v.Parse(tc.input, 1, nil)
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
		if tc.line_op != v.LineOp {
			t.Errorf("'%s' failed. line_op dont match ('%s', '%s')", tc.input, tc.line_op, v.LineOp)
		}
		if tc.record_op != v.RecordOp {
			t.Errorf("'%s' failed. record_op dont match ('%s', '%s')", tc.input, tc.record_op, v.RecordOp)
		}
		if tc.new_state != v.NewState {
			t.Errorf("'%s' failed. new_state dont match ('%s', '%s')", tc.input, tc.new_state, v.NewState)
		}
	}
	t.Logf("Executed %d test cases", len(ruleTestCases))
}

var ruleTestCases = []ruleTestCase{
	{
		input: `  `,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input:     "  ^A beer called ${beer}",
		match:     "^A beer called ${beer}",
		line_op:   "",
		record_op: "",
		new_state: "",
	},
	{
		input:     "  ^A $hi called ${beer}",
		match:     "^A $hi called ${beer}",
		line_op:   "",
		record_op: "",
		new_state: "",
	},
	{
		input:     "  ^A beer called ${beer} -> Next",
		match:     "^A beer called ${beer}",
		line_op:   "Next",
		record_op: "",
		new_state: "",
	},
	{
		input:     "  ^A beer called ${beer} -> Continue.Record",
		match:     "^A beer called ${beer}",
		line_op:   "Continue",
		record_op: "Record",
		new_state: "",
	},
	{
		input:     "  ^A beer called ${beer} -> Next.NoRecord End",
		match:     "^A beer called ${beer}",
		line_op:   "Next",
		record_op: "NoRecord",
		new_state: "End",
	},
	{
		input: "  ^A beer called ${beer} -> Next Next Next",
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: "  ^A beer called ${beer} -> Boo.hoo",
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: "  ^A beer called ${beer} -> Continue.Record $Hi",
		err:   regexp.MustCompile(`.+`),
	},
	{
		input:     "  ^A beer called ${beer} -> Record End",
		match:     "^A beer called ${beer}",
		record_op: "Record",
		new_state: "End",
	},
	{
		input:     "  ^A beer called ${beer} -> End",
		match:     "^A beer called ${beer}",
		record_op: "",
		new_state: "End",
	},
	{
		input:     "  ^A beer called ${beer} -> Next.NoRecord End",
		match:     "^A beer called ${beer}",
		line_op:   "Next",
		record_op: "NoRecord",
		new_state: "End",
	},
	{
		input:     "  ^A beer called ${beer} -> Clear End",
		match:     "^A beer called ${beer}",
		record_op: "Clear",
		new_state: "End",
	},
	{
		input:     `  ^A beer called ${beer} -> Error "Hello World"`,
		match:     "^A beer called ${beer}",
		line_op:   "Error",
		new_state: `"Hello World"`,
	},
	{
		input: `  ^A beer called ${beer} -> Next "Hello World`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: `  ^A beer called ${beer} -> Record.Next`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: `  ^A beer called ${beer} -> Continue End`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input: `  ^A beer called ${beer} -> Beer End`,
		err:   regexp.MustCompile(`.+`),
	},
	{
		input:     `  ^Hello World -> Boo`,
		match:     "^Hello World",
		new_state: "Boo",
	},
	{
		input:     `  ^Hello World ->  Boo`,
		match:     "^Hello World",
		new_state: "Boo",
	},
	{
		input:     `  ^Hello World ->   Boo`,
		match:     "^Hello World",
		new_state: "Boo",
	},
	// A '->' without a leading space is considered part of the matching line.
	{
		input:   `  A simple line-> Boo -> Next`,
		match:   "  A simple line-> Boo",
		line_op: "Next",
	},
	// Alpha numaric characters only in state names
	{
		input: `  A simple line -> Next +`,
		err:   regexp.MustCompile(`.+`),
	},
	// Unicode characters
	{
		input: `  ^A beer called ${beer}Δ`,
		match: `^A beer called ${beer}Δ`,
	},
	// Invalid regular expression
	{
		input: `  ^A beer called .+++ ${beer}Δ`,
		err:   regexp.MustCompile(`.+`),
	},
}
