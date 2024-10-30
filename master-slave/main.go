package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

const (
	masterConnStr = "postgres://postgres:my_password@localhost:5432/my_database?sslmode=disable"
	slaveConnStr  = "postgres://postgres:my_password@localhost:5433/my_database?sslmode=disable"
)

func main() {
	// Підключення до master-ноди для запису
	masterDB, err := sql.Open("postgres", masterConnStr)
	if err != nil {
		log.Fatalf("Не вдалося підключитися до master-ноди: %v", err)
	}
	defer masterDB.Close()

	// Підключення до slave-ноди для читання
	slaveDB, err := sql.Open("postgres", slaveConnStr)
	if err != nil {
		log.Fatalf("Не вдалося підключитися до slave-ноди: %v", err)
	}
	defer slaveDB.Close()

	// Створення таблиці, якщо вона не існує
	const createTableQuery = `
    CREATE TABLE IF NOT EXISTS test_table (
        id SERIAL PRIMARY KEY,
        data TEXT
    );`
	_, err = masterDB.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Не вдалося створити таблицю на master-ноді: %v", err)
	}
	fmt.Println("Таблиця перевірена або створена успішно.")

	// Приклад запису даних у master
	_, err = masterDB.Exec("INSERT INTO test_table (data) VALUES ($1)", "Hello, replication!")
	if err != nil {
		log.Fatalf("Не вдалося вставити дані в master-ноду: %v", err)
	}
	fmt.Println("Дані успішно записано в master-ноду.")

	time.Sleep(time.Second)

	// Приклад читання даних зі slave
	var data string
	err = slaveDB.QueryRow("SELECT data FROM test_table ORDER BY id DESC LIMIT 1").Scan(&data)
	if err != nil {
		log.Fatalf("Не вдалося прочитати дані зі slave-ноди: %v", err)
	}
	fmt.Printf("Дані, прочитані зі slave-ноди: %s\n", data)

	// Приклад запису даних у slave
	_, err = slaveDB.Exec("INSERT INTO test_table (data) VALUES ($1)", "Hello, replication!")
	if err == nil {
		log.Fatalf("Помилка: вдалося вставити дані в slave-ноду: %v", err)
	}
	fmt.Printf("Не вдалося вставити дані в slave-ноду: %s\n", err)
}
