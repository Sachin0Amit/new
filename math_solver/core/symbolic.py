import sympy
from typing import List, Dict, Any, Union

class SymbolicEngine:
    """
    Wraps SymPy to solve algebraic, calculus, and logic problems.
    Tracks transformation steps for reasoning output.
    """
    def __init__(self):
        self.steps = []

    def clear_steps(self):
        self.steps = []

    def add_step(self, description: str, expression: Any):
        self.steps.append({
            "description": description,
            "expression": sympy.latex(expression) if hasattr(expression, 'free_symbols') else str(expression)
        })

    def solve_equation(self, equation_str: str, variable_str: str = 'x') -> Dict[str, Any]:
        """
        Solves equations like x^2 - 5x + 6 = 0.
        """
        self.clear_steps()
        try:
            var = sympy.symbols(variable_str)
            # Parse equation
            if '=' in equation_str:
                lhs_str, rhs_str = equation_str.split('=')
                lhs = sympy.parse_expr(lhs_str)
                rhs = sympy.parse_expr(rhs_str)
                eq = sympy.Eq(lhs, rhs)
            else:
                eq = sympy.parse_expr(equation_str)
            
            self.add_step("Identify the equation to solve", eq)
            
            solutions = sympy.solve(eq, var)
            self.add_step("Find the roots of the equation", solutions)
            
            return {
                "result": solutions,
                "steps": self.steps
            }
        except Exception as e:
            return {"error": str(e), "steps": self.steps}

    def differentiate(self, expression_str: str, variable_str: str = 'x') -> Dict[str, Any]:
        """
        Calculates the derivative of an expression.
        """
        self.clear_steps()
        try:
            var = sympy.symbols(variable_str)
            expr = sympy.parse_expr(expression_str)
            self.add_step("Identify the expression to differentiate", expr)
            
            derivative = sympy.diff(expr, var)
            self.add_step("Apply differentiation rules", derivative)
            
            return {
                "result": derivative,
                "steps": self.steps
            }
        except Exception as e:
            return {"error": str(e), "steps": self.steps}

    def integrate(self, expression_str: str, variable_str: str = 'x', limits: tuple = None) -> Dict[str, Any]:
        """
        Calculates the indefinite or definite integral.
        """
        self.clear_steps()
        try:
            var = sympy.symbols(variable_str)
            expr = sympy.parse_expr(expression_str)
            
            if limits:
                self.add_step(f"Define integration of {expr} from {limits[0]} to {limits[1]}", expr)
                integral = sympy.integrate(expr, (var, limits[0], limits[1]))
            else:
                self.add_step(f"Define indefinite integration of {expr}", expr)
                integral = sympy.integrate(expr, var)
                
            self.add_step("Compute the integral result", integral)
            
            return {
                "result": integral,
                "steps": self.steps
            }
        except Exception as e:
            return {"error": str(e), "steps": self.steps}

    def matrix_ops(self, matrix_data: List[List[float]], op: str = 'det') -> Dict[str, Any]:
        """
        Performs matrix operations like determinant, inverse.
        """
        self.clear_steps()
        try:
            M = sympy.Matrix(matrix_data)
            self.add_step("Initialize matrix", M)
            
            result = None
            if op == 'det':
                result = M.det()
                self.add_step("Calculate determinant", result)
            elif op == 'inv':
                result = M.inv()
                self.add_step("Calculate matrix inverse", result)
            elif op == 'eigen':
                result = M.eigenvals()
                self.add_step("Calculate eigenvalues", result)
                
            return {
                "result": result,
                "steps": self.steps
            }
        except Exception as e:
            return {"error": str(e), "steps": self.steps}

    def simplify(self, expression_str: str) -> Dict[str, Any]:
        """
        Simplifies a mathematical expression.
        """
        self.clear_steps()
        try:
            expr = sympy.parse_expr(expression_str)
            self.add_step("Identify expression to simplify", expr)
            
            simplified = sympy.simplify(expr)
            self.add_step("Apply simplification rules", simplified)
            
            return {
                "result": simplified,
                "steps": self.steps
            }
        except Exception as e:
            return {"error": str(e), "steps": self.steps}
