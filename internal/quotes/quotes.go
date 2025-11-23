package quotes

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"
)

type QuoteData struct {
	Quotes []string `json:"quotes"`
}

type QuoteManager struct {
	quotes []string
	rand   *rand.Rand
}

func New(filePath string) (*QuoteManager, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		// Return default quotes if file doesn't exist
		return &QuoteManager{
			quotes: []string{
				"A cage went in search of a bird.",
				"The meaning of life is that it stops.",
				"Paths are made by walking.",
			},
			rand: rand.New(rand.NewSource(time.Now().UnixNano())),
		}, nil
	}

	var quoteData QuoteData
	if err := json.Unmarshal(data, &quoteData); err != nil {
		return nil, err
	}

	return &QuoteManager{
		quotes: quoteData.Quotes,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (q *QuoteManager) GetRandom() string {
	if len(q.quotes) == 0 {
		return "No quotes available."
	}
	return "💭 " + q.quotes[q.rand.Intn(len(q.quotes))]
}
