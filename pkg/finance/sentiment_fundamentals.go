package finance

/*
#cgo CXXFLAGS: -std=c++17 -O3 -march=native -Wall -Wextra
#cgo LDFLAGS: -L${SRCDIR}/../../internal/titan/cpp/finance -lfinance_engine -lstdc++ -lm
#include <stdlib.h>

char* finance_analyze_sentiment(const char* headline);
char* finance_analyze_fundamentals(const char* symbol);
void finance_free_string(char* s);
void* finance_engine_init();
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"unsafe"
)

// SentimentScore represents the result of the NLP Lexicon analyzer.
type SentimentScore struct {
	Score         float64 `json:"score"`
	PositiveWords int     `json:"positive_words"`
	NegativeWords int     `json:"negative_words"`
}

// FundamentalScore represents the elite chartered accountant evaluation.
type FundamentalScore struct {
	TotalScore float64 `json:"total_score"`
	Rating     string  `json:"rating"`
	Reasoning  string  `json:"reasoning"`
}

func init() {
	C.finance_engine_init()
}

// AnalyzeSentiment calls the C++ LexiconAnalyzer.
func AnalyzeSentiment(headline string) (*SentimentScore, error) {
	cHeadline := C.CString(headline)
	defer C.free(unsafe.Pointer(cHeadline))

	cResult := C.finance_analyze_sentiment(cHeadline)
	if cResult == nil {
		return nil, fmt.Errorf("sentiment analysis failed in C++ engine")
	}
	defer C.finance_free_string(cResult)

	goResult := C.GoString(cResult)

	var result SentimentScore
	if err := json.Unmarshal([]byte(goResult), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AnalyzeFundamentals calls the C++ FundamentalAnalyzer.
func AnalyzeFundamentals(symbol string) (*FundamentalScore, error) {
	cSymbol := C.CString(symbol)
	defer C.free(unsafe.Pointer(cSymbol))

	cResult := C.finance_analyze_fundamentals(cSymbol)
	if cResult == nil {
		return nil, fmt.Errorf("fundamental analysis failed in C++ engine")
	}
	defer C.finance_free_string(cResult)

	goResult := C.GoString(cResult)

	var result FundamentalScore
	if err := json.Unmarshal([]byte(goResult), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
