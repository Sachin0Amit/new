from .processor import InputProcessor
from .classifier import ProblemClassifier
from .symbolic import SymbolicEngine
from .numeric import NumericSolver
from .geometry import GeometryHandler
from typing import Dict, Any

class SolutionOrchestrator:
    """
    The central intelligence that parses, classifies, and solves mathematical problems.
    """
    def __init__(self):
        self.processor = InputProcessor()
        self.classifier = ProblemClassifier()
        self.symbolic_engine = SymbolicEngine()
        self.numeric_solver = NumericSolver()
        self.geometry_handler = GeometryHandler()

    def solve(self, user_input: str) -> Dict[str, Any]:
        """
        Full pipeline: Process -> Classify -> Solve -> Format
        """
        # 1. Input Processing
        proc_data = self.processor.process(user_input)
        query = proc_data['normalized']
        
        # 2. Classification
        category = self.classifier.classify(query)
        
        # 3. Solver Selection & Execution
        result = {}
        plot_data = None
        
        try:
            if category == 'calculus':
                if 'integral' in query:
                    # simplistic extraction logic - in real world use more robust regex
                    expr = proc_data['equations'][0] if proc_data['equations'] else query
                    result = self.symbolic_engine.integrate(expr)
                elif 'derivative' in query or 'differentiate' in query:
                    expr = proc_data['equations'][0] if proc_data['equations'] else query
                    result = self.symbolic_engine.differentiate(expr)
                else:
                    result = {"error": "Specific calculus operation not detected"}
            
            elif category == 'algebra':
                expr = proc_data['equations'][0] if proc_data['equations'] else query
                if '=' in expr:
                    result = self.symbolic_engine.solve_equation(expr)
                else:
                    result = self.symbolic_engine.simplify(expr)
            
            elif category == 'geometry':
                # Return a basic placeholder result for geometry
                result = {"result": "Geometry solving requires specific parameter extraction.", "type": "geometry"}
                # Example: plot a default circle
                plot_data = self.geometry_handler.plot_function("sqrt(25 - x**2)") 

            else:
                # Default to symbolic simplification or equation solving
                expr = proc_data['equations'][0] if proc_data['equations'] else query
                result = self.symbolic_engine.simplify(expr)
            
            # 4. Visualization (if applicable)
            if category in ['algebra', 'calculus'] and not plot_data:
                # Attempt to plot the result or primary expression
                expr_to_plot = proc_data['equations'][0] if proc_data['equations'] else None
                if expr_to_plot:
                    plot_data = self.geometry_handler.plot_function(expr_to_plot)

        except Exception as e:
            result = {"error": f"Internal solver error: {str(e)}"}

        # 5. Final Response Assembly
        return {
            "problem_type": category,
            "input_processed": proc_data,
            "solution": result.get("result", "No explicit result found."),
            "steps": result.get("steps", []),
            "plots": plot_data,
            "final_answer": str(result.get("result", "N/A"))
        }
