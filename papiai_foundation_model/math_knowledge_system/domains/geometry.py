import sympy as sp
import random
import math
from core import MathKnowledgeSystem

def generate_geometry_problems(num_problems: int = 100) -> list:
    """Procedurally generates geometry problems (Area, Volume, Pythagorean Theorem)."""
    records = []
    
    for _ in range(num_problems):
        prob_type = random.choice(['area_circle', 'volume_sphere', 'pythagoras'])
        
        if prob_type == 'area_circle':
            r = random.randint(1, 50)
            question = f"Calculate the area of a circle with radius r = {r}."
            steps = [
                f"1. The formula for the area of a circle is A = π * r^2.",
                f"2. Substitute r = {r} into the formula.",
                f"3. A = π * ({r})^2 = {r**2}π."
            ]
            ans = sp.pi * r**2
            latex = f"A = \\pi r^2 = \\pi ({r})^2 = {r**2}\\pi"
            final = f"Area = {r**2}π ≈ {round(math.pi * r**2, 2)}"
            
        elif prob_type == 'volume_sphere':
            r = random.randint(1, 30)
            question = f"Calculate the volume of a sphere with radius r = {r}."
            steps = [
                f"1. The formula for the volume of a sphere is V = (4/3) * π * r^3.",
                f"2. Substitute r = {r} into the formula.",
                f"3. V = (4/3) * π * ({r})^3 = (4/3) * {r**3} * π = {(4 * r**3)//3 if (4 * r**3)%3 == 0 else f'{4*r**3}/3'}π."
            ]
            ans = (sp.Rational(4, 3)) * sp.pi * r**3
            latex = f"V = \\frac{4}{3}\\pi r^3 = \\frac{4}{3}\\pi ({r})^3 = {sp.latex(ans)}"
            final = f"Volume = {sp.pretty(ans)} ≈ {round((4/3) * math.pi * r**3, 2)}"
            
        else: # Pythagoras
            a = random.randint(3, 20)
            b = random.randint(3, 20)
            c_sq = a**2 + b**2
            c = sp.sqrt(c_sq)
            
            question = f"In a right-angled triangle, the lengths of the two legs are a = {a} and b = {b}. Find the length of the hypotenuse c."
            steps = [
                f"1. According to the Pythagorean theorem, a^2 + b^2 = c^2.",
                f"2. Substitute a = {a} and b = {b}: {a}^2 + {b}^2 = c^2.",
                f"3. {a**2} + {b**2} = c^2.",
                f"4. {c_sq} = c^2.",
                f"5. c = √{c_sq}."
            ]
            latex = f"c = \\sqrt{{a^2 + b^2}} = \\sqrt{{{a}^2 + {b}^2}} = \\sqrt{{{c_sq}}} = {sp.latex(c)}"
            final = f"c = {sp.pretty(c)} ≈ {round(math.sqrt(c_sq), 2)}"

        records.append(MathKnowledgeSystem.format_record(
            topic=f"Geometry - {prob_type.replace('_', ' ').capitalize()}",
            question=question,
            steps=steps,
            final_answer=final,
            latex=latex
        ))
        
    return records
