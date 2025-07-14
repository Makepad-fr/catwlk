FROM python:3.11-slim

WORKDIR /app

# Install git first, then pip dependencies
RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

RUN pip install torch pillow git+https://github.com/openai/CLIP.git

COPY main.py .

CMD ["python", "main.py"]
