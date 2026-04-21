# FastAPI

## App Setup

```python
from fastapi import FastAPI, HTTPException, Depends, Query, Path, UploadFile
from pydantic import BaseModel

app = FastAPI()

@app.get("/")
async def root():
    return {"message": "Hello"}

@app.get("/items/{item_id}")
async def read_item(item_id: int, q: str | None = None):
    return {"item_id": item_id, "q": q}

@app.post("/items")
async def create_item(item: Item):
    return item
```

## Pydantic Models

```python
class Item(BaseModel):
    name: str
    price: float
    is_offer: bool = False

class ItemResponse(BaseModel):
    id: int
    name: str
    created_at: datetime
```

Request/response 자동 검증. Type hints 필수.

## Path & Query Parameters

```python
@app.get("/items/{item_id}")
async def get_item(
    item_id: int = Path(..., gt=0),           # Path param with validation
    skip: int = Query(0, ge=0),               # Query param
    limit: int = Query(10, le=100)
):
    return {"item_id": item_id, "skip": skip}
```

## Dependency Injection

```python
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

@app.get("/users")
async def list_users(db: Session = Depends(get_db)):
    return db.query(User).all()
```

## async vs def

```python
async def async_endpoint():  # Use for I/O-bound (DB, API calls)
    data = await fetch_data()
    return data

def sync_endpoint():         # Use for CPU-bound or simple logic
    return {"data": compute()}
```

## Error Handling

```python
@app.get("/items/{item_id}")
async def read_item(item_id: int):
    if item_id not in db:
        raise HTTPException(status_code=404, detail="Item not found")
    return db[item_id]
```

## File Upload

```python
@app.post("/upload")
async def upload_file(file: UploadFile):
    content = await file.read()
    return {"filename": file.filename, "size": len(content)}
```

## CORS

```python
from fastapi.middleware.cors import CORSMiddleware

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"]
)
```

## Run

```bash
uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

**패턴**: Type hints everywhere. Pydantic for validation. Depends() for shared logic. async for I/O. Auto docs at /docs.
