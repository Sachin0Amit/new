"""
Sovereign Intelligence Core - Custom Dataset & Tokenizer
========================================================
A completely custom tokenizer and data loader built from scratch.
"""

import os
import torch
from torch.utils.data import Dataset
import re

class SovereignTokenizer:
    def __init__(self):
        self.vocab = {}
        self.inverse_vocab = {}
        # Start with some basic tokens
        self.add_token("<pad>")
        self.add_token("<|unk|>")
        self.add_token("<|im_start|>")
        self.add_token("<|im_end|>")
        self.add_token("\n")
        self.add_token(" ")

    def add_token(self, token):
        if token not in self.vocab:
            idx = len(self.vocab)
            self.vocab[token] = idx
            self.inverse_vocab[idx] = token

    def build_vocab(self, text, max_vocab_size=50000):
        """
        Builds a basic word-level vocabulary from raw text.
        In a production environment, you would use BPE (Byte-Pair Encoding).
        """
        print("[*] Building custom Sovereign Vocabulary...")
        # Split by words and punctuation
        tokens = re.findall(r"\w+|[^\w\s]", text)
        
        # Count frequencies
        freqs = {}
        for t in tokens:
            freqs[t] = freqs.get(t, 0) + 1
            
        # Sort by frequency
        sorted_tokens = sorted(freqs.items(), key=lambda x: x[1], reverse=True)
        
        for token, _ in sorted_tokens[:max_vocab_size]:
            self.add_token(token)
            
        print(f"[*] Vocabulary built. Size: {len(self.vocab)} tokens.")

    def encode(self, text):
        # We try to match known tokens
        tokens = re.findall(r"\w+|[^\w\s]|\s+", text)
        encoded = []
        for t in tokens:
            if t in self.vocab:
                encoded.append(self.vocab[t])
            else:
                encoded.append(self.vocab["<|unk|>"])
        return encoded

    def decode(self, indices):
        return "".join([self.inverse_vocab.get(idx, "") for idx in indices])

    @property
    def vocab_size(self):
        return len(self.vocab)


class CodeDataset(Dataset):
    def __init__(self, data_path, tokenizer, seq_len=256):
        self.tokenizer = tokenizer
        self.seq_len = seq_len
        
        if not os.path.exists(data_path):
            print(f"[!] Warning: Data file {data_path} not found. Creating dummy dataset.")
            text = "def hello_world():\n    print('Welcome to PapiAi Sovereign Core!')\n\n" * 1000
            with open(data_path, "w") as f:
                f.write(text)
                
        with open(data_path, "r", encoding="utf-8") as f:
            raw_text = f.read()
            
        self.tokenizer.build_vocab(raw_text)
        print("[*] Encoding dataset into tensor matrix...")
        self.data = torch.tensor(self.tokenizer.encode(raw_text), dtype=torch.long)
        print(f"[*] Total dataset tokens: {len(self.data)}")

    def __len__(self):
        return len(self.data) - self.seq_len

    def __getitem__(self, idx):
        chunk = self.data[idx:idx + self.seq_len + 1]
        x = chunk[:-1]
        y = chunk[1:]
        return x, y
