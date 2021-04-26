package supervisor

// Logger is a simple interface for logging output during the execution
// of a supervision tree. Note that in an attempt at making this package
// agnostic, the function signatures are amongst the most common in the
// main logging packages.
type Logger interface {
	// Println is the standard level.
	Println(string)
}

var logger Logger

// WithLogger sets the `Logger` for this package; by default logging data
// is just discarded.
func WithLogger(l Logger) {
	logger = l
}

func log(msg string) {
	if logger != nil {
		logger.Println(msg)
	}
}
