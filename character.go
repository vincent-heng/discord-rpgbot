package main

type character struct {
	discordId    int
	class        Class
	experience   int
	level        int
	strength     int
	agility      int
	wisdom       int
	constitution int
	skillPoints  int
	currentHp    int
	stamina      int
}

type Class int

const (
	FIGHTER Class = iota
	MAGE
)

func getClass(classString string) Class {
	return map[string]Class{
		"Combattant": FIGHTER,
		"Magicien":   MAGE,
	}[classString]
}

func (s Class) String() string {
	return [...]string{"Combattant", "Magicien"}[s]
}
