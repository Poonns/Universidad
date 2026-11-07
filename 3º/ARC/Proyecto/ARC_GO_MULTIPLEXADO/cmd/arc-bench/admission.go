package main

func generatorAdmissionArgs(optimized bool) []string {
	if !optimized {
		return nil
	}
	return []string{"-optimize-data-admission"}
}
