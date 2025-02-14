FROM node:18-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the rest of the project files
COPY ./app/ .

# Install dependencies
RUN npm install

CMD ["npm", "run", "dev", "--", "--host"]