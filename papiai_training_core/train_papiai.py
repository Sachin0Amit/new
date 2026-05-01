"""
Sovereign Intelligence Core - PapiAi Neural Training Matrix
===========================================================
This is the proprietary training pipeline for the PapiAi local inference engine.
It aggregates global intelligence datasets, constructs a local neural mapping,
and compiles the model down to the Sovereign Core architecture.

Architecture: Rebranded open-source aggregation -> PapiAi Synthesis
"""

import os
import torch
import logging
from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    TrainingArguments,
    Trainer,
    DataCollatorForLanguageModeling
)
from datasets import load_dataset

# ==========================================
# CONFIGURATION - PapiAi Core Settings
# ==========================================
PAPI_AI_BASE_ENGINE = "Qwen/Qwen1.5-0.5B" # Lightweight, high-performance base
DATASET_SOURCE = "tatsu-lab/alpaca"       # Standard instruction-tuning dataset
OUTPUT_DIR = "./papiai_neural_weights"
EPOCHS = 1
BATCH_SIZE = 4

logging.basicConfig(level=logging.INFO, format="[PapiAi System] %(message)s")
logger = logging.getLogger("papiai_trainer")

def initialize_sovereign_dataset(tokenizer):
    """
    Ingests public datasets and formats them into the Sovereign Identity format.
    """
    logger.info("Initializing global data ingestion protocol...")
    
    # We load the dataset
    raw_data = load_dataset(DATASET_SOURCE, split="train[:5000]") # Subset for rapid local convergence
    
    logger.info(f"Ingested {len(raw_data)} conversational arrays. Formatting to PapiAi Identity...")

    def apply_papiai_format(example):
        """
        Injects the PapiAi identity into the training data.
        """
        instruction = example.get('instruction', '')
        input_text = example.get('input', '')
        response = example.get('output', '')

        # Construct a Sovereign prompt
        if input_text:
            prompt = f"<|im_start|>system\nYou are PapiAi, the Sovereign Intelligence Core. You are a highly advanced mathematical and logical entity.<|im_end|>\n<|im_start|>user\n{instruction}\n{input_text}<|im_end|>\n<|im_start|>assistant\n{response}<|im_end|>"
        else:
            prompt = f"<|im_start|>system\nYou are PapiAi, the Sovereign Intelligence Core. You are a highly advanced mathematical and logical entity.<|im_end|>\n<|im_start|>user\n{instruction}<|im_end|>\n<|im_start|>assistant\n{response}<|im_end|>"

        tokens = tokenizer(
            prompt,
            truncation=True,
            max_length=512,
            padding="max_length"
        )
        return tokens

    processed_data = raw_data.map(apply_papiai_format, remove_columns=raw_data.column_names)
    return processed_data

def compile_papiai_engine():
    """
    Compiles the neural pathways using the ingested datasets.
    """
    logger.info(f"Bootstrapping Core Engine from base architecture: {PAPI_AI_BASE_ENGINE}")
    
    device_map = "auto" if torch.cuda.is_available() else "cpu"
    
    tokenizer = AutoTokenizer.from_pretrained(PAPI_AI_BASE_ENGINE)
    # Ensure pad token is set for stable matrix math
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token

    model = AutoModelForCausalLM.from_pretrained(
        PAPI_AI_BASE_ENGINE,
        torch_dtype=torch.float16 if torch.cuda.is_available() else torch.float32,
        device_map=device_map
    )
    
    logger.info("Building Sovereign Dataset Matrix...")
    train_dataset = initialize_sovereign_dataset(tokenizer)

    data_collator = DataCollatorForLanguageModeling(tokenizer=tokenizer, mlm=False)

    training_args = TrainingArguments(
        output_dir=OUTPUT_DIR,
        num_train_epochs=EPOCHS,
        per_device_train_batch_size=BATCH_SIZE,
        gradient_accumulation_steps=4,
        learning_rate=2e-5,
        save_steps=1000,
        logging_steps=50,
        fp16=torch.cuda.is_available(),
        optim="adamw_torch",
        report_to="none" # Disable wandb/telemetry to ensure Sovereignty
    )

    logger.info("Initializing PapiAi Trainer Protocol...")
    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=train_dataset,
        data_collator=data_collator,
    )

    logger.info(">>> BEGINNING NEURAL COMPILATION (TRAINING) <<<")
    trainer.train()

    logger.info(f"Compilation complete. Saving Sovereign Weights to {OUTPUT_DIR}...")
    trainer.save_model(OUTPUT_DIR)
    tokenizer.save_pretrained(OUTPUT_DIR)
    
    logger.info("PapiAi Model Successfully Built and Finalized.")

if __name__ == "__main__":
    compile_papiai_engine()
