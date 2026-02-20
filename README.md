# 🧮 Distributed Calculator
A distributed calculator with support for complex expressions, asynchronous processing, and computation history.
It can calculate even ~(~2) + 3 and remembers that 2 + 2 * 2 = 6.

## ⚙️ Environment Requirements
Before running the project, make sure you have installed:

`Docker` and `Docker Compose` (for running)
## 🚀 Running the Project
### 1. Environment Setup
```bash
# Create .env file based on the example
cp .env.example .env
```
### 2. Running the Application
```bash
# Build the image
docker compose build --no-cache

# Run the stack
docker compose up
```
## 🔨 Technology Stack
| Component | Technology |
| :---:  | :---:     |
| Messaging | `Kafka` (3-node cluster)|
| Cache | `Redis` |
| Database | `PostgreSQL` |
| API | `gRPC` + `HTTP/JSON` |
| Authorization | `JWT` |
| Frontend | `React` + `Vite` + `Tailwind CSS` |
| Build | `Docker Compose` |

## 🧠 How It Works
1. User enters an expression in the web interface
2. Frontend sends a request to the `Gateway`
3. Gateway validates `JWT` and parses the expression
4. The task is broken down into steps and sent to `Kafka`
5. Workers process the steps, storing intermediate results in `Redis`
6. The final result is saved to `PostgreSQL`
7. User receives the result or an error message

## 📡 API Endpoints
| Method | URL | Description |
| :---: | :---: | :---: |
| `POST` | `/v1/calculate` | Start calculating an expression |
| `POST` | `/v1/result`    | Returns result by `task_id` |
| `POST` | `/v1/examples` | Returns computation history of the user |
| `POST` | `/v1/register` | User registration | 
| `POST` | `/v1/login`    | Authorization and JWT retrieval |

### 💡 Usage Examples
✅ Example: User Registration </br>
Request
```bash
curl --location 'http://localhost:8080/v1/register' \
--header 'Content-Type: application/json' \
--data-raw '{
  "email": "user@example.com",
  "password": "mysecretpassword123"
}'
```
Response
```json
{
  "success": true,
  "error": ""
}
```
✅ Example: User Login </br>
Request
```bash
curl --location 'http://localhost:8080/v1/login' \
--header 'Content-Type: application/json' \
--data-raw '{
  "email": "user@example.com",
  "password": "mysecretpassword123"
}'
```
Response
```json
{
    "success": true,
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZDk0ODE5NzEtMjMzMy00MzE2LWFkZDYtYWE1NmFlYzc1OTgxIiwiaXNzIjoiZGlzdHJpYnV0ZWRfY2FsY3VsYXRvciIsImV4cCI6MTc1Mzk4OTAxM30.pEd-1x3AfYreT4gzRWeo-oBcUzGdfDaBlUe31HIWddA",
    "userId": "d9481971-2333-4316-add6-aa56aec75981",
    "error": ""
}

```
✅ Example: Calculate an Expression <br>
Request 
```bash
curl --location 'http://localhost:8080/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: ••••••' \
--data '{
    "expression": "1000-7"
}'
```
Response
```json
{
    "taskId": "c352230c-802e-4158-b528-5b2365481179"
}

```
✅ Example: Get Result </br>
Request 
```bash
curl --location 'http://localhost:8080/v1/result' \
--header 'Content-Type: application/json' \
--header 'Authorization: ••••••' \
--data '{
    "taskId": "c352230c-802e-4158-b528-5b2365481179"
}'

```
Response
```json
{
    "value": 993
}
```
✅ Example: Computation History </br>
Request
```bash
curl --location --request POST 'http://localhost:8080/v1/examples' \
--header 'Authorization: ••••••'
```
Response
```json
{
    "examples": [
        {
            "id": "2b19f0d8-5674-4de3-b0fc-1ad09db15572",
            "expression": "Hello, World!",
            "calculated": true,
            "createdAt": "2025-07-30T19:15:40Z",
            "error": "line is not a mathematical expression or contains an error"
        },
        {
            "id": "8f70e303-44e5-46de-b1da-13f460d455af",
            "expression": "7 / 0",
            "calculated": true,
            "createdAt": "2025-07-30T19:15:21Z",
            "error": "division by zero"
        },
        {
            "id": "d136bce9-06f6-470d-a088-bda9a2e132be",
            "expression": "~(~3) + 8 ^ 0",
            "calculated": true,
            "result": 4,
            "createdAt": "2025-07-30T19:15:00Z"
        },
        {
            "id": "a8f71354-fb00-4034-a9ea-9515caf7bd77",
            "expression": "~3 + 8",
            "calculated": true,
            "result": 5,
            "createdAt": "2025-07-30T19:14:28Z"
        },
        {
            "id": "c352230c-802e-4158-b528-5b2365481179",
            "expression": "1000-7",
            "calculated": true,
            "result": 993,
            "createdAt": "2025-07-30T19:11:37Z"
        }
    ]
}
```
## 🗂️ Project Structure
```
distributed_calculator2/
├── cmd/
│   ├── main/        # Main server (gRPC + REST)
│   └── worker/      # Worker for task processing
├── internal/
│   ├── auth/        # JWT authorization
│   ├── models/      # Data models
│   ├── repository/  # Repositories (Postgres, Redis)
│   ├── service/     # Business logic
│   └── worker/      # Worker logic
├── pkg/
│   ├── api/         # gRPC proto
│   ├── config/      # Configuration
│   ├── db/          # Database connections
│   ├── logger/      # Logging
│   ├── messaging/   # Kafka
│   └── valueprovider/ # Value retrieval
├── migrations/      # Database migrations
├── my-calculator/   # React frontend
├── docker-compose.yml
├── Dockerfile
└── .env.example
```
## 🗃️ Database Structure
### Table examples
| Field | Type | Description |
| :---: | :---: | :---: |
|`id`|`TEXT`|Unique expression `ID`
|`expression`|`TEXT`|Original expression
|`response`|`TEXT`|Final variable
|`user_id`|`TEXT`|User `ID`
|`calculated`|`BOOLEAN`|Calculation completed
|`error`|`TEXT`|Error (if any)
|`created_at`|`TIMESTAMPTZ`|Creation time
|`updated_at`|`TIMESTAMPTZ`|Update time
### Table users
| Field | Type | Description |
| :---: | :---: | :---: |
|`id`|`TEXT`|Unique user ID
|`email`|`TEXT`|User email
|`password_hash`|`TEXT`|Password hash
|`role`|`TEXT`|Role (user/admin)
|`created_at`|`TIMESTAMPTZ`|Registration time
|`updated_at`|`TIMESTAMPTZ`|Update time

