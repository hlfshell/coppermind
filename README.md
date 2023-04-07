# coppermind 🤖💬🧠

This is a FIRST DRAFT project and is a WIP.

Coppermind is an additional "brain" that sits on top of an instruction based LLM model, programmed in golang.

It uses LLM services (currently just OpenAI) to try to create a conversational agent. Since LLMs only remember a given body of text, this handles several situations to "improve memory" of your AI agent. Specifically:

1. Remember conversation history to allow multiple exchanges with conversational memory
2. Allows specific control of the agent by specifying personality and how it should respond/act
3. Automatically handles conversation summarization so the agent has long term conversational memory and can recall what you talked about outside the typical token limit.
4. Extracts facts about people and objects that you talk about in order to remember them later efficiently.
