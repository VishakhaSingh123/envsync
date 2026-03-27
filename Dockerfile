# Use a lightweight Python 3.11 image
FROM python:3.11-slim

# Prevent Python from writing pyc files and enable unbuffered logging
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# Set the working directory inside the container
WORKDIR /app

# Copy the requirements file and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application code
COPY . .

# Set a fallback encryption key (should be overridden at runtime)
ENV ENVSYNC_KEY="docker_fallback_key_32_bytes_long"

# Make the main script the entrypoint
ENTRYPOINT ["python", "main.py"]
