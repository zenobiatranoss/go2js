package compiler

import (
	"fmt"
	"go/token"
	"strings"
)

type DiagnosticSeverity uint8

const (
	SeverityError DiagnosticSeverity = iota
	SeverityWarning
	SeverityInfo
)

type Diagnostic struct {
	Message  string
	Position token.Position
	Severity DiagnosticSeverity
}

type Diagnostics struct {
	Items []Diagnostic
}

func (d *Diagnostics) Add(message string) {
	d.AddAt(token.Position{}, SeverityError, message)
}

func (d *Diagnostics) AddAt(position token.Position, severity DiagnosticSeverity, message string) {
	if strings.TrimSpace(message) == "" {
		return
	}

	d.Items = append(d.Items, Diagnostic{
		Message:  message,
		Position: position,
		Severity: severity,
	})
}

func (d *Diagnostics) ErrorAt(position token.Position, message string) {
	d.AddAt(position, SeverityError, message)
}

func (d *Diagnostics) WarningAt(position token.Position, message string) {
	d.AddAt(position, SeverityWarning, message)
}

func (d *Diagnostics) InfoAt(position token.Position, message string) {
	d.AddAt(position, SeverityInfo, message)
}

func (d *Diagnostics) Empty() bool {
	return len(d.Items) == 0
}

func (d *Diagnostics) Len() int {
	return len(d.Items)
}

func (d *Diagnostics) Errors() int {
	count := 0

	for _, item := range d.Items {
		if item.Severity == SeverityError {
			count++
		}
	}

	return count
}

func (d *Diagnostics) Warnings() int {
	count := 0

	for _, item := range d.Items {
		if item.Severity == SeverityWarning {
			count++
		}
	}

	return count
}

func (d *Diagnostics) HasErrors() bool {
	return d.Errors() > 0
}

func (d *Diagnostics) First() *Diagnostic {
	if len(d.Items) == 0 {
		return nil
	}

	return &d.Items[0]
}

func (d *Diagnostics) Last() *Diagnostic {
	if len(d.Items) == 0 {
		return nil
	}

	return &d.Items[len(d.Items)-1]
}

func (d *Diagnostics) Reset() {
	d.Items = d.Items[:0]
}

func (d *Diagnostics) Error() error {
	if d.Empty() {
		return nil
	}

	if len(d.Items) == 1 {
		return fmt.Errorf("%s", formatDiagnostic(d.Items[0]))
	}

	return fmt.Errorf("%s", d.String())
}

func (d *Diagnostics) String() string {
	if d.Empty() {
		return ""
	}

	var builder strings.Builder

	for i, item := range d.Items {
		if i > 0 {
			builder.WriteByte('\n')
		}

		builder.WriteString(formatDiagnostic(item))
	}

	return builder.String()
}

func formatDiagnostic(d Diagnostic) string {
	prefix := severityName(d.Severity)

	if d.Position.IsValid() {
		return fmt.Sprintf("%s:%d:%d: %s: %s",
			d.Position.Filename,
			d.Position.Line,
			d.Position.Column,
			prefix,
			d.Message,
		)
	}

	return fmt.Sprintf("%s: %s", prefix, d.Message)
}

func severityName(severity DiagnosticSeverity) string {
	switch severity {
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "error"
	}
}
