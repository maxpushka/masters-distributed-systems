package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/lib/pq"
)

const iterations = 10

var (
	clickhouseQueries = []string{
		// 1. Порахувати кількість проданого товару
		"SELECT SUM(quantity) AS total_quantity_sold FROM purchases;",
		// 2. Порахувати вартість проданого товару
		"SELECT SUM(quantity * price) AS total_value_sold FROM purchases;",
		// 3. Порахувати вартість проданого товару за період
		"SELECT SUM(quantity * price) AS total_value_sold FROM purchases WHERE purchase_date BETWEEN '2024-01-01' AND '2024-01-31';",
		// 4. Порахувати скільки було придбано товару А в магазині В за період С
		"SELECT SUM(quantity) AS total_quantity_sold FROM purchases WHERE product = 'Чай' AND store = 'Магазин B' AND purchase_date BETWEEN '2024-01-01' AND '2024-01-31';",
		// 5. Порахувати скільки було придбано товару А в усіх магазинах за період С
		"SELECT SUM(quantity) AS total_quantity_sold FROM purchases WHERE product = 'Чай' AND purchase_date BETWEEN '2024-01-01' AND '2024-01-31';",
		// 6. Порахувати сумарну виручку магазинів за період С
		"SELECT store, SUM(quantity * price) AS total_revenue FROM purchases WHERE purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY store;",
		// 7. Вивести топ 10 купівель товарів по два за період С
		"SELECT least(p1.product, p2.product) AS product_1, greatest(p1.product, p2.product) AS product_2, COUNT(*) AS count FROM purchases p1 JOIN purchases p2 ON p1.receipt_id = p2.receipt_id WHERE p1.product != p2.product AND p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY product_1, product_2 ORDER BY count DESC LIMIT 10;",
		// 8. Вивести топ 10 купівель товарів по три за період С
		"SELECT least(p1.product, p2.product, p3.product) AS product_1, arraySort([p1.product, p2.product, p3.product])[2] AS product_2, greatest(p1.product, p2.product, p3.product) AS product_3, COUNT(*) AS count FROM purchases p1 JOIN purchases p2 ON p1.receipt_id = p2.receipt_id JOIN purchases p3 ON p1.receipt_id = p3.receipt_id WHERE p1.product < p2.product AND p2.product < p3.product AND p1.product != p3.product AND p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY product_1, product_2, product_3 ORDER BY count DESC LIMIT 10;",
		// 9. Вивести топ 10 купівель товарів по чотири за період С
		"SELECT least(p1.product, p2.product, p3.product, p4.product) AS product_1, arraySort([p1.product, p2.product, p3.product, p4.product])[2] AS product_2, arraySort([p1.product, p2.product, p3.product, p4.product])[3] AS product_3, greatest(p1.product, p2.product, p3.product, p4.product) AS product_4, COUNT(*) AS count FROM purchases p1 JOIN purchases p2 ON p1.receipt_id = p2.receipt_id JOIN purchases p3 ON p1.receipt_id = p3.receipt_id JOIN purchases p4 ON p1.receipt_id = p4.receipt_id WHERE p1.product < p2.product AND p2.product < p3.product AND p3.product < p4.product AND p1.product != p4.product AND p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY product_1, product_2, product_3, product_4 ORDER BY count DESC LIMIT 10;",
	}
	postgresQueries = []string{
		// 1. Порахувати кількість проданого товару
		"SELECT SUM(quantity) AS total_quantity_sold FROM purchases;",
		// 2. Порахувати вартість проданого товару
		"SELECT SUM(quantity * price) AS total_value_sold FROM purchases;",
		// 3. Порахувати вартість проданого товару за період
		"SELECT SUM(quantity * price) AS total_value_sold FROM purchases WHERE purchase_date BETWEEN '2024-01-01' AND '2024-01-31';",
		// 4. Порахувати скільки було придбано товару А в магазині В за період С
		"SELECT SUM(quantity) AS total_quantity_sold FROM purchases WHERE product = 'Чай' AND store = 'Магазин B' AND purchase_date BETWEEN '2024-01-01' AND '2024-01-31';",
		// 5. Порахувати скільки було придбано товару А в усіх магазинах за період С
		"SELECT SUM(quantity) AS total_quantity_sold FROM purchases WHERE product = 'Чай' AND purchase_date BETWEEN '2024-01-01' AND '2024-01-31';",
		// 6. Порахувати сумарну виручку магазинів за період С
		"SELECT store, SUM(quantity * price) AS total_revenue FROM purchases WHERE purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY store;",
		// 7. Вивести топ 10 купівель товарів по два за період С
		"SELECT LEAST(p1.product, p2.product) AS product_1, GREATEST(p1.product, p2.product) AS product_2, COUNT(*) AS count FROM purchases p1 JOIN purchases p2 ON p1.receipt_id = p2.receipt_id WHERE p1.product != p2.product AND p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY product_1, product_2 ORDER BY count DESC LIMIT 10;",
		// 8. Вивести топ 10 купівель товарів по три за період С
		"SELECT LEAST(p1.product, p2.product, p3.product) AS product_1, (SELECT product FROM (VALUES (p1.product), (p2.product), (p3.product)) AS t(product) ORDER BY product LIMIT 1 OFFSET 1) AS product_2, GREATEST(p1.product, p2.product, p3.product) AS product_3, COUNT(*) AS count FROM purchases p1 JOIN purchases p2 ON p1.receipt_id = p2.receipt_id AND p1.product < p2.product JOIN purchases p3 ON p1.receipt_id = p3.receipt_id AND p2.product < p3.product WHERE p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY product_1, product_2, product_3 ORDER BY count DESC LIMIT 10;",
		// 9. Вивести топ 10 купівель товарів по чотири за період С
		"SELECT LEAST(p1.product, p2.product, p3.product, p4.product) AS product_1, (SELECT product FROM (VALUES (p1.product), (p2.product), (p3.product), (p4.product)) AS t(product) ORDER BY product LIMIT 1 OFFSET 1) AS product_2, (SELECT product FROM (VALUES (p1.product), (p2.product), (p3.product), (p4.product)) AS t(product) ORDER BY product LIMIT 1 OFFSET 2) AS product_3, GREATEST(p1.product, p2.product, p3.product, p4.product) AS product_4, COUNT(*) AS count FROM purchases p1 JOIN purchases p2 ON p1.receipt_id = p2.receipt_id AND p1.product < p2.product JOIN purchases p3 ON p1.receipt_id = p3.receipt_id AND p2.product < p3.product JOIN purchases p4 ON p1.receipt_id = p4.receipt_id AND p3.product < p4.product WHERE p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY product_1, product_2, product_3, product_4 ORDER BY count DESC LIMIT 10;",
	}
)

