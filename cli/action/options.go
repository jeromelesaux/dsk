package action

type Options struct {
	quiet        bool
	format       bool
	force        bool
	analyze      bool
	dataFormat   bool
	vendorFormat bool
	stdout       bool
	hidden       bool
	removeHeader bool
	rawImport    bool
	rawExport    bool
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) WithQuiet(quiet bool) *Options {
	o.quiet = quiet
	return o
}

func (o *Options) WithFormat(format bool) *Options {
	o.format = format
	return o
}

func (o *Options) WithForce(force bool) *Options {
	o.force = force
	return o
}

func (o *Options) WithAnalyze(analyze bool) *Options {
	o.analyze = analyze
	return o
}

func (o *Options) WithDataFormat(dataFormat bool) *Options {
	o.dataFormat = dataFormat
	return o
}

func (o *Options) WithVendorFormat(vendorFormat bool) *Options {
	o.vendorFormat = vendorFormat
	return o
}

func (o *Options) WithStdout(stdout bool) *Options {
	o.stdout = stdout
	return o
}

func (o *Options) WithHidden(hidden bool) *Options {
	o.hidden = hidden
	return o
}

func (o *Options) WithRemoveHeader(removeHeader bool) *Options {
	o.removeHeader = removeHeader
	return o
}

func (o *Options) WithRawImport(rawImport bool) *Options {
	o.rawImport = rawImport
	return o
}

func (o *Options) WithRawExport(rawExport bool) *Options {
	o.rawExport = rawExport
	return o
}
