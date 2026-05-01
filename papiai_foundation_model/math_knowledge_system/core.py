import json
import os
from typing import List, Dict, Any

class MathKnowledgeSystem:
    """
    Core engine for generating a comprehensive mathematical knowledge dataset.
    Handles scaling, formatting (LaTeX, JSONL), and structuring the output.
    """
    def __init__(self, output_dir: str = "dataset"):
        self.output_dir = output_dir
        if not os.path.exists(output_dir):
            os.makedirs(output_dir)
            
    def save_batch(self, domain: str, records: List[Dict[str, Any]], batch_id: int):
        """Saves a batch of mathematical QA pairs in JSONL format."""
        file_path = os.path.join(self.output_dir, f"{domain}_batch_{batch_id}.jsonl")
        with open(file_path, "a") as f:
            for record in records:
                f.write(json.dumps(record) + "\n")
        print(f"[*] Saved {len(records)} records to {file_path}")

    @staticmethod
    def format_record(topic: str, question: str, steps: List[str], final_answer: str, latex: str) -> Dict[str, Any]:
        """Structures the data perfectly for LLM training."""
        return {
            "domain": topic,
            "instruction": question,
            "reasoning_chain": steps,
            "output": final_answer,
            "latex_representation": latex
        }
