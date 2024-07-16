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

type InputReservation struct {
	Message      string `json:"message"`
	ID           string `json:"ID"`
	Destinazione string `json:"Destinazione"`
	Posti        int    `json:"Posti"`
}

type Item struct {
	Destinazione string `json:"Destinazione"`
	Posti        int    `json:"Posti"`
	Prezzo       int    `json:"Prezzo"`
}

type ErrorOutput struct {
	Message string `json:"Message"`
}

type InputPayment struct {
	Message string `json:"Message"`
	Item    Item   `json:"item"`
	Posti   int    `json:"Posti"`
	ID      string `json:"ID"`
}

func HandleRequest(ctx context.Context, input InputReservation) (interface{}, error) {

	userErr := "Ops! Qualcosa non ha funzionato"
	inputErr := "Input utente errato"

	log.Printf("Ricevuto input: %+v", input)

	if input.Posti <= 0 {
		log.Println(inputErr)
		errMsg := "Bisogna selezionare almento un posto"
		return ErrorOutput{Message: errMsg}, nil
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))

	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Tabella_Voli"),
		Key: map[string]*dynamodb.AttributeValue{
			"Destinazione": {
				S: aws.String(input.Destinazione),
			},
		},
	})

	if err != nil {
		errMsg := fmt.Sprintf("Errore di DynamoDB: %s", err.Error())
		log.Println(errMsg)
		return ErrorOutput{userErr}, nil
	}

	if result.Item == nil {
		errMsg := fmt.Sprintf("Non voliamo ancora verso %s", input.Destinazione)
		log.Println(inputErr)
		return ErrorOutput{Message: errMsg}, nil
	}

	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	if err != nil {
		errMsg := fmt.Sprintf("Errore di DynamoDB: %s", err.Error())
		log.Println(errMsg)
		return ErrorOutput{Message: userErr}, nil
	}

	log.Printf("Elemento recuperato da DynamoDB: %+v", item)

	if item.Posti-input.Posti < 0 {
		errMsg := fmt.Sprintf("Non ci sono abbastanza posti disponibili. Posti disponibili %d", item.Posti)
		log.Println(errMsg)
		return ErrorOutput{Message: errMsg}, nil
	}

	return InputPayment{
		Message: "ok",
		Item:    item,
		ID:      input.ID,
		Posti:   input.Posti,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
