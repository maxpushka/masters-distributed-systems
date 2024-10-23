package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/streadway/amqp"
)

type Message struct {
	Num1 int `json:"num1"`
	Num2 int `json:"num2"`
}

type Result struct {
	Result int `json:"result"`
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	// Підписуємось на вхідну чергу
	msgs, err := ch.Consume(
		"input_queue", // queue
		"",            // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range msgs {
			var msg Message
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Printf("Error decoding JSON: %s", err)
				continue
			}

			// Робимо обчислення (наприклад, додаємо числа)
			result := msg.Num1 + msg.Num2

			// Готуємо результат для відправки
			res := Result{Result: result}
			resBody, _ := json.Marshal(res)

			// Кладемо результат в вихідну чергу
			err = ch.Publish(
				"",             // exchange
				"output_queue", // routing key
				false,          // mandatory
				false,          // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        resBody,
				})
			if err != nil {
				log.Printf("Failed to publish a message: %s", err)
			}
		}
	}()

	fmt.Println("Waiting for messages...")
	wg.Wait()
}
