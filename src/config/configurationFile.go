// Package config provides functions to read file and parse it into structure.
// Structure contains necessary information such as settings table name and SNS topic name.
package config

import (
	"encoding/json"
	"os"
	"log"
	"io/ioutil"
	)

type ConfigFile struct {
	DatabaseTable         string
	SnsTopic              string
	Region                string
	NumOfMinCheckSettings float64
	WatchdogId string
}

// Function parse file containing json into ConfigFile structure.
func ParseConfigFileJson(fileName string) (*ConfigFile,error){
	jsonFile, err := os.Open(fileName);
	if err != nil {
		log.Fatalln("Open file error: ",err)
		return nil,err
	}
	defer jsonFile.Close()
	byteFile, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatalln("Read file error: ",err)
		return nil,err
	}
	configFile := &ConfigFile{}
	json.Unmarshal(byteFile, configFile)
	return configFile, nil
}

// Function validate config file
func (cf ConfigFile) Validate() bool {
	test := true
	test = test && cf.DatabaseTable != ""
	test = test && cf.SnsTopic!= ""
	test = test && cf.Region != ""
	test = test && cf.NumOfMinCheckSettings > 0.0
	test = test && cf.WatchdogId != ""
	return test
}