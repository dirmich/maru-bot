# MaruBot REST API Documentation

This document describes the REST API endpoints provided by the MaruBot embedded server. These APIs are used by the Web Dashboard (Admin Panel) to interact with the MaruBot engine.

## Base URL
The API is served at `http://<host>:8080/api`.

---

## Chat

### Get Chat History (Mock)
Retrieves the chat history (currently returns a mock/empty list as persistence is not fully implemented for web chat).

- **Endpoint**: `GET /api/chat`
- **Response**: `200 OK`
  ```json
  []
  ```

### Send Message
Sends a message to the MaruBot agent and receives a response.

- **Endpoint**: `POST /api/chat`
- **Content-Type**: `application/json`
- **Request Body**:
  ```json
  {
    "message": "Hello, MaruBot!"
  }
  ```
- **Response**: `200 OK`
  ```json
  {
    "response": "Hello! How can I assist you today?"
  }
  ```

---

## Configuration

### Get Configuration
Retrieves the current MaruBot configuration, including agent settings and provider keys.

- **Endpoint**: `GET /api/config`
- **Response**: `200 OK`
  ```json
  {
    "agents": {
      "defaults": {
        "model": "gemini-1.5-pro",
        "workspace": "~/.marubot/workspace"
      }
    },
    "providers": {
      "openai": { "api_key": "sk-...", "api_base": "" },
      "gemini": { "api_key": "...", "api_base": "" }
    }
  }
  ```

### Update Configuration (Planned)
Updates the configuration. (Implementation in progress)

- **Endpoint**: `POST /api/config`
- **Content-Type**: `application/json`
- **Request Body**: (Same structure as GET response)
- **Response**: `200 OK` on success.

---

## Skills

### List Installed Skills
Lists all skills currently installed in the workspace.

- **Endpoint**: `GET /api/skills`
- **Response**: `200 OK`
  ```json
  {
    "output": "- weather (github.com/sipeed/marubot-skills/weather)\n- news (builtin)\n"
  }
  ```
  *(Currently returns formatted text output for CLI compatibility, will migrate to structured JSON object in future versions)*.

### Install/Remove Skill
Installs or removes a skill.

- **Endpoint**: `POST /api/skills`
- **Content-Type**: `application/json`
- **Request Body**:
  ```json
  {
    "action": "install", // or "remove"
    "skill": "sipeed/marubot-skills/weather"
  }
  ```
- **Response**: `200 OK`
  ```json
  {
    "stdout": "✓ Skill 'weather' installed successfully!\n",
    "stderr": ""
  }
  ```

---

## GPIO (Planned)

### Get GPIO Status
Retrieves the current status of GPIO pins.

- **Endpoint**: `GET /api/gpio`
- **Response**: `200 OK`
  ```json
  {
    "pins": [
      { "pin": 7, "mode": "OUT", "value": 1, "label": "Status LED" }
    ]
  }
  ```

### Set GPIO Pin
Controls a specific GPIO pin.

- **Endpoint**: `POST /api/gpio`
- **Request Body**:
  ```json
  {
    "pin": 7,
    "value": 0
  }
  ```

