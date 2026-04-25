package titan

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/Sachin0Amit/new/internal/models"
	"github.com/Sachin0Amit/new/pkg/logger"
)

// SovereignKnowledgeEntry represents a single knowledge item.
type SovereignKnowledgeEntry struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Instruction string   `json:"instruction"`
	Response    string   `json:"response"`
	Tags        []string `json:"tags"`
}

// SovereignKnowledgeFile represents the knowledge JSON structure.
type SovereignKnowledgeFile struct {
	Metadata struct {
		Source     string `json:"source"`
		Version   string `json:"version"`
		TotalB    int    `json:"total_params_b"`
		ActiveB   int    `json:"activated_params_b"`
		Created   string `json:"created"`
	} `json:"metadata"`
	Entries []SovereignKnowledgeEntry `json:"knowledge_entries"`
}

// SovereignTrainingPair represents an instruction-response pair.
type SovereignTrainingPair struct {
	Instruction string `json:"instruction"`
	Response    string `json:"response"`
}

// SovereignTrainingFile represents the training corpus JSON structure.
type SovereignTrainingFile struct {
	Metadata struct {
		Source     string   `json:"source"`
		Version   string   `json:"version"`
		TotalPairs int     `json:"total_pairs"`
		Domains   []string `json:"domains"`
	} `json:"metadata"`
	Pairs []SovereignTrainingPair `json:"training_pairs"`
}

// KnowledgeIngestor handles loading Sovereign datasets into the Sovereign knowledge store.
type KnowledgeIngestor struct {
	storage models.StorageManager
	log     logger.Logger
}

// NewKnowledgeIngestor creates a new ingestor.
func NewKnowledgeIngestor(storage models.StorageManager) *KnowledgeIngestor {
	return &KnowledgeIngestor{
		storage: storage,
		log:     logger.New(),
	}
}

// IngestKnowledge loads the deepseek_knowledge.json file into the store.
func (ki *KnowledgeIngestor) IngestKnowledge(ctx context.Context, dataDir string) (int, error) {
	path := filepath.Join(dataDir, "sovereign_knowledge.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read knowledge file: %w", err)
	}

	var kf SovereignKnowledgeFile
	if err := json.Unmarshal(data, &kf); err != nil {
		return 0, fmt.Errorf("failed to parse knowledge file: %w", err)
	}

	count := 0
	for _, entry := range kf.Entries {
		chunk := models.Chunk{
			ID:      uuid.New(),
			DocID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Content: entry.Instruction + "\n\n" + entry.Response,
			Metadata: map[string]interface{}{
				"source":   "deepseek_neural_core",
				"category": entry.Category,
				"entry_id": entry.ID,
				"tags":     entry.Tags,
			},
		}

		key := fmt.Sprintf("knowledge:%s", entry.ID)
		if err := ki.storage.Save(ctx, key, chunk); err != nil {
			ki.log.Error("Failed to ingest knowledge entry", logger.String("id", entry.ID), logger.ErrorF(err))
			continue
		}
		count++
	}

	ki.log.Info("Knowledge ingestion complete",
		logger.Int("entries", count),
		logger.String("source", kf.Metadata.Source),
		logger.String("version", kf.Metadata.Version))

	return count, nil
}

// IngestTrainingCorpus loads the deepseek_training_corpus.json file.
func (ki *KnowledgeIngestor) IngestTrainingCorpus(ctx context.Context, dataDir string) (int, error) {
	path := filepath.Join(dataDir, "sovereign_training.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read training corpus: %w", err)
	}

	var tf SovereignTrainingFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return 0, fmt.Errorf("failed to parse training corpus: %w", err)
	}

	count := 0
	for i, pair := range tf.Pairs {
		chunk := models.Chunk{
			ID:      uuid.New(),
			DocID:   uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Content: pair.Instruction + "\n\n" + pair.Response,
			Metadata: map[string]interface{}{
				"source": "sovereign_training_pipeline",
				"index":  i,
				"type":   "instruction_pair",
			},
		}

		key := fmt.Sprintf("training:%d", i)
		if err := ki.storage.Save(ctx, key, chunk); err != nil {
			continue
		}
		count++
	}

	ki.log.Info("Training corpus ingestion complete",
		logger.Int("pairs", count),
		logger.String("version", tf.Metadata.Version))

	return count, nil
}

// IngestAll loads all Sovereign datasets from the data directory.
func (ki *KnowledgeIngestor) IngestAll(ctx context.Context, dataDir string) error {
	ki.log.Info("Starting Sovereign knowledge pipeline ingestion", logger.String("dir", dataDir))

	kCount, err := ki.IngestKnowledge(ctx, dataDir)
	if err != nil {
		ki.log.Error("Knowledge ingestion failed", logger.ErrorF(err))
	}

	tCount, err := ki.IngestTrainingCorpus(ctx, dataDir)
	if err != nil {
		ki.log.Error("Training corpus ingestion failed", logger.ErrorF(err))
	}

	ki.log.Info("Full Sovereign pipeline ingestion complete",
		logger.Int("knowledge_entries", kCount),
		logger.Int("training_pairs", tCount),
		logger.Int("total", kCount+tCount))

	return nil
}
