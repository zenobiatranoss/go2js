package javascript

import "strings"

// Export is a declaration of a program that was asked to be handed over to
// whatever runs it, under the name it is handed over under.
type Export struct {
	// Name is the name the declaration was written under in the program, which is
	// what it is called inside it.
	Name string

	// Alias is the name it is reached under from outside it, which is another name
	// when one was asked for and the name of the declaration when none was.
	Alias string

	// Func says whether the declaration is a function, which matters because a
	// function of a Go program is written as a generator and a generator is not
	// something JavaScript calls as though it were a function: it has to be run to
	// the end before it has a value to hand back.
	Func bool
}

// exportSuffix is what the declaration of a function is renamed to where it is
// handed over from, so that what is handed over is a function JavaScript can call
// rather than the generator the program was written with. A declaration of
// anything else is handed over as it stands.
const exportSuffix = "$go2js_export"

// ExportLocalName is the name a declaration is known by inside the program it was
// declared in, which is a function of a Go program under a name of its own, since
// it is the generator the program was written with that is run to its end here.
func ExportLocalName(export Export) string {
	if export.Func {
		return export.Name + exportSuffix
	}

	return export.Name
}

// ExportExportedName is the name a declaration is reached under from outside the
// program it was declared in.
func ExportExportedName(export Export) string {
	if export.Alias != "" {
		return export.Alias
	}

	return export.Name
}

// ExportDeclarations are the declarations of what a program hands over, written
// where the declarations they hand over are all in place, so that a function of a
// Go program is handed over as something JavaScript can call: a function of a Go
// program is written as a generator, which is run to its end here rather than
// being handed over as it stands.
func ExportDeclarations(exports []Export) string {
	var out strings.Builder
	seen := make(map[string]bool)

	for _, export := range exports {
		local := ExportLocalName(export)

		if export.Name == "" || seen[local] {
			continue
		}

		seen[local] = true

		if !export.Func {
			continue
		}

		// What is handed over is a function JavaScript can call, which runs the
		// call as a goroutine of the program would and hands back what it came to,
		// so a Go function called from JavaScript returns what a Go function
		// returns rather than a generator that has not been run.
		out.WriteString("const " + local + " = (...go2jsArgs) => go2jsCallFromJavaScript(" + export.Name + ", null, go2jsArgs);\n")
	}

	return out.String()
}

// ExportHandlers are the exports of a program written the way the module they are
// in hands things over, which is one line per declaration: a module written as an
// ES module names what it exports, a module written for CommonJS puts each of
// them on what it is reached under, and a program written as an expression has
// nowhere of its own to hand anything to, so it hands nothing over.
func ExportHandlers(exports []Export, options ModuleOptions) string {
	options = options.Normalize()

	var out strings.Builder
	seen := make(map[string]bool)

	for _, export := range exports {
		exported := ExportExportedName(export)
		local := ExportLocalName(export)

		if export.Name == "" || seen[exported] {
			continue
		}

		seen[exported] = true
		out.WriteString(moduleExportAs(local, exported, options.Format))
	}

	return out.String()
}
