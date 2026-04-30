#!/usr/bin/env python3
"""
Sovereign Intelligence Core — Distributed Multi-GPU Fine-Tuning
Uses DeepSpeed ZeRO-3 + LoRA for multi-GPU training across nodes.

Usage:
  # Single node, multi-GPU
  torchrun --nproc_per_node=4 scripts/distributed_train.py \
      --model mistralai/Mistral-7B-v0.1 \
      --data data/training/train.jsonl \
      --output checkpoints/distributed_lora

  # Multi-node (launch on each node)
  torchrun --nnodes=2 --nproc_per_node=4 \
      --rdzv_backend=c10d --rdzv_endpoint=master:29500 \
      scripts/distributed_train.py --model meta-llama/Llama-3-8B
"""

import argparse
import json
import os
import sys
from pathlib import Path

def parse_args():
    p = argparse.ArgumentParser(description="Distributed Multi-GPU LoRA Fine-Tuning")
    p.add_argument("--model", type=str, default="mistralai/Mistral-7B-v0.1",
                    help="HuggingFace model ID or local path")
    p.add_argument("--data", type=str, default="data/training/train.jsonl",
                    help="Path to training data (JSONL)")
    p.add_argument("--val_data", type=str, default="data/training/val.jsonl",
                    help="Path to validation data (JSONL)")
    p.add_argument("--output", type=str, default="checkpoints/distributed_lora",
                    help="Output directory for checkpoints")
    p.add_argument("--epochs", type=int, default=3)
    p.add_argument("--batch_size", type=int, default=4,
                    help="Per-device batch size")
    p.add_argument("--gradient_accumulation", type=int, default=4)
    p.add_argument("--lr", type=float, default=2e-4)
    p.add_argument("--lora_rank", type=int, default=16)
    p.add_argument("--lora_alpha", type=int, default=32)
    p.add_argument("--lora_dropout", type=float, default=0.05)
    p.add_argument("--max_length", type=int, default=2048)
    p.add_argument("--warmup_ratio", type=float, default=0.03)
    p.add_argument("--save_steps", type=int, default=500)
    p.add_argument("--deepspeed_stage", type=int, default=3, choices=[2, 3],
                    help="DeepSpeed ZeRO stage (2 or 3)")
    p.add_argument("--fp16", action="store_true", default=True)
    p.add_argument("--bf16", action="store_true", default=False)
    return p.parse_args()


def generate_deepspeed_config(args):
    """Generate DeepSpeed ZeRO config based on selected stage."""
    config = {
        "train_batch_size": "auto",
        "train_micro_batch_size_per_gpu": args.batch_size,
        "gradient_accumulation_steps": args.gradient_accumulation,
        "gradient_clipping": 1.0,
        "steps_per_print": 50,
        "optimizer": {
            "type": "AdamW",
            "params": {
                "lr": args.lr,
                "betas": [0.9, 0.95],
                "eps": 1e-8,
                "weight_decay": 0.01
            }
        },
        "scheduler": {
            "type": "WarmupDecayLR",
            "params": {
                "warmup_min_lr": 0,
                "warmup_max_lr": args.lr,
                "warmup_num_steps": "auto",
                "total_num_steps": "auto"
            }
        },
        "fp16": {"enabled": args.fp16 and not args.bf16},
        "bf16": {"enabled": args.bf16},
        "zero_optimization": {
            "stage": args.deepspeed_stage,
            "offload_optimizer": {"device": "cpu", "pin_memory": True},
            "overlap_comm": True,
            "contiguous_gradients": True,
            "reduce_bucket_size": 5e7,
            "stage3_prefetch_bucket_size": 5e7,
            "stage3_param_persistence_threshold": 1e5,
            "gather_16bit_weights_on_model_save": True,
        },
        "activation_checkpointing": {
            "partition_activations": True,
            "contiguous_memory_optimization": True,
            "cpu_checkpointing": True
        },
        "wall_clock_breakdown": False
    }
    return config


