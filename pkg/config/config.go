package config

type Config struct {
	Chat    ChatConfig    `json:"chat"`
	Summary SummaryConfig `json:"summary"`
}

var DefaultConfig Config = Config{
	Chat:    DefaultChatConfig,
	Summary: DefaultSummaryConfig,
}

type ChatConfig struct {
	ConversationMaintainanceDurationSeconds int `json:"conversation_maintainance_duration_seconds"`
	MaxConversationIdleTimeSeconds          int `json:"max_conversation_idle_time_seconds"`
	MaxSummariesToInclude                   int `json:"max_summaries_to_include"`
}

var DefaultChatConfig ChatConfig = ChatConfig{
	ConversationMaintainanceDurationSeconds: 5,
	MaxConversationIdleTimeSeconds:          6,
	MaxSummariesToInclude:                   25,
}

type SummaryConfig struct {
	SummaryDaemonIntervalSeconds     int `json:"summary_daemon_interval_seconds"`
	MinMessagesToSummarize           int `json:"min_messages_to_summarize"`
	MinConversationTimeToWaitSeconds int `json:"min_conversation_time_to_wait_seconds"`
	MinMessagesToForceSummarization  int `json:"min_messages_to_force_summarization"`
}

var DefaultSummaryConfig SummaryConfig = SummaryConfig{
	SummaryDaemonIntervalSeconds:     60,
	MinMessagesToSummarize:           5,
	MinConversationTimeToWaitSeconds: 5,
	MinMessagesToForceSummarization:  15,
}
