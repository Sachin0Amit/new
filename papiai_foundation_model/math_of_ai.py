import torch
import math

# ==============================================================================
# LESSON 1: THE DOT PRODUCT (The Foundation of Intelligence)
# ==============================================================================
"""
In AI, every "thought" starts as a Dot Product. 
Formula: A · B = Σ (a_i * b_i)

Think of vector A as "Input Data" and vector B as "AI Knowledge" (Weights).
The Dot Product measures how much they ALIGN.
"""

def teach_dot_product():
    print("--- Lesson 1: The Dot Product ---")
    
    # Input Vector (e.g., features of a word)
    input_vector = torch.tensor([1.0, 2.0, 3.0])
    
    # Weight Vector (what the AI has learned)
    weight_vector = torch.tensor([0.5, -1.0, 2.0])
    
    # Manual Calculation: (1.0*0.5) + (2.0*-1.0) + (3.0*2.0) = 0.5 - 2.0 + 6.0 = 4.5
    dot_product = torch.dot(input_vector, weight_vector)
    
    print(f"Input:  {input_vector}")
    print(f"Weight: {weight_vector}")
    print(f"Dot Product (Alignment): {dot_product.item()}\n")


# ==============================================================================
# LESSON 2: SOFTMAX (Turning Raw Math into Probabilities)
# ==============================================================================
"""
A Neural Network outputs "Logits" (raw scores). We use the SOFTMAX function
to turn these scores into probabilities that sum to 100%.

Formula: σ(z)_i = e^(z_i) / Σ e^(z_j)
"""

def teach_softmax():
    print("--- Lesson 2: Softmax (Probability) ---")
    
    # Raw scores for three words: ["Apple", "Banana", "Cat"]
    logits = torch.tensor([2.0, 1.0, 0.1])
    
    # Apply Softmax
    probabilities = torch.softmax(logits, dim=0)
    
    print(f"Raw Scores (Logits): {logits}")
    print(f"Probabilities:      {probabilities}")
    print(f"Sum:                {torch.sum(probabilities).item()} (Should be 1.0)\n")


# ==============================================================================
# LESSON 3: GRADIENTS (How AI Learns via Calculus)
# ==============================================================================
"""
The Gradient (Derivative) tells the AI how to change its weights to 
reduce its "Loss" (error). If the gradient is positive, the weight 
should decrease. If negative, it should increase.
"""

def teach_gradients():
    print("--- Lesson 3: Gradients (Calculus) ---")
    
    # 1. Initialize a weight (random knowledge)
    w = torch.tensor([2.0], requires_grad=True)
    
    # 2. Define a Goal: We want the weight to become 10.0
    goal = torch.tensor([10.0])
    
    # 3. Calculate Loss (Error squared): L = (w - goal)^2
    loss = (w - goal)**2
    
    # 4. Calculus: Compute the Gradient (dL/dw)
    # d/dw of (w - 10)^2 is 2*(w - 10). 
    # At w=2, gradient should be 2*(2 - 10) = -16.
    loss.backward()
    
    print(f"Current Weight: {w.item()}")
    print(f"Target Goal:    {goal.item()}")
    print(f"Calculated Loss (Error): {loss.item()}")
    print(f"Gradient (Direction to move): {w.grad.item()}")
    print("Interpretation: A negative gradient means 'Increase the weight to lower the loss'.\n")

# ==============================================================================
# LESSON 4: MATRIX MULTIPLICATION (Parallel Thinking)
# ==============================================================================
"""
A Dot Product handles 1 comparison. 
A Matrix Multiplication handles THOUSANDS of comparisons simultaneously.

When an AI reads a sentence, it doesn't process one word at a time in 
isolation. It uses Matrix Multiplication to compare EVERY word in the 
sentence with EVERY weight in its brain at the same time.
"""

def teach_matrices():
    print("--- Lesson 4: Matrix Multiplication (The Speed Demon) ---")
    
    # A Batch of Inputs (3 words, each with 2 features)
    # Shape: (3, 2)
    inputs = torch.tensor([
        [1.0, 0.5],  # Word 1
        [0.2, 1.0],  # Word 2
        [-0.5, 2.0]  # Word 3
    ])
    
    # A Matrix of Weights (2 features mapped to 4 hidden concepts)
    # Shape: (2, 4)
    weights = torch.tensor([
        [0.1, 0.2, 0.3, 0.4],
        [0.5, 0.6, 0.7, 0.8]
    ])
    
    # Matrix Multiplication: (3x2) * (2x4) = (3x4)
    # This calculates 12 Dot Products in one single operation!
    output = torch.matmul(inputs, weights)
    
    print(f"Inputs (Words):\n{inputs}")
    print(f"Weights (Brain Patterns):\n{weights}")
    print(f"Output (Parallel Thoughts):\n{output}")
    print("\nConclusion: Without Matrix Multiplication, AI would be 1,000,000x slower.")

if __name__ == "__main__":
    teach_dot_product()
    teach_softmax()
    teach_gradients()
    teach_matrices()
