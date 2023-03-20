package komi

// drain will remove any pending values from the channel.
func drain[T any](v chan T) {
	for {
		select {
		case vv := <-v:
			nop(vv)
		default:
			return
		}
	}
}

// nop is a no-op (does nothing).
func nop(v any) {}
