from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Any
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
from core.orchestrator import SolutionOrchestrator

app = FastAPI(
    title="Universal Math Intelligence API",
    description="A step-by-step mathematical reasoning and solving API.",
    version="1.0.0"
)

# Enable CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

orchestrator = SolutionOrchestrator()

class SolveRequest(BaseModel):
    input: str

class SolveResponse(BaseModel):
    problem_type: str
    solution: Any
    steps: list
    final_answer: str
    plots: str = None

@app.post("/solve")
async def solve_problem(request: SolveRequest):
    try:
        result = orchestrator.solve(request.input)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    return {"status": "active", "engine": "symbolic+numeric"}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
