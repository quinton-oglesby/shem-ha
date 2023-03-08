package main

var (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

// Function to check if a string is in an array, returns true or false.
func stringInArray(str string, list []string) bool {
	for _, i := range list {
		if i == str {
			return true
		}
	}

	return false
}

// Function to remove a string from an array, returning the newly updated array.
func removeStringFromArray(str string, list []string) []string {
	for i, j := range list {
		if j == str {
			return append(list[:i], list[i+1:]...)
		}
	}

	return list
}
