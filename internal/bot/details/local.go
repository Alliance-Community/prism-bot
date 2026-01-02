package details

type level struct {
	Name string `json:"Name"`
	Key  string `json:"Key"`
	Size int    `json:"Size"`
}

var (
	layers = map[int]string{
		16:  "INF",
		32:  "ALT",
		64:  "STD",
		128: "LRG",
	}

	gameModes = map[string]string{
		"gpm_cq":         "AAS",
		"gpm_cnc":        "CNC",
		"gpm_coop":       "Co-Op",
		"gpm_insurgency": "INS",
		"gpm_skirmish":   "Skirmish",
		"gpm_vehicles":   "Vehicle Warfare",
		"gpm_gungame":    "Gungame",
	}
)
