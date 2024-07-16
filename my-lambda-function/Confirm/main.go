package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Item struct {
	Destinazione string `json:"Destinazione"`
	Posti        int    `json:"Posti"`
	Prezzo       int    `json:"Prezzo"`
}

type InputConfirm struct {
	Message      string `json:"Message"`
	Item         Item   `json:"Item"`
	ID           string `json:"ID"`
	Backup_saldo int    `json:"Backup_saldo"`
	Posti        int    `json:"Posti"`
}

type ErrorInputCompPayment struct {
	Message      string `json:"Message"`
	Destinazione string `json:"Destinazione"`
	ID           string `json:"ID"`
	Backup_saldo int    `json:"Backup_saldo"`
	Posti        int    `json:"Posti"`
}

type SuccessOutput struct {
	Message string `json:"Message"`
}

func HandleRequest(ctx context.Context, input InputConfirm) (interface{}, error) {

	destinazione := input.Item.Destinazione

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Tabella_Voli"),
		Key: map[string]*dynamodb.AttributeValue{
			"Destinazione": {
				S: aws.String(input.Item.Destinazione),
			},
		},
	})

	if err != nil || result.Item == nil {
		log.Printf("Errore di DynamoDB: %s", err.Error())
		return ErrorInputCompPayment{
			Message:      "Siamo spiacenti ma qualcosa è andato storto",
			Destinazione: destinazione,
			ID:           input.ID,
			Backup_saldo: input.Backup_saldo,
			Posti:        input.Posti,
		}, nil
	}

	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	if err != nil {
		log.Printf("Errore di DynamoDB: %s", err.Error())
		return ErrorInputCompPayment{
			Message:      "Ci dispiace me qualcosa è andato storto",
			Destinazione: destinazione,
			ID:           input.ID,
			Backup_saldo: input.Backup_saldo,
			Posti:        input.Posti,
		}, nil
	}
	if item.Posti-input.Posti < 0 {
		log.Printf("Numero di posti non più sufficiente")
		return ErrorInputCompPayment{
			Message:      "Ops! Qualcosa è andato storto",
			Destinazione: destinazione,
			ID:           input.ID,
			Backup_saldo: input.Backup_saldo,
			Posti:        input.Posti,
		}, nil
	}

	item.Posti -= input.Posti
	av, err := dynamodbattribute.MarshalMap(item)

	if err != nil {
		log.Printf("Errore di DynamoDB: %s", err.Error())
		return ErrorInputCompPayment{
			Message:      "Qualcosa è andato storto, riprova più tardi",
			Destinazione: destinazione,
			ID:           input.ID,
			Backup_saldo: input.Backup_saldo,
			Posti:        input.Posti,
		}, nil
	}
	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("Tabella_Voli"),
		Item:      av,
	})

	if err != nil {
		log.Printf("Errore di DynamoDB: %s", err.Error())
		return ErrorInputCompPayment{
			Message:      "Ops! Qualcosa non è andato per il verso giusto.",
			Destinazione: destinazione,
			ID:           input.ID,
			Backup_saldo: input.Backup_saldo,
			Posti:        input.Posti,
		}, nil
	}

	log.Printf("Successo")

	return SuccessOutput{
		Message: "Prenotazione confermata. Buon Viaggio!",
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
