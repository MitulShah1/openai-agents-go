package tools

import (
	"context"

	"github.com/MitulShah1/openai-agents-go/pkg/embeddings"
	"github.com/MitulShah1/openai-agents-go/pkg/vectorstore"
)

// NewEmbedTextTool creates a tool that generates embeddings for text.
func NewEmbedTextTool(client *embeddings.Client) Tool {
	return FromFunc(
		"embed_text",
		"Generate embeddings for a list of text strings.",
		func(args struct {
			Texts []string `json:"texts" jsonschema:"description=List of texts to embed"`
			Model string   `json:"model,omitempty" jsonschema:"description=Embedding model to use (default: text-embedding-3-small)"`
		}) (any, error) {
			ctx := context.Background()
			vectors, err := client.Generate(ctx, args.Texts, args.Model)
			if err != nil {
				return nil, err
			}
			return vectors, nil
		},
	)
}

// NewCreateVectorStoreTool creates a tool that creates a new vector store.
func NewCreateVectorStoreTool(client *vectorstore.Client) Tool {
	return FromFunc(
		"create_vector_store",
		"Create a new vector store.",
		func(args struct {
			Name string `json:"name" jsonschema:"description=Name of the vector store"`
		}) (any, error) {
			ctx := context.Background()
			vs, err := client.Create(ctx, args.Name)
			if err != nil {
				return nil, err
			}
			return vs, nil
		},
	)
}

// NewAddFilesToVectorStoreTool creates a tool that adds files to a vector store.
func NewAddFilesToVectorStoreTool(client *vectorstore.Client) Tool {
	return FromFunc(
		"add_files_to_vector_store",
		"Add files to a vector store.",
		func(args struct {
			VectorStoreID string   `json:"vector_store_id" jsonschema:"description=ID of the vector store"`
			FileIDs       []string `json:"file_ids" jsonschema:"description=List of file IDs to add"`
		}) (any, error) {
			ctx := context.Background()
			err := client.AddFiles(ctx, args.VectorStoreID, args.FileIDs)
			if err != nil {
				return nil, err
			}
			return map[string]string{"status": "batch_created", "vector_store_id": args.VectorStoreID}, nil
		},
	)
}
