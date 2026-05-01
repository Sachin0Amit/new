"""
Sovereign Intelligence Core - Inference Engine
============================================
Loads the from-scratch compiled PapiAi model and generates tokens.
"""

import torch
import sys
import os
from model import PapiAiModel, PapiAiConfig
from dataset import SovereignTokenizer

def generate():
    print("==================================================")
    print("   PapiAi | Foundation Inference Engine")
    print("==================================================")
    
    if not os.path.exists('papiai_foundation.pt'):
        print("[!] Neural Weights not found. Please run 'python3 train.py' first.")
        sys.exit(1)

    print("[*] Loading Sovereign Neural Matrix...")
    device = 'cuda' if torch.cuda.is_available() else 'cpu'
    
    # Load state
    state = torch.load('papiai_foundation.pt', map_location=device)
    
    # Reconstruct tokenizer
    tokenizer = SovereignTokenizer()
    tokenizer.vocab = state['vocab']
    tokenizer.inverse_vocab = state['inverse_vocab']
    
    # Reconstruct model
    config = PapiAiConfig(
        vocab_size=tokenizer.vocab_size,
        d_model=256,
        n_heads=8,
        n_layers=6,
        max_seq_len=256
    )
    
    model = PapiAiModel(config).to(device)
    model.load_state_dict(state['model_state'])
    model.eval()
    
    print("[+] Core Online. Type your prompt below (or 'exit' to quit).")
    
    while True:
        try:
            prompt = input("\nUser >> ")
            if prompt.lower() in ['exit', 'quit']:
                break
                
            # Encode prompt
            encoded = tokenizer.encode(prompt)
            if not encoded:
                continue
                
            idx = torch.tensor([encoded], dtype=torch.long).to(device)
            
            # Generate
            out_idx = model.generate(idx, max_new_tokens=100, temperature=0.8, top_k=10)
            
            # Decode
            response = tokenizer.decode(out_idx[0].tolist())
            print(f"\nPapiAi >> {response}")
            
        except KeyboardInterrupt:
            break

if __name__ == "__main__":
    generate()
