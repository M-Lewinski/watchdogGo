// Package provides functions to monitor settings source and checks if settings changed.
// If new settings are invalid than the old ones are kept.
package watchdog

import (
	"../settings"
	"sync"
	"time"
	"log"
)

// Watchdog function that checks wheter settings changed
func WatchSettings(currentSettings settings.Settings, src settings.Source, waitTime time.Duration, settingsChannel chan settings.Settings, stop chan bool, watchdogGroup *sync.WaitGroup) {
	defer watchdogGroup.Done()
	terminate := false
	for !terminate{
		newSettings, err := src.GetSettings()
		if err != nil {
			log.Fatalln(err)
			log.Fatalln("Keeping old settings")
		} else if !currentSettings.Equal(*newSettings) {
			log.Println("Detected settings change")
			currentSettings = *newSettings
			settingsChannel <- currentSettings
		}
		waitLoop: for {
			select {
			case terminate = <- stop:
				stop <- true
				break waitLoop
			case <- time.After(waitTime):
				break waitLoop
			}
		}
	}
}