import sympy as sp
import random
import statistics
from core import MathKnowledgeSystem

def generate_statistics_problems(num_problems: int = 100) -> list:
    """Procedurally generates statistics problems (Mean, Median, Variance)."""
    records = []
    
    for _ in range(num_problems):
        prob_type = random.choice(['central_tendency', 'variance'])
        
        # Generate a random sample of data
        sample_size = random.randint(5, 15)
        data = [random.randint(1, 100) for _ in range(sample_size)]
        
        if prob_type == 'central_tendency':
            question = f"For the following dataset, find the mean and median: {data}"
            steps = [
                f"1. To find the mean, sum all the elements and divide by the count ({sample_size}).",
                f"2. To find the median, sort the data and find the middle value."
            ]
            mean_val = statistics.mean(data)
            median_val = statistics.median(data)
            
            latex = f"\\mu = \\frac{{\\sum x_i}}{{n}} = {mean_val}, \\text{{Median}} = {median_val}"
            final = f"Mean = {mean_val}, Median = {median_val}"
            
        else: # Variance & Std Dev
            question = f"Calculate the population variance and standard deviation for the dataset: {data}"
            steps = [
                f"1. Calculate the mean (μ).",
                f"2. For each number, subtract the mean and square the result.",
                f"3. Find the mean of those squared differences (Variance σ²).",
                f"4. Take the square root of the variance (Standard Deviation σ)."
            ]
            var_val = round(statistics.pvariance(data), 2)
            stdev_val = round(statistics.pstdev(data), 2)
            
            latex = f"\\sigma^2 = {var_val}, \\sigma = {stdev_val}"
            final = f"Variance = {var_val}, Std Dev = {stdev_val}"

        records.append(MathKnowledgeSystem.format_record(
            topic=f"Statistics - {prob_type.replace('_', ' ').capitalize()}",
            question=question,
            steps=steps,
            final_answer=final,
            latex=latex
        ))
        
    return records
