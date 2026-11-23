FROM node:26.13.0@sha256:e7d748ae5e42807dde2a43559ed366ec6ecf8a4ebe79d31c3cb7da4eaf1282db
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
USER node
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]
