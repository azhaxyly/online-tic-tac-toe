# 1
FROM node:22-slim AS builder

WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# 2
FROM node:22-slim

WORKDIR /app
RUN npm install -g serve
COPY --from=builder /app/dist /app/dist
EXPOSE 5173
CMD ["serve", "-s", "dist", "-l", "5173"]
