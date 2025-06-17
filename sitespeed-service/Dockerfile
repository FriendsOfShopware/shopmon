FROM sitespeedio/sitespeed.io:latest

# Install Node.js for our API server
USER root
RUN apt-get update && apt-get install -y curl && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs

# Create app directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy application code
COPY . .

# Create results directory with proper permissions
RUN mkdir -p /app/results && chmod 755 /app/results

# Expose port
EXPOSE 3001

# Override entrypoint and use absolute path to node
ENTRYPOINT []
CMD ["/usr/bin/node", "server.js"]