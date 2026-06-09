package models

var ValidSnackTypes = []string{
	"Candy",
	"Baked Goods",
	"Beverages",
	"Pantry",
	"Chips/Crackers",
}

func IsValidSnackType(value string) bool {
	for _, t := range ValidSnackTypes {
		if t == value {
			return true
		}
	}
	return false
}
