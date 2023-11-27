package deps

// stringsliceContain verifies if 'item' is a part of string slice
func stringsliceContain(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}
