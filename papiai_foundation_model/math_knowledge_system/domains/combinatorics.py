import sympy as sp
import random
import math
from core import MathKnowledgeSystem

def generate_combinatorics_problems(num_problems: int = 100) -> list:
    """Procedurally generates combinatorics problems (Permutations, Combinations)."""
    records = []
    
    for _ in range(num_problems):
        prob_type = random.choice(['permutation', 'combination'])
        
        n = random.randint(5, 15)
        r = random.randint(1, n)
        
        if prob_type == 'permutation':
            question = f"How many ways can you arrange {r} items out of a set of {n} distinct items (Permutations P(n, r))?"
            steps = [
                f"1. The formula for permutations is P(n, r) = n! / (n-r)!.",
                f"2. Substitute n = {n} and r = {r}.",
                f"3. P({n}, {r}) = {n}! / ({n}-{r})! = {n}! / {n-r}!."
            ]
            ans = math.perm(n, r)
            latex = f"P({n}, {r}) = \\frac{{{n}!}}{{({n}-{r})!}} = {ans}"
            final = f"Permutations = {ans}"
            
        else: # Combination
            question = f"How many ways can you choose {r} items out of a set of {n} distinct items (Combinations C(n, r))?"
            steps = [
                f"1. The formula for combinations is C(n, r) = n! / (r!(n-r)!).",
                f"2. Substitute n = {n} and r = {r}.",
                f"3. C({n}, {r}) = {n}! / ({r}! * ({n}-{r})!) = {n}! / ({r}! * {n-r}!)."
            ]
            ans = math.comb(n, r)
            latex = f"C({n}, {r}) = \\binom{{{n}}}{{{r}}} = \\frac{{{n}!}}{{{r}!({n}-{r})!}} = {ans}"
            final = f"Combinations = {ans}"

        records.append(MathKnowledgeSystem.format_record(
            topic=f"Combinatorics - {prob_type.capitalize()}",
            question=question,
            steps=steps,
            final_answer=final,
            latex=latex
        ))
        
    return records
