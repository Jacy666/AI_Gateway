import time
from fastapi import FastAPI

app = FastAPI(title="AI Gateway Service", description="A simulated AI model service")


@app.get("/")
def read_root():
    """Health check endpoint"""
    return {"status": "ok", "message": "AI Gateway Service is running"}


@app.post("/run_model")
def run_model():
    """
    Simulates running an AI model that takes 10 seconds to complete.
    This endpoint uses time.sleep(10) to simulate a slow model execution.
    """
    # Simulate a slow AI model
    time.sleep(10)
    
    return {
        "status": "success",
        "message": "Model execution completed",
        "execution_time": "10 seconds"
    }
