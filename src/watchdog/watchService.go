// Package provides functions to monitor services and attempt to restart them if needed.
// If it's not possible to start service than appopriate notification is send.
package watchdog

import (
	"../settings"
	"log"
	"sync"
	"time"
	"os/exec"
	"../notification"
	"bytes"
	"strings"
)

const(
	nameCmd = "systemctl"
	startCmd = "start"
	statusCmd = "is-active"
)

type serviceManager struct {
	service string // service name
	taskSettings settings.Settings // current settings
	settingChannel chan settings.Settings // channel which indicate whether settings changed
	stop chan bool // channel needed for terminating watchdog
	notify notification.Notification // notification sender
}

// Function execute os command and returns it's result
func (sm serviceManager) executeCmd(operation string) (string, bool) {
	cmd := exec.Command(nameCmd,operation,sm.service)
	var result, stderr bytes.Buffer
	cmd.Stdout = &result
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		var builder strings.Builder
		builder.WriteString(err.Error())
		builder.WriteString("\n")
		builder.WriteString(stderr.String())
		return builder.String(), false
	}
	return result.String(), true
}

// Function check if service is running
func (sm serviceManager) checkRunning() bool{
	_, result := sm.executeCmd(statusCmd)
	return result
}

// Function attempts to restart service
func (sm serviceManager) restart() (string, bool){
	return sm.executeCmd(startCmd)
}

// Function manages services.
// It checks wheter service is still running. If not than it attempts to restart.
func (sm *serviceManager) manage(sg *sync.WaitGroup){
	defer sg.Done()
	if isRunning := sm.checkRunning(); isRunning {
		return
	}
	sm.notify.NotifyShutdown(sm.service)
	log.Println("Service", sm.service, "is not running. Attempting restart...")
	var i int
	result := ""
	attemptLoop: for i = 0; i < sm.taskSettings.NumOfAttempts; i++{
		var restartSuccesful bool
		result, restartSuccesful = sm.restart()
		if restartSuccesful {
			i++
			sm.notify.NotifyRestart(sm.service,i,notification.SuccededStatus,"")
			log.Println("Restarting service ",sm.service," SUCCEDED after ", i, " attempts")
			return
		}
		waitLoop: for {
			select {
			case sm.taskSettings = <- sm.settingChannel:
				if _, ok := sm.taskSettings.SetOfServices[sm.service] ; !ok || sm.taskSettings.NumOfAttempts <= i {
					i++
					break attemptLoop
				}
				continue
			case <- sm.stop:
				i++
				sm.stop <- true
				break attemptLoop
			case <- time.After(sm.taskSettings.NumOfSecWait):
				break waitLoop
			}
		}
	}
	sm.notify.NotifyRestart(sm.service,i,notification.FailedStatus,result)
	log.Println("Restarting service",sm.service,"FAILED after", i, "attempts","with last error:", result)
}

// Watchdog function that launch service managers and wait for their execution.
func WatchServices(watchSettings settings.Settings,settingChannel chan settings.Settings, notif notification.Notification, stop chan bool,watchdogGroup *sync.WaitGroup){
	defer watchdogGroup.Done()
	terminate := false
	log.Println("Staring watchdog service")
	for !terminate {
		waitChannel := make(chan bool,1)
		var serviceGroup sync.WaitGroup
		serviceNumber := len(watchSettings.SetOfServices)
		log.Println("Number of services: ",serviceNumber)
		serviceGroup.Add(serviceNumber)
		arrayChannels := make([]chan settings.Settings,serviceNumber)
		for i := range arrayChannels {
			arrayChannels[i] = make(chan settings.Settings,1)
		}
		i := 0
		for k, _ := range watchSettings.SetOfServices {
			sm := &serviceManager{
				service:      k,
				taskSettings: watchSettings,
				settingChannel: arrayChannels[i],
				stop: stop,
				notify: notif,
			}
			go sm.manage(&serviceGroup)
			i++
		}
		go func (){
			serviceGroup.Wait()
			waitChannel <- true
		} ()
		serviceWait: for {
			select {
			case watchSettings = <- settingChannel:
				log.Println("Settings change detected during execution of services check")
				for i = range arrayChannels {
					arrayChannels[i] <- watchSettings
				}
				continue
			case <- waitChannel:
				log.Println("Finished waiting for services")
				break serviceWait
			}
		}
		checkWait: for {
			select {
			case terminate = <- stop:
				stop <- true
				break checkWait
			case watchSettings = <-settingChannel:
				log.Println("Settings change detected during wait time")
				continue
			case <-time.After(watchSettings.NumOfSecCheck):
				break checkWait
			}
		}
	}
	log.Println("Watchdog terminated")
}

