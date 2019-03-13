# WatchdogGo
WatchdogGo is a simple program written in Go language that launches as daemon to monitor services. During execution if one of services is not running then watchdog attempts to restart it. In addition, watchdog notifies via email if something goes wrong with specific service. Program was written with _systemd_ in mind that's why it will only work on those systems that uses systemd as daemon manager.

## Installation
Required packets:
- go version 1.10 or newer
- go packets: [aws/aws-sdk-go, google/go-cmp, sevlyar/go-daemon]
Script **install.sh** will get all required packets, build program and will automatically copy template files that will be used when starting watchdog.
If it's only necessary to build program, use script **build.sh**.

## Features
- Monitors service and attempts to restart it if necessary. Program uses systemctl command to do so. (REQUIRED ROOT)
- Checks if settings were changed during execution of watchdog. If they were then program react to it accordingly.
- Notifies via email if something is wrong with monitored service.
- Uses AWS services to store settings in DynamoDB database and publish notifications via SNS.

## Configuration

#### Local configuration file
It's necessary to modify configuration file with appropriate data.
Configuration file consists of:
```
- DatabaseTable : name of DynamoDB database that will be used to acquire settings
- SnsTopic : name of SNS topic that will be used to publish notifications
- Region : region of AWS services
- NumOfMinCheckSettings : number of minutes that watchdog will wait in order to check if settings were modified
- WatchdogId : id of a watchdog daemon. It's used for identifying which watchdog instance sent notification
```

Remember to also configure credentials for AWS services (check AWS documentation for more info).

#### Database configuration
Currently watchdog communicate with DynamoDB to acquire necessary information about list of services and settings.
DynamoDB scheme consists of:
```
- Id : identification used to find specific settings
- ListOfServices : list of services that watchdog should monitor
- NumOfSecCheck : number of seconds between consecutive checks of services
- NumOfSecWait : number of seconds that watchdog wait for between checking attempts
- NumOfAttempts : number of check attempts. If attempts exceed provided value, than watchdog will consider service as faulty.
```
##### Example
```
{
  “Id”: “1”,
  “ListOfServices”: [“mysqld”, “docker”],
  “NumOfSecCheck” : 60,
  “NumOfSecWait”: 10,
  “NumOfAttempts”: 4
}
```

File **settings.json** in directory **example** contains example of DynamoDB scheme configuration.

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
Sending signal **SIGTERM** or **SIGQUIT** will shut down watchdog instance cleanly.
It's possible to run multiple daemon instances simultaneously on the same system. 
#### Example
```
Launch watchdogGo with value 1 as primary key, use newConfigFile.json as config file and logfile.log as log file.
# ./watchdogGo -i 1 -c newConfigFile.json -l logfile.log 
```
