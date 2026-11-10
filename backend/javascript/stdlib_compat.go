package javascript

// Packages that are the same functions under another path: the older io/ioutil
// is what os and io have now, and math/rand/v2 is math/rand with the names it
// was given later.
func init() {
	ioutil := map[string]string{}

	for name, helper := range map[string]string{
		"ReadFile": "go2jsOSReadFile", "WriteFile": "go2jsOSWriteFile",
		"ReadAll": "go2jsIOReadAll", "NopCloser": "go2jsIONopCloser",
	} {
		ioutil[name] = helper
	}

	if helper, ok := osFuncsMap()["MkdirTemp"]; ok {
		ioutil["TempDir"] = helper
	}

	if helper, ok := osFuncsMap()["CreateTemp"]; ok {
		ioutil["TempFile"] = helper
	}

	stdlibFuncMaps["io/ioutil"] = ioutil
	stdlibPkgAliases["io/ioutil"] = []string{"ioutil"}

	if value, ok := packageConstants["io.Discard"]; ok {
		packageConstants["ioutil.Discard"] = value
	}

	stdlibFuncMaps["math/rand/v2"] = map[string]string{
		"IntN": "go2jsRandV2IntN", "Int64N": "go2jsRandV2IntN", "Int32N": "go2jsRandV2IntN",
		"UintN": "go2jsRandV2IntN", "Uint64N": "go2jsRandV2IntN", "Uint32N": "go2jsRandV2IntN",
		"Int": "go2jsRandInt", "Int64": "go2jsRandInt63", "Int32": "go2jsRandV2Int32", "Uint32": "go2jsRandV2Uint32",
		"Float64": "go2jsRandFloat64", "Float32": "go2jsRandFloat64", "Perm": "go2jsRandPerm", "Shuffle": "go2jsRandShuffle",
	}
	stdlibPkgAliases["math/rand/v2"] = []string{"rand"}
}

func osFuncsMap() map[string]string {
	return stdlibFuncMaps["os"]
}
