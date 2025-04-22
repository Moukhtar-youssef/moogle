FROM python:3.13-slim

# Set working directory
WORKDIR /app

# Set NLTK environment so it saves data in a known location
ENV NLTK_DATA=/usr/local/nltk_data

# Copy and install dependencies
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Download required NLTK resources ahead of time
RUN python3 -m nltk.downloader -d /usr/local/nltk_data stopwords punkt punkt_tab

# Copy the rest of the application code
COPY . .

# Optional: add NLTK path in case your code runs before downloading
ENV PYTHONPATH=/usr/local/nltk_data

# Run the application
CMD ["python", "./main.py"]
