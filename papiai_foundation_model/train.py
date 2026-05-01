"""
Sovereign Intelligence Core - Training Loop
===========================================
Trains the PapiAi Foundation Model from scratch using PyTorch.
"""

import torch
from torch.utils.data import DataLoader
from model import PapiAiModel, PapiAiConfig
from dataset import CodeDataset, SovereignTokenizer
import time
import os

# Hyperparameters
BATCH_SIZE = 16
SEQ_LEN = 256
LEARNING_RATE = 3e-4
MAX_ITERATIONS = 500
EVAL_INTERVAL = 50
DEVICE = 'cuda' if torch.cuda.is_available() else 'cpu'

def train():
    print("==================================================")
    print("   PapiAi | Foundation Model Compiler (From Scratch)")
    print("==================================================")
    print(f"[*] Booting Compiler on device: {DEVICE}")

    # 1. Initialize Dataset & Tokenizer
    tokenizer = SovereignTokenizer()
    dataset = CodeDataset("training_data.txt", tokenizer, seq_len=SEQ_LEN)
    
    # Simple data loader
    dataloader = DataLoader(dataset, batch_size=BATCH_SIZE, shuffle=True, drop_last=True)
    
    # 2. Initialize Model
    print("\n[*] Initializing Neural Matrix Architecture...")
    config = PapiAiConfig(
        vocab_size=tokenizer.vocab_size,
        d_model=256,   # Scaled down for rapid local testing
        n_heads=8,
        n_layers=6,
        max_seq_len=SEQ_LEN
    )
    
    model = PapiAiModel(config).to(DEVICE)
    num_params = sum(p.numel() for p in model.parameters())
    print(f"[*] Model initialized with {num_params:,} parameters.")

    # 3. Setup Optimizer
    optimizer = torch.optim.AdamW(model.parameters(), lr=LEARNING_RATE)

    # 4. Training Loop
    print("\n>>> BEGINNING COMPILATION LOOP <<<")
    
    model.train()
    step = 0
    start_time = time.time()
    
    for epoch in range(1): # Keep epoch low for demo
        for x, y in dataloader:
            x, y = x.to(DEVICE), y.to(DEVICE)
            
            # Forward pass
            logits, loss = model(x, targets=y)
            
            # Backward pass
            optimizer.zero_grad(set_to_none=True)
            loss.backward()
            torch.nn.utils.clip_grad_norm_(model.parameters(), 1.0)
            optimizer.step()
            
            if step % EVAL_INTERVAL == 0:
                dt = time.time() - start_time
                print(f"[Step {step}] Loss: {loss.item():.4f} | Time: {dt:.2f}s")
                
            step += 1
            if step >= MAX_ITERATIONS:
                break
                
    # 5. Save the compiled model
    print("\n[*] Compilation Complete. Saving Sovereign Weights...")
    torch.save({
        'model_state': model.state_dict(),
        'vocab': tokenizer.vocab,
        'inverse_vocab': tokenizer.inverse_vocab
    }, 'papiai_foundation.pt')
    print("[*] Weights saved to 'papiai_foundation.pt'.")

if __name__ == "__main__":
    train()
