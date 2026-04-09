# Security Essentials

## OWASP Top 10 Summary

### 1. Injection

**SQL Injection**:
```python
# ❌ Bad
query = f"SELECT * FROM users WHERE id = {user_id}"

# ✅ Good (parameterized)
cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))
```

**NoSQL Injection**:
```js
// ❌ Bad
db.users.find({ username: req.body.username })

// ✅ Good
db.users.find({ username: { $eq: req.body.username } })
```

### 2. Cross-Site Scripting (XSS)

```js
// ❌ Bad
document.innerHTML = userInput

// ✅ Good (escape)
document.textContent = userInput
// Or use DOMPurify.sanitize(userInput)
```

**Content Security Policy (CSP)**:
```http
Content-Security-Policy: default-src 'self'; script-src 'self' 'nonce-random123'
```

### 3. CSRF (Cross-Site Request Forgery)

```python
# Token validation
@app.before_request
def csrf_protect():
    if request.method == "POST":
        token = session.get('csrf_token')
        if token != request.form.get('csrf_token'):
            abort(403)
```

**SameSite cookie**:
```python
response.set_cookie('session', value, samesite='Lax', secure=True, httponly=True)
```

### 4. Broken Authentication

```python
# ✅ Password hashing
import bcrypt
hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt())

# ✅ JWT rotation
access_token = create_token(user_id, expires_in=900)  # 15min
refresh_token = create_token(user_id, expires_in=604800)  # 7days
```

**Multi-factor authentication (MFA)** 권장.

### 5. SSRF (Server-Side Request Forgery)

```python
# ✅ URL allowlist
ALLOWED_HOSTS = ['api.example.com', 'cdn.example.com']

def fetch_url(url):
    parsed = urlparse(url)
    if parsed.hostname not in ALLOWED_HOSTS:
        raise SecurityError("Host not allowed")
    return requests.get(url)
```

## Secret Management

```bash
# ❌ Never hardcode
API_KEY = "sk-1234567890abcdef"

# ✅ Use environment variables
API_KEY = os.getenv('API_KEY')
```

**.env file in .gitignore**. Use vault (AWS Secrets Manager, HashiCorp Vault) in production.

## HTTPS Everywhere

```nginx
# Force HTTPS redirect
server {
    listen 80;
    return 301 https://$host$request_uri;
}
```

**HSTS header**:
```http
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

## Input Validation

```python
from pydantic import BaseModel, validator

class UserInput(BaseModel):
    email: str
    age: int

    @validator('age')
    def validate_age(cls, v):
        if v < 0 or v > 150:
            raise ValueError('Invalid age')
        return v
```

**Validate at boundaries** (API entry, form submit). Sanitize output (HTML escape, URL encode).

## Principle of Least Privilege

- DB user: read-only for queries, write for mutations
- File permissions: 600 for secrets, 644 for public
- IAM roles: minimal required permissions

**원칙**: Never trust user input. Defense in depth (multiple layers). Security by default.
