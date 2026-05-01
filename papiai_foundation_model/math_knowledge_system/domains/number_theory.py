import sympy as sp
import random
from core import MathKnowledgeSystem

def generate_number_theory_problems(num_problems: int = 100) -> list:
    """Procedurally generates Number Theory problems (GCD, Prime Factors, Mod Inverse)."""
    records = []
    
    for _ in range(num_problems):
        prob_type = random.choice(['gcd', 'prime_factors', 'mod_inverse'])
        
        if prob_type == 'gcd':
            a, b = random.randint(100, 10000), random.randint(100, 10000)
            question = f"Find the Greatest Common Divisor (GCD) of {a} and {b}."
            steps = [
                f"1. We need to find GCD({a}, {b}).",
                "2. Apply the Euclidean algorithm.",
                "3. Divide the larger number by the smaller number and find the remainder.",
                "4. Repeat the process with the divisor and the remainder until the remainder is 0.",
                "5. The last non-zero remainder is the GCD."
            ]
            ans = sp.gcd(a, b)
            latex = f"\\gcd({a}, {b}) = {ans}"
            final = f"GCD = {ans}"
            
        elif prob_type == 'prime_factors':
            n = random.randint(1000, 100000)
            question = f"Find the prime factorization of {n}."
            steps = [
                f"1. We need to find the prime factors of {n}.",
                "2. Systematically divide the number by prime numbers starting from 2.",
                "3. Continue dividing until the quotient is 1."
            ]
            ans = sp.factorint(n)
            factors = " * ".join([f"{p}^{e}" if e > 1 else f"{p}" for p, e in ans.items()])
            latex_factors = " \\cdot ".join([f"{p}^{{{e}}}" if e > 1 else f"{p}" for p, e in ans.items()])
            latex = f"{n} = {latex_factors}"
            final = f"{n} = {factors}"
            
        else: # mod inverse
            a = random.randint(2, 50)
            m = random.randint(51, 200)
            while sp.gcd(a, m) != 1:
                a = random.randint(2, 50)
                m = random.randint(51, 200)
                
            question = f"Find the modular inverse of {a} modulo {m}."
            steps = [
                f"1. We need to find x such that {a}x ≡ 1 (mod {m}).",
                "2. Use the Extended Euclidean Algorithm.",
                f"3. Find coefficients s and t such that {a}s + {m}t = 1.",
                "4. The modular inverse is s (mod m)."
            ]
            ans = sp.mod_inverse(a, m)
            latex = f"{a}^{{-1}} \\pmod{{{m}}} \\equiv {ans}"
            final = f"Inverse = {ans}"

        records.append(MathKnowledgeSystem.format_record(
            topic=f"Number Theory - {prob_type.replace('_', ' ').capitalize()}",
            question=question,
            steps=steps,
            final_answer=final,
            latex=latex
        ))
        
    return records
