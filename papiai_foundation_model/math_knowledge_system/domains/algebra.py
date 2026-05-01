import sympy as sp
import random
from core import MathKnowledgeSystem

def generate_algebra_problems(num_problems: int = 100) -> list:
    """Procedurally generates algebra problems (Equations, Systems, Matrices)."""
    records = []
    x, y, z = sp.symbols('x y z')
    
    for _ in range(num_problems):
        prob_type = random.choice(['equation', 'system', 'matrix_det'])
        
        if prob_type == 'equation':
            # Generate a cubic or quadratic equation
            root1, root2, root3 = random.randint(-5, 5), random.randint(-5, 5), random.randint(-5, 5)
            expr = sp.expand((x - root1) * (x - root2) * (x - root3))
            
            question = f"Find all roots of the polynomial equation: {sp.pretty(expr)} = 0"
            steps = [
                f"1. Let P(x) = {sp.latex(expr)}",
                "2. We need to find values of x such that P(x) = 0.",
                "3. We factor the polynomial using algebraic grouping or the rational root theorem.",
            ]
            ans = sp.solve(expr, x)
            latex = f"\\text{{Roots of }} {sp.latex(expr)} = 0 \\text{{ are }} x \\in \\{{ {', '.join([sp.latex(r) for r in ans])} \\}}"
            final = f"x = {ans}"
            
        elif prob_type == 'system':
            # Generate a 2x2 linear system
            a1, b1, c1 = random.randint(-10,10), random.randint(-10,10), random.randint(-20,20)
            a2, b2, c2 = random.randint(-10,10), random.randint(-10,10), random.randint(-20,20)
            eq1 = sp.Eq(a1*x + b1*y, c1)
            eq2 = sp.Eq(a2*x + b2*y, c2)
            
            question = f"Solve the following system of linear equations:\n1) {sp.pretty(eq1)}\n2) {sp.pretty(eq2)}"
            steps = [
                f"1. We have a system of two linear equations.",
                f"2. Eq 1: {sp.latex(eq1)}",
                f"3. Eq 2: {sp.latex(eq2)}",
                "4. We can solve this using substitution, elimination, or matrix inversion (Cramer's rule)."
            ]
            ans = sp.solve((eq1, eq2), (x, y))
            if not ans or not isinstance(ans, dict):
                final = "No unique solution (parallel or coincident lines)"
                latex = "\\text{No unique solution}"
            else:
                final = f"x = {ans.get(x, 'N/A')}, y = {ans.get(y, 'N/A')}"
                latex = f"x = {sp.latex(ans.get(x))}, y = {sp.latex(ans.get(y))}"
            
        else: # Matrix Determinant
            dim = random.choice([2, 3])
            mat_data = [[random.randint(-5, 5) for _ in range(dim)] for _ in range(dim)]
            mat = sp.Matrix(mat_data)
            
            question = f"Calculate the determinant of the {dim}x{dim} matrix:\n{sp.pretty(mat)}"
            steps = [
                f"1. Let A be the matrix: {sp.latex(mat)}",
                "2. Apply the determinant formula (e.g., ad-bc for 2x2, or cofactor expansion for 3x3).",
                "3. Sum the products of the elements and their corresponding cofactors."
            ]
            ans = mat.det()
            latex = f"\\det(A) = {ans}"
            final = f"Determinant = {ans}"

        records.append(MathKnowledgeSystem.format_record(
            topic=f"Algebra - {prob_type.replace('_', ' ').capitalize()}",
            question=question,
            steps=steps,
            final_answer=final,
            latex=latex
        ))
        
    return records
