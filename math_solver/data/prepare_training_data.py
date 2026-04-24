import json
import csv
import os
import random
from pathlib import Path
from collections import Counter

# Internal Label Mapping
CATEGORY_MAP = {
    "Prealgebra": "algebra",
    "Algebra": "algebra",
    "Intermediate Algebra": "algebra",
    "Geometry": "geometry",
    "Counting & Probability": "statistics",
    "Number Theory": "number_theory",
    "Precalculus": "calculus",
    "Calculus": "calculus",
}

BLACKLIST_TERMS = [
    "Wolfram", "Alpha", "Mathway", "Symbolab", "Microsoft Math"
]

def clean_text(text: str) -> str:
    """Normalizes whitespace and removes extraneous formatting."""
    text = text.replace('\n', ' ').replace('\r', ' ')
    # Check blacklist
    for term in BLACKLIST_TERMS:
        if term.lower() in text.lower():
            return ""
    return ' '.join(text.split()).strip()

def parse_math_dataset(base_path: Path):
    """Parses the MATH dataset from JSON files."""
    data = []
    if not base_path.exists():
        print("[Warning] MATH dataset path not found.")
        return data

    # MATH dataset has 'train' and 'test' folders
    for split in ['train', 'test']:
        split_path = base_path / split
        if not split_path.exists(): continue
        
        for category_dir in split_path.iterdir():
            if not category_dir.is_dir(): continue
            internal_label = CATEGORY_MAP.get(category_dir.name, "general")
            
            for file in category_dir.glob("*.json"):
                try:
                    with open(file, 'r', encoding='utf-8') as f:
                        item = json.load(f)
                        problem = clean_text(item.get("problem", ""))
                        if problem:
                            data.append((problem, internal_label))
                except Exception:
                    continue
    return data

def parse_gsm8k(file_path: Path):
    """Parses GSM8K from JSONL."""
    data = []
    if not file_path.exists():
        print("[Warning] GSM8K file not found.")
        return data

    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            for line in f:
                item = json.loads(line)
                problem = clean_text(item.get("question", ""))
                if problem:
                    data.append((problem, "arithmetic"))
    except Exception:
        pass
    return data

def balance_dataset(data: list, target_count: int = None):
    """Balances classes by oversampling minority classes."""
    if not data: return []
    
    counts = Counter(label for _, label in data)
    if not target_count:
        target_count = max(counts.values())
        
    balanced = []
    by_label = {}
    for p, l in data:
        if l not in by_label: by_label[l] = []
        by_label[l].append(p)
        
    for label, problems in by_label.items():
        # Basic oversampling
        resampled = random.choices(problems, k=target_count)
        for p in resampled:
            balanced.append((p, label))
            
    random.shuffle(balanced)
    return balanced

def prepare_data():
    """Main pipeline for preparing training and test data."""
    cache_path = Path("math_solver/data/.cache")
    synthetic_path = Path("math_solver/data/synthetic_data.csv")
    
    all_data = []
    
    # 1. Add Synthetic
    if synthetic_path.exists():
        with open(synthetic_path, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f)
            for row in reader:
                all_data.append((row['problem'], row['label']))
                
    # 2. Add MATH
    all_data.extend(parse_math_dataset(cache_path / "MATH_RAW" / "train")) # Simplified path check
    
    # 3. Add GSM8K
    all_data.extend(parse_gsm8k(cache_path / "gsm8k_train.jsonl"))
    
    # Deduplicate
    unique_data = list(set(all_data))
    print(f"[Prepare] Total unique problems collected: {len(unique_data)}")
    
    # Balance
    balanced_data = balance_dataset(unique_data)
    
    # Split
    random.shuffle(balanced_data)
    split_idx = int(len(balanced_data) * 0.8)
    train_data = balanced_data[:split_idx]
    test_data = balanced_data[split_idx:]
    
    # Save
    save_csv("math_solver/data/train.csv", train_data)
    save_csv("math_solver/data/test.csv", test_data)
    print(f"[Prepare] Saved {len(train_data)} train samples and {len(test_data)} test samples.")

def save_csv(path, data):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', newline='', encoding='utf-8') as f:
        writer = csv.writer(f)
        writer.writerow(["problem", "label"])
        writer.writerows(data)

if __name__ == "__main__":
    prepare_data()
