package knowledge

import (
	"sort"
	"strings"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
)

type KeywordRetriever struct {
	store *postgres.Store
}

func NewKeywordRetriever(store *postgres.Store) *KeywordRetriever {
	return &KeywordRetriever{store: store}
}

type scoredEntry struct {
	entry knowledge.KnowledgeEntry
	score int
}

func (r *KeywordRetriever) Retrieve(query string) ([]knowledge.KnowledgeEntry, error) {
	queryLower := strings.ToLower(query)
	queryWords := tokenize(queryLower)

	if len(queryWords) == 0 {
		return nil, nil
	}

	allEntries, err := r.store.FindAllKnowledgeEntries()
	if err != nil {
		return nil, err
	}

	var scored []scoredEntry
	for _, entry := range allEntries {
		score := calculateScore(queryWords, queryLower, &entry)
		if score > 0 {
			scored = append(scored, scoredEntry{entry: entry, score: score})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	maxResults := 3
	if len(scored) < maxResults {
		maxResults = len(scored)
	}

	results := make([]knowledge.KnowledgeEntry, maxResults)
	for i := 0; i < maxResults; i++ {
		results[i] = scored[i].entry
	}
	return results, nil
}

func (r *KeywordRetriever) IsStrongMatch(query string) (*knowledge.KnowledgeEntry, bool) {
	queryLower := strings.TrimSpace(strings.ToLower(query))
	queryWords := tokenize(queryLower)

	if len(queryWords) == 0 {
		return nil, false
	}

	allEntries, err := r.store.FindAllKnowledgeEntries()
	if err != nil {
		return nil, false
	}

	var bestEntry *knowledge.KnowledgeEntry
	bestScore := 0

	for i, entry := range allEntries {
		if strings.ToLower(strings.TrimSpace(entry.Title)) == queryLower {
			return &allEntries[i], true
		}

		score := calculateScore(queryWords, queryLower, &entry)
		if score > bestScore {
			bestScore = score
			bestEntry = &allEntries[i]
		}
	}

	if bestScore >= 8 {
		return bestEntry, true
	}
	return nil, false
}

func calculateScore(queryWords []string, queryLower string, entry *knowledge.KnowledgeEntry) int {
	score := 0
	entryTags := strings.ToLower(entry.Tags)
	entryTitle := strings.ToLower(entry.Title)
	entryContent := strings.ToLower(entry.Content)

	tags := strings.Split(entryTags, ",")
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if strings.Contains(queryLower, tag) {
			score += 2
		}
		for _, word := range queryWords {
			if strings.Contains(tag, word) || strings.Contains(word, tag) {
				score++
			}
		}
	}

	for _, word := range queryWords {
		if len(word) > 2 && strings.Contains(entryTitle, word) {
			score += 2
		}
	}

	for _, word := range queryWords {
		if len(word) > 3 && strings.Contains(entryContent, word) {
			score += 1
		}
	}

	return score
}

func tokenize(text string) []string {
	stopWords := map[string]bool{
		"yang": true, "dan": true, "di": true, "ke": true, "dari": true,
		"ini": true, "itu": true, "untuk": true, "dengan": true, "pada": true,
		"adalah": true, "saya": true, "apa": true, "bagaimana": true, "mau": true,
		"bisa": true, "ada": true, "tidak": true, "ya": true, "atau": true,
		"the": true, "is": true, "a": true, "an": true, "how": true,
	}

	words := strings.Fields(text)
	var result []string
	for _, w := range words {
		w = strings.TrimFunc(w, func(r rune) bool {
			return r == '?' || r == '!' || r == '.' || r == ','
		})
		if len(w) > 1 && !stopWords[w] {
			result = append(result, w)
		}
	}
	return result
}
