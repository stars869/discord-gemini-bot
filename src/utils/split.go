package utils

// SplitLongText splits a long string into chunks of maxLength
func SplitLongText(text string, maxLength int) []string {
	if maxLength <= 0 {
		return []string{text}
	}
	var chunks []string
	for len(text) > maxLength {
		chunks = append(chunks, text[:maxLength])
		text = text[maxLength:]
	}
	if len(text) > 0 {
		chunks = append(chunks, text)
	}
	return chunks
}
