package client

func splitInto(n int) func([]byte, bool) (int, []byte, error) {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if len(data) >= n {
			return n, data[0:n], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	}
}
