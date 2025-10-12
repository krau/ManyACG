package shared

//go:generate go-enum --values --names --nocase

// R18Type
/*
ENUM(
none
r18
all
)
*/
type R18Type uint

func R18TypeFromInt(i int) R18Type {
	switch i {
	case 0:
		return R18TypeNone
	case 1:
		return R18TypeR18
	default:
		return R18TypeAll
	}
}
