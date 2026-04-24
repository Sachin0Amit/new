import numpy as np
from scipy import integrate, optimize, stats
from typing import Callable, List, Dict, Any, Union

class NumericSolver:
    """
    Handles numerical computation: integration, root finding, and ODEs.
    """
    def __init__(self):
        pass

    def numerical_integration(self, func_str: str, lower: float, upper: float) -> Dict[str, Any]:
        """
        Computes definite integral numerically.
        """
        try:
            # Dangerous eval, in production use a safer parser like numexpr or a restricted dict
            f = eval(f"lambda x: {func_str}", {"np": np, "math": __import__('math')})
            result, error = integrate.quad(f, lower, upper)
            return {
                "result": result,
                "estimated_error": error,
                "method": "scipy.integrate.quad"
            }
        except Exception as e:
            return {"error": str(e)}

    def find_roots(self, func_str: str, x0: Union[float, List[float]]) -> Dict[str, Any]:
        """
        Finds roots of a function numerically.
        """
        try:
            f = eval(f"lambda x: {func_str}", {"np": np, "math": __import__('math')})
            result = optimize.fsolve(f, x0)
            return {
                "result": result.tolist(),
                "method": "scipy.optimize.fsolve"
            }
        except Exception as e:
            return {"error": str(e)}

    def solve_ode(self, deriv_func_str: str, y0: float, t: List[float]) -> Dict[str, Any]:
        """
        Solves first-order ordinary differential equations.
        dy/dt = f(t, y)
        """
        try:
            # f(y, t) format for odeint
            f = eval(f"lambda y, t: {deriv_func_str}", {"np": np, "math": __import__('math')})
            result = integrate.odeint(f, y0, t)
            return {
                "result": result.flatten().tolist(),
                "time_points": t,
                "method": "scipy.integrate.odeint"
            }
        except Exception as e:
            return {"error": str(e)}

    def linear_regression(self, x: List[float], y: List[float]) -> Dict[str, Any]:
        """
        Performs simple linear regression.
        """
        try:
            slope, intercept, r_value, p_value, std_err = stats.linregress(x, y)
            return {
                "slope": slope,
                "intercept": intercept,
                "r_squared": r_value**2,
                "p_value": p_value,
                "std_err": std_err
            }
        except Exception as e:
            return {"error": str(e)}
