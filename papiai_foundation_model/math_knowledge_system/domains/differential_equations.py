import sympy as sp
import random
from core import MathKnowledgeSystem

def generate_ode_problems(num_problems: int = 100) -> list:
    """Procedurally generates Ordinary Differential Equation problems."""
    records = []
    x = sp.Symbol('x')
    f = sp.Function('f')
    
    for _ in range(num_problems):
        prob_type = random.choice(['separable', 'linear'])
        
        if prob_type == 'separable':
            # f'(x) = g(x)h(f(x)) -> simple form f'(x) = k * x^n * f(x)
            k = random.randint(-5, 5)
            if k == 0: k = 1
            n = random.randint(1, 3)
            
            eq = sp.Eq(f(x).diff(x), k * x**n * f(x))
            question = f"Solve the separable differential equation: {sp.pretty(eq)}"
            steps = [
                f"1. The given differential equation is {sp.latex(eq)}",
                "2. Separate the variables by moving all terms involving f(x) to one side and x to the other.",
                "3. Integrate both sides with respect to their respective variables.",
                "4. Solve for f(x) and include the constant of integration C."
            ]
            try:
                ans = sp.dsolve(eq, f(x))
                latex = sp.latex(ans)
                final = sp.pretty(ans)
            except:
                continue
            
        else: # simple linear first order: f'(x) + p(x)f(x) = q(x)
            p = random.randint(1, 3) * x
            q = random.randint(1, 5) * sp.exp(x)
            
            eq = sp.Eq(f(x).diff(x) + p * f(x), q)
            question = f"Solve the first-order linear differential equation: {sp.pretty(eq)}"
            steps = [
                f"1. The given differential equation is {sp.latex(eq)}",
                "2. Identify the integrating factor μ(x) = exp(∫ p(x) dx).",
                "3. Multiply the entire equation by the integrating factor.",
                "4. Rewrite the left side as the derivative of the product (μ(x)f(x))'.",
                "5. Integrate both sides and solve for f(x)."
            ]
            try:
                ans = sp.dsolve(eq, f(x))
                latex = sp.latex(ans)
                final = sp.pretty(ans)
            except:
                continue

        records.append(MathKnowledgeSystem.format_record(
            topic=f"Differential Equations - {prob_type.capitalize()}",
            question=question,
            steps=steps,
            final_answer=final,
            latex=latex
        ))
        
    return records
