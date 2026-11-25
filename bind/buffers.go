package bind

// bufferElements are the typed arrays a function that takes bytes (a BufferSource)
// is given a variant for, each named by what it holds: BufferDataFloat32,
// BufferDataUint16.
var bufferElements = []struct{ suffix, array string }{
	{"Float32", "Float32Array"},
	{"Uint16", "Uint16Array"},
	{"Uint32", "Uint32Array"},
	{"Uint8", "Uint8Array"},
	{"Int32", "Int32Array"},
}

// expandBuffers replaces each function that takes a buffer, which Go cannot say
// without knowing what it holds, by one function for each kind of array.
func expandBuffers(m *Module) {
	m.Functions = expandFunctions(m.Functions)

	for i := range m.Classes {
		m.Classes[i].Methods = expandFunctions(m.Classes[i].Methods)
		m.Classes[i].Statics = expandFunctions(m.Classes[i].Statics)
	}
}

func hasBuffer(f Function) bool {
	for _, p := range f.Params {
		if p.Type.Kind == Typed && p.Type.Buffer {
			return true
		}
	}

	return false
}

func expandFunctions(list []Function) []Function {
	var out []Function

	for _, f := range list {
		if !hasBuffer(f) {
			out = append(out, f)

			continue
		}

		for _, element := range bufferElements {
			variant := f
			variant.Name = f.Name + element.suffix

			if variant.JS == "" {
				variant.JS = f.Name
			}

			variant.Params = append([]Param(nil), f.Params...)

			for i := range variant.Params {
				if t := variant.Params[i].Type; t.Kind == Typed && t.Buffer {
					variant.Params[i].Type = Type{Kind: Typed, Name: element.array}
				}
			}

			out = append(out, variant)
		}
	}

	return out
}
