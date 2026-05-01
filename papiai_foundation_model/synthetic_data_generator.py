"""
Sovereign Intelligence Core - Infinite Synthetic Data Generator
=============================================================
This module programmatically generates the "terabytes of data" required
to train the PapiAi Foundation Model. It algorithmically constructs
endless permutations of code, mathematics, and logical reasoning.
"""

import random
import time
import os
import itertools

DATA_FILE = "training_data.txt"

def generate_math_problem():
    """Generates a step-by-step mathematical derivation."""
    a = random.randint(10, 1000)
    b = random.randint(10, 1000)
    op = random.choice(["+", "-", "*"])
    
    if op == "+":
        ans = a + b
    elif op == "-":
        ans = a - b
    else:
        ans = a * b
        
    prompt = f"User: Calculate {a} {op} {b} step-by-step.\n"
    response = f"PapiAi: To solve {a} {op} {b}, we break it down.\n"
    response += f"1. Identify the operands: {a} and {b}.\n"
    response += f"2. Apply the operation '{op}'.\n"
    response += f"3. The final computed result is {ans}.\n"
    return prompt + response + "\n"

def generate_code_snippet():
    """Generates structural Python code algorithms."""
    funcs = ["sort_array", "find_max", "calculate_entropy", "invert_matrix", "hash_string"]
    func = random.choice(funcs)
    var1 = random.choice(["data", "arr", "matrix", "payload"])
    
    prompt = f"User: Write a python function to {func.replace('_', ' ')}.\n"
    response = f"PapiAi: Here is the optimal implementation for {func}:\n"
    response += f"```python\ndef {func}({var1}):\n"
    response += f"    # Autonomous algorithmic generation\n"
    response += f"    if not {var1}:\n        return None\n"
    
    if func == "sort_array":
        response += f"    return sorted({var1})\n"
    elif func == "find_max":
        response += f"    return max({var1})\n"
    else:
        response += f"    result = [x * {random.randint(2,9)} for x in {var1}]\n"
        response += f"    return result\n"
    response += "```\n"
    
    return prompt + response + "\n"

def generate_logic_puzzle():
    """Generates a boolean logic derivation."""
    vars = ["P", "Q", "R"]
    v1, v2 = random.sample(vars, 2)
    gate = random.choice(["AND", "OR", "XOR", "IMPLIES"])
    
    prompt = f"User: Evaluate the logical statement {v1} {gate} {v2} when both are True.\n"
    response = f"PapiAi: Let's analyze {v1} {gate} {v2}.\n"
    response += f"Given {v1} = True, and {v2} = True.\n"
    
    if gate == "AND" or gate == "OR":
        res = "True"
    elif gate == "XOR":
        res = "False"
    else:
        res = "True"
        
    response += f"According to boolean logic, True {gate} True evaluates to {res}.\n"
    return prompt + response + "\n"

def generate_terabytes(target_mb=100):
    """
    Runs an infinite loop to generate training data.
    Defaults to writing until a specific MB limit is reached for safety,
    but can be set to infinity for actual Terabyte generation.
    """
    print("==================================================")
    print("   PapiAi | Synthetic Data Matrix Generator")
    print("==================================================")
    print(f"[*] Target Generation: {target_mb} MB of pure intelligence.")
    print("[*] Generating permutations of Math, Code, and Logic...")
    
    generators = [generate_math_problem, generate_code_snippet, generate_logic_puzzle]
    
    written_bytes = 0
    target_bytes = target_mb * 1024 * 1024
    
    # Check existing size
    if os.path.exists(DATA_FILE):
        written_bytes = os.path.getsize(DATA_FILE)
    
    try:
        with open(DATA_FILE, "a", encoding="utf-8") as f:
            while written_bytes < target_bytes:
                # Randomly pick a generation algorithm
                gen_func = random.choice(generators)
                data_chunk = gen_func()
                
                f.write(data_chunk)
                written_bytes += len(data_chunk.encode('utf-8'))
                
                if random.random() < 0.001: # Print status occasionally
                    print(f"    -> Matrix expansion: {written_bytes / (1024*1024):.2f} MB / {target_mb} MB")
                    
    except KeyboardInterrupt:
        print("\n[!] Data generation interrupted by user.")
        
    print(f"\n[+] Generation complete. Total Dataset Size: {written_bytes / (1024*1024):.2f} MB")
    print(f"[+] You may now run 'python3 train.py' to compile this data.")

if __name__ == "__main__":
    # Change this to 1000000 to generate 1 Terabyte of data!
    # Currently set to 50MB for immediate testing.
    generate_terabytes(target_mb=50)