def main():
    args = parse_args()

    # Validate data paths
    if not Path(args.data).exists():
        print(f"❌ Training data not found: {args.data}")
        print("   Run: make download-data")
        sys.exit(1)

    os.makedirs(args.output, exist_ok=True)

    # Write DeepSpeed config
    ds_config = generate_deepspeed_config(args)
    ds_config_path = os.path.join(args.output, "ds_config.json")
    with open(ds_config_path, "w") as f:
        json.dump(ds_config, f, indent=2)
    print(f"📝 DeepSpeed config written to {ds_config_path}")

    # Check for required packages
    try:
        import torch
        import transformers
        from peft import LoraConfig, get_peft_model, TaskType
        from datasets import load_dataset
        print(f"✅ PyTorch {torch.__version__} | Transformers {transformers.__version__}")
        print(f"✅ GPUs available: {torch.cuda.device_count()}")
    except ImportError as e:
        print(f"❌ Missing dependency: {e}")
        print("   Install: pip install torch transformers peft datasets deepspeed")
        sys.exit(1)

    # Load tokenizer & model
    print(f"\n🔄 Loading model: {args.model}")
    tokenizer = transformers.AutoTokenizer.from_pretrained(args.model, trust_remote_code=True)
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token

    model = transformers.AutoModelForCausalLM.from_pretrained(
        args.model,
        torch_dtype=torch.bfloat16 if args.bf16 else torch.float16,
        trust_remote_code=True,
    )

    # Apply LoRA
    lora_config = LoraConfig(
        task_type=TaskType.CAUSAL_LM,
        r=args.lora_rank,
        lora_alpha=args.lora_alpha,
        lora_dropout=args.lora_dropout,
        target_modules=["q_proj", "k_proj", "v_proj", "o_proj", "gate_proj", "up_proj", "down_proj"],
        bias="none",
    )
    model = get_peft_model(model, lora_config)
    trainable, total = model.get_nb_trainable_parameters()
    print(f"✅ LoRA applied: {trainable:,} / {total:,} params trainable ({100*trainable/total:.2f}%)")

    # Load dataset
    print(f"📂 Loading dataset: {args.data}")
    dataset = load_dataset("json", data_files={"train": args.data, "validation": args.val_data})

    def tokenize(examples):
        texts = []
        for item in zip(examples.get("instruction", [""]*len(examples["output"])),
                        examples.get("input", [""]*len(examples["output"])),
                        examples["output"]):
            inst, inp, out = item
            prompt = f"### Instruction:\n{inst}"
            if inp:
                prompt += f"\n### Input:\n{inp}"
            prompt += f"\n### Response:\n{out}"
            texts.append(prompt)
        return tokenizer(texts, truncation=True, max_length=args.max_length, padding="max_length")

    tokenized = dataset.map(tokenize, batched=True, remove_columns=dataset["train"].column_names)

    # Training arguments
    training_args = transformers.TrainingArguments(
        output_dir=args.output,
        num_train_epochs=args.epochs,
        per_device_train_batch_size=args.batch_size,
        per_device_eval_batch_size=args.batch_size,
        gradient_accumulation_steps=args.gradient_accumulation,
        learning_rate=args.lr,
        warmup_ratio=args.warmup_ratio,
        lr_scheduler_type="cosine",
        logging_steps=10,
        save_steps=args.save_steps,
        eval_strategy="steps",
        eval_steps=args.save_steps,
        save_total_limit=3,
        load_best_model_at_end=True,
        deepspeed=ds_config_path,
        fp16=args.fp16 and not args.bf16,
        bf16=args.bf16,
        report_to="none",
        dataloader_num_workers=4,
        gradient_checkpointing=True,
    )

    # Train
    trainer = transformers.Trainer(
        model=model,
        args=training_args,
        train_dataset=tokenized["train"],
        eval_dataset=tokenized["validation"],
        data_collator=transformers.DataCollatorForLanguageModeling(tokenizer, mlm=False),
    )

    print(f"\n🚀 Starting distributed training on {torch.cuda.device_count()} GPU(s)...")
    trainer.train()

    # Save final model
    final_path = os.path.join(args.output, "final_model")
    trainer.save_model(final_path)
    tokenizer.save_pretrained(final_path)
    print(f"\n✅ Training complete. Model saved to {final_path}")


if __name__ == "__main__":
    main()
