package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var templateFuncs = template.FuncMap{
	"add":                      func(a, b int) int { return a + b },
	"trim":                     strings.TrimSpace,
	"trimRightSpace":           trimRightSpace,
	"trimTrailingWhitespaces":  trimRightSpace,
	"rpad":                     rpad,
	"gt":                       cobra.Gt,
	"hasSubCommands":           hasSubCommands,
	"hasManagementSubCommands": hasManagementSubCommands,
	"operationSubCommands":     operationSubCommands,
	"managementSubCommands":    managementSubCommands,
	"decoratedName":            decoratedName,
	"useLine":                  useLine,
	"hasArgs":                  hasArgs,
	"argUsages":                argUsages,
	"flagUsages":               flagUsages,
}

// Tmpl executes the given template text on data, writing the result to w.
func Tmpl(w io.Writer, text string, data interface{}) error {
	t := template.New("top")
	t.Funcs(templateFuncs)
	template.Must(t.Parse(text))
	return t.Execute(w, data)
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

func hasSubCommands(cmd *cobra.Command) bool {
	return len(operationSubCommands(cmd)) > 0
}

func hasManagementSubCommands(cmd *cobra.Command) bool {
	return len(managementSubCommands(cmd)) > 0
}

func operationSubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && !sub.HasSubCommands() {
			cmds = append(cmds, sub)
		}
	}
	return cmds
}

func managementSubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, sub := range (*cmd).Commands() {
		if sub.IsAvailableCommand() && sub.HasSubCommands() {
			cmds = append(cmds, sub)
		}
	}
	return cmds
}

func decoratedName(cmd cobra.Command) string {
	return cmd.Name() + " "
}

func hasArgs(args []Argument) bool {
	return len(args) > 0
}

func argUsages(args []Argument) string {
	buf := new(bytes.Buffer)
	lines := make([]string, 0, len(args))
	maxlen := 0

	for _, arg := range args {
		line := "  "
		if arg.Placeholder != "" {
			line += arg.Placeholder
		} else {
			line += "<" + arg.Name + ">"
		}

		line += "\x00"
		if len(line) > maxlen {
			maxlen = len(line)
		}

		line += arg.Description
		lines = append(lines, line)
	}

	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		spacing := strings.Repeat(" ", maxlen-sidx)
		// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
		fmt.Fprintln(buf, line[:sidx], spacing, strings.Replace(line[sidx+1:], "\n", "\n"+strings.Repeat(" ", maxlen+2), -1))
	}

	return buf.String()
}

func useLine(c *cobra.Command, args []Argument) string {
	var useLine string

	if c.HasParent() {
		useLine = c.Parent().CommandPath() + " " + c.Use
	} else {
		useLine = c.Use
	}

	if c.HasAvailableFlags() && !strings.Contains(useLine, "[flags]") {
		useLine += " [flags]"
	}

	for _, arg := range args {
		if arg.Placeholder != "" {
			useLine += " " + arg.Placeholder
		} else {
			useLine += " <" + arg.Name + ">"
		}
	}

	return useLine
}

func flagUsages(c *CommandClause) string {
	flagSet := c.Flags()
	buf := new(bytes.Buffer)

	lines := make([]string, 0, flagSet.NFlag())

	maxlen := 0

	flagSet.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}

		line := ""
		if f.Shorthand != "" && f.ShorthandDeprecated == "" {
			line = fmt.Sprintf("  -%s, --%s", f.Shorthand, f.Name)
		} else {
			line = fmt.Sprintf("      --%s", f.Name)
		}
		if f.NoOptDefVal != "" {
			switch f.Value.Type() {
			case "string":
				line += fmt.Sprintf("[=\"%s\"]", f.NoOptDefVal)
			case "bool":
				if f.NoOptDefVal != "true" {
					line += fmt.Sprintf("[=%s]", f.NoOptDefVal)
				}
			case "count":
				if f.NoOptDefVal != "+1" {
					line += fmt.Sprintf("[=%s]", f.NoOptDefVal)
				}
			default:
				line += fmt.Sprintf("[=%s]", f.NoOptDefVal)
			}
		}

		_, usage := pflag.UnquoteUsage(f)
		// This special character will be replaced with spacing once the
		// correct alignment is calculated
		line += "\x00"
		if len(line) > maxlen {
			maxlen = len(line)
		}
		line += usage

		if c.Flag(f.Name).envVar != "" && f.Name != "help" {
			line += " ($" + c.Flag(f.Name).envVar + ")"
		}

		lines = append(lines, line)
	})

	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		spacing := strings.Repeat(" ", maxlen-sidx)
		// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
		fmt.Fprintln(buf, line[:sidx], spacing, strings.Replace(line[sidx+1:], "\n", "\n"+strings.Repeat(" ", maxlen+2), -1))
	}

	return buf.String()
}

var UsageTemplate = `Usage:
{{if not .Cmd.HasSubCommands}} {{(useLine .Cmd .Args)}}{{end}}
{{- if .Cmd.HasSubCommands}}  {{ .Cmd.CommandPath}}{{- if .Cmd.HasAvailableFlags}} [flags]{{end}} [command]{{end}}

{{if ne .Cmd.Long ""}}{{ .Cmd.Long | trim }}{{ else }}{{ .Cmd.Short | trim }}{{end}}
{{- if gt (len .Cmd.Aliases) 0}}

Aliases:
  {{.Cmd.NameAndAliases}}
{{- end}}
{{- if .Cmd.HasExample}}
Examples:
{{ .Cmd.Example }}
{{- end}}
{{- if hasManagementSubCommands .Cmd }}

Management Commands:
{{- range managementSubCommands .Cmd }}
  {{rpad (decoratedName .) (add .NamePadding 1)}}{{.Short}}
{{- end}}
{{- end}}
{{- if hasSubCommands .Cmd}}

Commands:
{{- range operationSubCommands .Cmd }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}
{{- if .Cmd.HasAvailableLocalFlags}}

Flags:
{{flagUsages . | trimTrailingWhitespaces}}
{{- end}}
{{- if .Cmd.HasAvailableInheritedFlags}}

Global Flags:
{{flagUsages .Cmd.InheritedFlags | trimTrailingWhitespaces}}
{{- end}}
{{- if hasArgs .Args}}

Arguments:
{{argUsages .Args}}
{{- end}}
{{- if .Cmd.HasAvailableSubCommands}}

Use "{{.Cmd.CommandPath}} [command] --help" for more information about a command.
{{- end}}
`

var HelpTemplate = `
{{if or .Cmd.Runnable .Cmd.HasSubCommands}}{{.Cmd.UsageString}}{{end}}`
