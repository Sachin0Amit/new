#!/usr/bin/env python3
"""
Evaluation Harness for Sovereign Intelligence Core
Measures: Perplexity, BLEU, ROUGE-L, Task Accuracy
"""

import json
import logging
import math
import torch
from pathlib import Path
from typing import List, Dict, Tuple
from dataclasses import dataclass
from collections import defaultdict

from transformers import AutoModelForCausalLM, AutoTokenizer
from datasets import load_dataset
import numpy as np

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@dataclass
class EvaluationMetrics:
    """Container for evaluation metrics"""
    perplexity: float
    bleu_score: float
    rouge_l_score: float
    task_accuracy: float
    inference_time: float  # seconds per sample
    
    def to_dict(self) -> Dict:
        return {
            "perplexity": self.perplexity,
            "bleu_score": self.bleu_score,
            "rouge_l_score": self.rouge_l_score,
            "task_accuracy": self.task_accuracy,
            "inference_time_sec": self.inference_time,
        }


class ModelEvaluator:
    """Evaluates model performance"""
    
    def __init__(self, model_name: str, device: str = "cuda" if torch.cuda.is_available() else "cpu"):
        self.model_name = model_name
        self.device = device
        
        logger.info(f"Loading model: {model_name}")
        self.tokenizer = AutoTokenizer.from_pretrained(model_name)
        if self.tokenizer.pad_token is None:
            self.tokenizer.pad_token = self.tokenizer.eos_token
        
        self.model = AutoModelForCausalLM.from_pretrained(
            model_name,
            device_map="auto" if device == "cuda" else "cpu",
            torch_dtype=torch.float16 if device == "cuda" else torch.float32,
        )
        self.model.eval()
    
    def calculate_perplexity(self, texts: List[str]) -> float:
        """Calculate perplexity on a set of texts"""
        logger.info("Calculating perplexity...")
        
        total_loss = 0.0
        total_tokens = 0
        
        with torch.no_grad():
            for text in texts:
                # Tokenize
                inputs = self.tokenizer(text, return_tensors="pt", truncation=True, max_length=512)
                inputs = {k: v.to(self.device) for k, v in inputs.items()}
                
                # Forward pass
                outputs = self.model(**inputs, labels=inputs["input_ids"])
                loss = outputs.loss
                
                if loss is not None:
                    total_loss += loss.item() * inputs["input_ids"].shape[1]
                    total_tokens += inputs["input_ids"].shape[1]
        
        if total_tokens == 0:
            return float('inf')
        
        avg_loss = total_loss / total_tokens
        perplexity = math.exp(avg_loss)
        
        logger.info(f"Perplexity: {perplexity:.4f}")
        return perplexity
    
    def bleu_score(self, reference: str, candidate: str, max_n: int = 4) -> float:
        """Calculate BLEU score"""
        ref_tokens = reference.lower().split()
        cand_tokens = candidate.lower().split()
        
        if len(cand_tokens) == 0:
            return 0.0
        
        score = 1.0
        for n in range(1, min(max_n + 1, len(cand_tokens) + 1)):
            # Get n-grams
            ref_ngrams = self._get_ngrams(ref_tokens, n)
            cand_ngrams = self._get_ngrams(cand_tokens, n)
            
            # Count matches
            matches = sum(min(cand_ngrams[ng], ref_ngrams[ng]) 
                         for ng in cand_ngrams if ng in ref_ngrams)
            total = max(len(cand_tokens) - n + 1, 0)
            
            if total == 0:
                continue
            
            precision = matches / total if total > 0 else 0
            score *= precision ** (1 / max_n)
        
        # Brevity penalty
        if len(cand_tokens) < len(ref_tokens):
            bp = math.exp(1 - len(ref_tokens) / len(cand_tokens))
            score *= bp
        
        return score
    
    def rouge_l_score(self, reference: str, candidate: str) -> float:
        """Calculate ROUGE-L score (F-measure)"""
        ref_tokens = reference.lower().split()
        cand_tokens = candidate.lower().split()
        
        if len(ref_tokens) == 0 or len(cand_tokens) == 0:
            return 0.0
        
        # Longest common subsequence
        lcs_len = self._lcs_length(ref_tokens, cand_tokens)
        
        precision = lcs_len / len(cand_tokens) if len(cand_tokens) > 0 else 0
        recall = lcs_len / len(ref_tokens) if len(ref_tokens) > 0 else 0
        
        if precision + recall == 0:
            return 0.0
        
        f_score = (2 * precision * recall) / (precision + recall)
        return f_score
    
    def evaluate_on_dataset(self, val_data_path: str, max_samples: int = 100) -> EvaluationMetrics:
        """Evaluate on a validation dataset"""
        logger.info(f"Loading validation data from {val_data_path}")
        
        # Load dataset
        dataset = load_dataset("json", data_files=val_data_path, split="train")
        
        if max_samples:
            dataset = dataset.select(range(min(max_samples, len(dataset))))
        
        # Collect metrics
        perplexities = []
        bleu_scores = []
        rouge_scores = []
        accuracies = []
        inference_times = []
        
        logger.info(f"Evaluating on {len(dataset)} samples...")
        
        for idx, sample in enumerate(dataset):
            if idx % 10 == 0:
                logger.info(f"  {idx}/{len(dataset)}")
            
            instruction = sample.get('instruction', '')
            output = sample.get('output', '')
            
            if not output:
                continue
            
            # Generate model output
            import time
            start = time.time()
            
            with torch.no_grad():
                inputs = self.tokenizer(
                    instruction,
                    return_tensors="pt",
                    truncation=True,
                    max_length=256,
                ).to(self.device)
                
                outputs = self.model.generate(
                    **inputs,
                    max_length=512,
                    temperature=0.7,
                    top_p=0.9,
                    do_sample=True,
                )
                
                generated_text = self.tokenizer.decode(outputs[0], skip_special_tokens=True)
            
            inference_times.append(time.time() - start)
            
            # Extract just the generated part (not the prompt)
            generated_output = generated_text[len(instruction):].strip()
            
            # Calculate metrics
            bleu = self.bleu_score(output, generated_output)
            rouge = self.rouge_l_score(output, generated_output)
            
            bleu_scores.append(bleu)
            rouge_scores.append(rouge)
            
            # Task accuracy: simple check if output is relevant (contains key words)
            output_lower = output.lower()
            gen_lower = generated_output.lower()
            accuracy = 1.0 if any(word in gen_lower for word in output_lower.split()[:3]) else 0.0
            accuracies.append(accuracy)
        
        # Calculate perplexity on all outputs
        all_texts = [sample.get('output', '') for sample in dataset if sample.get('output', '')]
        perplexity = self.calculate_perplexity(all_texts[:min(50, len(all_texts))])
        
        # Aggregate metrics
        metrics = EvaluationMetrics(
            perplexity=perplexity,
            bleu_score=np.mean(bleu_scores) if bleu_scores else 0.0,
            rouge_l_score=np.mean(rouge_scores) if rouge_scores else 0.0,
            task_accuracy=np.mean(accuracies) if accuracies else 0.0,
            inference_time=np.mean(inference_times) if inference_times else 0.0,
        )
        
        return metrics
    
    def _get_ngrams(self, tokens: List[str], n: int) -> Dict[Tuple, int]:
        """Get n-gram counts"""
        ngrams = defaultdict(int)
        for i in range(len(tokens) - n + 1):
            ng = tuple(tokens[i:i+n])
            ngrams[ng] += 1
        return ngrams
    
    def _lcs_length(self, seq1: List[str], seq2: List[str]) -> int:
        """Calculate longest common subsequence length"""
        m, n = len(seq1), len(seq2)
        dp = [[0] * (n + 1) for _ in range(m + 1)]
        
        for i in range(1, m + 1):
            for j in range(1, n + 1):
                if seq1[i-1] == seq2[j-1]:
                    dp[i][j] = dp[i-1][j-1] + 1
                else:
                    dp[i][j] = max(dp[i-1][j], dp[i][j-1])
        
        return dp[m][n]


