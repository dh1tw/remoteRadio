package utils

func SliceDiff(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for _, s1 := range slice1 {
		found := false
		for _, s2 := range slice2 {
			if s1 == s2 {
				found = true
				break
			}
		}
		// String not found. We add it to return slice
		if !found {
			diff = append(diff, s1)
		}
	}
	return diff
}

func StringMapDiff(map1 map[string]float64, map2 map[string]float64) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for key, _ := range map1 {
		_, found := map2[key]
		// String not found. We add it to return slice
		if !found {
			diff = append(diff, key)
		}
	}
	return diff
}
