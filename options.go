package ecsgo

type options struct {
	// tick count per a second
	fps int

	// sets tick interval time always same
	fixedTick bool
}

type option func(opts *options)

// FPS sets tick count per a second
func FPS(fps int) option {
	return func(opts *options) {
		opts.fps = fps
	}
}

// FixedTick sets deltaseconds same even tick is slow
func FixedTick(fixedTick bool) option {
	return func(opts *options) {
		opts.fixedTick = fixedTick
	}
}
