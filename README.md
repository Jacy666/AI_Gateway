# AI_Gateway

A simulated AI model service built with FastAPI.

## Features

- FastAPI-based REST API
- `/run_model` endpoint that simulates a slow AI model (10 seconds execution time)
- Health check endpoint at `/`

## Installation

1. Install dependencies:
```bash
pip install -r requirements.txt
```

## Running the Service

Start the FastAPI server with uvicorn:
```bash
uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

The service will be available at `http://localhost:8000`

## API Endpoints

### GET /
Health check endpoint that returns the service status.

**Response:**
```json
{
  "status": "ok",
  "message": "AI Gateway Service is running"
}
```

### POST /run_model
Simulates running an AI model. This endpoint sleeps for 10 seconds to simulate a slow model execution.

**Response:**
```json
{
  "status": "success",
  "message": "Model execution completed",
  "execution_time": "10 seconds"
}
```

## Testing

You can test the API using curl:

```bash
# Health check
curl http://localhost:8000/

# Run model (will take 10 seconds)
curl -X POST http://localhost:8000/run_model
```

Or visit the interactive API docs at `http://localhost:8000/docs`