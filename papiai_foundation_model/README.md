# Sovereign Intelligence Core - Foundation Model

This directory contains the ultimate realization of your request: **An AI system built entirely from scratch, from zero to high-level.**

It does **not** rely on external AI APIs. It does **not** rely on pre-built `transformers` models. It is a raw, pure PyTorch implementation of a Causal Language Model (Transformer).

## Architecture Details
- **`model.py`**: A custom GPT-style Decoder-only Transformer. Includes Multi-Head Attention, Feed-Forward Networks, and custom weight initialization.
- **`dataset.py`**: A custom sub-word Tokenizer and Dataset loader built from absolute zero. It parses raw text, builds a vocabulary mathematically based on word frequency, and converts text into tensors.
- **`train.py`**: The neural training loop. It initializes the empty neural matrix and trains it on your local data using Backpropagation and AdamW optimization.
- **`inference.py`**: The neural execution environment. It boots the locally trained `.pt` weights and generates text tokens step-by-step.

## How to Compile Your Own AI from Scratch:

**1. Provide Training Data**
By default, the script looks for a `training_data.txt` file. You can paste any text, code, or datasets into this file. If it doesn't exist, the system will auto-generate a dummy dataset to prove the pipeline works.

**2. Train the Network**
Run the training loop to build the neural weights from absolute zero:
```bash
python3 train.py
```
*(This will output `papiai_foundation.pt` once complete.)*

**3. Run Inference**
Talk to the model you just created:
```bash
python3 inference.py
```

This is your very own, 100% sovereign foundation model, created entirely from first principles.
