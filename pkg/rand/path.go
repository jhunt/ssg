package rand

func Path(lens ...int) string {
	if len(lens) == 0 {
		lens = []int{4, 4, 16, 48}
	}
	s := String(lens[0])
	for i := range lens[1:] {
		s += "/" + String(lens[1+i])
	}
	return s
}
