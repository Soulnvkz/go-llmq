# LLM Chat System with Queue

**Full-stack implementation for LLM chat with RabbitMQ queue system**  
*Purpose: scalable solution for deploying multiple LLM agents across machines*

## Quick Start Guide

### Prerequisites
- Linux environment
- Docker installed
- LLM model files (GGUF format requred)

---

### Setup

1. **Clone Repository**

2. **Update submodule**
   ```bash
   git submodule update --init --recursive
   ```

2. **Configure Docker**  
   Edit `fullstack-compose.yaml`:
   ```yaml
   services:
     llm:
       volumes:
         - /path/to/your/models:/app/models  # Mount model directory
       environment:
         - MODEL_PATH=/app/models/your_model.gguf  # Specify model filename
   ```

3. **Launch System**
   ```bash
   docker compose -f fullstack-compose.yaml up --build -d
   ```

---

### Access Interface
Open `http://localhost:5000` in your browser

---

## License

MIT License