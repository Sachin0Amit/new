from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Dict, Any, Optional
import sympy as sp
from .solver import SymbolicSolver

app = FastAPI(title="Sovereign Math Solver", version="1.0.0")
solver = SymbolicSolver()

class SolveRequest(BaseModel):
    expression: Optional[str] = None
    matrix: Optional[List[List[Any]]] = None
    variables: List[str] = ["x"]
    options: Dict[str, Any] = {}

@app.get("/health")
async def health():
    return {
        "status": "ok",
        "sympy_version": sp.__version__
    }

@app.post("/solve/algebra")
async def solve_algebra(req: SolveRequest):
    try:
        return solver.solve_algebra(req.expression, req.variables, req.options)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.post("/solve/calculus")
async def solve_calculus(req: SolveRequest, op: str = "diff"):
    try:
        return solver.solve_calculus(req.expression, req.variables, op)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.post("/solve/linear-algebra")
async def solve_linear_algebra(req: SolveRequest, op: str = "determinant"):
    try:
        return solver.solve_linear_algebra(req.matrix, op)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))
