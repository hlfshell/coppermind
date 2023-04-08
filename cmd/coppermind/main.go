package main

import (
	"log"
	"os"

	"github.com/hlfshell/coppermind/internal/quickstart"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "coppermind",
		Usage: "All in one backend for smarter AI friends",
		Commands: []*cli.Command{
			quickstartCommand,
			agentCommand,
			userCommand,
			serverCommand,
			chatCommand,
		},
		// Action: func(cli *cli.Context) error {
		// 	args := cli.Args()
		// 	sqliteFile := args.Get(0)
		// 	port := args.Get(1)
		// 	if sqliteFile == "" {
		// 		fmt.Println("You must pass the database file path")
		// 		return nil
		// 	}
		// 	if port == "" {
		// 		port = ":8080"
		// 	}

		// 	serve(sqliteFile, port)
		// 	return nil
		// },
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var quickstartCommand *cli.Command = &cli.Command{
	Name:  "quickstart",
	Usage: "Guided quickstart tool for new setups",
	Action: func(cli *cli.Context) error {
		return quickstart.Survey()
	},
}

var agentCommand *cli.Command = &cli.Command{
	Name:  "agent",
	Usage: "Agent creation and configuration tooling",
	Subcommands: []*cli.Command{
		{
			Name:  "create",
			Usage: "Create a new agent",
		},
		{
			Name:  "load",
			Usage: "Load an agent config file into the database",
		},
	},
}

var serverCommand *cli.Command = &cli.Command{
	Name:  "server",
	Usage: "Launch coppermind as a backend server",
	Subcommands: []*cli.Command{
		{
			Name:  "config",
			Usage: "Create or validate a config for server usage",
			Action: func(cli *cli.Context) error {
				return nil
			},
		},
	},
}

var userCommand *cli.Command = &cli.Command{
	Name:  "user",
	Usage: "Manage authorized users",
	Subcommands: []*cli.Command{
		{
			Name:  "create",
			Usage: "Create a new user",
			Action: func(cli *cli.Context) error {
				return nil
			},
		},
		{
			Name:  "delete",
			Usage: "Delete a user",
			Action: func(cli *cli.Context) error {
				return nil
			},
		},
		{
			Name:  "token",
			Usage: "Generate a user token",
			Action: func(cli *cli.Context) error {
				return nil
			},
		},
	},
}

var chatCommand *cli.Command = &cli.Command{
	Name:  "chat",
	Usage: "Chat with an agent in your terminal",
	Action: func(cli *cli.Context) error {
		// Do stuff
		return nil
	},
}
