package rag

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

type KeywordIndex struct {
	chunks   []domain.Chunk
	termFreq []map[string]int
	docLen   []int
	docFreq  map[string]int
	avgLen   float64
	k1       float64
	b        float64
}

func NewKeywordIndex(chunks []domain.Chunk) *KeywordIndex {
	ki := &KeywordIndex{
		chunks:   chunks,
		termFreq: make([]map[string]int, len(chunks)),
		docLen:   make([]int, len(chunks)),
		docFreq:  make(map[string]int),
		k1:       1.2,
		b:        0.75,
	}

	var totalLen int
	for i, ch := range chunks {
		tokens := tokenize(ch.Text)
		tf := make(map[string]int, len(tokens))
		for _, tok := range tokens {
			tf[tok]++
		}
		ki.termFreq[i] = tf
		ki.docLen[i] = len(tokens)
		totalLen += len(tokens)
		for tok := range tf {
			ki.docFreq[tok]++
		}
	}
	if len(chunks) > 0 {
		ki.avgLen = float64(totalLen) / float64(len(chunks))
	}
	return ki
}

func (ki *KeywordIndex) Search(query string, k int) []Result {
	if k <= 0 || len(ki.chunks) == 0 {
		return nil
	}
	terms := tokenize(query)
	if len(terms) == 0 {
		return nil
	}

	n := float64(len(ki.chunks))
	results := make([]Result, 0, len(ki.chunks))
	for i, ch := range ki.chunks {
		var score float64
		for _, term := range terms {
			tf, ok := ki.termFreq[i][term]
			if !ok {
				continue
			}
			df := ki.docFreq[term]
			idf := math.Log(1 + (n-float64(df)+0.5)/(float64(df)+0.5))
			denom := float64(tf) + ki.k1*(1-ki.b+ki.b*float64(ki.docLen[i])/ki.avgLen)
			score += idf * (float64(tf) * (ki.k1 + 1)) / denom
		}
		if score > 0 {
			results = append(results, Result{Chunk: ch, Score: float32(score)})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > k {
		results = results[:k]
	}
	return results
}

func tokenize(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}
