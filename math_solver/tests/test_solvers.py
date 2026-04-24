import pytest
from core.symbolic import SymbolicEngine
from core.numeric import NumericSolver

def test_symbolic_solve():
    engine = SymbolicEngine()
    result = engine.solve_equation("x**2 - 5*x + 6 = 0")
    assert 2 in result['result']
    assert 3 in result['result']

def test_symbolic_differentiate():
    engine = SymbolicEngine()
    result = engine.differentiate("x**3")
    # result is a sympy object, so we convert to string or check its properties
    assert str(result['result']) == "3*x**2"

def test_numeric_integration():
    solver = NumericSolver()
    result = solver.numerical_integration("x**2", 0, 1)
    assert abs(result['result'] - 1/3) < 1e-5

def test_linear_regression():
    solver = NumericSolver()
    x = [1, 2, 3, 4, 5]
    y = [2, 4, 6, 8, 10]
    result = solver.linear_regression(x, y)
    assert abs(result['slope'] - 2.0) < 1e-5
