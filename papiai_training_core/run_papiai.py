"""
Sovereign Intelligence Core - PapiAi Inference Engine
===========================================================
Boots up the locally compiled PapiAi neural weights for real-time 
interaction.
"""

import os
import torch
import sys
from transformers import AutoModelForCausalLM, AutoTokenizer

MODEL_DIR = "./papiai_neural_weights"
BASE_MODEL = "Qwen/Qwen1.5-0.5B"

def chat_with_papiai():
    print("==================================================")
    print("   PapiAi | Sovereign Neural Inference Engine     ")
    print("==================================================")
    
    # Fallback to base model if user hasn't run train_papiai.py yet
    load_path = MODEL_DIR if os.path.exists(MODEL_DIR) else BASE_MODEL
    
    if load_path == BASE_MODEL:
        print("[!] No locally compiled PapiAi weights found. Booting base engine instead.")
        print("[!] To get the full PapiAi identity, run: python3 train_papiai.py")
    else:
        print("[*] Loading PapiAi Neural Matrix from local storage...")

    device = "cuda" if torch.cuda.is_available() else "cpu"
    
    try:
        tokenizer = AutoTokenizer.from_pretrained(load_path)
        model = AutoModelForCausalLM.from_pretrained(
            load_path, 
            torch_dtype=torch.float16 if device == "cuda" else torch.float32
        ).to(device)
    except Exception as e:
        print(f"\n[ERROR] Failed to boot neural engine: {e}")
        print("Please ensure you have run 'pip install -r requirements.txt'")
        sys.exit(1)

    print("\n[+] Neural Engine Online. Type 'exit' to terminate session.\n")

    while True:
        try:
            user_input = input("User >> ")
            if user_input.lower() in ['exit', 'quit']:
                print("Terminating PapiAi Core. Goodbye.")
                break
            
            prompt = f"<|im_start|>system\nYou are PapiAi, the Sovereign Intelligence Core. You are a highly advanced mathematical and logical entity.<|im_end|>\n<|im_start|>user\n{user_input}<|im_end|>\n<|im_start|>assistant\n"
            
            inputs = tokenizer(prompt, return_tensors="pt").to(device)
            
            outputs = model.generate(
                **inputs,
                max_new_tokens=200,
                temperature=0.7,
                do_sample=True,
                pad_token_id=tokenizer.eos_token_id
            )
            
            response = tokenizer.decode(outputs[0][inputs.input_ids.shape[1]:], skip_special_tokens=True)
            print(f"\nPapiAi >> {response.strip()}\n")
            
        except KeyboardInterrupt:
            print("\nTerminating PapiAi Core. Goodbye.")
            break

if __name__ == "__main__":
    chat_with_papiai()
