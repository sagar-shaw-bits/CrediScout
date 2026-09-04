package ask

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
	"github.com/sagar-shaw-bits/CrediScout/internal/rag"
)

// fakeRetriever returns canned hits and records the k it was asked for.
type fakeRetriever struct {
	hits   []rag.Result
	err    error
	gotK   int
	gotQry string
}

func (f *fakeRetriever) Retrieve(_ context.Context, query string, k int) ([]rag.Result, error) {
	f.gotK = k
	f.gotQry = query
	return f.hits, f.err
}

// fakeLLM echoes a fixed answer and captures the prompt it received.
type fakeLLM struct {
	answer     string
	err        error
	gotSystem  string
	gotUserMsg string
}

func (f *fakeLLM) Generate(_ context.Context, system, user string) (string, error) {
	f.gotSystem = system
	f.gotUserMsg = user
	return f.answer, f.err
}

func TestAsk_HappyPath(t *testing.T) {
	ret := &fakeRetriever{hits: []rag.Result{
		{Chunk: domain.Chunk{URL: "https://hdfc/millennia/fees", Title: "Fees", Text: "Joining fee is 1000."}, Score: 0.9},
		{Chunk: domain.Chunk{URL: "https://hdfc/millennia/fees", Title: "Fees", Text: "Renewal fee is 1000."}, Score: 0.8},
		{Chunk: domain.Chunk{URL: "https://hdfc/millennia", Title: "Millennia", Text: "Overview."}, Score: 0.5},
	}}
	model := &fakeLLM{answer: "The joining fee is 1000."}

	svc := NewService(ret, model, 5)
	resp, err := svc.Ask(context.Background(), AskRequest{Question: "What is the fee?"})
	require.NoError(t, err)

	assert.Equal(t, "The joining fee is 1000.", resp.Answer)
	// Two hits share a URL => deduped to 2 sources.
	require.Len(t, resp.Sources, 2, "expected deduped sources")
	assert.Equal(t, float32(0.9), resp.Sources[0].Score, "highest score kept")
	// Prompt should contain the question and the retrieved context.
	assert.Contains(t, model.gotUserMsg, "What is the fee?", "prompt missing question")
	assert.Contains(t, model.gotUserMsg, "Joining fee is 1000.", "prompt missing retrieved context")
}

func TestAsk_EmptyQuestion(t *testing.T) {
	svc := NewService(&fakeRetriever{}, &fakeLLM{}, 5)
	_, err := svc.Ask(context.Background(), AskRequest{Question: "   "})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyQuestion)
}

func TestAsk_NoHitsShortCircuits(t *testing.T) {
	model := &fakeLLM{answer: "should not be called"}
	svc := NewService(&fakeRetriever{hits: nil}, model, 5)

	resp, err := svc.Ask(context.Background(), AskRequest{Question: "unknown"})
	require.NoError(t, err)
	assert.Empty(t, model.gotUserMsg, "LLM should not be called when there are no hits")
	assert.Empty(t, resp.Sources)
}

func TestAsk_RetrieveErrorPropagates(t *testing.T) {
	svc := NewService(&fakeRetriever{err: errors.New("boom")}, &fakeLLM{}, 5)
	_, err := svc.Ask(context.Background(), AskRequest{Question: "q"})
	assert.Error(t, err)
}

func TestAsk_UsesRequestTopK(t *testing.T) {
	ret := &fakeRetriever{hits: []rag.Result{{Chunk: domain.Chunk{URL: "u"}, Score: 1}}}
	svc := NewService(ret, &fakeLLM{answer: "a"}, 5)

	_, err := svc.Ask(context.Background(), AskRequest{Question: "q", TopK: 8})
	require.NoError(t, err)
	assert.Equal(t, 8, ret.gotK, "k should be passed through to retriever")
}
