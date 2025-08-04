package config

import "errors"

var (
	// ErrUnsupportedAIProvider is returned when an unsupported AI provider is specified
	ErrUnsupportedAIProvider = errors.New("unsupported AI provider")
)

const (
	// DefaultResourceTagKey and DefaultResourceTagValue are used for tagging AWS resources
	DefaultResourceTagKey = "project"

	// DefaultResourceTagValue is the default value for the resource tag
	DefaultResourceTagValue = "CodeRefactoring"

	// DefaultRepositoryTagKey get repository tag key per code base.
	DefaultRepositoryTagKey = "repository"

	// CodeRefactoringDatabaseName is the name of the database for this project
	CodeRefactoringDatabaseName = "code_refactoring_db"

	// AWSBedrockAgentModel model used for Bedrock Agent.
	AWSBedrockAgentModel = "amazon.titan-tg1-large"

	// AWSBedrockRAGEmbeddingModel model used for Bedrock Knowledge Base embedding.
	AWSBedrockRAGEmbeddingModel = "amazon.titan-embed-text-v1"

	// AWSBedrockDataStoreEnrichmentModelARN is the ARN of the model used for context enrichment in the RAG pipeline.
	AWSBedrockDataStoreEnrichmentModelARN = "amazon.titan-text-express-v1:0"

	// AWSBedrockDataStoreParsingModelARN is the ARN of the model used for parsing in the RAG pipeline.
	AWSBedrockDataStoreParsingModelARN = "amazon.titan-text-express-v1:0"

	// AWSRegion used for aws.
	AWSRegion = "us-east-1"

	// AIProviderBedrock uses AWS Bedrock (default for SaaS)
	AIProviderBedrock = "bedrock"

	// AIProviderLocal uses local Ollama + ChromaDB (for development/enterprise)
	AIProviderLocal = "local"

	// AIProviderOpenAI uses OpenAI APIs (future extension)
	AIProviderOpenAI = "openai"

	// DefaultAgentsTableName is the default name for the agents table
	DefaultAgentsTableName = "agents"

	// DefaultProjectsTableName is the default name for the projects table
	DefaultProjectsTableName = "projects"

	// DefaultCodebasesTableName is the default name for the codebases table
	DefaultCodebasesTableName = "codebases"

	// DefaultTasksTableName is the default name for the tasks table
	DefaultTasksTableName = "tasks"

	// DefaultUsersTableName is the default name for the users table
	DefaultUsersTableName = "users"
)

var (
	// FoundationModels is a list of foundation models to be used in the application.
	FoundationModels = []string{
		// Anthropic Claude
		"anthropic.claude-instant-v1",
		"anthropic.claude-v2",
		"anthropic.claude-v2:1",
		"anthropic.claude-3-sonnet-20240229-v1:0",
		"anthropic.claude-3-5-sonnet-20240620-v1:0",

		// Mistral
		"mistral.mistral-7b-instruct-v0:2",
		"mistral.mistral-large-2402-v1:0",

		// Meta (Llama)
		"meta.llama2-13b-chat-v1",
		"meta.llama2-70b-chat-v1",

		// Cohere
		"cohere.command-r-v1",
		"cohere.command-r-plus-v1",

		// AI21 Labs
		"ai21.j2-mid-v1",
		"ai21.j2-ultra-v1",
		"ai21.j2-light-v1",

		// Amazon Titan (Text and Embeddings)
		"amazon.titan-text-lite-v1",
		"amazon.titan-text-express-v1",
		"amazon.titan-embed-text-v1",
	}
)
