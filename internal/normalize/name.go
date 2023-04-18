package normalize

import (
	"strings"
)

func Name(str string) string {
	str = normalize(str)
	str = strings.ToLower(str)
	return str
}
