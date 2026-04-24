import re
try:
    import spacy
except ImportError:
    spacy = None
from typing import List, Dict, Any

class InputProcessor:
    """
    Handles raw text input, normalization, LaTeX detection, and tokenization.
    """
    def __init__(self):
        try:
            self.nlp = spacy.load("en_core_web_sm")
        except:
            # Fallback if model is not downloaded yet
            self.nlp = None
        
        self.latex_pattern = re.compile(r'\$(.*?)\$|\\\[(.*?)\\\]')
        self.math_operators = ['+', '-', '*', '/', '^', '=', '>', '<', '(', ')']

    def process(self, raw_input: str) -> Dict[str, Any]:
        """
        Main entry point for input processing.
        """
        normalized = self.normalize_text(raw_input)
        latex_segments = self.extract_latex(raw_input)
        
        # Detect intent and extract equations
        equations = self.extract_equations(normalized)
        
        return {
            "original": raw_input,
            "normalized": normalized,
            "latex_segments": latex_segments,
            "equations": equations,
            "is_multilingual": False # Placeholder for future expansion
        }

    def normalize_text(self, text: str) -> str:
        """
        Cleans and normalizes the input text.
        """
        text = text.strip()
        # Remove common unnecessary phrases
        text = re.sub(r'(?i)please solve|calculate|find the value of|what is', '', text)
        return text.strip()

    def extract_latex(self, text: str) -> List[str]:
        """
        Finds all LaTeX segments in the input.
        """
        matches = self.latex_pattern.findall(text)
        # flatten results from groups
        segments = []
        for m in matches:
            segments.append(m[0] if m[0] else m[1])
        return segments

    def extract_equations(self, text: str) -> List[str]:
        """
        Uses regex and logic to pull potential mathematical expressions.
        """
        # Look for patterns containing digits, variables, and math operators
        equation_pattern = re.compile(r'([a-zA-Z0-9\s\+\-\*\/\^\=\(\)\.\,]+)')
        potential_matches = equation_pattern.findall(text)
        
        refined = []
        for match in potential_matches:
            # Only keep if it contains at least one operator or an equals sign
            if any(op in match for op in self.math_operators) or re.search(r'\d', match):
                clean = match.strip()
                if len(clean) > 1:
                    refined.append(clean)
        
        return refined

    def split_questions(self, text: str) -> List[str]:
        """
        Splits text into multiple questions if delimiters like 'and' or ';' are used.
        """
        if self.nlp:
            doc = self.nlp(text)
            return [sent.text.strip() for sent in doc.sents]
        
        # Fallback split
        return [q.strip() for q in re.split(r'[;.?\n]|\band\b', text) if q.strip()]
