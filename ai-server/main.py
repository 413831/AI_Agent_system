from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()

class Request(BaseModel):
    prompt: str

@app.post("/ask")
def ask_ai(req: Request):
    # Simulación AI
    return {
        "result": f"AI response to: {req.prompt}"
    }