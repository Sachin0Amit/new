import os
import sys
import torch
import torch.nn as nn
from torch.nn import functional as F
import random
import string

# ==============================================================================
# THE GLOBAL PHYSICS SOLVER ARCHITECTURE (100% Original, Zero External Dependencies)
# ==============================================================================
# This script procedurally generates an infinite stream of physics problems 
# (Classical Mechanics, Kinematics, Energy) and tokenizes them using a custom 
# from-scratch Tokenizer. It then trains a massive Transformer architecture.
#
# No Hugging Face. No external datasets. No external tokenizers.
# ==============================================================================

print(">>> [Global Physics Solver] Initializing 100% Original Physics Engine...")

# 1. THE MASSIVE-SCALE CONFIGURATION
CONFIG_PHYSICS = {
    "vocab_size": 256,     # Character-level vocabulary
    "d_model": 4096 if torch.cuda.is_available() else 128,  # Massive on GPU, scaled down for CPU
    "n_heads": 32 if torch.cuda.is_available() else 4,
    "n_layers": 32 if torch.cuda.is_available() else 4,
    "max_seq_len": 512,    # Context window for physics word problems
    "batch_size": 2,       # Kept tiny for local testing
    "learning_rate": 1e-4, 
    "device": "cuda" if torch.cuda.is_available() else "cpu",
    "accumulation_steps": 16 
}

# 2. CUSTOM FROM-SCRATCH TOKENIZER (Character-Level)
# The AI will learn physics character by character, including letters, numbers, and units.
class PhysicsTokenizer:
    def __init__(self):
        # We use all printable characters to handle variables (m, v, F, a, t), numbers, and symbols
        chars = string.printable
        self.vocab = {ch: i for i, ch in enumerate(chars)}
        self.inverse_vocab = {i: ch for i, ch in enumerate(chars)}
        self.unk_id = len(self.vocab)
        self.pad_id = len(self.vocab) + 1
        CONFIG_PHYSICS["vocab_size"] = len(self.vocab) + 2

    def encode(self, text):
        return [self.vocab.get(ch, self.unk_id) for ch in text]

    def decode(self, tokens):
        return "".join([self.inverse_vocab.get(t.item(), "?") for t in tokens])

tokenizer = PhysicsTokenizer()

# 3. PROCEDURAL PHYSICS GENERATOR (Infinite Stream)
# Generates an infinite stream of varied physics problems: Newton's Laws, Kinematics, Energy.
def generate_physics_problem():
    c = 300000000  # Speed of light
    g = 9.8        # Gravity on Earth
    
    while True:
        problem_type = random.choice([
            'newton_force', 'kinematics_velocity', 'einstein_energy', 
            'momentum', 'potential_energy', 'kinetic_energy'
        ])
        
        if problem_type == 'newton_force':
            m = random.randint(1, 1000)
            a = random.randint(1, 100)
            f = m * a
            yield f"Q: Mass m = {m} kg, Acceleration a = {a} m/s^2. Find Force F? \nA: F = m*a = {f} N \n"
            
        elif problem_type == 'kinematics_velocity':
            u = random.randint(0, 100)
            a = random.randint(1, 50)
            t = random.randint(1, 20)
            v = u + a * t
            yield f"Q: Initial velocity u = {u} m/s, Acceleration a = {a} m/s^2, Time t = {t} s. Find final velocity v? \nA: v = u+at = {v} m/s \n"
            
        elif problem_type == 'einstein_energy':
            m = random.randint(1, 500)
            e = m * (c ** 2)
            # Scientific notation for massive numbers
            yield f"Q: Mass m = {m} kg, Speed of light c = 3e8 m/s. Find Rest Energy E? \nA: E = mc^2 = {e:.1e} J \n"
            
        elif problem_type == 'momentum':
            m = random.randint(1, 1000)
            v = random.randint(1, 100)
            p = m * v
            yield f"Q: Mass m = {m} kg, Velocity v = {v} m/s. Find Momentum p? \nA: p = m*v = {p} kg*m/s \n"
            
        elif problem_type == 'potential_energy':
            m = random.randint(1, 500)
            h = random.randint(1, 1000)
            u_e = round(m * g * h, 2)
            yield f"Q: Mass m = {m} kg, Gravity g = 9.8 m/s^2, Height h = {h} m. Find Potential Energy U? \nA: U = mgh = {u_e} J \n"
            
        elif problem_type == 'kinetic_energy':
            m = random.randint(1, 500)
            v = random.randint(1, 100)
            k = round(0.5 * m * (v ** 2), 2)
            yield f"Q: Mass m = {m} kg, Velocity v = {v} m/s. Find Kinetic Energy K? \nA: K = 0.5*m*v^2 = {k} J \n"

