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

type Utente struct {
	ID            string `json:"ID"`
	Nome          string `json:"Nome"`
	Cognome       string `json:"Cognome"`
	Saldo         int    `json:"Saldo"`
	Destinazione  string `json:"Destinazione"`
	Num_Biglietti int    `json:"Num_Biglietti"`
}

type InputPayment struct {
	Message string `json:"Message"`
	Item    Item   `json:"item"`
	Posti   int    `json:"Posti"`
	ID      string `json:"ID"`
}

type ErrorOutput struct {
	Message string `json:"Message"`
}

type InputConfirm struct {
	Message      string `json:"Message"`
	Item         Item   `json:"Item"`
	ID           string `json:"ID"`
	Backup_saldo int    `json:"Backup_saldo"`
	Posti        int    `json:"Posti"`
}

func HandleRequest(ctx context.Context, input InputPayment) (interface{}, error) {

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

	utente := Utente{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &utente)
	if err != nil {
		log.Printf("Errore di DynamoDB: %s", err.Error())
		return ErrorOutput{Message: userErr}, nil
	}

	costo := input.Item.Prezzo * input.Posti
	if utente.Saldo < costo {
		return ErrorOutput{Message: "Saldo insufficiente"}, nil
	} else {
		backup_saldo := utente.Saldo
		resto := utente.Saldo - costo

		utente.Destinazione += input.Item.Destinazione + "\n"
		utente.Num_Biglietti += input.Posti
		utente.Saldo = resto

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

		return InputConfirm{
			Message:      "Pagamento completato.",
			Item:         input.Item,
			Backup_saldo: backup_saldo,
			ID:           input.ID,
			Posti:        input.Posti,
		}, nil
	}
}

func main() {
	lambda.Start(HandleRequest)
}
