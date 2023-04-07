package agent

import (
	"fmt"

	"github.com/hlfshell/coppermind/pkg/memory"
)

func (agent *Agent) Summarize(conversationId string) (*memory.Summary, error) {
	// First we get all of the messages in that conversation
	// that we'll be trying to summarize
	conversation, err := agent.db.GetConversation(conversationId)
	if err != nil {
		return nil, err
	}

	//Determine if a summary already exists for this conversation
	existingSummary, err := agent.db.GetSummaryByConversation(conversationId)
	if err != nil {
		return nil, err
	}

	// Ask the llm to generate the summaries
	summary, err := agent.llm.Summarize(conversation, existingSummary)
	if err != nil {
		return nil, err
	} else if summary == nil {
		agent.db.ExcludeConversationFromSummary(conversationId)
		return nil, nil
	}

	err = agent.db.SaveSummary(summary)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func (agent *Agent) SummaryDaemon() error {
	conversations, err := agent.db.GetConversationsToSummarize(
		agent.summaryMinMessages,
		agent.summaryMinConversationTime,
		agent.summaryMinMessagesToForceSummarization,
	)
	fmt.Println("convos", conversations)
	if err != nil {
		return err
	}

	for _, conversation := range conversations {
		summary, err := agent.Summarize(conversation)
		if err != nil {
			return err
		}

		fmt.Println("summary")
		fmt.Println(summary)
	}

	return nil
}