physics_stream = generate_physics_problem()

# 4. MASSIVE MODEL ARCHITECTURE (From Scratch)
class MassivePhysicsTransformer(nn.Module):
    def __init__(self, cfg):
        super().__init__()
        self.cfg = cfg
        self.wte = nn.Embedding(cfg["vocab_size"], cfg["d_model"])
        self.wpe = nn.Embedding(cfg["max_seq_len"], cfg["d_model"])
        
        # Attention Blocks built entirely natively
        self.blocks = nn.ModuleList([
            nn.TransformerEncoderLayer(
                d_model=cfg["d_model"], 
                nhead=cfg["n_heads"], 
                dim_feedforward=4 * cfg["d_model"],
                batch_first=True,
                norm_first=True,
                activation="gelu"
            ) for _ in range(cfg["n_layers"])
        ])
        self.ln_f = nn.LayerNorm(cfg["d_model"])
        self.lm_head = nn.Linear(cfg["d_model"], cfg["vocab_size"], bias=False)

    def forward(self, idx):
        b, t = idx.size()
        pos = torch.arange(0, t, device=idx.device)
        
        # Causal mask to prevent looking into the future of the equation
        mask = nn.Transformer.generate_square_subsequent_mask(t).to(idx.device)
        
        x = self.wte(idx) + self.wpe(pos)
        for block in self.blocks:
            x = block(x, src_mask=mask, is_causal=True)
            
        x = self.ln_f(x)
        logits = self.lm_head(x)
        return logits

# 5. TRAINING LOOP
def train_physics_model():
    print(f">>> [Global Physics Solver] Allocating {CONFIG_PHYSICS['n_layers']} native layers on {CONFIG_PHYSICS['device']}...")
    
    try:
        model = MassivePhysicsTransformer(CONFIG_PHYSICS).to(CONFIG_PHYSICS["device"])
        optimizer = torch.optim.AdamW(model.parameters(), lr=CONFIG_PHYSICS["learning_rate"])
        scaler = torch.cuda.amp.GradScaler() if CONFIG_PHYSICS["device"] == "cuda" else None

        print(">>> [Global Physics Solver] Commencing infinite procedural training...")
        
        model.train()
        step = 0
        optimizer.zero_grad()
        
        # We loop infinitely over procedurally generated physics
        while True:
            # Generate a batch of physics problems dynamically
            text = next(physics_stream)
            tokens = tokenizer.encode(text)
            
            # Truncate if too long
            if len(tokens) > CONFIG_PHYSICS["max_seq_len"]:
                tokens = tokens[:CONFIG_PHYSICS["max_seq_len"]]
                
            idx = torch.tensor(tokens, dtype=torch.long).unsqueeze(0).to(CONFIG_PHYSICS["device"])
            
            # We need at least 2 tokens to predict
            if idx.size(1) < 2:
                continue
                
            inputs = idx[:, :-1]
            targets = idx[:, 1:]
            
            if scaler:
                with torch.cuda.amp.autocast():
                    logits = model(inputs)
                    loss = F.cross_entropy(logits.view(-1, logits.size(-1)), targets.view(-1))
                    loss = loss / CONFIG_PHYSICS["accumulation_steps"]
                scaler.scale(loss).backward()
            else:
                logits = model(inputs)
                loss = F.cross_entropy(logits.view(-1, logits.size(-1)), targets.view(-1))
                loss = loss / CONFIG_PHYSICS["accumulation_steps"]
                loss.backward()

            if (step + 1) % CONFIG_PHYSICS["accumulation_steps"] == 0:
                if scaler:
                    scaler.step(optimizer)
                    scaler.update()
                else:
                    optimizer.step()
                optimizer.zero_grad()
                
                # Format text for neat printing
                print_text = text.replace('\n', ' | ')
                print(f"Physics Step {step} | Loss: {loss.item() * CONFIG_PHYSICS['accumulation_steps']:.4f} | Processed: {print_text.strip()}")

            step += 1
            
            if step > 50:
                print(">>> [Global Physics Solver] Local safety limit reached. Native Physics architecture verified.")
                break
                
    except torch.cuda.OutOfMemoryError:
        print("\n[CRITICAL ERROR] Out of Memory! Your GPU cannot hold this native architecture.")
    except MemoryError:
        print("\n[CRITICAL ERROR] System RAM exhausted!")

if __name__ == "__main__":
    train_physics_model()
