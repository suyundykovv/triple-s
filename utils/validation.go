package utils

const (
	maxObjectKeyLength = 1024
)

func isValidCharacter(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '.' || c == '-' ||
		c == '_' || c == '/' || c == '+'
}

func IsValidObjectKey(key string) bool {
	if len(key) == 0 || len(key) > maxObjectKeyLength {
		return false
	}
	for _, c := range key {
		if !isValidCharacter(c) {
			return false
		}
	}
	return true
}
