package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/papi-ai/sovereign-core/internal/models"
	"github.com/papi-ai/sovereign-core/internal/storage"
	"github.com/papi-ai/sovereign-core/pkg/errors"
	"github.com/papi-ai/sovereign-core/pkg/logger"
	"github.com/papi-ai/sovereign-core/pkg/audit"
	"github.com/papi-ai/sovereign-core/pkg/etl"
	"github.com/papi-ai/sovereign-core/pkg/p2p"
	ksync "github.com/papi-ai/sovereign-core/pkg/sync"
	"github.com/papi-ai/sovereign-core/pkg/reflex"
	"github.com/papi-ai/sovereign-core/pkg/fleet"
	"github.com/papi-ai/sovereign-core/pkg/metrics"
	"github.com/papi-ai/sovereign-core/pkg/sandbox"
	"github.com/papi-ai/sovereign-core/pkg/sensory"
	"github.com/papi-ai/sovereign-core/pkg/vector"
	"github.com/papi-ai/sovereign-core/pkg/retrieval"
	"github.com/papi-ai/sovereign-core/pkg/plugins"
	"github.com/papi-ai/sovereign-core/pkg/capabilities"
	"github.com/papi-ai/sovereign-core/pkg/security"
	"github.com/papi-ai/sovereign-core/internal/api"
)

// SovereignOrchestrator implements the models.Orchestrator interface.
type SovereignOrchestrator struct {
	engine   models.InferenceEngine
	storage  models.StorageManager
	security security.SecurityManager
	pipeline *etl.Pipeline
	sandbox  *sandbox.SandboxManager
	enforcer *capabilities.Enforcer
	vision   *sensory.VisionProcessor
	mesh     *retrieval.KnowledgeMesh
	plugins  *plugins.Registry
	hub      *api.Hub
	broker   *ksync.KnowledgeBroker
	reflex   *reflex.ReflexEngine
	fleet    *fleet.FleetScheduler
	gossip   *p2p.GossipNode
	tasks    sync.Map // Map[uuid.UUID]*models.Task
	logger   logger.Logger
}

func New(ctx context.Context, engine models.InferenceEngine, storage models.StorageManager, sec security.SecurityManager, hub *api.Hub, gossip *p2p.GossipNode) *SovereignOrchestrator {
	// Start Ingestion logic (Background)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Placeholder for future multi-node task listening
			}
		}
	}()

	store := &vector.BadgerFlatStore{
		SaveFunc: storage.Save,
		LoadFunc: storage.Load,
	}

	broker := ksync.NewKnowledgeBroker(store, gossip)

	return &SovereignOrchestrator{
		engine:   engine,
		storage:  storage,
		security: sec,
		vision:   sensory.NewVisionProcessor(224, 224),
		mesh:     retrieval.NewKnowledgeMesh(store, broker),
		plugins:  plugins.NewRegistry(logger.New()),
		hub:      hub,
		broker:   broker,
		reflex:   reflex.NewEngine(3), // Max 3 reflex iterations
		fleet:    fleet.NewScheduler(0.8), // 80% Load threshold
		gossip:   gossip,
		logger:   logger.New(),
		pipeline: etl.NewPipeline(4),
		sandbox:  sandbox.NewManager(10 * time.Second),
		enforcer: capabilities.NewEnforcer(),
	}
}

// RecoverTasks scans the persistence layer for interrupted tasks and re-queues them.
func (o *SovereignOrchestrator) RecoverTasks(ctx context.Context) error {
	o.logger.Info("Initiating state reconciliation...")
	
	// Recover PENDING tasks
	pending, _ := o.storage.(*storage.SovereignStorage).QueryTasks(ctx, models.StatusPending, 100)
	for _, t := range pending {
		o.tasks.Store(t.ID, t)
		go o.processTask(ctx, t.ID)
	}

	// Recover RUNNING tasks (treat as pending since they were interrupted)
	running, _ := o.storage.(*storage.SovereignStorage).QueryTasks(ctx, models.StatusRunning, 100)
	for _, t := range running {
		t.Status = models.StatusPending
		o.tasks.Store(t.ID, t)
		go o.processTask(ctx, t.ID)
	}

	o.logger.Info("Reconciliation complete", logger.Int("recovered", len(pending)+len(running)))
	return nil
}

