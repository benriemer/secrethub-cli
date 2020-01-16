package secrethub

import (
	"testing"

	"github.com/secrethub/secrethub-cli/internals/cli/ui"

	"github.com/secrethub/secrethub-cli/internals/secrethub/tpl"
	"github.com/secrethub/secrethub-go/internals/assert"
)

func TestVariableReader(t *testing.T) {
	cases := map[string]struct {
		osEnv               map[string]string
		commandTemplateVars map[string]string
		constructorErr      error
		variableToRead      string
		expectedValue       string
		readErr             error
	}{
		"os_environment_variable_success": {
			osEnv: map[string]string{
				"DIFFERENT_PREFIX_TEST":          "other_test_value",
				templateVarEnvVarPrefix + "TEST": "test_value",
				"TEST":                           "yet_another_test_value",
			},
			commandTemplateVars: nil,
			constructorErr:      nil,
			variableToRead:      "test",
			expectedValue:       "test_value",
			readErr:             nil,
		},
		"command_template_vars_success": {
			osEnv: nil,
			commandTemplateVars: map[string]string{
				templateVarEnvVarPrefix + "TEST": "test_value",
				"DIFFERENT_PREFIX_TEST":          "other_test_value",
				"TEST":                           "yet_another_test_value",
			},
			constructorErr: nil,
			variableToRead: "test",
			expectedValue:  "yet_another_test_value",
			readErr:        nil,
		},
		"variable_not_existent": {
			osEnv: map[string]string{
				templateVarEnvVarPrefix + "TEST1": "testA",
				templateVarEnvVarPrefix + "TEST2": "testB",
			},
			commandTemplateVars: map[string]string{
				"test3": "testC",
				"test4": "testD",
			},
			constructorErr: nil,
			variableToRead: "test5",
			expectedValue:  "",
			readErr:        tpl.ErrTemplateVarNotFound("test5"),
		},
		"os_var_name_not_posix": {
			osEnv: map[string]string{
				templateVarEnvVarPrefix + "TEST-1": "testA",
				templateVarEnvVarPrefix + "TEST2":  "testB",
			},
			commandTemplateVars: map[string]string{
				"test3": "testC",
				"test4": "testD",
			},
			constructorErr: ErrInvalidTemplateVar("test-1"),
			variableToRead: "",
			expectedValue:  "",
			readErr:        nil,
		},
		"command_var_name_not_posix": {
			osEnv: map[string]string{
				templateVarEnvVarPrefix + "TEST1": "testA",
				templateVarEnvVarPrefix + "TEST2": "testB",
			},
			commandTemplateVars: map[string]string{
				"3test3": "testC",
				"test4":  "testD",
			},
			constructorErr: ErrInvalidTemplateVar("3test3"),
			variableToRead: "",
			expectedValue:  "",
			readErr:        nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			reader, err := newVariableReader(tc.osEnv, tc.commandTemplateVars)
			if err != nil {
				assert.Equal(t, err, tc.constructorErr)
				return
			}

			value, err := reader.ReadVariable(tc.variableToRead)
			if err != nil {
				assert.Equal(t, err, tc.readErr)
			}

			assert.Equal(t, value, tc.expectedValue)
		})
	}
}

func TestPromptVariableReader(t *testing.T) {
	osEnv := map[string]string{
		templateVarEnvVarPrefix + "TEST1": "testA",
		templateVarEnvVarPrefix + "TEST2": "testB",
	}
	commandTemplateVars := map[string]string{
		"test3": "testC",
		"test4": "testD",
		"test1": "testAA",
	}

	reader, err := newVariableReader(osEnv, commandTemplateVars)
	assert.OK(t, err)

	t.Run("prompt", func(t *testing.T) {
		io := ui.NewFakeIO()
		io.PromptIn.Reads = []string{"foobar\n"}
		reader = newPromptMissingVariableReader(reader, io)

		val, err := reader.ReadVariable("test5")
		assert.Equal(t, val, "foobar")
		assert.Equal(t, err, nil)
	})

	t.Run("no prompt", func(t *testing.T) {
		io := ui.NewFakeIO()
		reader = newPromptMissingVariableReader(reader, io)

		val, err := reader.ReadVariable("test4")
		assert.Equal(t, val, "testD")
		assert.Equal(t, err, nil)
	})

	t.Run("from os env", func(t *testing.T) {
		io := ui.NewFakeIO()
		reader = newPromptMissingVariableReader(reader, io)

		val, err := reader.ReadVariable("test2")
		assert.Equal(t, val, "testB")
		assert.Equal(t, err, nil)
	})

	t.Run("template vars shadow os env", func(t *testing.T) {
		io := ui.NewFakeIO()
		reader = newPromptMissingVariableReader(reader, io)

		val, err := reader.ReadVariable("test1")
		assert.Equal(t, val, "testAA")
		assert.Equal(t, err, nil)
	})
}
