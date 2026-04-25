import time
import sympy as sp
from typing import List, Dict, Any, Union

class SymbolicSolver:
    @staticmethod
    def _to_json(obj: Any) -> str:
        if hasattr(obj, 'tolist'):
            return str(obj.tolist())
        return str(obj)

    def solve_algebra(self, expression: str, variables: List[str], options: Dict[str, Any]) -> Dict[str, Any]:
        start_time = time.time()
        expr = sp.sympify(expression)
        syms = [sp.Symbol(v) for v in variables]
        
        result = sp.simplify(expr)
        
        return {
            "result": self._to_json(result),
            "latex": sp.latex(result),
            "steps": ["Simplified expression using SymPy"],
            "computation_time_ms": int((time.time() - start_time) * 1000)
        }

    def solve_calculus(self, expression: str, variables: List[str], op_type: str) -> Dict[str, Any]:
        start_time = time.time()
        expr = sp.sympify(expression)
        x = sp.Symbol(variables[0])
        
        if op_type == "diff":
            result = sp.diff(expr, x)
        elif op_type == "integrate":
            result = sp.integrate(expr, x)
        elif op_type == "limit":
            result = sp.limit(expr, x, 0)
        else:
            result = expr
            
        return {
            "result": self._to_json(result),
            "latex": sp.latex(result),
            "steps": [f"Performed {op_type} with respect to {variables[0]}"],
            "computation_time_ms": int((time.time() - start_time) * 1000)
        }

    def solve_linear_algebra(self, matrix_data: List[List[Any]], op_type: str) -> Dict[str, Any]:
        start_time = time.time()
        m = sp.Matrix(matrix_data)
        
        if op_type == "inverse":
            result = m.inv()
        elif op_type == "eigenvalues":
            result = m.eigenvals()
        elif op_type == "determinant":
            result = m.det()
        else:
            result = m
            
        return {
            "result": self._to_json(result),
            "latex": sp.latex(result),
            "steps": [f"Calculated {op_type} of the provided matrix"],
            "computation_time_ms": int((time.time() - start_time) * 1000)
        }