// IngestLocalData triggers a parallel ETL pipeline to vectorize local documents.
func (s *SovereignOrchestrator) IngestLocalData(ctx context.Context, data []string) error {
	s.pipeline.AddStage(etl.NewTextChunker(256))       // 256 word chunks
	s.pipeline.AddStage(etl.NewMockEmbeddingStage(128)) // 128 Dimension vectors

	inputs := make([]interface{}, len(data))
	for i, v := range data {
		inputs[i] = v
	}

	results := s.pipeline.Process(ctx, inputs)
	for job := range results {
		if job.Error != nil {
			continue // Log and move on
		}
		
		chunks := job.Payload.([]models.Chunk)
		for _, chunk := range chunks {
			key := fmt.Sprintf("chunk:%s", chunk.ID.String())
			s.storage.Save(ctx, key, chunk)
		}
	}

	return nil
}

func (o *SovereignOrchestrator) SubmitTask(ctx context.Context, task *models.Task) (uuid.UUID, error) {
	// 1. Rate Limiting Check
	if !o.security.Allow() {
		return uuid.Nil, errors.New(errors.CodeInternal, "rate limit exceeded - cognitive core saturated", nil)
	}

	// 2. Payload Validation
	if err := o.security.Validate(task.Payload); err != nil {
		return uuid.Nil, err
	}

	// 3. Input Sanitization
	for k, v := range task.Payload {
		if s, ok := v.(string); ok {
			task.Payload[k] = o.security.Sanitize(s)
		}
	}

	task.ID = uuid.New()
	task.Payload = task.Payload
	task.Status = models.StatusPending
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	// 1. Check for autonomous offloading
	if o.fleet.ShouldOffload(0.5) { // Pass current load tracking metrics
		if targetID, ok := o.fleet.SelectMigrationTarget(); ok {
			o.logger.Info("Delegating task to fleet matrix", logger.String("target", targetID))
			// Node delegation logic would go here
		}
	}

	o.tasks.Store(task.ID, task)
	o.logger.Info("Task submitted and secured", logger.String("task_id", task.ID.String()))

	// Route task based on type
	if task.Type == "tool_use" {
		go o.processToolTask(context.Background(), task.ID)
	} else if task.Type == "multimodal" {
		go o.processMultimodalTask(context.Background(), task.ID)
	} else {
		go o.processTask(context.Background(), task.ID)
	}

	return task.ID, nil
}

func (o *SovereignOrchestrator) GetTask(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	val, ok := o.tasks.Load(id)
	if !ok {
		return nil, errors.New(errors.CodeNotFound, "task not found", nil)
	}
	return val.(*models.Task), nil
}

