package news

import (
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Article represents a single parsed news item.
type Article struct {
	Title       string
	Description string
	Link        string
	PubDate     time.Time
	Source      string
}

// Aggregator handles fetching and caching internet RSS feeds.
type Aggregator struct {
	mu       sync.RWMutex
	articles []Article
	feeds    []string
}

// RSS represents the structure of an RSS 2.0 XML feed.
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// NewAggregator creates a new internet-connected news aggregator.
func NewAggregator() *Aggregator {
	return &Aggregator{
		articles: make([]Article, 0),
		feeds: []string{
			"https://feeds.finance.yahoo.com/rss/2.0/headline?s=AAPL,MSFT,TSLA,BTC-USD,ETH-USD",
			// "https://www.investing.com/rss/news_25.rss", // Crypto news
			// Add more standard RSS feeds here
		},
	}
}

// FetchNews connects to the internet to download and parse the latest news.
func (a *Aggregator) FetchNews() error {
	client := &http.Client{Timeout: 10 * time.Second}
	var newArticles []Article

	for _, url := range a.feeds {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		// Mimic a standard browser to avoid basic blocks
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		var rss RSS
		if err := xml.Unmarshal(body, &rss); err != nil {
			continue
		}

		for _, item := range rss.Channel.Items {
			// Try to parse the pubDate
			pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
			if err != nil {
				pubDate, err = time.Parse(time.RFC1123, item.PubDate)
				if err != nil {
					pubDate = time.Now()
				}
			}

			// Clean description of HTML tags
			desc := stripHTML(item.Description)

			newArticles = append(newArticles, Article{
				Title:       item.Title,
				Description: desc,
				Link:        item.Link,
				PubDate:     pubDate,
				Source:      rss.Channel.Title,
			})
		}
	}

	a.mu.Lock()
	a.articles = newArticles
	a.mu.Unlock()

	return nil
}

// GetNewsForSymbol returns recent articles that mention the ticker symbol.
func (a *Aggregator) GetNewsForSymbol(symbol string) []Article {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var filtered []Article
	symLower := strings.ToLower(symbol)
	
	// Add common name mappings
	aliases := []string{symLower}
	if symbol == "AAPL" { aliases = append(aliases, "apple") }
	if symbol == "MSFT" { aliases = append(aliases, "microsoft") }
	if symbol == "BTC" { aliases = append(aliases, "bitcoin", "crypto") }
	if symbol == "ETH" { aliases = append(aliases, "ethereum") }

	for _, article := range a.articles {
		titleLower := strings.ToLower(article.Title)
		descLower := strings.ToLower(article.Description)
		
		match := false
		for _, alias := range aliases {
			if strings.Contains(titleLower, alias) || strings.Contains(descLower, alias) {
				match = true
				break
			}
		}
		
		if match {
			filtered = append(filtered, article)
		}
	}

	// Limit to top 5 most recent
	if len(filtered) > 5 {
		filtered = filtered[:5]
	}

	return filtered
}

// GetRecent returns the latest global financial news.
func (a *Aggregator) GetRecent(limit int) []Article {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	if len(a.articles) == 0 {
		return []Article{}
	}

	if limit > len(a.articles) {
		limit = len(a.articles)
	}

	return a.articles[:limit]
}

// Simple regex-less HTML stripper for descriptions
func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, char := range s {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(char)
		}
	}
	return strings.TrimSpace(result.String())
}
