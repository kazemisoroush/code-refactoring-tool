// Package ai contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package ai

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
)

// BedrockRAGBuilder is an implementation of RAGBuilder that uses AWS Bedrock for building the RAG pipeline.
type BedrockRAGBuilder struct {
	s3Client *s3.Client
	kbClient *bedrockagent.Client
}

// NewBedrockRAGBuilder creates a new instance of BedrockRAGBuilder.
func NewBedrockRAGBuilder(cfg aws.Config) (RAGBuilder, error) {
	return &BedrockRAGBuilder{
		s3Client: s3.NewFromConfig(cfg),
		kbClient: bedrockagent.NewFromConfig(cfg),
	}, nil
}

// Build implements RAGBuilder.
func (b BedrockRAGBuilder) Build(_ context.Context, _ repository.Repository) (RAGMetadata, error) {
	panic("unimplemented")
}
