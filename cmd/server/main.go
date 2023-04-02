package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hlfshell/coppermind/internal/agent"
	"github.com/hlfshell/coppermind/internal/llm"
	"github.com/hlfshell/coppermind/internal/protocol/http"
	"github.com/hlfshell/coppermind/internal/service"
	"github.com/hlfshell/coppermind/internal/store/sqlite"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "coppermind-http-server",
		Usage: "Simple HTTP endpoint",
		Action: func(cli *cli.Context) error {
			args := cli.Args()
			sqliteFile := args.Get(0)
			port := args.Get(1)
			if sqliteFile == "" {
				fmt.Println("You must pass the database file path")
				return nil
			}
			if port == "" {
				port = ":8080"
			}

			serve(sqliteFile, port)
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func serve(sqliteFile string, port string) {
	apiKey := os.Getenv("OPENAI_KEY")
	if apiKey == "" {
		fmt.Println("Open AI Key must be specified")
		os.Exit(3)
	}

	client := llm.NewOpenAI(apiKey)

	db, err := sqlite.NewSqliteStore(sqliteFile)
	if err != nil {
		fmt.Println("SQL error")
		fmt.Println(err)
		os.Exit(3)
	}

	agent := agent.NewAgent("Rose", db, client)

	service := service.NewService(agent)

	server := http.NewHttpAPI(service, port)
	err = server.Serve()
	if err != nil {
		fmt.Println("Server error")
		fmt.Println(err)
		os.Exit(3)
	}
}
