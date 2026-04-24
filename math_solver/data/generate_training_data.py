import sys
import os

# Add the project root to the python path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from core.classifier import generate_synthetic_data, train_classifier

if __name__ == "__main__":
    print("Generating synthetic math dataset...")
    generate_synthetic_data("math_solver/data/math_dataset.csv")
    print("Training classifier model...")
    train_classifier("math_solver/data/math_dataset.csv", "math_solver/data/classifier.pkl")
    print("Setup complete.")
