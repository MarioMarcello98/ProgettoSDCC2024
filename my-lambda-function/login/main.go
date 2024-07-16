package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Utente struct {
	ID            string `json:"ID"`
	Nome          string `json:"Nome"`
	Cognome       string `json:"Cognome"`
	Saldo         int    `json:"Saldo"`
	Destinazione  string `json:"Destinazione"`
	Num_Biglietti int    `json:"Num_Biglietti"`
}

type Input struct {
	ID           string `json:"ID"`
	Nome         string `json:"Nome"`
	Cognome      string `json:"Cognome"`
	Ricarica     int    `json:"Ricarica"`
	Destinazione string `json:"Destinazione"`
	Posti        int    `json:"Posti"`
}

type ErrorOutput struct {
	Message string `json:"message"`
}

type InputReservation struct {
	Message      string `json:"message"`
	ID           string `json:"ID"`
	Destinazione string `json:"Destinazione"`
	Posti        int    `json:"Posti"`
}

func HandleRequest(ctx context.Context, input Input) (interface{}, error) {

	log.Printf("Ricevuto input: %+v", input)

	userErr := "Ops! Qualcosa non ha funzionato"

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Utenti"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(input.ID),
			},
		},
	})

	if err != nil {
		log.Printf("Errore di DynamoDB: %s", err.Error())
		return ErrorOutput{Message: userErr}, nil
	}

	if result.Item == nil {
		utente := Utente{
			ID:            input.ID,
			Nome:          input.Nome,
			Cognome:       input.Cognome,
			Saldo:         input.Ricarica,
			Destinazione:  "",
			Num_Biglietti: 0,
		}

		av, err := dynamodbattribute.MarshalMap(utente)
		if err != nil {
			log.Printf("Errore di DynamoDB: %s", err.Error())
			return ErrorOutput{Message: userErr}, nil
		}

		inputDynamo := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String("Utenti"),
		}

		_, err = svc.PutItem(inputDynamo)
		if err != nil {
			log.Printf("Errore di DynamoDB: %s", err.Error())
			return ErrorOutput{Message: userErr}, nil
		}

		log.Printf("Utente aggiunto con successo: %+v", utente)

		return InputReservation{
			Message:      fmt.Sprintf("Benvenuto %s %s. Saldo attuale: %d", utente.Nome, utente.Cognome, utente.Saldo),
			ID:           utente.ID,
			Destinazione: input.Destinazione,
			Posti:        input.Posti,
		}, nil

	} else {
		utente := Utente{}
		err = dynamodbattribute.UnmarshalMap(result.Item, &utente)
		if err != nil {
			log.Printf("Errore di DynamoDB: %s", err.Error())
			return ErrorOutput{Message: userErr}, nil
		}

		utente.Saldo += input.Ricarica

		av, err := dynamodbattribute.MarshalMap(utente)
		if err != nil {
			log.Printf("Errore di DynamoDB: %s", err.Error())
			return ErrorOutput{Message: userErr}, nil
		}

		inputDynamo := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String("Utenti"),
		}

		_, err = svc.PutItem(inputDynamo)
		if err != nil {
			log.Printf("Errore di DynamoDB: %s", err.Error())
			return ErrorOutput{Message: userErr}, nil
		}

		log.Printf("Utente aggiornato con successo: %+v", utente)

		return InputReservation{
			Message:      fmt.Sprintf("Bentornato %s %s. Saldo attuale: %d", utente.Nome, utente.Cognome, utente.Saldo),
			ID:           utente.ID,
			Destinazione: input.Destinazione,
			Posti:        input.Posti,
		}, nil
	}
}

func main() {
	lambda.Start(HandleRequest)
}
