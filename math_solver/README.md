# Universal Math Intelligence Core

A production-grade, original AI-powered system designed to solve a vast range of mathematical problems, from simple arithmetic to advanced calculus, linear algebra, and more.

## Core Features
- **Natural Language Understanding**: Parse problems written in plain English, LaTeX, or mathematical notation.
- **Problem Classification**: Automatically categorize problems into domains like Algebra, Calculus, Statistics, etc.
- **Symbolic Reasoning**: Step-by-step derivation using an internal symbolic engine.
- **Numerical Approximation**: High-precision solvers for ODEs, integration, and optimization.
- **Geometric Visualization**: Automatically generate diagrams and plots for relevant problems.
- **Multi-Modal API**: REST API via FastAPI and a comprehensive CLI.

## Installation

1. Clone the repository and navigate to the project directory.
2. Create a virtual environment:
   ```bash
   python -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```
3. Install dependencies:
   ```bash
   pip install -r requirements.txt
   python -m spacy download en_core_web_sm
   ```

## Usage

### Command Line Interface
Run the `solve.py` script:
```bash
python solve.py --input "Solve x^2 - 5x + 6 = 0"
```

### REST API
Start the server:
```bash
python main.py
```
Then send a POST request to `http://localhost:8000/solve`:
```json
{
  "input": "Calculate the derivative of x^3 * sin(x)"
}
```

## Architecture
- `core/processor.py`: Input parsing and LaTeX normalization.
- `core/classifier.py`: Domain classification using machine learning.
- `core/symbolic.py`: Symbolic manipulation and step-tracking.
- `core/numeric.py`: Numerical solvers and physics-based approximations.
- `core/orchestrator.py`: The central logic that selects solvers and compiles results.
- `core/geometry.py`: Geometric parsing and visualization.

## Legal and Attestation
This system is an original work and does not derive from or reference any commercial math-solving products. All logic is implemented from scratch using open-source libraries.
