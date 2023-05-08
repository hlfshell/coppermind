package config

import "time"

type Config struct {
	Chat    ChatConfig    `json:"chat"`
	Summary SummaryConfig `json:"summary"`
}

type ChatConfig struct {
	ConversationMaintainanceDuration time.Duration `json:"conversation_maintainance_duration"`
	MaxConversationIdleTime          time.Duration `json:"max_conversation_idle_time"`
	MaxSummariesToInclude            int           `json:"max_summaries_to_include"`
}

var DefaultChatConfig ChatConfig = ChatConfig{
	ConversationMaintainanceDuration: 5,
	MaxConversationIdleTime:          6,
	MaxSummariesToInclude:            25,
}

type SummaryConfig struct {
	SummaryDaemonInterval           time.Duration `json:"summary_daemon_interval"`
	MinMessagesToSummarize          int           `json:"min_messages_to_summarize"`
	MinConversationTimeToWait       time.Duration `json:"min_conversation_time_to_wait"`
	MinMessagesToForceSummarization int           `json:"min_messages_to_force_summarization"`
}

var DefaultSummaryConfig SummaryConfig = SummaryConfig{
	SummaryDaemonInterval:           60,
	MinMessagesToSummarize:          5,
	MinConversationTimeToWait:       5,
	MinMessagesToForceSummarization: 15,
}
