package focas

import "strings"

func trimNull(s string) string {
	return strings.TrimRight(s, "\x00")
}
