FROM oven/bun:latest

WORKDIR /app

# Copy everything from current directory
COPY . .

# Install dependencies
RUN bun install

# Elysia usually runs on 3000 by default
EXPOSE 3000

CMD ["bun", "run", "src/index.ts"]