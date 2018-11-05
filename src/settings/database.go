// Package provide functions that connect to settings and manage communication with it.
package settings

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"log"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"errors"
)

type SettingsDatabase struct {
	db *dynamodb.DynamoDB
	tableName string
	primaryKey string
}

// Function connects to database
func ConnectToDatabase(sess *session.Session, table string, pk string) *SettingsDatabase {
	log.Println("Connecting to database")
	return &SettingsDatabase{
		db: dynamodb.New(sess),
		tableName: table,
		primaryKey: pk,
	}
}

func (sd SettingsDatabase) GetSettings() (*Settings, error){
	return sd.GetSettingsByIndex(sd.primaryKey)
}

// Function gets settings with id
func (sd SettingsDatabase) GetSettingsByIndex(index string) (*Settings, error){
	result, err := sd.db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(sd.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(index),
			}},
	})
	if err != nil {
		return nil, err
	}
	sett := &Settings{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &sett)
	if err != nil {
		return nil,err
	}
	if sett.Id == "" {
		return nil, errors.New("Couldn't find settings with provided primary key")
	}
	if !sett.validate(){
		return nil, errors.New("Acquired settings are not correct")
	}
	sett.Unmarshal()
	return sett, nil
}