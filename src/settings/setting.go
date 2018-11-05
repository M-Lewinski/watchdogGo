package settings

import (
	"github.com/google/go-cmp/cmp"
	"time"
	)

type Settings struct {
	Id            string              `json:"id"`
	SetOfServices map[string]struct{} `json:",omitempty"`
	ListOfServices []string				`json:"listOfServices"`
	NumOfSecCheck time.Duration       `json:"numOfSecCheck"`
	NumOfSecWait  time.Duration       `json:"numOfSecWait"`
	NumOfAttempts int                 `json:"numOfAttempts"`
}

type Source interface {
	GetSettings() (*Settings, error)
	GetSettingsByIndex(string) (*Settings, error)
}

func (primary Settings) Equal(secondary Settings) bool{
	test := true
	test = test && primary.Id == secondary.Id
	test = test && cmp.Equal(primary.SetOfServices,secondary.SetOfServices)
	test = test && primary.NumOfSecCheck == secondary.NumOfSecCheck
	test = test && primary.NumOfSecWait == secondary.NumOfSecWait
	test = test && primary.NumOfAttempts == secondary.NumOfAttempts
	return test
}

func (sett Settings) validate() bool{
	test := true
	test = test && sett.Id != ""
	test = test && len(sett.ListOfServices) > 0
	test = test && sett.NumOfSecWait > 0
	test = test && sett.NumOfSecCheck > 0
	test = test && sett.NumOfAttempts > 0
	return test
}

func (sett *Settings) Unmarshal(){
	sett.SetOfServices = map[string]struct{}{}
	sett.NumOfSecWait *= time.Second
	sett.NumOfSecCheck *= time.Second
	for _, elem := range sett.ListOfServices {
		sett.SetOfServices[elem] = struct{}{}
	}
}

