package plugin

import (
	"encoding/json"
	"fmt"
	"strings"
)

type mapStrings map[string]string

type customSettings struct {
	jsonData                mapStrings
	decryptedSecureJSONData mapStrings
	headers                 mapStrings
	queryParameters         mapStrings
}

func NewCustomSettings(jsonData json.RawMessage, decryptedSecureJSONData mapStrings) (*customSettings, error) {
	var jsonMap mapStrings
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		return nil, fmt.Errorf("error converting json.RawMessage to map[string]string")
	}

	customSettings := &customSettings{
		jsonData:                jsonMap,
		decryptedSecureJSONData: decryptedSecureJSONData,
		headers:                 make(mapStrings),
		queryParameters:         make(mapStrings),
	}

	customSettings.getDataSourceCustomSettings()

	return customSettings, nil
}

func (cs *customSettings) getDataSourceCustomSettings() {
	// jsonData has the headers and queryParameters names
	for key, value := range cs.jsonData {
		if strings.HasPrefix(key, "headerName") {
			cs.headers[value] = cs.getCustomSettingValue(key)
		} else if strings.HasPrefix(key, "queryParamName") {
			cs.queryParameters[value] = cs.getCustomSettingValue(key)
		}
	}
}

func (cs *customSettings) getCustomSettingValue(settingName string) string {
	prefixSettingName := ""
	suffixSettingName := ""

	if strings.HasPrefix(settingName, "headerName") || strings.HasPrefix(settingName, "queryParamName") {
		prefixAndSuffix := strings.Split(settingName, "Name")
		prefixSettingName = prefixAndSuffix[0]
		suffixSettingName = prefixAndSuffix[1]
	}

	// decryptedSecureJSONData has the respective values of each header and query parameter name in the jsonData
	settingValue, exists := cs.decryptedSecureJSONData[prefixSettingName+"Value"+suffixSettingName]
	if !exists {
		return "custom setting not found"
	}

	return settingValue
}
