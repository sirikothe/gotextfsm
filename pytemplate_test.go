package gotextfsm

import (
	"testing"
)

type pyTemplateTestcase struct {
	input  string
	vars   map[string]interface{}
	output string
	err    bool
}

func TestPyTemplate(t *testing.T) {
	for _, tc := range pyTemplateTestcases {
		output, err := ExecutePythonTemplate(tc.input, tc.vars)
		if tc.err {
			if err == nil {
				t.Errorf("'%s' failed. Expected error, but none found", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("'%s' failed. Expected no error. But found error %s", tc.input, err)
			continue
		}
		if output != tc.output {
			t.Errorf("'%s' failed. Expected outputs dont match (%s, %s)", tc.input, output, tc.output)
		}
	}
	t.Logf("Executed %d test cases", len(pyTemplateTestcases))
}

var pyTemplateTestcases = []pyTemplateTestcase{
	{
		input:  `Hello ${world}`,
		vars:   map[string]interface{}{"world": "Siri"},
		output: `Hello Siri`,
	},
	{
		input:  `Hello $world`,
		vars:   map[string]interface{}{"world": "Siri"},
		output: `Hello Siri`,
	},
	{
		input:  `Hello ${intVal} Hi ${floatVal} ${never ending`,
		vars:   map[string]interface{}{"intVal": 10, "floatVal": 5.2, "never ending": "Dummy"},
		output: `Hello 10 Hi 5.2 ${never ending`,
	},
	{
		input:  `Hello ${top.bottom} Hi ${floatVal}`,
		vars:   map[string]interface{}{"top": map[string]interface{}{"bottom": "Structure"}, "floatVal": 5.2},
		output: `Hello Structure Hi 5.2`,
	},
	{
		input:  `Hello ${top.bottom} Hi $no_variable`,
		vars:   map[string]interface{}{"top": map[string]interface{}{"bottom": "Structure"}},
		output: `Hello Structure Hi $no_variable`,
	},
	{
		input:  `Escape $$ with $ $$$temp`,
		vars:   map[string]interface{}{"world": "Siri", "world123": "Bigger", "temp": "Dummy"},
		output: `Escape $ with $ $Dummy`,
	},
	{
		input:  `Hello $world123 Hi ${world}`,
		vars:   map[string]interface{}{"world": "Siri", "world123": "Bigger"},
		output: `Hello Bigger Hi Siri`,
	},
	{
		input:  `Escape {{ and }}`,
		vars:   map[string]interface{}{"world": "Siri", "world123": "Bigger"},
		output: `Escape {{ and }}`,
	},
}
