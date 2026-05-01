import torch
from model import PapiAiConfig, PapiAiModel

# ==============================================================================
# APPROACH 1: LEGITIMATELY USING OPEN-SOURCE (Hugging Face)
# We respect the open-source licenses by using their official libraries to download
# and process data, rather than scraping or copying without attribution.
# ==============================================================================
print("--- [Approach 1] Legitimate Open-Source Integration ---")
try:
    from datasets import load_dataset
    from transformers import AutoTokenizer, AutoModelForCausalLM

    print("1. Downloading a legitimate Open-Source Dataset (e.g., wikitext)...")
    # This automatically handles licensing, attribution, and formatting
    dataset = load_dataset("wikitext", "wikitext-2-raw-v1", split="train[:1%]")
    print(f"   Successfully loaded {len(dataset)} samples from Wikitext-2.")

    print("2. Downloading a legitimate Open-Source Tokenizer (e.g., GPT-2)...")
    tokenizer = AutoTokenizer.from_pretrained("gpt2")
    tokenizer.pad_token = tokenizer.eos_token
    print("   Tokenizer loaded successfully.")

    # You could also load an open-source model directly if you wanted to use it out of the box:
    # print("3. Loading pre-trained Open-Source Model (GPT-2)...")
    # hf_model = AutoModelForCausalLM.from_pretrained("gpt2")

except ImportError:
    print("Please install 'transformers' and 'datasets' using: pip install transformers datasets")
    exit(1)


# ==============================================================================
# APPROACH 2: WRITING ORIGINAL CODE FROM SCRATCH
# We take the data processed by the open-source tokenizer and feed it into YOUR 
# original, custom-built AI architecture (PapiAiModel).
# ==============================================================================
print("\n--- [Approach 2] Original Custom AI Training Pipeline ---")
print("1. Initializing your custom from-scratch model (PapiAiModel)...")

# Setup configuration for your original model
config = PapiAiConfig(
    vocab_size=tokenizer.vocab_size, 
    d_model=256,   # Smaller size for demonstration
    n_heads=4,     
    n_layers=4,    
    max_seq_len=128
)

# Instantiate the custom model from model.py
custom_model = PapiAiModel(config)
print("   Custom model initialized successfully with", sum(p.numel() for p in custom_model.parameters()), "parameters.")

print("2. Preparing Open-Source Data for Custom Training...")
# Tokenize a sample text
sample_text = dataset[10]['text'] if len(dataset[10]['text'].strip()) > 0 else "Artificial Intelligence is fascinating."
print(f"   Sample Text: '{sample_text}'")

inputs = tokenizer(sample_text, return_tensors="pt", max_length=config.max_seq_len, truncation=True)
input_ids = inputs["input_ids"]

print(f"   Tokenized Input Shape: {input_ids.shape}")

print("3. Running a forward pass through your original architecture...")
# Set model to evaluation mode for demonstration
custom_model.eval()

with torch.no_grad():
    logits, _ = custom_model(input_ids)
    
print(f"   Output Logits Shape: {logits.shape}")

print("\nSuccess! You have combined legitimate open-source data processing with your original AI code architecture.")
