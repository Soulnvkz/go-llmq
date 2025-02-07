FROM node:18-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy package.json and package-lock.json before copying the entire project
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the project files
COPY . .
