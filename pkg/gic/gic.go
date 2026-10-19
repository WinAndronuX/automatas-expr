package gic

type Productions map[rune]string

type GIC struct {
	VN []rune
	VT []rune
	P  Productions
	S  rune
}
