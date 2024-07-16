Step 1:
La cartella my-lambda-function contiene sei cartelle ognuna dedicata ad una funzione. 
In queste cartelle sono presente sia i codici sorgente delle singole funzioni che le loro versioni compresse.
Creare su AWS Lambda sei funzioni e caricare in ognuna di esse il file .zip della relativa funzione.
(Importante non modificare i nomi delle funzioni).

Step 2:
Bisogna creare le due tabelle "Utenti" e "Tabella_Voli" in DynamoDB. Per fare ciò bisogna:
* Creare in S3 due bucket che conterranno rispettivamente Tabella_Voli.csv e Utenti.csv
* Andare in DynamoDB e creare le due tabelle importandole da S3.
(Importante non modificare i nomi delle tabelle).

Step 3:
Creare una macchina a stati in Step Functions copiando il contenuto di StateMachine_GestioneAerei.asl.json
(Importante modificare gli arn delle funzioni con quelli corretti).

Step 4: 
Procedere alla fase di test.
