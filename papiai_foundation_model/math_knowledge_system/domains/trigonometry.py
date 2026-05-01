import sympy as sp
import random
import math
from core import MathKnowledgeSystem

def generate_trigonometry_problems(num_problems: int = 100) -> list:
    """Procedurally generates trigonometry problems (Identities, Values, Conversions)."""
    records = []
    theta = sp.Symbol('theta')
    
    for _ in range(num_problems):
        prob_type = random.choice(['values', 'identity', 'conversion'])
        
        if prob_type == 'values':
            angle_deg = random.choice([0, 30, 45, 60, 90, 120, 150, 180])
            func = random.choice(['sin', 'cos', 'tan'])
            
            question = f"Find the exact value of {func}({angle_deg}°)."
            steps = [
                f"1. Identify the angle in the unit circle: {angle_deg}°.",
                f"2. Use the standard trigonometric values for common angles."
            ]
            
            angle_rad = sp.Rational(angle_deg, 180) * sp.pi
            if func == 'sin': ans = sp.sin(angle_rad)
            elif func == 'cos': ans = sp.cos(angle_rad)
            else: ans = sp.tan(angle_rad)
            
            latex = f"\\{func}({angle_deg}^\\circ) = {sp.latex(ans)}"
            final = f"{func}({angle_deg}°) = {sp.pretty(ans)}"
            
        elif prob_type == 'identity':
            # Simplify sin^2 + cos^2 type identities
            expr = sp.sin(theta)**2 + sp.cos(theta)**2
            question = f"Simplify the trigonometric expression: sin²(θ) + cos²(θ)"
            steps = [
                "1. Recall the fundamental Pythagorean trigonometric identity.",
                "2. For any angle θ, sin²(θ) + cos²(θ) = 1."
            ]
            ans = sp.simplify(expr)
            latex = "\\sin^2(\\theta) + \\cos^2(\\theta) = 1"
            final = "1"
            
        else: # Conversion
            deg = random.randint(1, 360)
            question = f"Convert {deg} degrees to radians."
            steps = [
                f"1. The conversion factor from degrees to radians is π/180.",
                f"2. Multiply {deg} by π/180."
            ]
            ans = sp.Rational(deg, 180) * sp.pi
            latex = f"{deg}^\\circ = \\frac{{{deg}}}{{180}}\\pi = {sp.latex(ans)} \\text{{ radians}}"
            final = f"{sp.pretty(ans)} radians"

        records.append(MathKnowledgeSystem.format_record(
            topic=f"Trigonometry - {prob_type.capitalize()}",
            question=question,
            steps=steps,
            final_answer=final,
            latex=latex
        ))
        
    return records