// Executes a query on the given database and returns the duration of the query in seconds
func runQuery(db *sql.DB, query string) (float64, error) {
	start := time.Now()

	if _, err := db.Exec(query); err != nil {
		return 0, fmt.Errorf("failed to execute PostgreSQL query: %v", err)
	}

	elapsed := time.Since(start).Seconds()
	return elapsed, nil
}

// Computes the average time for a given query
func averageQueryTime(db *sql.DB, query string) (float64, error) {
	var total float64
	for i := 0; i < iterations; i++ {
		timeTaken, err := runQuery(db, query)
		if err != nil {
			return 0, err
		}
		total += timeTaken
	}
	return total / float64(iterations), nil
}

func main() {
	dbType := flag.String("db", "", "Specify which database to run queries against: clickhouse or postgres")
	flag.Parse()

	var db *sql.DB
	var queries []string
	var err error
	if *dbType == "clickhouse" {
		db, err = sql.Open("clickhouse", "clickhouse://default:@localhost:9000/default")
		queries = clickhouseQueries
	} else if *dbType == "postgres" {
		db, err = sql.Open("postgres", "user=user password=password dbname=postgres sslmode=disable")
		queries = postgresQueries
	} else {
		log.Fatal("Invalid database type specified. Use either 'clickhouse' or 'postgres'")
	}
	if err != nil {
		log.Fatalf("failed to connect to %s: %v", *dbType, err)
	}
	defer db.Close()

	for i, query := range queries {
		averageTime, err := averageQueryTime(db, query)
		if err != nil {
			fmt.Printf("Error running %s query %d: %v\n", *dbType, i+1, err)
			continue
		}
		fmt.Printf("%s query %d average time: %.5f seconds\n", *dbType, i+1, averageTime)
	}
}
