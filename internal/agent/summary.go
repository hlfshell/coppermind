package agent

import (
	"fmt"

	"github.com/hlfshell/coppermind/internal/memory"
)

func (agent *Agent) Summarize(conversation string) (*memory.Summary, error) {
	// First we get all of the messages in that conversation
	// that we'll be trying to summarize
	msgs, err := agent.db.LoadConversationMessages(conversation)
	if err != nil {
		return nil, err
	}

	//Determine if a summary already exists for this conversation
	existingSummary, err := agent.db.GetSummaryByConversation(conversation)
	if err != nil {
		return nil, err
	}

	// Ask the llm to generate the summaries
	summary, err := agent.llm.Summarize(agent.summaryInstructions, msgs, existingSummary)

	return summary, err
}

func (agent *Agent) SummaryDaemon() error {
	conversations, err := agent.db.GetConversationsToUpdate()
	if err != nil {
		return err
	}

	for _, conversation := range conversations {
		summary, err := agent.Summarize(conversation)
		if err != nil {
			return err
		}
		// err = agent.db.SaveSummary(summary)
		// if err != nil {
		// 	return err
		// }
		fmt.Println("summary")
		fmt.Println(summary)
	}

	return nil
}
