package main

import "strconv"

// Applied only to clientArgs, never to the shared/server arguments. Preserve
// historical default invocations and pass only the existing client CLI knob.
func generatorPriorityArgs(burst int) []string {
	if burst == 8 {
		return nil
	}
	return []string{"-high-burst", strconv.Itoa(burst)}
}

func generatorHighCapacityArgs(capacity int) []string {
	if capacity == 1024 {
		return nil
	}
	return []string{"-high-capacity", strconv.Itoa(capacity)}
}
