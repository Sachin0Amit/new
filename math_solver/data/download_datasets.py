import urllib.request
import zipfile
import os
import shutil
from pathlib import Path

# Configuration for datasets
DATASETS = {
    "MATH": "https://github.com/hendrycks/math/archive/refs/heads/main.zip",
    "GSM8K": "https://raw.githubusercontent.com/openai/grade-school-math/master/grade_school_math/data/train.jsonl"
}

CACHE_DIR = Path("math_solver/data/.cache")

def download_file(url: str, dest_path: Path):
    """Downloads a file from a URL to a destination path."""
    if dest_path.exists():
        print(f"[Cache] {dest_path.name} already exists. Skipping download.")
        return True
    
    print(f"[Download] Fetching {url}...")
    try:
        with urllib.request.urlopen(url) as response, open(dest_path, 'wb') as out_file:
            shutil.copyfileobj(response, out_file)
        return True
    except Exception as e:
        print(f"[Error] Failed to download {url}: {e}")
        return False

def setup_cache():
    """Ensures the cache directory exists."""
    CACHE_DIR.mkdir(parents=True, exist_ok=True)

def extract_zip(zip_path: Path, extract_to: Path):
    """Extracts a zip file to a directory."""
    print(f"[Extract] Unzipping {zip_path} to {extract_to}...")
    try:
        with zipfile.ZipFile(zip_path, 'r') as zip_ref:
            zip_ref.extractall(extract_to)
        return True
    except Exception as e:
        print(f"[Error] Failed to extract {zip_path}: {e}")
        return False

def run_downloader():
    """Main execution point for downloading all required datasets."""
    setup_cache()
    
    # 1. Download MATH dataset
    math_zip = CACHE_DIR / "math_repo.zip"
    if download_file(DATASETS["MATH"], math_zip):
        extract_to = CACHE_DIR / "MATH_RAW"
        if not extract_to.exists():
            extract_zip(math_zip, CACHE_DIR)
            # GitHub zips have a folder like 'math-main'
            repo_folder = CACHE_DIR / "math-main"
            if repo_folder.exists():
                # MATH repo has 'train' and 'test' in its root
                repo_folder.rename(extract_to)
            else:
                # Try to find any folder if it's not 'math-main'
                for d in CACHE_DIR.iterdir():
                    if d.is_dir() and d.name.startswith("math-"):
                        d.rename(extract_to)
                        break
            
    # 2. Download GSM8K
    gsm8k_jsonl = CACHE_DIR / "gsm8k_train.jsonl"
    download_file(DATASETS["GSM8K"], gsm8k_jsonl)

if __name__ == "__main__":
    run_downloader()
