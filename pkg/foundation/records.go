package foundation

// DictionaryImportRECTypes is the shared REC allow-list used by dictionary import
// and terminology target extraction.
var DictionaryImportRECTypes = []string{
	"BOOK:FULL",
	"NPC_:FULL",
	"NPC_:SHRT",
	"ARMO:FULL",
	"WEAP:FULL",
	"LCTN:FULL",
	"CELL:FULL",
	"CONT:FULL",
	"MISC:FULL",
	"ALCH:FULL",
	"FURN:FULL",
	"DOOR:FULL",
	"RACE:FULL",
	"INGR:FULL",
	"FLOR:FULL",
	"SHOU:FULL",
}

// IsDictionaryImportREC reports whether recType is part of the shared import allow-list.
func IsDictionaryImportREC(recType string) bool {
	for _, allowed := range DictionaryImportRECTypes {
		if recType == allowed {
			return true
		}
	}
	return false
}
