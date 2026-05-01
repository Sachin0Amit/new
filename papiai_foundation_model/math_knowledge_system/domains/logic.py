import sympy as sp
import random
from core import MathKnowledgeSystem

def generate_logic_problems(num_problems: int = 100) -> list:
    """Procedurally generates logic and set theory problems (Truth tables, Sets)."""
    records = []
    
    for _ in range(num_problems):
        prob_type = random.choice(['truth_table', 'set_ops'])
        
        if prob_type == 'truth_table':
            # Logic: P AND Q, P OR Q, P XOR Q, P -> Q
            p, q = sp.symbols('P Q')
            op_name = random.choice(['AND', 'OR', 'XOR', 'Implies'])
            
            if op_name == 'AND': expr = p & q
            elif op_name == 'OR': expr = p | q
            elif op_name == 'XOR': expr = p ^ q
            else: expr = sp.Implies(p, q)
            
            question = f"Construct the truth table for the logical expression: {p} {op_name} {q}"
            steps = [
                f"1. Create columns for P and Q with all possible truth value combinations: (T,T), (T,F), (F,T), (F,F).",
                f"2. Apply the {op_name} operator rule to each pair.",
                f"3. Fill in the resulting column for {p} {op_name} {q}."
            ]
            
            # Simple manual truth table result
            results = []
            for pv in [True, False]:
                for qv in [True, False]:
                    val = expr.subs({p: pv, q: qv})
                    results.append(f"({pv},{qv}) -> {val}")
            
            ans = " | ".join(results)
            latex = f"P \\text{{ {op_name} }} Q"
            final = f"Truth Table: {ans}"
            
        else: # Set Operations
            set_a = set(random.sample(range(1, 20), random.randint(4, 8)))
            set_b = set(random.sample(range(1, 20), random.randint(4, 8)))
            op = random.choice(['Union', 'Intersection', 'Difference'])
            
            question = f"Given sets A = {set_a} and B = {set_b}, find the {op} of A and B."
            steps = [
                f"1. To find the Union (A ∪ B), combine all unique elements from both sets.",
                f"2. To find the Intersection (A ∩ B), find the elements that appear in both sets.",
                f"3. To find the Difference (A - B), take elements in A that are NOT in B."
            ]
            
            if op == 'Union': res = set_a.union(set_b)
            elif op == 'Intersection': res = set_a.intersection(set_b)
            else: res = set_a.difference(set_b)
            
            ans = sorted(list(res))
            latex = f"A \\text{{ {op} }} B = {ans}"
            final = f"{op} = {ans}"

        records.append(MathKnowledgeSystem.format_record(
            topic=f"Logic/Set Theory - {prob_type.capitalize()}",
            question=question,
            steps=steps,
            final_answer=final,
            latex=latex
        ))
        
    return records