## 🧩 Implementation Features
1. Unary minus through ~
```
// ~5 becomes (0-5)
// ~(~2) + 3 = 5
```
2. Division by zero handling
```
5 / 0 → error "division by zero"
```
3. Asynchronous processing
* Expression is broken down into steps
* Each step is sent to `Kafka`
* Workers process steps in parallel
* Result is assembled from intermediate values
4. Support for complex expressions
```
~(~2) + 3 * (4 - 1) ^ 2
```
## 🖥️ Frontend
Frontend on `React` with a dark theme (black background, purple accents):

* Home page — project description and technologies
* Calculator — input expressions and get results
* History — view previous calculations
* Authorization — registration and login
* Available at [`http://localhost:3000`](http://localhost:3000) after running 
```bash
npm run dev
```

### 🛠️ How Expression Is Calculated
1. User enters expression: `~(~2) + 3`
2. System parses it to reverse Polish notation:
```
2 ~ 2 ~ 3 + → 2 (0 - 2) (0 - 3) + 
```
#### Breaks into steps:
* Step 1: `~2 = -2`
* Step 2: `~(-2) = 2`
* Step 3: `2 + 3 = 5`
* Each step is sent to `Kafka`
* Workers process steps and store results in `Redis`
* Final result is saved to `PostgreSQL`
## 📊 Workers Monitoring
The frontend has a "`Workers`" section that displays:

* Worker status (online/offline)
* Number of tasks processed
* Current load
* Time of last activity
## 🔐 Security
* All requests require `JWT` authorization (except registration and login)
* Passwords are stored in hashed form (`bcrypt`)
* Input validation at all stages
* No use of `eval()` — safe expression parsing
## 📁 Docker Compose
The project uses powerful `Docker Compose` with:

* 3-node `Kafka` cluster (`KRaft`)
* `Redis` for caching
* `PostgreSQL` for storing history
* Automatic migrations
* Creating `Kafka` topics on startup
* Building frontend on `React`