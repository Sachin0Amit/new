#!/usr/bin/env python3
"""
LoRA Fine-tuning Script for Sovereign Intelligence Core
Fine-tunes Mistral-7B or LLaMA-3-8B using HuggingFace PEFT/LoRA
"""

import os
import json
import logging
import torch
from pathlib import Path
from typing import Optional

import transformers
from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    TrainingArguments,
    Trainer,
    DataCollatorForLanguageModeling
)
from datasets import load_dataset
from peft import get_peft_model, LoraConfig, TaskType

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class LoRAFineTuner:
    """Fine-tunes models using LoRA"""
    
    def __init__(
        self,
        model_name: str = "mistralai/Mistral-7B",
        output_dir: str = "checkpoints/lora_finetuned",
        lora_rank: int = 16,
        lora_alpha: int = 32,
        lora_dropout: float = 0.05,
    ):
        self.model_name = model_name
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)
        
        self.lora_rank = lora_rank
        self.lora_alpha = lora_alpha
        self.lora_dropout = lora_dropout
        
        self.model = None
        self.tokenizer = None
        self.trainer = None
        
    def load_model_and_tokenizer(self) -> None:
        """Load the base model and tokenizer"""
        logger.info(f"Loading model: {self.model_name}")
        
        # Load tokenizer
        self.tokenizer = AutoTokenizer.from_pretrained(self.model_name, trust_remote_code=True)
        if self.tokenizer.pad_token is None:
            self.tokenizer.pad_token = self.tokenizer.eos_token
        
        # Load model
        self.model = AutoModelForCausalLM.from_pretrained(
            self.model_name,
            device_map="auto",
            load_in_8bit=True,  # Use 8-bit quantization to save memory
            torch_dtype=torch.float16 if torch.cuda.is_available() else torch.float32,
            trust_remote_code=True,
        )
        
        logger.info(f"Model loaded. Trainable params: {self.get_trainable_params()}")
    
    def setup_lora(self) -> None:
        """Setup LoRA adapters"""
        logger.info(f"Setting up LoRA with rank={self.lora_rank}, alpha={self.lora_alpha}")
        
        lora_config = LoraConfig(
            r=self.lora_rank,
            lora_alpha=self.lora_alpha,
            target_modules=["q_proj", "v_proj"],  # For Mistral/LLaMA
            lora_dropout=self.lora_dropout,
            bias="none",
            task_type=TaskType.CAUSAL_LM,
        )
        
        self.model = get_peft_model(self.model, lora_config)
        logger.info(f"LoRA setup complete. Trainable params: {self.get_trainable_params()}")
    
    def get_trainable_params(self) -> int:
        """Count trainable parameters"""
        return sum(p.numel() for p in self.model.parameters() if p.requires_grad)
    
    def format_conversation(self, example: dict) -> dict:
        """Format conversation into prompt-response format"""
        instruction = example.get('instruction', '').strip()
        input_text = example.get('input', '').strip()
        output_text = example.get('output', '').strip()
        
        # Build prompt
        if input_text:
            prompt = f"### Instruction:\n{instruction}\n\n### Input:\n{input_text}\n\n### Response:\n"
        else:
            prompt = f"### Instruction:\n{instruction}\n\n### Response:\n"
        
        # Full example
        text = prompt + output_text + self.tokenizer.eos_token
        
        return {"text": text}
    
    def preprocess_function(self, examples: dict) -> dict:
        """Tokenize and format training examples"""
        # Format conversations
        formatted = [self.format_conversation({"instruction": instr, "input": inp, "output": out})
                     for instr, inp, out in zip(
                         examples.get('instruction', []),
                         examples.get('input', []),
                         examples.get('output', [])
                     )]
        
        texts = [ex['text'] for ex in formatted]
        
        # Tokenize
        tokenized = self.tokenizer(
            texts,
            padding="max_length",
            max_length=512,
            truncation=True,
            return_overflowing_tokens=False,
        )
        
        # Set labels (same as input_ids for language modeling)
        tokenized['labels'] = tokenized['input_ids'].copy()
        
        return tokenized
    
    def load_and_prepare_data(self, train_path: str, val_path: str):
        """Load training and validation data"""
        logger.info(f"Loading training data from {train_path}")
        
        # Load from JSONL files
        train_dataset = load_dataset("json", data_files=train_path, split="train")
        val_dataset = load_dataset("json", data_files=val_path, split="train")
        
        logger.info(f"Train size: {len(train_dataset)}, Val size: {len(val_dataset)}")
        
        # Preprocess
        logger.info("Tokenizing datasets...")
        train_dataset = train_dataset.map(
            self.preprocess_function,
            batched=True,
            remove_columns=train_dataset.column_names,
            desc="Tokenizing train",
        )
        
        val_dataset = val_dataset.map(
            self.preprocess_function,
            batched=True,
            remove_columns=val_dataset.column_names,
            desc="Tokenizing val",
        )
        
        return train_dataset, val_dataset
    
    def train(
        self,
        train_path: str,
        val_path: str,
        num_epochs: int = 3,
        batch_size: int = 8,
        learning_rate: float = 2e-4,
        warmup_steps: int = 100,
    ) -> None:
        """Fine-tune the model"""
        
        # Load and prepare data
        train_dataset, val_dataset = self.load_and_prepare_data(train_path, val_path)
        
        # Training arguments
        training_args = TrainingArguments(
            output_dir=str(self.output_dir),
            overwrite_output_dir=True,
            num_train_epochs=num_epochs,
            per_device_train_batch_size=batch_size,
            per_device_eval_batch_size=batch_size,
            warmup_steps=warmup_steps,
            learning_rate=learning_rate,
            weight_decay=0.01,
            logging_steps=10,
            save_steps=500,
            eval_steps=500,
            save_total_limit=3,
            evaluation_strategy="steps",
            gradient_accumulation_steps=4,
            gradient_checkpointing=True,  # Memory optimization
            optim="paged_adamw_8bit",  # Memory-efficient optimizer
            bf16=torch.cuda.is_available() and torch.cuda.get_device_capability()[0] >= 8,
            tf32=True,
        )
        
        # Data collator
        data_collator = DataCollatorForLanguageModeling(
            self.tokenizer,
            mlm=False,
        )
        
        # Create trainer
        self.trainer = Trainer(
            model=self.model,
            args=training_args,
            train_dataset=train_dataset,
            eval_dataset=val_dataset,
            data_collator=data_collator,
            callbacks=[
                transformers.EarlyStoppingCallback(
                    early_stopping_patience=3,
                    early_stopping_threshold=0.0,
                ),
            ],
        )
        
        # Train
        logger.info("Starting training...")
        self.trainer.train()
        
        logger.info("Training complete!")
        
        # Save model
        self.save_model()
    
    def save_model(self) -> None:
        """Save the fine-tuned model"""
        logger.info(f"Saving model to {self.output_dir}")
        self.model.save_pretrained(self.output_dir / "final_model")
        self.tokenizer.save_pretrained(self.output_dir / "final_model")
        logger.info("Model saved successfully!")
    
    def evaluate(self) -> dict:
        """Evaluate the model"""
        if self.trainer is None:
            raise ValueError("Must train the model first")
        
        logger.info("Evaluating...")
        eval_results = self.trainer.evaluate()
        
        # Log metrics
        logger.info(f"Evaluation Results:")
        for key, value in eval_results.items():
            logger.info(f"  {key}: {value}")
        
        return eval_results


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Fine-tune a model using LoRA")
    parser.add_argument("--model", default="mistralai/Mistral-7B", help="Model name or path")
    parser.add_argument("--train-data", default="data/training/train.jsonl", help="Training data path")
    parser.add_argument("--val-data", default="data/training/val.jsonl", help="Validation data path")
    parser.add_argument("--output-dir", default="checkpoints/lora_finetuned", help="Output directory")
    parser.add_argument("--epochs", type=int, default=3, help="Number of epochs")
    parser.add_argument("--batch-size", type=int, default=8, help="Batch size")
    parser.add_argument("--lr", type=float, default=2e-4, help="Learning rate")
    parser.add_argument("--rank", type=int, default=16, help="LoRA rank")
    parser.add_argument("--alpha", type=int, default=32, help="LoRA alpha")
    
    args = parser.parse_args()
    
    # Check if training data exists
    if not os.path.exists(args.train_data):
        logger.error(f"Training data not found at {args.train_data}")
        logger.info("Run: python scripts/download_datasets.py")
        exit(1)
    
    # Initialize fine-tuner
    finetuner = LoRAFineTuner(
        model_name=args.model,
        output_dir=args.output_dir,
        lora_rank=args.rank,
        lora_alpha=args.alpha,
    )
    
    # Load model and setup LoRA
    finetuner.load_model_and_tokenizer()
    finetuner.setup_lora()
    
    # Train
    finetuner.train(
        train_path=args.train_data,
        val_path=args.val_data,
        num_epochs=args.epochs,
        batch_size=args.batch_size,
        learning_rate=args.lr,
    )
    
    # Evaluate
    finetuner.evaluate()
