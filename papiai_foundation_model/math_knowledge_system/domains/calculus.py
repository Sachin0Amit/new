import sympy as sp
import random
from core import MathKnowledgeSystem

def generate_calculus_problems(num_problems: int = 100) -> list:
    """Procedurally generates calculus problems (Derivatives & Integrals) with step-by-step proofs."""
    records = []
    x = sp.Symbol('x')
    
    for _ in range(num_problems):
        # Generate a random polynomial or trigonometric function
        terms = []
        for _ in range(random.randint(2, 4)):
            coef = random.randint(-10, 10)
            if coef == 0: coef = 1
            power = random.randint(1, 5)
            func_type = random.choice(['poly', 'sin', 'cos', 'exp'])
            
            if func_type == 'poly': terms.append(coef * x**power)
            elif func_type == 'sin': terms.append(coef * sp.sin(power * x))
            elif func_type == 'cos': terms.append(coef * sp.cos(power * x))
            elif func_type == 'exp': terms.append(coef * sp.exp(power * x))
            
        expr = sum(terms)
        prob_type = random.choice(['derivative', 'integral'])
        
        if prob_type == 'derivative':
            question = f"Calculate the first derivative of the function f(x) = {sp.pretty(expr)} with respect to x."
            steps = [
                f"1. Let f(x) = {sp.latex(expr)}",
                f"2. Apply the linearity of differentiation: d/dx [a*f(x) + b*g(x)] = a*d/dx[f(x)] + b*d/dx[g(x)]",
                f"3. Differentiate each term individually using power, chain, and trigonometric rules.",
            ]
            ans = sp.diff(expr, x)
            latex = f"\\frac{{d}}{{dx}} \\left( {sp.latex(expr)} \\right) = {sp.latex(ans)}"
            final = sp.pretty(ans)
            
        else: # integral
            question = f"Calculate the indefinite integral of the function f(x) = {sp.pretty(expr)} with respect to x."
            steps = [
                f"1. Let I = ∫ ({sp.latex(expr)}) dx",
                f"2. Apply the linearity of integration: ∫ [a*f(x) + b*g(x)] dx = a*∫ f(x) dx + b*∫ g(x) dx",
                f"3. Integrate each term individually using power, substitution, and trigonometric rules.",
                f"4. Add the constant of integration C."
            ]
            ans = sp.integrate(expr, x)
            latex = f"\\int \\left( {sp.latex(expr)} \\right) dx = {sp.latex(ans)} + C"
            final = f"{sp.pretty(ans)} + C"

        records.append(MathKnowledgeSystem.format_record(
            topic=f"Calculus - {prob_type.capitalize()}",
            question=question,
            steps=steps,
            final_answer=final,
            latex=latex
        ))
        
    return records
