package common

import (
	"encoding/json"
	"io/ioutil"

	"github.com/moxiaomomo/configcenter/logger"
)

// LoadJSON load jsondata from file
func LoadJSON(filename string) (data map[string][]string, err error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Errorf("Load json file: %s", err.Error())
		return nil, err
	}

	if err := json.Unmarshal(bytes, &data); err != nil {
		logger.Errorf("Unmarshal: %s", err.Error())
		return nil, err
	}
	return
}
