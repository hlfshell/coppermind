package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/coppermind/internal/agent"
	"github.com/hlfshell/coppermind/internal/llm/openai"
	"github.com/hlfshell/coppermind/internal/store/sqlite"
	"github.com/hlfshell/coppermind/pkg/chat"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "coppermind-tester",
		Usage: "talk with the ai",
		Action: func(cli *cli.Context) error {
			args := cli.Args()
			name := args.Get(0)
			sqliteFile := args.Get(1)
			if name == "" || sqliteFile == "" {
				fmt.Println("You must pass your name and the database file path")
				return nil
			}
			run(name, sqliteFile)
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(name string, sqliteFile string) {
	store, err := sqlite.NewSqliteStore(sqliteFile)
	if err != nil {
		fmt.Println("SQL error")
		fmt.Println(err)
		os.Exit(3)
	}
	err = store.Migrate()
	if err != nil {
		fmt.Println("Migrate error")
		fmt.Println(err)
		os.Exit(3)
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Open AI Key must be specified")
		os.Exit(3)
	}
	openai := openai.NewOpenAI(apiKey)

	agent := agent.NewAgent("Rose", store, openai)

	fmt.Println("Talk to your bot")

	for {
		message := InputPrompt(">> ")

		msg := &chat.Message{
			ID:           uuid.New().String(),
			Conversation: "",
			User:         name,
			Tone:         "",
			CreatedAt:    time.Now(),
			Content:      message,
		}

		response, err := agent.SendMessage(msg)
		if err != nil {
			fmt.Println("Chat error")
			fmt.Println(err)
			os.Exit(3)
		}
		fmt.Println(response)

	}
}

func InputPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}
