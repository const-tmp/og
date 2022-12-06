package names

import "strings"

func MatchProto(name, pb string) bool {
	if name == pb {
		return true
	}
	if strings.Contains(name, "ID") && strings.Replace(name, "ID", "Id", 1) == pb {
		return true
	}
	return false
}