func (o *SovereignOrchestrator) processTask(ctx context.Context, id uuid.UUID) {
	val, _ := o.tasks.Load(id)
	task := val.(*models.Task)

	task.Status = models.StatusRunning
	task.UpdatedAt = time.Now()

	// 0. Initialize Auditor
	auditor, _ := audit.NewAuditor()
	auditor.Record("INITIALIZATION", "Cognitive derivation loop started", nil)

	// 1. Pre-Processing Hook Point
	o.plugins.InvokeHook(models.HookPreProcessing, task)
	auditor.Record("PRE_PROCESSING", "Dynamic plugin middleware executed", nil)

	// 1.5 Tier Propagation & Assignment
	if tier, ok := task.Payload["tier"].(string); !ok || tier == "" {
		// Assign fallback tier if not specified
		task.Payload["tier"] = "local"
		auditor.Record("ORCHESTRATION", "No tier specified; defaulting to 'local'", nil)
	} else {
		auditor.Record("ORCHESTRATION", fmt.Sprintf("Propagating architectural selection: %s", tier), nil)
	}

	// 2. Semantic Retrieval Stage (RAG)
	if _, exists := task.Payload["prompt"]; exists {
		o.logger.Debug("Initiating semantic memory retrieval", logger.String("task_id", id.String()))
		queryVec := make([]float32, 128) 
		chunks, _ := o.mesh.Search(ctx, queryVec, 3)
		
		contextStr := retrieval.BuildContext(chunks)
		if contextStr != "" {
			task.Payload["prompt_context"] = contextStr
			auditor.Record("RETRIEVAL", "Injected semantic memory context", map[string]interface{}{
				"chunks": len(chunks),
			})
		}
	}

	start := time.Now()
	auditor.Record("INFERENCE", "Dispatching payload to Titan C++ Core", nil)
	result, err := o.engine.Infer(ctx, task.Payload)
	duration := time.Since(start)

	if err != nil {
		task.Status = models.StatusFailed
		metrics.RecordTaskCompletion("failed")
		o.logger.Error("Inference failed", logger.String("task_id", id.String()), logger.ErrorF(err))
	} else {
		task.Status = models.StatusCompleted
		task.Result = result
		metrics.RecordTaskCompletion("success")
		metrics.RecordLatency("titan_v1", duration.Seconds())
		o.logger.Info("Inference completed", logger.String("task_id", id.String()))
		
		// 3. Post-Inference Hook Point
		o.plugins.InvokeHook(models.HookPostInference, task)

		// 4. Finalize Audit Trail & Sign Derivation
		auditor.Record("FINALIZATION", "Derivation complete. Generating proof-of-authenticity.", nil)
		auditor.Sign(result)

		// 5. Autonomous Reflex Check
		eval := o.reflex.Evaluate(task)
		if eval.Action != reflex.ActionNone && eval.Action != reflex.ActionAbort {
			o.logger.Info("Self-Correction Reflex Triggered", 
				logger.String("task_id", id.String()), 
				logger.String("reason", eval.Reason),
				logger.String("validator", eval.ValidatorID))
			
			task.ReflexDepth++
			task.Payload["reflex_correction"] = eval.Correction
			
			// Record reflex in audit trail
			auditor.Record("REFLEX_ADJUSTMENT", fmt.Sprintf("Autonomous correction triggered by %s", eval.ValidatorID), map[string]interface{}{
				"reason": eval.Reason,
				"depth":  task.ReflexDepth,
			})
			auditor.Sign(result) // Sign the intermediate failed result as well for proof-of-reflex

			// Broadcast Reflex Event
			o.hub.Broadcast(api.Message{
				Type: api.EventMetrics,
				Payload: map[string]interface{}{
					"event":      "REFLEX",
					"task_id":    id,
					"reason":     eval.Reason,
					"validator":  eval.ValidatorID,
					"depth":      task.ReflexDepth,
					"correction": eval.Correction,
				},
				Timestamp: time.Now().UnixMilli(),
			})
			
			go o.processTask(ctx, id)
			return
		}

		// Broadcast completion to Command Center
		o.hub.Broadcast(api.Message{
			Type:      api.EventTask,
			Payload:   task,
			Timestamp: time.Now().UnixMilli(),
		})
	}

	task.UpdatedAt = time.Now()
	
	// Persist to local LSM-tree with Indexing
	if s, ok := o.storage.(*storage.SovereignStorage); ok {
		if err := s.SaveIndexedTask(ctx, task); err != nil {
			o.logger.Error("Failed to persist task state", logger.ErrorF(err))
		}
	} else {
		o.storage.Save(ctx, fmt.Sprintf("task:%s", task.ID), task)
	}
}