def log_metrics_to_json(metrics: EvaluationMetrics, output_path: str) -> None:
    """Log metrics to JSON for tracking"""
    output_path = Path(output_path)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    
    with open(output_path, 'w') as f:
        json.dump(metrics.to_dict(), f, indent=2)
    
    logger.info(f"Metrics saved to {output_path}")


def log_metrics_to_mlflow(metrics: EvaluationMetrics, run_name: str = "evaluation") -> None:
    """Log metrics to MLflow (if available)"""
    try:
        import mlflow
        
        with mlflow.start_run(run_name=run_name):
            mlflow.log_metrics(metrics.to_dict())
        
        logger.info("Metrics logged to MLflow")
    except ImportError:
        logger.warning("MLflow not installed. Skipping MLflow logging.")


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Evaluate a fine-tuned model")
    parser.add_argument("--model", required=True, help="Model name or path")
    parser.add_argument("--val-data", default="data/training/val.jsonl", help="Validation data path")
    parser.add_argument("--output-json", default="results/eval_metrics.json", help="Output JSON path")
    parser.add_argument("--max-samples", type=int, default=100, help="Max samples to evaluate")
    parser.add_argument("--mlflow", action="store_true", help="Log to MLflow")
    
    args = parser.parse_args()
    
    # Check if validation data exists
    if not Path(args.val_data).exists():
        logger.error(f"Validation data not found at {args.val_data}")
        exit(1)
    
    # Evaluate
    evaluator = ModelEvaluator(args.model)
    metrics = evaluator.evaluate_on_dataset(args.val_data, args.max_samples)
    
    # Log results
    logger.info("\n" + "="*60)
    logger.info("EVALUATION RESULTS")
    logger.info("="*60)
    for key, value in metrics.to_dict().items():
        logger.info(f"{key:.<40} {value:.4f}")
    logger.info("="*60 + "\n")
    
    # Save to JSON
    log_metrics_to_json(metrics, args.output_json)
    
    # Save to MLflow if requested
    if args.mlflow:
        log_metrics_to_mlflow(metrics)
