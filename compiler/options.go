package compiler

type Options struct {
	PackageName string
	Module      string
	SourceMap   bool
	Minify      bool
	Strict      bool
}

func DefaultOptions() Options {
	return Options{
		Module: "esm",
		Strict: true,
	}
}

func (o Options) WithModule(module string) Options {
	o.Module = module
	return o
}

func (o Options) WithSourceMap(enabled bool) Options {
	o.SourceMap = enabled
	return o
}

func (o Options) WithMinify(enabled bool) Options {
	o.Minify = enabled
	return o
}

func (o Options) WithStrict(enabled bool) Options {
	o.Strict = enabled
	return o
}
