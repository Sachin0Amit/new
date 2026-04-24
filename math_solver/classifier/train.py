import pandas as pd
import pickle
import os
import sys
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline
from sklearn.metrics import classification_report

# Add project root to path
sys.path.append(os.getcwd())

from math_solver.data.download_datasets import run_downloader
from math_solver.data.synthetic_generator import generate_synthetic_data
from math_solver.data.prepare_training_data import prepare_data

MODEL_PATH = "math_solver/data/classifier.pkl"
TRAIN_DATA = "math_solver/data/train.csv"
TEST_DATA = "math_solver/data/test.csv"

def train():
    """Execution pipeline for downloading, preparing, and training the classifier."""
    
    # 1. Ensure data exists
    if not os.path.exists(TRAIN_DATA):
        print("[Train] Training data not found. Initializing collection pipeline...")
        run_downloader()
        generate_synthetic_data()
        prepare_data()
        
    if not os.path.exists(TRAIN_DATA):
        print("[Error] Failed to prepare training data. Aborting.")
        return

    # 2. Load Data
    print("[Train] Loading dataset...")
    df_train = pd.read_csv(TRAIN_DATA)
    df_test = pd.read_csv(TEST_DATA)

    # 3. Build Pipeline
    # LogisticRegression is often more robust than NaiveBayes for larger datasets
    pipeline = Pipeline([
        ('tfidf', TfidfVectorizer(max_features=10000, stop_words='english')),
        ('clf', LogisticRegression(max_iter=1000, multi_class='auto'))
    ])

    # 4. Train
    print(f"[Train] Training model on {len(df_train)} samples...")
    pipeline.fit(df_train['problem'], df_train['label'])

    # 5. Evaluate
    print("[Train] Evaluating model...")
    y_pred = pipeline.predict(df_test['problem'])
    print(classification_report(df_test['label'], y_pred))

    # 6. Save
    os.makedirs(os.path.dirname(MODEL_PATH), exist_ok=True)
    with open(MODEL_PATH, 'wb') as f:
        pickle.dump(pipeline, f)
    
    print(f"[Train] Model successfully saved to {MODEL_PATH}")

if __name__ == "__main__":
    train()
