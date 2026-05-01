import sympy as sp
import random
from core import MathKnowledgeSystem

def generate_probability_problems(num_problems: int = 100) -> list:
    """Procedurally generates probability problems (Basic, Bayes, Distributions)."""
    records = []
    
    for _ in range(num_problems):
        prob_type = random.choice(['basic', 'bayes', 'dice'])
        
        if prob_type == 'basic':
            total = random.randint(20, 100)
            target = random.randint(1, total - 1)
            question = f"In a bag of {total} marbles, {target} are red. If you pick one at random, what is the probability it is red?"
            steps = [
                f"1. The probability of an event is (number of favorable outcomes) / (total number of outcomes).",
                f"2. Number of favorable outcomes (red marbles) = {target}.",
                f"3. Total outcomes = {total}.",
                f"4. P(Red) = {target}/{total}."
            ]
            ans = sp.Rational(target, total)
            latex = f"P(\\text{{Red}}) = \\frac{{{target}}}{{{total}}} = {sp.latex(ans)}"
            final = f"Probability = {ans} ≈ {round(target/total, 4)}"
            
        elif prob_type == 'bayes':
            p_a = random.randint(1, 10) / 100 # P(Disease)
            p_b_given_a = random.randint(90, 99) / 100 # P(Test+|Disease)
            p_b_given_not_a = random.randint(1, 5) / 100 # P(Test+|No Disease)
            
            question = f"A disease affects {p_a*100}% of the population. A test for it is {p_b_given_a*100}% accurate (true positive). The false positive rate is {p_b_given_not_a*100}%. If a person tests positive, what is the probability they have the disease?"
            steps = [
                "1. Apply Bayes' Theorem: P(A|B) = [P(B|A) * P(A)] / P(B).",
                f"2. P(A) = {p_a} (Probability of disease).",
                f"3. P(B|A) = {p_b_given_a} (True positive rate).",
                f"4. P(B|not A) = {p_b_given_not_a} (False positive rate).",
                f"5. Calculate P(B) (Total probability of testing positive): P(B) = P(B|A)P(A) + P(B|not A)P(not A).",
                f"6. P(B) = ({p_b_given_a} * {p_a}) + ({p_b_given_not_a} * {1-p_a})."
            ]
            p_b = (p_b_given_a * p_a) + (p_b_given_not_a * (1 - p_a))
            p_a_given_b = (p_b_given_a * p_a) / p_b
            
            ans = round(p_a_given_b, 4)
            latex = f"P(A|B) = \\frac{{{p_b_given_a} \\cdot {p_a}}}{{{p_b_given_a} \\cdot {p_a} + {p_b_given_not_a} \\cdot {1-p_a}}} = {ans}"
            final = f"Probability = {ans}"
            
        else: # Dice
            n = random.randint(1, 3)
            target_sum = random.randint(n, 6*n)
            question = f"If you roll {n} fair six-sided dice, what is the probability that the sum is exactly {target_sum}?"
            steps = [
                f"1. Determine the total number of outcomes: 6^{n} = {6**n}.",
                f"2. Find the number of ways to get a sum of {target_sum} using {n} dice.",
                f"3. Probability = (Ways) / (Total Outcomes)."
            ]
            
            # Simple way to count outcomes for small n
            def count_ways(dice, remaining_sum):
                if dice == 0: return 1 if remaining_sum == 0 else 0
                return sum(count_ways(dice - 1, remaining_sum - i) for i in range(1, 7))
            
            ways = count_ways(n, target_sum)
            ans = sp.Rational(ways, 6**n)
            latex = f"P(\\text{{Sum}}={target_sum}) = \\frac{{{ways}}}{{{6**n}}} = {sp.latex(ans)}"
            final = f"Probability = {ans} ≈ {round(ways/(6**n), 4)}"

        records.append(MathKnowledgeSystem.format_record(
            topic=f"Probability - {prob_type.capitalize()}",
            question=question,
            steps=steps,
            final_answer=final,
            latex=latex
        ))
        
    return records