func (o *SovereignOrchestrator) processToolTask(ctx context.Context, id uuid.UUID) {
	val, _ := o.tasks.Load(id)
	task := val.(*models.Task)

	task.Status = models.StatusRunning
	task.UpdatedAt = time.Now()

	// 1. Extract Capabilities
	rawCaps, _ := task.Payload["required_capabilities"].([]interface{})
	var caps []capabilities.Capability
	for _, c := range rawCaps {
		caps = append(caps, capabilities.Capability(c.(string)))
	}

	// 2. Enforce Permissions
	if err := o.enforcer.Authorize(caps); err != nil {
		task.Status = models.StatusFailed
		o.logger.Error("Capability authorization failed", logger.ErrorF(err))
		return
	}

	// 3. Execute in Sandbox
	cmd, _ := task.Payload["command"].(string)
	argsRaw, _ := task.Payload["args"].([]interface{})
	var args []string
	for _, a := range argsRaw {
		args = append(args, a.(string))
	}

	record, err := o.sandbox.Execute(ctx, cmd, args...)
	if err != nil {
		task.Status = models.StatusFailed
		o.logger.Error("Sandbox execution failed", logger.ErrorF(err))
	} else {
		task.Status = models.StatusCompleted
		task.Result = &models.TaskResult{
			Data: map[string]interface{}{
				"stdout":    record.Stdout,
				"stderr":    record.Stderr,
				"exit_code": record.ExitCode,
				"duration":  record.Duration.String(),
			},
			Completed: time.Now(),
		}
		o.logger.Info("Tool execution completed", logger.String("task_id", id.String()))
	}

	task.UpdatedAt = time.Now()
	
	// Reflex Evaluation for Tool Use
	eval := o.reflex.Evaluate(task)
	if eval.Action == reflex.ActionRetry && task.ReflexDepth < 3 {
		o.logger.Info("Self-Correction Reflex Triggered for Tool", 
			logger.String("task_id", id.String()), 
			logger.String("reason", eval.Reason),
			logger.String("validator", eval.ValidatorID))
			
		task.ReflexDepth++
		task.Payload["command_correction"] = eval.Correction
		
		// Broadcast Reflex Event for Tools
		o.hub.Broadcast(api.Message{
			Type: api.EventMetrics,
			Payload: map[string]interface{}{
				"event":      "REFLEX",
				"subtype":    "TOOL",
				"task_id":    id,
				"reason":     eval.Reason,
				"validator":  eval.ValidatorID,
				"depth":      task.ReflexDepth,
				"correction": eval.Correction,
			},
			Timestamp: time.Now().UnixMilli(),
		})

		// Re-submit for tool processing with feedback
		go o.processToolTask(ctx, id)
		return
	}

	o.storage.Save(ctx, fmt.Sprintf("task:%s", id), task)
}

func (o *SovereignOrchestrator) processMultimodalTask(ctx context.Context, id uuid.UUID) {
	val, _ := o.tasks.Load(id)
	task := val.(*models.Task)

	task.Status = models.StatusRunning
	task.UpdatedAt = time.Now()

	// 1. Check for sensory input
	if raw, ok := task.Payload["sensory_input"].(models.SensoryData); ok {
		if raw.Type == "image" {
			frame, err := o.vision.ProcessImage(raw.Buffer)
			if err != nil {
				o.logger.Error("Sensory processing failed", logger.ErrorF(err))
				task.Status = models.StatusFailed
				return
			}
			task.Payload["visual_tensor"] = frame
		}
	}

	// 2. Pass to engine (Same unified Infer call)
	o.processTask(ctx, id)
}

func (o *SovereignOrchestrator) Shutdown() error {
	o.logger.Info("Orchestrator shutting down...")
	return nil
}

// GetTasks retrieves a slice of recently submitted tasks.
func (o *SovereignOrchestrator) GetTasks(ctx context.Context, status models.TaskStatus, limit int) ([]*models.Task, error) {
	return o.storage.(*storage.SovereignStorage).QueryTasks(ctx, status, limit)
}
