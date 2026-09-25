package javascript

func genericRuntimeSource() string {
	return `
function go2jsZero(type) {
	switch (type) {
	case "bool":
		return false
	case "string":
		return ""
	case "int":
	case "int8":
	case "int16":
	case "int32":
	case "int64":
	case "uint":
	case "uint8":
	case "uint16":
	case "uint32":
	case "uint64":
	case "uintptr":
	case "float32":
	case "float64":
	case "complex64":
	case "complex128":
		return 0
	case "pointer":
	case "interface":
	case "slice":
	case "map":
	case "chan":
	case "function":
		return null
	case "array":
	case "struct":
		return {}
	default:
		return undefined
	}
}

function go2jsConvert(type, value) {
	switch (type) {
	case "bool":
		return Boolean(value)
	case "string":
		return String(value)
	case "int":
	case "int8":
	case "int16":
	case "int32":
	case "int64":
	case "uint":
	case "uint8":
	case "uint16":
	case "uint32":
	case "uint64":
	case "uintptr":
		return Math.trunc(value)
	case "float32":
	case "float64":
		return Number(value)
	case "complex64":
	case "complex128":
		return value
	default:
		return value
	}
}
`
}
