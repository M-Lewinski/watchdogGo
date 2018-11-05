# WatchdogGo
WatchdogGo is a simple program written in Go that launches as daemon to monitor services. During execution if one of services shuts down than watchdog attempts to restart it. In addition, watchdog notifies via email if something goes wrong. Program was written with _systemd_ in mind that's why it will only work on those systems that uses systemd as daemon manager.

## Installation
Required packets:
- go version 1.10 or newer
- go packets: [aws/aws-sdk-go, google/go-cmp, sevlyar/go-daemon]
Script **install.sh** will get all required packets, build program and will automatically copy template files that will be used when starting watchdog.
If it's only necessary to build program, use script **build.sh**.

## Features
- Monitors service and attempts to restart it if necessary, Uses systemctl command to do so. (REQUIRED ROOT)
- Checks if settings were changed during execution of watchdog. If they were than program reacts to it accordingly.
- Notifies via mail if something is wrong with monitored service.
- Uses AWS services to store settings in DynamoDB database and publishes notification via SNS.

## Configuration
It's necessary to modify configuration file with appropriate data.
Configuration file consists of:
```
- DatabaseTable : name of DynamoDB database that will be used to acquire settings
- SnsTopic : name of SNS topic that will be used to publish notifications
- Region : region of AWS services
- NumOfMinCheckSettings : number of minutes that watchdog will wait in order to check if settings were modified
- WatchdogId : id of a watchdog daemon. It's useful for identifying which watchdog instance sent notification
```

Remember to also configure credentials for AWS services (check AWS documentation for more info).

## Usage
Program monitors services across system, that's why **IT'S IMPORTANT TO RUN PROGRAM AS ROOT OR OTHER PRIVILEGED USER**.
```
Usage of watchdogGo:
  -i string [Required]
    	Primary key of a table entry containing configuration of watchdog
  -c string
    	Name of a config file containing for example settings table name (default "config.json")
  -l string
    	Name of a log file (default "watchdogGo.log")
  -m string
    	Name of a mail directory containing mail templates (default "mail")
```
Sending signals **SIGTERM** or **SIGQUIT** properly turns off watchdog instance.
It's possible to run multiple daemon instances simultaneously on the same system. 
#### Example
```
Launch watchdogGo with value 1 as primary key, use newConfigFile.json as config file and logfile.log as log file.
# ./watchdogGo -i 1 -c newConfigFile.json -l logfile.log 
```