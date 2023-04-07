package agent

import (
	"fmt"
	"os"
	"sync"
)

func (agent *Agent) RunDaemons() {

	var wait sync.WaitGroup

	wait.Add(1)
	go func() {
		defer wait.Done()
		fmt.Println("Summary Daemon triggered")
		err := agent.SummaryDaemon()
		if err != nil {
			fmt.Println("Summary error")
			fmt.Println(err)
			os.Exit(3)
		}
	}()

	wait.Add(1)
	func() {
		defer wait.Done()
		fmt.Println("Knowledge Daemon triggered")
		err := agent.KnowledgeDaemon()
		if err != nil {
			fmt.Println("Knowledge error")
			fmt.Println(err)
			os.Exit(3)
		}
	}()

	wait.Wait()
}
