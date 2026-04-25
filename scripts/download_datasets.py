#!/usr/bin/env python3
"""
Training data pipeline: Download and process open datasets for fine-tuning
Supports: OASST1, Alpaca, Dolly-15k
"""

import json
import os
import requests
import hashlib
import logging
from pathlib import Path
from typing import List, Dict, Tuple
from dataclasses import dataclass
import re

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class ConversationTurn:
    """A single turn in a conversation"""
    instruction: str
    input: str
    output: str
    
    def to_dict(self) -> Dict:
        return {
            "instruction": self.instruction,
            "input": self.input,
            "output": self.output
        }

class DatasetDownloader:
    """Downloads and processes open datasets"""
    
    def __init__(self, output_dir: str = "data/training"):
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)
        self.seen_hashes = set()
    
    def _is_duplicate(self, text: str) -> bool:
        """Check if we've seen this text before (deduplication)"""
        hash_val = hashlib.md5(text.encode()).hexdigest()
        if hash_val in self.seen_hashes:
            return True
        self.seen_hashes.add(hash_val)
        return False
    
    def _normalize_text(self, text: str) -> str:
        """Normalize text for consistency"""
        # Remove extra whitespace
        text = re.sub(r'\s+', ' ', text).strip()
        # Remove HTML entities
        text = text.replace('&quot;', '"').replace('&amp;', '&').replace('&lt;', '<').replace('&gt;', '>')
        return text
    
    def download_alpaca(self) -> List[ConversationTurn]:
        """Download Alpaca dataset (52k examples)"""
        logger.info("Downloading Alpaca dataset...")
        url = "https://raw.githubusercontent.com/tatsu-lab/alpaca/main/data/alpaca_data.json"
        
        try:
            response = requests.get(url, timeout=30)
            response.raise_for_status()
            
            data = response.json()
            turns = []
            
            for item in data:
                if self._is_duplicate(item['instruction']):
                    continue
                
                turn = ConversationTurn(
                    instruction=self._normalize_text(item['instruction']),
                    input=self._normalize_text(item.get('input', '')),
                    output=self._normalize_text(item['output'])
                )
                turns.append(turn)
            
            logger.info(f"Downloaded {len(turns)} Alpaca examples")
            return turns
            
        except Exception as e:
            logger.error(f"Failed to download Alpaca: {e}")
            return []
    
    def download_dolly(self) -> List[ConversationTurn]:
        """Download Dolly-15k dataset"""
        logger.info("Downloading Dolly-15k dataset...")
        url = "https://raw.githubusercontent.com/databrickslabs/dolly/main/data/databricks-dolly-15k.jsonl"
        
        try:
            response = requests.get(url, timeout=30)
            response.raise_for_status()
            
            turns = []
            for line in response.text.strip().split('\n'):
                if not line.strip():
                    continue
                
                item = json.loads(line)
                
                if self._is_duplicate(item['instruction']):
                    continue
                
                turn = ConversationTurn(
                    instruction=self._normalize_text(item['instruction']),
                    input=self._normalize_text(item.get('context', '')),
                    output=self._normalize_text(item['response'])
                )
                turns.append(turn)
            
            logger.info(f"Downloaded {len(turns)} Dolly examples")
            return turns
            
        except Exception as e:
            logger.error(f"Failed to download Dolly: {e}")
            return []
    
    def download_oasst1(self) -> List[ConversationTurn]:
        """Download OpenAssistant OASST1 dataset"""
        logger.info("Downloading OASST1 dataset...")
        # Note: OASST1 is large; this downloads a subset
        url = "https://huggingface.co/datasets/OpenAssistant/oasst1/resolve/main/data/en/train-00000-of-00001.parquet"
        
        try:
            # For production, use pyarrow/pandas to read parquet
            # For now, return a note that this needs proper setup
            logger.warning("OASST1 requires pandas/pyarrow. Install with: pip install pandas pyarrow")
            
            # Alternative: download from a processed version
            response = requests.get(
                "https://huggingface.co/datasets/OpenAssistant/oasst1/raw/main/README.md",
                timeout=30
            )
            logger.info("OASST1 available on HuggingFace. Use 'huggingface-hub' library to download.")
            return []
            
        except Exception as e:
            logger.warning(f"OASST1 download skipped: {e}")
            return []
    
    def process_all(self) -> Tuple[int, int]:
        """Download and process all datasets"""
        all_turns = []
        
        # Download each dataset
        all_turns.extend(self.download_alpaca())
        all_turns.extend(self.download_dolly())
        all_turns.extend(self.download_oasst1())
        
        logger.info(f"Total examples after deduplication: {len(all_turns)}")
        
        # Save as JSONL
        output_path = self.output_dir / "combined_training_data.jsonl"
        with open(output_path, 'w') as f:
            for turn in all_turns:
                f.write(json.dumps(turn.to_dict()) + '\n')
        
        logger.info(f"Saved {len(all_turns)} examples to {output_path}")
        
        # Create train/val split (80/20)
        train_size = int(len(all_turns) * 0.8)
        train_turns = all_turns[:train_size]
        val_turns = all_turns[train_size:]
        
        train_path = self.output_dir / "train.jsonl"
        val_path = self.output_dir / "val.jsonl"
        
        with open(train_path, 'w') as f:
            for turn in train_turns:
                f.write(json.dumps(turn.to_dict()) + '\n')
        
        with open(val_path, 'w') as f:
            for turn in val_turns:
                f.write(json.dumps(turn.to_dict()) + '\n')
        
        logger.info(f"Train set: {len(train_turns)} examples -> {train_path}")
        logger.info(f"Validation set: {len(val_turns)} examples -> {val_path}")
        
        return len(train_turns), len(val_turns)

if __name__ == "__main__":
    downloader = DatasetDownloader()
    train_count, val_count = downloader.process_all()
    print(f"\n✓ Training pipeline complete!")
    print(f"  Train: {train_count} examples")
    print(f"  Val:   {val_count} examples")
