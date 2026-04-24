import pytest
import os
from pathlib import Path
import pandas as pd

def test_data_files_exist():
    # Note: These might only exist after running the training pipeline
    train_path = Path("math_solver/data/train.csv")
    test_path = Path("math_solver/data/test.csv")
    
    # We check if the preparation logic is sound by checking the CSV structure if they exist
    if train_path.exists():
        df = pd.read_csv(train_path)
        assert "problem" in df.columns
        assert "label" in df.columns
        assert not df.empty

def test_label_consistency():
    train_path = Path("math_solver/data/train.csv")
    if train_path.exists():
        df = pd.read_csv(train_path)
        labels = set(df['label'].unique())
        expected_labels = {'algebra', 'calculus', 'geometry', 'statistics', 'number_theory', 'arithmetic'}
        # Ensure at least some of our target labels are present
        assert len(labels.intersection(expected_labels)) > 0
