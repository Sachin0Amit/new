import csv
import os

def generate_synthetic_data(output_path="math_solver/data/synthetic_data.csv"):
    """
    Generates a synthetic dataset for training the classifier.
    This provides a baseline for all categories even if internet download fails.
    """
    data = [
        ("Solve for x: 2x + 5 = 10", "algebra"),
        ("Factorize x^2 - 5x + 6", "algebra"),
        ("Simplify the expression 3(x+2) - 4x", "algebra"),
        ("Calculate the integral of x^2 from 0 to 1", "calculus"),
        ("Find the derivative of sin(x) * cos(x)", "calculus"),
        ("What is the limit of (1/x) as x approaches infinity?", "calculus"),
        ("Find the determinant of the 3x3 matrix", "linear_algebra"),
        ("Calculate the dot product of [1,2,3] and [4,5,6]", "linear_algebra"),
        ("Find the area of a circle with radius 5", "geometry"),
        ("What is the hypotenuse of a right triangle with sides 3 and 4?", "geometry"),
        ("Calculate the probability of drawing an ace from a deck", "statistics"),
        ("Find the mean and standard deviation of the set [1, 2, 5, 10]", "statistics"),
        ("Is 17 a prime number?", "number_theory"),
        ("Solve the differential equation dy/dx = y", "calculus"),
        ("Expand (a + b)^3", "algebra"),
        ("Find the inverse of the matrix A", "linear_algebra"),
        ("Calculate the volume of a sphere with diameter 10", "geometry"),
        ("Perform a t-test on the following samples", "statistics")
    ]
    
    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    
    with open(output_path, 'w', newline='', encoding='utf-8') as f:
        writer = csv.writer(f)
        writer.writerow(["problem", "label"])
        # Create variations
        for text, label in data:
            writer.writerow([text, label])
            # Add some variations
            writer.writerow([f"Calculate: {text}", label])
            writer.writerow([f"Please solve {text.lower()}", label])
            
    print(f"[Synthetic] Generated {len(data)*3} problems at {output_path}")

if __name__ == "__main__":
    generate_synthetic_data()
