package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type InputCompPayment struct {
	Message      string `json:"Message"`
	Destinazione string `json:"Destinazione"`
	ID           string `json:"ID"`
	Backup_saldo int    `json:"Backup_saldo"`
	Posti        int    `json:"Posti"`
}

type Utente struct {
	ID            string `json:"ID"`
	Nome          string `json:"Nome"`
	Cognome       string `json:"Cognome"`
	Saldo         int    `json:"Saldo"`
	Destinazione  string `json:"Destinazione"`
	Num_Biglietti int    `json:"Num_Biglietti"`
}

type SuccessOutput struct {
	Message string `json:"Message"`
}

func HandleRequest(ctx context.Context, input InputCompPayment) (interface{}, error) {
	counter := 0
	for counter < 50 {
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
			counter++
			continue
		}

		utente := Utente{}
		err = dynamodbattribute.UnmarshalMap(result.Item, &utente)
		if err != nil {
			log.Printf("Errore di DynamoDB: %s", err.Error())
			counter++
			continue
		}

		destinazioni := strings.Split(utente.Destinazione, "\n")
		for i := len(destinazioni) - 1; i >= 0; i-- {
			if strings.TrimSpace(destinazioni[i]) == input.Destinazione {
				destinazioni = append(destinazioni[:i], destinazioni[i+1:]...)
				break
			}
		}

		utente.Destinazione = strings.Join(destinazioni, "\n")
		utente.Num_Biglietti -= input.Posti
		utente.Saldo = input.Backup_saldo
		av, err := dynamodbattribute.MarshalMap(utente)

		if err != nil {
			log.Printf("Errore di DynamoDB: %s", err.Error())
			counter++
			continue
		}

		inputDynamo := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String("Utenti"),
		}

		_, err = svc.PutItem(inputDynamo)
		if err != nil {
			log.Printf("Errore di DynamoDB: %s", err.Error())
			counter++
			continue
		}

		log.Printf("Utente aggiornato con successo: %+v", utente)
		counter = 100
		return SuccessOutput{
			Message: fmt.Sprintf("Prenotazione cancellata; ci scusiamo per il disagio. Saldo attuale %d", input.Backup_saldo),
		}, nil
	}
	return nil, fmt.Errorf("impossibile elaborare la richiesta")
}

func main() {
	lambda.Start(HandleRequest)
}
