FROM golang:1.21-alpine AS build-stage

WORKDIR /build

# The Go image does not include git, add it to Alpine
RUN apk add git

RUN git clone https://github.com/newrelic-experimental/SlackLogsIntegration.git

WORKDIR SlackLogsIntegration

COPY SlackConfig.yaml /

# Build the executable
RUN GOARCH=amd64 GOOS=linux go build -o /slackLogger internal/main.go

# Step 2: Deploy the application binary into a lean image
FROM alpine AS build-release-stage

WORKDIR /

COPY --from=build-stage /slackLogger /slackLogger

COPY --from=build-stage /SlackConfig.yaml /SlackConfig.yaml

ENTRYPOINT ["/slackLogger" ]
