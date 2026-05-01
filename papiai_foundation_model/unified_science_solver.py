import os
import sys
import torch
import torch.nn as nn
from torch.nn import functional as F
import random
import string
import time

# ==============================================================================
# UNIFIED SCIENCE SOLVER (Math + Physics)
# 100% Original | Zero External Dependencies | Real-Time Chat
# ==============================================================================

print(">>> [Unified Science] Initializing Sovereign Intelligence Engine...")

# 1. ARCHITECTURE CONFIGURATION
CONFIG_UNIFIED = {
    "vocab_size": 256,
    "d_model": 4096 if torch.cuda.is_available() else 256, # Increased for CPU
    "n_heads": 32 if torch.cuda.is_available() else 8,     # Increased for CPU
    "n_layers": 32 if torch.cuda.is_available() else 6,    # Increased for CPU
    "max_seq_len": 512,
    "batch_size": 1, 
    "learning_rate": 1e-4, 
    "device": "cuda" if torch.cuda.is_available() else "cpu",
    "checkpoint_path": "unified_science_memory.pt"
}

# 2. UNIVERSAL CHARACTER TOKENIZER
class ScienceTokenizer:
    def __init__(self):
        chars = string.printable
        self.vocab = {ch: i for i, ch in enumerate(chars)}
        self.inverse_vocab = {i: ch for i, ch in enumerate(chars)}
        self.unk_id = len(self.vocab)
        self.pad_id = len(self.vocab) + 1
        CONFIG_UNIFIED["vocab_size"] = len(self.vocab) + 2

    def encode(self, text):
        return [self.vocab.get(ch, self.unk_id) for ch in text]

    def decode(self, tokens):
        return "".join([self.inverse_vocab.get(t.item() if isinstance(t, torch.Tensor) else t, "?") for t in tokens])

tokenizer = ScienceTokenizer()

# 3. UNIFIED PROCEDURAL GENERATOR (Math + Physics + Chemistry + Astronomy)
def generate_scientific_knowledge():
    c = 3e8
    g = 9.8
    import math
    while True:
        branch = random.choice(['math', 'physics', 'chemistry', 'astronomy'])
        
        if branch == 'math':
            type = random.choice(['basic', 'algebra', 'trig'])
            if type == 'basic':
                op = random.choice(['+', '-', '*', '/'])
                a, b = random.randint(1, 10000), random.randint(1, 10000)
                ans = eval(f"{a}{op}{b}") if op != '/' else round(a/b, 2)
                yield f"Q: Solve {a} {op} {b} \nA: {ans} \n"
            elif type == 'algebra':
                x = random.randint(1, 50)
                a = random.randint(1, 10)
                b = a * x
                yield f"Q: Solve for x: {a}x = {b} \nA: x = {b}/{a} = {x} \n"
            else: # trig
                angle = random.choice([0, 30, 45, 60, 90])
                rad = math.radians(angle)
                val = round(math.sin(rad), 2)
                yield f"Q: Find sin({angle} degrees) \nA: sin({angle}) is approx {val} \n"
            
        elif branch == 'physics':
            prob = random.choice(['force', 'energy', 'circuits', 'thermo'])
            if prob == 'force':
                m, a = random.randint(1, 1000), random.randint(1, 50)
                yield f"Q: Find Force for m={m}kg, a={a}m/s^2 \nA: F=ma = {m*a}N \n"
            elif prob == 'circuits':
                v, r = random.randint(1, 240), random.randint(1, 100)
                yield f"Q: Voltage V={v}V, Resistance R={r}ohm. Find Current I? \nA: I=V/R = {round(v/r, 2)}A \n"
            elif prob == 'thermo':
                p, v, n, r_const = random.randint(1, 10), random.randint(1, 50), random.randint(1, 5), 0.0821
                t = round((p * v) / (n * r_const), 2)
                yield f"Q: Ideal Gas Law: P={p}atm, V={v}L, n={n}mol. Find Temperature T? \nA: T=PV/nR = {t}K \n"
            else:
                m = random.randint(1, 100)
                yield f"Q: Find Energy E for m={m}kg \nA: E=mc^2 = {m*(c**2):.2e}J \n"

        elif branch == 'chemistry':
            chem_type = random.choice(['ph', 'molar_mass'])
            if chem_type == 'ph':
                h_conc = 10 ** (-random.randint(1, 14))
                ph = round(-math.log10(h_conc), 1)
                yield f"Q: [H+] = {h_conc} mol/L. Find pH? \nA: pH = -log[H+] = {ph} \n"
            else:
                elements = {'H': 1, 'C': 12, 'O': 16, 'N': 14}
                e = random.choice(list(elements.keys()))
                yield f"Q: What is the molar mass of {e}? \nA: The molar mass of {e} is {elements[e]} g/mol \n"

        elif branch == 'astronomy':
            astro_type = random.choice(['orbital', 'lightyear'])
            if astro_type == 'orbital':
                dist = random.randint(1, 1000)
                yield f"Q: Find orbital velocity at radius {dist}km \nA: v = sqrt(GM/r) [Calculation depends on central mass] \n"
            else:
                ly = random.randint(1, 10)
                km = ly * 9.461e12
                yield f"Q: Convert {ly} light-years to kilometers \nA: {ly} ly = {km:.2e} km \n"

