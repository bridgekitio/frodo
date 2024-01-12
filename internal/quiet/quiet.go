package quiet

import "io"

// Close releases the given resource without worrying about whether that was
// successful or not. You typically do this in defers or in situations where
// trying to handle errors doesn't make a damn bit of difference.
func Close(closer io.Closer) {
	if closer != nil {
		_ = closer.Close()
	}
}
