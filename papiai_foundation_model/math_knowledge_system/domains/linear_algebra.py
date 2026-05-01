import sympy as sp
import random
from core import MathKnowledgeSystem

def generate_linear_algebra_problems(num_problems: int = 100) -> list:
    """Procedurally generates linear algebra problems (Eigenvalues, Inverses)."""
    records = []
    
    for _ in range(num_problems):
        prob_type = random.choice(['eigen', 'inverse'])
        dim = random.choice([2, 3])
        
        # Generate a random matrix
        mat_data = [[random.randint(-5, 5) for _ in range(dim)] for _ in range(dim)]
        mat = sp.Matrix(mat_data)
        
        if prob_type == 'eigen':
            question = f"Find the eigenvalues of the {dim}x{dim} matrix:\n{sp.pretty(mat)}"
            steps = [
                f"1. Let A be the matrix: {sp.latex(mat)}",
                "2. To find eigenvalues, we solve the characteristic equation det(A - λI) = 0.",
                "3. Calculate the determinant and solve for λ."
            ]
            
            try:
                ans = list(mat.eigenvals().keys())
                latex = f"\\text{{Eigenvalues: }} \\lambda \\in \\{{ {', '.join([sp.latex(val) for val in ans])} \\}}"
                final = f"Eigenvalues: {ans}"
            except:
                continue # Skip if too complex symbolically
                
        else: # inverse
            question = f"Find the inverse of the {dim}x{dim} matrix, if it exists:\n{sp.pretty(mat)}"
            steps = [
                f"1. Let A be the matrix: {sp.latex(mat)}",
                "2. First, check if the determinant is non-zero.",
                "3. If det(A) ≠ 0, calculate the inverse using the adjugate matrix and determinant."
            ]
            
            if mat.det() == 0:
                ans = "Matrix is singular (not invertible)."
                latex = "\\text{Matrix is singular}"
                final = ans
            else:
                try:
                    ans = mat.inv()
                    latex = f"A^{{-1}} = {sp.latex(ans)}"
                    final = sp.pretty(ans)
                except:
                    continue

        records.append(MathKnowledgeSystem.format_record(
            topic=f"Linear Algebra - {prob_type.capitalize()}",
            question=question,
            steps=steps,
            final_answer=final,
            latex=latex
        ))
        
    return records