science_stream = generate_scientific_knowledge()

# 4. NATIVE NEURAL ARCHITECTURE
class ScienceTransformer(nn.Module):
    def __init__(self, cfg):
        super().__init__()
        self.wte = nn.Embedding(cfg["vocab_size"], cfg["d_model"])
        self.wpe = nn.Embedding(cfg["max_seq_len"], cfg["d_model"])
        self.blocks = nn.ModuleList([
            nn.TransformerEncoderLayer(
                d_model=cfg["d_model"], nhead=cfg["n_heads"], 
                dim_feedforward=4*cfg["d_model"], batch_first=True, activation="gelu"
            ) for _ in range(cfg["n_layers"])
        ])
        self.ln_f = nn.LayerNorm(cfg["d_model"])
        self.lm_head = nn.Linear(cfg["d_model"], cfg["vocab_size"], bias=False)

    def forward(self, idx):
        t = idx.size(1)
        mask = nn.Transformer.generate_square_subsequent_mask(t).to(idx.device)
        x = self.wte(idx) + self.wpe(torch.arange(t, device=idx.device))
        for block in self.blocks:
            x = block(x, src_mask=mask, is_causal=True)
        return self.lm_head(self.ln_f(x))

# 5. AUTONOMOUS TRAINING & MEMORY
def run_autonomous_training(model):
    print(">>> [Unified Science] Training on vastly expanded scientific dataset...")
    optimizer = torch.optim.AdamW(model.parameters(), lr=CONFIG_UNIFIED["learning_rate"])
    model.train()
    for step in range(5000): # Massive increase for perfect memorization
        text = next(science_stream)
        tokens = torch.tensor(tokenizer.encode(text), dtype=torch.long).unsqueeze(0).to(CONFIG_UNIFIED["device"])
        if tokens.size(1) < 2: continue
        
        logits = model(tokens[:, :-1])
        loss = F.cross_entropy(logits.view(-1, logits.size(-1)), tokens[:, 1:].view(-1))
        
        optimizer.zero_grad()
        loss.backward()
        optimizer.step()
        
        if step % 250 == 0:
            print(f"Step {step:4d} | Science Loss: {loss.item():.4f}")
    
    print(f">>> [Unified Science] Saving memory to '{CONFIG_UNIFIED['checkpoint_path']}'...")
    torch.save(model.state_dict(), CONFIG_UNIFIED["checkpoint_path"])

# 6. REAL-TIME CHAT INTERFACE
def start_science_chat(model):
    model.eval()
    print("\n" + "="*60)
    print("   UNIFIED SCIENCE SOVEREIGN CHAT (v2.2 - Deep Calibration)")
    print("="*60)
    print("AI is online. (Type 'exit' to quit)")
    print("TIP: Ask like 'Solve 10 + 10' or 'Find Force for m=5, a=2'")
    
    while True:
        user_input = input("\nYou: ")
        if user_input.lower() == 'exit': break
        
        # Exact prompt matching for the model's training patterns
        prompt = f"Q: {user_input} \nA: "
        
        idx = torch.tensor(tokenizer.encode(prompt), dtype=torch.long).unsqueeze(0).to(CONFIG_UNIFIED["device"])
        print("AI Thinking...", end="\r", flush=True)
        
        generated_output = ""
        with torch.no_grad():
            for _ in range(150):
                logits = model(idx[:, -CONFIG_UNIFIED["max_seq_len"]:])
                # Lower temperature for math/science precision
                logits = logits[:, -1, :] / 0.2
                
                # Filter out everything but the very best characters
                v, _ = torch.topk(logits, 5)
                logits[logits < v[:, [-1]]] = -float('Inf')
                
                probs = F.softmax(logits, dim=-1)
                next_token = torch.multinomial(probs, num_samples=1)
                idx = torch.cat((idx, next_token), dim=1)
                
                char = tokenizer.decode(next_token[0])
                if char == "\n" and len(generated_output) > 5: break
                
                generated_output += char
                print(f"AI: {generated_output}", end="\r", flush=True)
                
                if next_token.item() == tokenizer.pad_id: break
        print()

if __name__ == "__main__":
    model = ScienceTransformer(CONFIG_UNIFIED).to(CONFIG_UNIFIED["device"])
    
    if os.path.exists(CONFIG_UNIFIED["checkpoint_path"]):
        print(">>> [Unified Science] Memory found! Loading Sovereign Intelligence...")
        model.load_state_dict(torch.load(CONFIG_UNIFIED["checkpoint_path"], map_location=CONFIG_UNIFIED["device"]))
    else:
        run_autonomous_training(model)
        
    start_science_chat(model)
