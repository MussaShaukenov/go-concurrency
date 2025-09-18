package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MussaShaukenov/go-concurrency/internal/config"
	"github.com/MussaShaukenov/go-concurrency/internal/network"
	"go.uber.org/zap"
)

func main() {
	address := flag.String("address", "localhost:3223", "address to listen on")
	idleTimeout := flag.Duration("idle_timeout", time.Minute, "idle timeout")
	maxMessageSizeStr := flag.String("max_message_size", "4KB", "max message size")
	flag.Parse()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	maxMessageSize, err := config.ParseSize(*maxMessageSizeStr)
	if err != nil {
		logger.Warn("Failed to parse max message size", zap.Error(err))
	}

	reader := bufio.NewReader(os.Stdin)
	client, err := network.NewTCPClient(config.NetworkConfig{
		Address:        *address,
		IdleTimeout:    *idleTimeout,
		MaxMessageSize: maxMessageSize,
	})
	if err != nil {
		logger.Fatal("error creating client", zap.Error(err))
	}
	defer client.Close()

	fmt.Println("available commands: SET <key> <value>, GET <key>, DEL <key>")
	fmt.Println("type 'exit' to quit")

	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			logger.Error(fmt.Sprintf("read error: %v", err))
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

		response, err := client.SendCommand(text)
		if err != nil {
			logger.Error(fmt.Sprintf("send command error: %v", err))
		}

		fmt.Println(response)
	}

}
