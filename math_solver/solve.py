import argparse
import json
from core.orchestrator import SolutionOrchestrator

def main():
    parser = argparse.ArgumentParser(description="Universal Math Intelligence CLI")
    parser.add_argument("--input", type=str, required=True, help="The math problem to solve")
    parser.add_argument("--json", action="store_true", help="Output result as JSON")
    
    args = parser.parse_args()
    
    orchestrator = SolutionOrchestrator()
    result = orchestrator.solve(args.input)
    
    if args.json:
        print(json.dumps(result, indent=2))
    else:
        print("\n" + "="*50)
        print(f"Problem Type: {result['problem_type'].upper()}")
        print("="*50)
        print(f"\nQUERY: {args.input}")
        print("\nSTEPS:")
        for i, step in enumerate(result['steps'], 1):
            print(f"{i}. {step['description']}")
            print(f"   => {step['expression']}")
        
        print("\n" + "-"*50)
        print(f"FINAL ANSWER: {result['final_answer']}")
        print("-"*50 + "\n")
        
        if result['plots']:
            print("[System] Visual diagram generated and available in API response.")

if __name__ == "__main__":
    main()
