package quickstart

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
)

/*
Survey presents a series of user prompts to build and setup
the appropriate resources required to run coppermind.

The survey is expected to do the following:
 1. Create a configuration file in the specified spot; when
    we are finished here it will be what coppermind uses to
    run
 2. Ask which of the database options the user wants to setup.
 3. Given the choice, set the database up. For SQLite, we ask
    where its database file. For PostgreSQL and MySQL, we ask
    for its host, port, username, and password. After setting
    up the database, we attempt a connection. If it works, we
    then ask if they'd like us to migrate the database
    immediately.
 4. We then ask about the services coppermind should run. The
    user can choose run any combination of the http, socket,
    or sms services. Optionally none of these can be chosen
    if they are just going to run coppermind as a standalone
    terminal chat. For each chosen service to run, additional
    questions for each service will be asked.
 5. If the use chose for any self hosted services, the option
    to enable authentication is asked. If they choose to have
    coppermind utilize authentication, then we ask them if
    they'd like to create a user. Either way we tell them
    how to create a user with the cli tool.
 6. We then ask whether or not we should preload the default
    AI agents into their new database.
 7. Finally we save all of this to a JSON configuration file.

On the way out, we remind them to create an agent with the
coppermind agent command.
*/
func Survey() error {
	fmt.Println("This tool will walk you through setting up Coppermind for the first time")
	fmt.Println("We will setup a configuration file and database for you to run your server or application with.")

	config, err := config()
	if err != nil {
		return err
	}
	fmt.Println(config)

	_, err = database()
	if err != nil {
		return err
	}

	return err
}

func config() (*os.File, error) {
	var configFile string

	var configInput *survey.Input = &survey.Input{
		Message: `Where do you want to store your created configuraton? (config.json)`,
	}
	survey.AskOne(configInput, &configFile)
	if configFile == "" {
		configFile = "config.json"
	}
	fmt.Println("chosen", configFile)

	return os.OpenFile(configFile, os.O_RDWR|os.O_CREATE, 0666)
}

func database() (*sql.DB, error) {
	var database string
	choices := &survey.Select{
		Message: "Choose a database:",
		Options: []string{"SQLite", "PostgreSQL", "MySQL"},
	}
	survey.AskOne(choices, &database)

	switch database {
	case "SQLite":
		return sqlite()
	case "PostgreSQL":
		return postgres()
	case "MySQL":
		return mysql()
	default:
		return nil, fmt.Errorf("invalid database type")
	}
}

func sqlite() (*sql.DB, error) {
	return nil, nil
}

func mysql() (*sql.DB, error) {
	return nil, nil
}

func postgres() (*sql.DB, error) {
	return nil, nil
}
