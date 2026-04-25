import pytest
from math_solver.solver import SymbolicSolver

@pytest.fixture
def solver():
    return SymbolicSolver()

def test_algebra_simplification(solver):
    res = solver.solve_algebra("x**2 + 2*x + 1", ["x"], {})
    assert "x + 1" in res["result"] # (x+1)**2 simplifies to x**2+2x+1, but sympy might keep it expanded or factor
    assert res["computation_time_ms"] < 50

def test_calculus_differentiation(solver):
    res = solver.solve_calculus("sin(x)", ["x"], "diff")
    assert res["result"] == "cos(x)"
    assert res["computation_time_ms"] < 500

def test_calculus_integration(solver):
    res = solver.solve_calculus("exp(x)", ["x"], "integrate")
    assert res["result"] == "exp(x)"

def test_linear_algebra_determinant(solver):
    matrix = [[1, 2], [3, 4]]
    res = solver.solve_linear_algebra(matrix, "determinant")
    assert res["result"] == "-2"

def test_edge_division_by_zero(solver):
    with pytest.raises(Exception):
        solver.solve_algebra("1/0", ["x"], {})

def test_unicode_input(solver):
    # Testing symbolic π
    res = solver.solve_algebra("sin(pi)", ["x"], {})
    assert res["result"] == "0"
