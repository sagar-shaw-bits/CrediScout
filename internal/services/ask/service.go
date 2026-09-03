package ask

import (
	"context"
	"fmt"
	"strings"

	"github.com/sagar-shaw-bits/CrediScout/internal/llm"
	"github.com/sagar-shaw-bits/CrediScout/internal/rag"
)

const systemPrompt = `You are CrediScout, an assistant that answers questions about HDFC Bank credit cards.
Answer ONLY using the provided context. If the context does not contain the answer, say you don't have that information.
Be concise and factual. Do not invent fees, features, or eligibility details.`

type Retriever interface {
	Retrieve(ctx context.Context, query string, k int) ([]rag.Result, error)
}
type Service interface {
	Ask(ctx context.Context, req AskRequest) (AskResponse, error)
}

type service struct {
	retriever   Retriever
	model       llm.LLM
	defaultTopK int
}

func NewService(retriever Retriever, model llm.LLM, defaultTopK int) Service {
	if defaultTopK <= 0 {
		defaultTopK = 5
	}
	return &service{retriever: retriever, model: model, defaultTopK: defaultTopK}
}

func (s *service) Ask(ctx context.Context, req AskRequest) (AskResponse, error) {
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return AskResponse{}, fmt.Errorf("ask: question is required")
	}

	k := req.TopK
	if k <= 0 {
		k = s.defaultTopK
	}

	hits, err := s.retriever.Retrieve(ctx, question, k)
	if err != nil {
		return AskResponse{}, fmt.Errorf("ask: retrieve: %w", err)
	}
	if len(hits) == 0 {
		return AskResponse{Answer: "I don't have any information on that.", Sources: []Source{}}, nil
	}

	userPrompt := buildPrompt(question, hits)
	answer, err := s.model.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return AskResponse{}, fmt.Errorf("ask: generate: %w", err)
	}

	return AskResponse{
		Answer:  strings.TrimSpace(answer),
		Sources: dedupSources(hits),
	}, nil
}

func buildPrompt(question string, hits []rag.Result) string {
	var b strings.Builder
	b.WriteString("Context:\n")
	for i, h := range hits {
		fmt.Fprintf(&b, "[%d] %s\n%s\n\n", i+1, h.Chunk.URL, h.Chunk.Text)
	}
	b.WriteString("Question: ")
	b.WriteString(question)
	b.WriteString("\nAnswer:")
	return b.String()
}

func dedupSources(hits []rag.Result) []Source {
	seen := make(map[string]int)
	out := make([]Source, 0, len(hits))
	for _, h := range hits {
		if idx, ok := seen[h.Chunk.URL]; ok {
			if h.Score > out[idx].Score {
				out[idx].Score = h.Score
			}
			continue
		}
		seen[h.Chunk.URL] = len(out)
		out = append(out, Source{URL: h.Chunk.URL, Title: h.Chunk.Title, Score: h.Score})
	}
	return out
}
