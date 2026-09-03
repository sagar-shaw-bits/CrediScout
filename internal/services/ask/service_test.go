package ask

import (
	"context"
	"errors"
	"strings"
	"testing"

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
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}

	if resp.Answer != "The joining fee is 1000." {
		t.Errorf("answer = %q", resp.Answer)
	}
	// Two hits share a URL => deduped to 2 sources.
	if len(resp.Sources) != 2 {
		t.Fatalf("expected 2 deduped sources, got %d", len(resp.Sources))
	}
	if resp.Sources[0].Score != 0.9 {
		t.Errorf("expected highest score kept, got %v", resp.Sources[0].Score)
	}
	// Prompt should contain the question and the retrieved context.
	if !strings.Contains(model.gotUserMsg, "What is the fee?") {
		t.Error("prompt missing question")
	}
	if !strings.Contains(model.gotUserMsg, "Joining fee is 1000.") {
		t.Error("prompt missing retrieved context")
	}
}

func TestAsk_EmptyQuestion(t *testing.T) {
	svc := NewService(&fakeRetriever{}, &fakeLLM{}, 5)
	if _, err := svc.Ask(context.Background(), AskRequest{Question: "   "}); err == nil {
		t.Fatal("expected error for blank question")
	}
}

func TestAsk_NoHitsShortCircuits(t *testing.T) {
	model := &fakeLLM{answer: "should not be called"}
	svc := NewService(&fakeRetriever{hits: nil}, model, 5)

	resp, err := svc.Ask(context.Background(), AskRequest{Question: "unknown"})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if model.gotUserMsg != "" {
		t.Error("LLM should not be called when there are no hits")
	}
	if len(resp.Sources) != 0 {
		t.Errorf("expected no sources, got %d", len(resp.Sources))
	}
}

func TestAsk_RetrieveErrorPropagates(t *testing.T) {
	svc := NewService(&fakeRetriever{err: errors.New("boom")}, &fakeLLM{}, 5)
	if _, err := svc.Ask(context.Background(), AskRequest{Question: "q"}); err == nil {
		t.Fatal("expected retrieve error to propagate")
	}
}

func TestAsk_UsesRequestTopK(t *testing.T) {
	ret := &fakeRetriever{hits: []rag.Result{{Chunk: domain.Chunk{URL: "u"}, Score: 1}}}
	svc := NewService(ret, &fakeLLM{answer: "a"}, 5)

	if _, err := svc.Ask(context.Background(), AskRequest{Question: "q", TopK: 8}); err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if ret.gotK != 8 {
		t.Errorf("expected k=8 passed to retriever, got %d", ret.gotK)
	}
}
