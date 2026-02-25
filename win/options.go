package win

// Option is a functional option to the window constructor New.
type Option func(*options)

type options struct {
	title         string
	width, height int
	resizable     bool
	borderless    bool
	maximized     bool
}

// Title option sets the title (caption) of the window.
func Title(title string) Option {
	return func(o *options) {
		o.title = title
	}
}

// Size option sets the width and height of the window.
func Size(width, height int) Option {
	return func(o *options) {
		o.width = width
		o.height = height
	}
}

// Resizable option makes the window resizable by the user.
func Resizable() Option {
	return func(o *options) {
		o.resizable = true
	}
}

// Borderless option makes the window borderless.
func Borderless() Option {
	return func(o *options) {
		o.borderless = true
	}
}

// Maximized option makes the window start maximized.
func Maximized() Option {
	return func(o *options) {
		o.maximized = true
	}
}
