import sys
import os
import argparse

# Ensure domains can be imported
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from core import MathKnowledgeSystem
from domains.calculus import generate_calculus_problems
from domains.algebra import generate_algebra_problems
from domains.linear_algebra import generate_linear_algebra_problems
from domains.differential_equations import generate_ode_problems
from domains.number_theory import generate_number_theory_problems
from domains.geometry import generate_geometry_problems
from domains.trigonometry import generate_trigonometry_problems
from domains.statistics import generate_statistics_problems
from domains.probability import generate_probability_problems
from domains.logic import generate_logic_problems
from domains.combinatorics import generate_combinatorics_problems

def main():
    parser = argparse.ArgumentParser(description="Comprehensive Mathematics Knowledge System Generator")
    parser.add_argument("--batch-size", type=int, default=50, help="Number of problems to generate per batch")
    parser.add_argument("--num-batches", type=int, default=2, help="Number of batches to generate")
    parser.add_argument("--domain", type=str, default="all", help="Domain to generate")
    
    args = parser.parse_args()
    
    engine = MathKnowledgeSystem(output_dir="math_dataset")
    
    print(f"==================================================")
    print(f"  COMPREHENSIVE MATH ENGINE INITIALIZED")
    print(f"  Target: {args.batch_size * args.num_batches} problems per domain")
    print(f"==================================================\n")
    
    for batch_id in range(1, args.num_batches + 1):
        if args.domain in ["all", "calculus"]:
            calc_data = generate_calculus_problems(args.batch_size)
            engine.save_batch("calculus", calc_data, batch_id)
            
        if args.domain in ["all", "algebra"]:
            alg_data = generate_algebra_problems(args.batch_size)
            engine.save_batch("algebra", alg_data, batch_id)
            
        if args.domain in ["all", "linear_algebra"]:
            la_data = generate_linear_algebra_problems(args.batch_size)
            engine.save_batch("linear_algebra", la_data, batch_id)
            
        if args.domain in ["all", "ode"]:
            ode_data = generate_ode_problems(args.batch_size)
            engine.save_batch("ode", ode_data, batch_id)
            
        if args.domain in ["all", "number_theory"]:
            nt_data = generate_number_theory_problems(args.batch_size)
            engine.save_batch("number_theory", nt_data, batch_id)
            
        if args.domain in ["all", "geometry"]:
            geo_data = generate_geometry_problems(args.batch_size)
            engine.save_batch("geometry", geo_data, batch_id)
            
        if args.domain in ["all", "trigonometry"]:
            trig_data = generate_trigonometry_problems(args.batch_size)
            engine.save_batch("trigonometry", trig_data, batch_id)
            
        if args.domain in ["all", "statistics"]:
            stat_data = generate_statistics_problems(args.batch_size)
            engine.save_batch("statistics", stat_data, batch_id)
            
        if args.domain in ["all", "probability"]:
            prob_data = generate_probability_problems(args.batch_size)
            engine.save_batch("probability", prob_data, batch_id)
            
        if args.domain in ["all", "logic"]:
            logic_data = generate_logic_problems(args.batch_size)
            engine.save_batch("logic", logic_data, batch_id)
            
        if args.domain in ["all", "combinatorics"]:
            comb_data = generate_combinatorics_problems(args.batch_size)
            engine.save_batch("combinatorics", comb_data, batch_id)

    print("\n[SUCCESS] Generation complete.")
    print(f"The data is structured in JSONL format, ready for LLM fine-tuning.")
    print(f"Check the 'math_dataset' folder.")

if __name__ == "__main__":
    main()
