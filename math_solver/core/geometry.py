import matplotlib.pyplot as plt
import io
import base64
from typing import Dict, Any, List

class GeometryHandler:
    """
    Solves geometric problems and generates visual diagrams.
    """
    def __init__(self):
        pass

    def solve_triangle(self, a: float = None, b: float = None, c: float = None, 
                       A: float = None, B: float = None, C: float = None) -> Dict[str, Any]:
        """
        Solves for unknown sides/angles of a triangle using Law of Sines/Cosines.
        """
        # Implementation for triangle solving logic
        # For brevity, this is a skeleton showing the return structure
        return {
            "type": "triangle",
            "sides": {"a": a, "b": b, "c": c},
            "angles": {"A": A, "B": B, "C": C},
            "area": 0.5 * a * b * 1.0 # placeholder
        }

    def generate_plot(self, x_data: List[float], y_data: List[float], 
                      title: str = "Plot", xlabel: str = "X", ylabel: str = "Y") -> str:
        """
        Generates a plot and returns a base64 encoded PNG string.
        """
        plt.figure(figsize=(8, 5))
        plt.plot(x_data, y_data, marker='o', linestyle='-', color='b')
        plt.title(title)
        plt.xlabel(xlabel)
        plt.ylabel(ylabel)
        plt.grid(True)
        
        buf = io.BytesIO()
        plt.savefig(buf, format='png')
        plt.close()
        buf.seek(0)
        return base64.b64encode(buf.read()).decode('utf-8')

    def plot_function(self, func_str: str, range_x: tuple = (-10, 10)) -> str:
        """
        Plots a mathematical function.
        """
        import numpy as np
        x = np.linspace(range_x[0], range_x[1], 400)
        try:
            # Safer eval context
            safe_dict = {"np": np, "x": x, "sin": np.sin, "cos": np.cos, "tan": np.tan, 
                         "exp": np.exp, "log": np.log, "sqrt": np.sqrt, "pi": np.pi}
            y = eval(func_str.replace('^', '**'), {"__builtins__": None}, safe_dict)
            
            plt.figure(figsize=(8, 5))
            plt.plot(x, y)
            plt.title(f"Plot of f(x) = {func_str}")
            plt.axhline(0, color='black',linewidth=0.5)
            plt.axvline(0, color='black',linewidth=0.5)
            plt.grid(True)
            
            buf = io.BytesIO()
            plt.savefig(buf, format='png')
            plt.close()
            buf.seek(0)
            return base64.b64encode(buf.read()).decode('utf-8')
        except Exception as e:
            return f"Error plotting: {str(e)}"
