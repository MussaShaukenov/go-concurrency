package main

import (
	"bufio"
	"fmt"
	"github.com/MussaShaukenov/go-concurrency/internal/compute/parser"
	"github.com/MussaShaukenov/go-concurrency/internal/storage/engine"
	"go.uber.org/zap"
	"os"
	"strings"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	log := logger.Sugar()

	reader := bufio.NewReader(os.Stdin)
	storageEngine := engine.NewEngine(log)
	queryParser := parser.NewParser(log, storageEngine)

	fmt.Println("available commands: SET <key> <value>, GET <key>, DEL <key>")
	fmt.Println("type 'exit' to quit")

	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Errorf("read error: %v", err)
			continue
		}

		text = strings.TrimSpace(text)

		// handle exit
		if strings.ToLower(text) == "exit" {
			fmt.Println("program ended")
			break
		}

		// skip empty input
		if text == "" {
			continue
		}

		result, err := queryParser.ParseQuery(text)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}

		if result != nil {
			fmt.Printf("result: %v\n", result)
		} else {
			fmt.Println("OK")
		}
	}
}
