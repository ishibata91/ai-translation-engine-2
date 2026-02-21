package dictionary_builder

// Config holds the configuration for the DictionaryBuilderSlice.
type Config struct {
	// AllowedRECTypes contains the list of REC types (e.g., "BOOK:FULL", "NPC_:FULL")
	// that should be extracted from the XML file and saved to the dictionary.
	AllowedRECTypes []string `json:"allowed_rec_types" mapstructure:"allowed_rec_types"`
}

// DefaultConfig returns a Config populated with standard default values
// suitable for typical Skyrim Mod Translation.
func DefaultConfig() Config {
	return Config{
		AllowedRECTypes: []string{
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
		},
	}
}

// IsAllowedREC checks if a given REC type string is in the AllowedRECTypes list.
func (c *Config) IsAllowedREC(recType string) bool {
	for _, allowed := range c.AllowedRECTypes {
		if recType == allowed {
			return true
		}
	}
	return false
}
