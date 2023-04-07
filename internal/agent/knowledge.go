package agent

import (
	"github.com/hlfshell/coppermind/pkg/memory"
)

func (agent *Agent) KnowledgeDaemon() error {
	conversations, err := agent.db.GetConversationsToExtractKnowledge()
	if err != nil {
		return err
	}
	err = agent.generateNewKnowledge(conversations)
	if err != nil {
		return err
	}

	err = agent.db.ExpireKnowledge()
	if err != nil {
		return err
	}

	return nil
}

func (agent *Agent) generateNewKnowledge(conversations []string) error {
	for _, conversation := range conversations {
		facts, err := agent.ExtractKnowledge(conversation)
		if err != nil {
			return err
		}
		for _, facts := range facts {
			err = agent.db.SaveKnowledge(facts)
			if err != nil {
				return err
			}
		}
		agent.db.SetConversationAsKnowledgeExtracted(conversation)
	}

	return nil
}

// func (agent *Agent) compressKnowledge(conversationIds []string) error {
// users := []string{}
// for _, id := range conversationIds {
// 	conversation, err := agent.db.GetConversation(id)
// 	if err != nil {
// 		return err
// 	}
// 	users = append(users, conversation.User)
// }

// for _, user := range users {
// 	facts, err := agent.db.GetKnowlegeByAgentAndUser(agent.Name, user)
// 	if err != nil {
// 		return err
// 	}

// 	// agent.llm.CompressFacts(agent)
// }
// 	return nil
// }

func (agent *Agent) ExtractKnowledge(conversation string) ([]*memory.Knowledge, error) {
	history, err := agent.db.GetConversation(conversation)
	if err != nil {
		return nil, err
	}

	summary, err := agent.db.GetSummaryByConversation(conversation)
	if err != nil {
		return nil, err
	}

	knowledge, err := agent.llm.Learn(
		history,
		summary,
	)

	return knowledge, err
}
