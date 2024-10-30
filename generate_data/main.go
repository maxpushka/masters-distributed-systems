package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

// Define product and store lists
var products = []string{"Молоко", "Хліб", "Масло", "Сир", "Цукор", "Кава", "Чай"}
var stores = []string{"Магазин A", "Магазин B", "Магазин C", "Магазин D"}

// Generate random purchase data for a receipt
func generatePurchaseData(receiptID int, results chan<- []string) {
	rand.Seed(time.Now().UnixNano() + int64(receiptID)) // Seed the random number generator based on worker ID

	// Randomly select the number of items in the receipt
	numItems := rand.Intn(5) + 1 // Each receipt will contain 1 to 5 goods
	store := stores[rand.Intn(len(stores))]

	for i := 0; i < numItems; i++ {
		product := products[rand.Intn(len(products))]
		quantity := rand.Intn(10) + 1
		price := fmt.Sprintf("%.2f", rand.Float64()*(100.0-5.0)+5.0)

		// Ensure valid and properly formatted date
		randomDaysAgo := rand.Intn(365)
		purchaseDate := time.Now().AddDate(0, 0, -randomDaysAgo).Format("2006-01-02 15:04:05")

		results <- []string{strconv.Itoa(receiptID), store, product, strconv.Itoa(quantity), price, purchaseDate}
	}
}

// Worker function to process data concurrently
func worker(id int, wg *sync.WaitGroup, jobs <-chan int, results chan<- []string) {
	defer wg.Done()
	for receiptID := range jobs {
		generatePurchaseData(receiptID, results)
	}
}

func main() {
	start := time.Now()

	// Parse command-line arguments
	numRecords := flag.Int("numRecords", 1000000, "Number of records to generate")
	numWorkers := flag.Int("numWorkers", 10, "Number of workers to use")
	flag.Parse()
	fmt.Printf("Generating %d records with %d workers\n", *numRecords, *numWorkers)

	// Create CSV file
	file, err := os.Create("purchase_data.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"receipt_id", "store", "product", "quantity", "price", "purchase_date"})

	// Channels and WaitGroup for concurrency
	jobs := make(chan int, *numRecords)
	results := make(chan []string, *numRecords)
	var wg sync.WaitGroup

	// Start workers
	for i := 1; i <= *numWorkers; i++ {
		wg.Add(1)
		go worker(i, &wg, jobs, results)
	}

	// Feed jobs into the jobs channel
	go func() {
		for i := 0; i < *numRecords; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Close the results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Write results to CSV as they come in
	for record := range results {
		writer.Write(record)
	}

	elapsed := time.Since(start)
	fmt.Printf("Data generation completed in %s\n", elapsed)
}
