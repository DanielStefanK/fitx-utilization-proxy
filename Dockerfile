FROM node:16.18.1-slim as frontend
ADD ./frontend /build/frontend
WORKDIR /build/frontend
RUN npm ci
RUN npm run build

FROM golang:1.19.3-bullseye
ADD . /app
COPY --from=frontend /build/frontend/dist/ /app/static/
WORKDIR /app
RUN go build -o main .

EXPOSE 8080
CMD ["/app/main"]