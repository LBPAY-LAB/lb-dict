package entities

import "regexp"

// isValidISPB checks if an ISPB (participant code) is valid (8 digits)
func isValidISPB(ispb string) bool {
	return regexp.MustCompile(`^\d{8}$`).MatchString(ispb)
}
