package crawler

import (
	"encoding/json"
	"os"
)

func NewWebsite(path string) *Website {

	w, err := LoadConfig(path)
	if err != nil {
		return nil
	}
	return w
}

// LoadConfig loads the configuration from a JSON file
func LoadConfig(configPath string) (*Website, error) {
	w := &Website{}

	// os.ReadFile is the direct replacement for ioutil.ReadFile
	data, err := os.ReadFile("./config/" + configPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}
