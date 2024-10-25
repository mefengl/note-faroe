package main

func padEnd(s string, n int) string {
	for len(s) < n {
		s += " "
	}
	return s
}
