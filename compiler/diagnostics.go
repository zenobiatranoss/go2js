package compiler

import (
	"errors"
	"fmt"
	"go/scanner"
	gotypes "go/types"

	"github.com/zenobiatranoss/go2js/backend/javascript"
)

type DiagnosticSeverity string

const (
	DiagnosticError   DiagnosticSeverity = "error"
	DiagnosticWarning DiagnosticSeverity = "warning"
)

type Diagnostic struct {
	Severity DiagnosticSeverity
	Phase    string
	Message  string
	Filename string
	Line     int
	Column   int
}

type Diagnostics struct {
	Items []Diagnostic
}

func (d *Diagnostics) Add(diagnostic Diagnostic) {
	if diagnostic.Message == "" {
		return
	}
	d.Items = append(d.Items, diagnostic)
}

func (d Diagnostics) Empty() bool {
	return len(d.Items) == 0
}

func (d Diagnostics) Error() error {
	if d.Empty() {
		return nil
	}

	return errors.New(d.Items[0].String())
}

func (d Diagnostic) String() string {
	location := ""

	if d.Filename != "" {
		location = d.Filename

		if d.Line > 0 {
			location += fmt.Sprintf(":%d", d.Line)

			if d.Column > 0 {
				location += fmt.Sprintf(":%d", d.Column)
			}
		}

		location += ": "
	}

	if d.Phase != "" {
		return fmt.Sprintf("%s: %s%s", d.Phase, location, d.Message)
	}

	return location + d.Message
}

func DiagnosticFromError(err error, phase, filename string) Diagnostic {
	result := Diagnostic{
		Severity: DiagnosticError,
		Phase:    phase,
		Message:  err.Error(),
		Filename: filename,
	}

	var scanErrors scanner.ErrorList
	if errors.As(err, &scanErrors) && len(scanErrors) > 0 {
		pos := scanErrors[0].Pos
		result.Filename = pos.Filename
		result.Line = pos.Line
		result.Column = pos.Column
		result.Message = scanErrors[0].Msg
		return result
	}

	var typeError *gotypes.Error
	if errors.As(err, &typeError) && typeError != nil {
		pos := typeError.Fset.Position(typeError.Pos)
		result.Filename = pos.Filename
		result.Line = pos.Line
		result.Column = pos.Column
		result.Message = typeError.Msg
		return result
	}

	return result
}

type CompileResult struct {
	Code        string
	Diagnostics Diagnostics
}

func (r CompileResult) Err() error {
	return r.Diagnostics.Error()
}

func (c *Compiler) CompileFileDetailed(path string) CompileResult {
	result := CompileResult{}

	if c == nil {
		result.Diagnostics.Add(Diagnostic{
			Severity: DiagnosticError,
			Phase:    "compiler",
			Message:  "nil compiler",
			Filename: path,
		})
		return result
	}

	if path == "" {
		result.Diagnostics.Add(Diagnostic{
			Severity: DiagnosticError,
			Phase:    "input",
			Message:  "empty source path",
		})
		return result
	}

	parsed, err := ParseFile(path)
	if err != nil {
		result.Diagnostics.Add(DiagnosticFromError(err, "parse", path))
		return result
	}

	analysis, err := Analyze(parsed)
	if err != nil {
		result.Diagnostics.Add(DiagnosticFromError(err, "typecheck", path))
		return result
	}

	code, err := javascriptEmitDetailed(c, parsed, analysis)
	if err != nil {
		result.Diagnostics.Add(DiagnosticFromError(err, "emit", path))
		return result
	}

	result.Code = code
	return result
}

func javascriptEmitDetailed(c *Compiler, parsed *ParsedFile, analysis *Analysis) (string, error) {
	return c.compileParsedFile(parsed, analysis)
}

func (c *Compiler) compileParsedFile(parsed *ParsedFile, analysis *Analysis) (string, error) {
	if c == nil {
		return "", fmt.Errorf("nil compiler")
	}

	if parsed == nil || parsed.File == nil {
		return "", fmt.Errorf("invalid parsed file")
	}

	if analysis == nil || analysis.Types == nil {
		return "", fmt.Errorf("missing analysis")
	}

	output, err := javascript.EmitWithContextOptionsTarget(
		parsed.File,
		analysis.Types,
		analysis.Semantic,
		c.Options.Runtime,
		c.Options.Target,
	)
	if err != nil {
		return "", err
	}

	return c.wrap(output), nil
}
