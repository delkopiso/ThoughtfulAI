package sort

const (
	MaxDimension = 150
	MaxVolume    = 1_000_000
	MaxWeight    = 20
)

const (
	Standard Stack = iota
	Special
	Rejected
)

var stackTypeStrings = [...]string{"STANDARD", "SPECIAL", "REJECTED"}

type Stack uint

func (s Stack) String() string { return stackTypeStrings[s] }

func Sort(width, height, length, mass int) Stack {
	heavy := mass >= MaxWeight
	bulky := width > MaxDimension || height > MaxDimension || length > MaxDimension || width*height*length > MaxVolume

	if !heavy && !bulky {
		return Standard
	}
	if heavy && bulky {
		return Rejected
	}
	return Special
}
