package steam

var WearCategoryName = map[int]string{ // TODO игнорируется factory new надо исправить
	0: "Factory New",
	1: "Minimal Wear",
	2: "Field-Tested",
	3: "Well-Worn",
	4: "Battle-Scarred",
}
var WearCategoryValue = map[string]int{
	"Factory New":    0,
	"Minimal Wear":   1,
	"Field-Tested":   2,
	"Well-Worn":      3,
	"Battle-Scarred": 4,
}
