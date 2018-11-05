package notification

import (
	"text/template"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
			"strings"
	"log"
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"errors"
)

const (
	messageFile    = "message"
	subjectFile    = "subject"
	shutdownMail   = "shutdown"
	restartMail    = "restart"
	subjectSize    = 100 // Mail subject has to be less than 100 CHARACTERS (According to AWS documentation)
	messageSize    = 262144 // Mail message has to be less than 262144 BYTES (According to AWS documentation)
	FailedStatus   = iota
	SuccededStatus = iota

)

var mailNames = []string{shutdownMail,restartMail}

type mail struct {
	Subject *template.Template
	Message *template.Template
}

type SnsNotification struct {
	Name string
	Id string
	mails map[string] mail
	svc *sns.SNS
	topicARN *string
	workdir string
}

type Shutdown struct {
	Notify SnsNotification
	Service string
}

type Restart struct {
	Notify    SnsNotification
	Service   string
	Attempts  int
	Status    string
	ErrorInfo string
}

type Notification interface {
	notify(subject string, message string)
	NotifyRestart(service string, attempts int, status int, errInfo string)
	NotifyShutdown(service string)
}

// Function executes template with data to get string
func (sn SnsNotification) getTemplateString(tmp *template.Template, data interface{}) (string, error){
	var stringBytes bytes.Buffer
	if err := tmp.Execute(&stringBytes,data) ; err != nil {
		return "", err
	}
	return stringBytes.String(), nil
}

func (sn SnsNotification) GetMailMessage(topic string, data interface{}) (string, string, error) {
	m := sn.mails[topic]
	sbj, err  := sn.getTemplateString(m.Subject,data)
	if err != nil {
		return "", "", err
	}
	mss, err  := sn.getTemplateString(m.Message,data)
	if err != nil {
		return "", "", err
	}
	return sbj, mss, nil

}

// Function creates mail notification
func (sn SnsNotification) notifyMail(mailName string, data interface{}){
	sbj, mss, err := sn.GetMailMessage(mailName,data)
	if err != nil {
		log.Fatalln(err)
		return
	}
	sn.notify(sbj, mss)
}

// Function notify about service shutdown
func (sn SnsNotification) NotifyShutdown(service string) {
	sh := Shutdown{
		Notify: sn,
		Service: service,
	}
	sn.notifyMail(shutdownMail,sh)
}

// Function notify about restart attempts and results
func (sn SnsNotification) NotifyRestart(service string, attempts int, status int, errInfo string) {
	var statusInfo string
	switch status {
	case SuccededStatus:
		errInfo = ""
		statusInfo = "SUCCESS"
	case FailedStatus:
		statusInfo = "FAILURE"
	}
	rest := Restart{
		Notify:    sn,
		Service:   service,
		Attempts:  attempts,
		Status:    statusInfo,
		ErrorInfo: errInfo,
	}
	sn.notifyMail(restartMail,rest)
}

// Function publish notification to SNS service
func (sn SnsNotification) notify(subject string, message string){
	if len(subject) >= subjectSize {
		subject = subject[:subjectSize]
	}
	messageBytes := []byte(message)
	if len(messageBytes) >= messageSize {
		message = string(messageBytes[:messageSize])
	}
	params := &sns.PublishInput{
		Subject: aws.String(subject),
		Message: aws.String(message),
		TopicArn: sn.topicARN,
	}
	_, err := sn.svc.Publish(params)
	if err != nil {
		log.Fatalln(err)
		return
	}
}

// Function creates mail by reading files with templates
func (sn *SnsNotification) ReadMailsFromFiles() error{
	for _, elem := range mailNames {
		tpc, err := parseFile(elem, messageFile, sn.workdir)
		if err != nil {
			return err
		}
		sbj, err := parseFile(elem, subjectFile, sn.workdir)
		if err != nil {
			return err
		}
		sn.mails[elem] = mail{
			Message: tpc,
			Subject: sbj,
		}
	}
	return nil
}

// Function parse file with template
func parseFile(name string, extension string, dir string) (*template.Template, error){
	var builder strings.Builder
	builder.WriteString(dir)
	builder.WriteString("/")
	builder.WriteString(name)
	builder.WriteString(".")
	builder.WriteString(extension)
	fileName := builder.String()
	tmp, err := template.ParseFiles(fileName)
	if err != nil {
		return nil, err
	}
	return tmp, nil

}

// Function creates session with AWS SNS service and creates templates necessary to send mails
func ConnectToNotificationService(sess *session.Session, snsTopicName string, id string, mailDir string) (*SnsNotification,error){
	notify := &SnsNotification{
		svc:      sns.New(sess),
		Name:     "Watchdog",
		Id:       id,
		workdir:  mailDir,
		mails:    map[string]mail{},
		topicARN: nil,
	}
	listInput := &sns.ListTopicsInput{}
	err :=  notify.svc.ListTopicsPages(listInput, func(output *sns.ListTopicsOutput, lastPage bool) bool {
		for _, elem := range output.Topics {
			if strings.Contains(*elem.TopicArn, snsTopicName) {
				notify.topicARN = elem.TopicArn
				return false
			}
		}
		if lastPage == true {
			return false
		}
		return true
	})
	if err != nil {
		return nil, err
	} else if notify.topicARN == nil {
		return nil, errors.New("couldn't find specified SNS topic with provided name")
	}
		err = notify.ReadMailsFromFiles()
	if err != nil {
		return nil, err
	}
	return notify, nil
}