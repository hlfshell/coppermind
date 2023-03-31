package agent

import (
	"fmt"

	"github.com/hlfshell/coppermind/internal/memory"
)

func (agent *Agent) KnowledgeDaemon() error {
	conversations, err := agent.identifyConversationsToExtractKnowledge()
	fmt.Println("conversatons", conversations)
	if err != nil {
		return err
	}

	for _, conversation := range conversations {
		knowledge, err := agent.ExtractKnowledge(conversation)
		if err != nil {
			return err
		}
		fmt.Println("post extraction")
		fmt.Println(knowledge)
		//TODO - save to db
		// return err
	}
	return nil
}

func (agent *Agent) ExtractKnowledge(conversaton string) ([]*memory.Knowledge, error) {
	history, err := agent.db.GetConversation(conversaton)
	if err != nil {
		return nil, err
	}

	knowledge, err := agent.llm.Learn(agent.knowledgeInstructions, history)

	return knowledge, err
}

func (agent *Agent) identifyConversationsToExtractKnowledge() ([]string, error) {
	return []string{"61f5ea47-7e1f-4326-8b13-042b32b83c0a"}, nil
}
