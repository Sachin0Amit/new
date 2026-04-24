import pickle
import os

class ProblemClassifier:
    """
    Categorizes math problems into domains using the pre-trained pipeline.
    """
    def __init__(self, model_path="math_solver/data/classifier.pkl"):
        self.model_path = model_path
        self.pipeline = None
        self.load_model()

    def load_model(self):
        if os.path.exists(self.model_path):
            try:
                with open(self.model_path, 'rb') as f:
                    self.pipeline = pickle.load(f)
            except Exception:
                self.pipeline = None

    def classify(self, text: str) -> str:
        """
        Returns the predicted category for the problem.
        """
        if self.pipeline:
            return self.pipeline.predict([text])[0]
        
        # Rule-based fallback
        text_lower = text.lower()
        if any(w in text_lower for w in ['integrate', 'derivative', 'integral', 'limit', 'calc']):
            return 'calculus'
        if any(w in text_lower for w in ['matrix', 'vector', 'eigen', 'determinant']):
            return 'linear_algebra'
        if any(w in text_lower for w in ['triangle', 'circle', 'angle', 'area', 'perimeter']):
            return 'geometry'
        if any(w in text_lower for w in ['probability', 'mean', 'median', 'distribution']):
            return 'statistics'
        if any(w in text_lower for w in ['solve for x', 'equation', 'simplify']):
            return 'algebra'
        
        return 'general'
