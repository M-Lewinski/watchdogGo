package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"sync"

	"./config"
	"./settings"
	"./watchdog"
	"./notification"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/sevlyar/go-daemon"
)

const (
	parameterError     = iota             // Return error code when something goes wrong while parsing parameters
	configFileError    = iota             // Return error code when something goes wrong while reading config file
	settingsError      = iota             // Return error code when something goes wrong while getting settings info
	mailError          = iota             // Return error code when something goes wrong with mail notification
	logFileNameDefault = "watchdogGo.log" // watchdog log file
	mailDirDefault     = "mail"           // mail directory with templates
	logWorkingDir      = "./"             // working directory for log file

)

var (
	configFileName	= flag.String("c","config.json", "Name of a config file containing for example settings table name")
	primaryKey	= flag.String("i","", "Primary key of a table entry containing configuration of watchdog")
	logFile = flag.String("l", logFileNameDefault,"Name of a log file")
	mailDir = flag.String("m", mailDirDefault, "Name of a mail directory containing mail templates")
	stop = make(chan bool,1)
	settingsChannel = make(chan settings.Settings,1)
)


// Main function
func main()  {
	flag.Parse()
	if *primaryKey == "" {
		log.Fatalln("Passed primary key is missing: ",*primaryKey)
		os.Exit(parameterError)
	}
	if strings.TrimSpace(*mailDir) == ""{
		log.Fatalln("Passed mail directory is incorrect: ",*mailDir)
	}
	if strings.TrimSpace(*configFileName) == "" {
		log.Fatalln("Passed config file name is incorrect: ",*configFileName)
	}
	configFile, err := config.ParseConfigFileJson(*configFileName)
	if err != nil {
		os.Exit(configFileError)
	}
	if !configFile.Validate(){
		log.Fatalln("Configuration is not valid")
		os.Exit(configFileError)
	}
	if err != nil {
		log.Fatalln("Session error: ", err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(configFile.Region)},
	)
	sd:= settings.ConnectToDatabase(sess,configFile.DatabaseTable,*primaryKey)
	sett, err  := sd.GetSettings()
	if err != nil {
		log.Fatalln(err)
		os.Exit(settingsError)
	}
	notif, err := notification.ConnectToNotificationService(sess,configFile.SnsTopic,configFile.WatchdogId, *mailDir)
	if err != nil {
		log.Fatalln(err)
		os.Exit(mailError)
	}
	context := &daemon.Context{
		LogFileName: *logFile,
		LogFilePerm: 0644,
		WorkDir:     logWorkingDir,
	}
	child, err := context.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if child != nil {
		return // Exit parent process, it's no longer needed
	}

	defer context.Release()

	sig := make(chan os.Signal,1)
	signal.Notify(sig,syscall.SIGTERM,syscall.SIGQUIT)
	go func () {
		sc := <- sig
		log.Println("Received signal: ", sc)
		log.Println("Shutting down in progress...")
		stop <- true
	} ()

	var watchGroup sync.WaitGroup
	watchGroup.Add(2)
	go watchdog.WatchServices(*sett,settingsChannel,*notif,stop,&watchGroup)
	go watchdog.WatchSettings(*sett,sd,time.Duration(configFile.NumOfMinCheckSettings)*time.Minute,settingsChannel,stop,&watchGroup)
	watchGroup.Wait()
	log.Println("Shut down completed")
}