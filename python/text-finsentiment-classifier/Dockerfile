FROM python:3.11-slim-bullseye

# Set the working directory inside the container
WORKDIR /app

# Copy only the requirements file first to leverage Docker's layer caching
COPY requirements.txt .

# Create the virtual environment
RUN python3 -m venv .venv

# Activate the virtual environment and install dependencies
RUN . .venv/bin/activate && pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application code
COPY . .

CMD [".venv/bin/python", "main.py"]

# TODO: add distroless stage
