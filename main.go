package main

import (
	"flag"
	"fmt"
	"log"
	"whatAmIBuying/internal/services"

	_ "modernc.org/sqlite"
)

func main() {
	assignFlag := flag.Bool("a", false, "assign mode (shorthand)")
	assignFlagLong := flag.Bool("assign", false, "assign mode")
	readFlag := flag.Bool("r", false, "read mode (shorthand)")
	readFlagLong := flag.Bool("read", false, "read mode")
	predictFlag := flag.Bool("p", false, "predict mode (shorthand)")
	predictFlagLong := flag.Bool("predict", false, "predict mode")
	llmFlag := flag.Bool("l", false, "LLM mode (shorthand)")
	llmFlagLong := flag.Bool("llm", false, "LLM mode")

	flag.Parse()

	if *assignFlag || *assignFlagLong {
		fmt.Println("Assign mode activated")
		err := services.AssignPurchases()
		if err != nil {
			log.Fatal("Error assigning purchases: ", err)
		}
	} else if *readFlag || *readFlagLong {
		fmt.Println("Read mode activated")
		services.ReadReceipts()
	} else if *predictFlag || *predictFlagLong {
		fmt.Println("Predict mode activated")
		services.PredictPurchases()
	} else if *llmFlag || *llmFlagLong {
		fmt.Println("LLM mode activated")
		services.TestLLM()
	}
}
