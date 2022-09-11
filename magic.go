package domainset

import _ "unsafe"

//go:linkname parseCapacityHint github.com/database64128/shadowsocks-go/domainset.parseCapacityHint
//go:noescape
func parseCapacityHint(line string) ([4]int, bool, error)
