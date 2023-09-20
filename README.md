<a href="https://opensource.newrelic.com/oss-category/#new-relic-experimental"><picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/dark/Experimental.png"><source media="(prefers-color-scheme: light)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Experimental.png"><img alt="New Relic Open Source experimental project banner." src="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Experimental.png"></picture></a>


# Slack Logs to New Relic Log API
Send follwoing Slack API logs to New Relic's Log API. 🚧 This project is currently work in progress and supprots following logs collection
- [ChannelDetail](https://api.slack.com/methods/conversations.list)
- [UserLogs](https://api.slack.com/methods/users.list)
- [AccessLogs](https://api.slack.com/methods/team.accessLogs)

### Prerequisites
- Install slack APP with required permissions and collect user token. Current solution requires following permissions.<br> 
       admin, users:read, channels:read, teams:read

  Please refer [Development](#Development) if you need help to create a slack app

### Installation
Slack logs integration can be installed in two ways

#### Option 1: Docker Container
- Build docker image using [Dockerfile](https://github.com/newrelic-experimental/SlackLogsIntegration/blob/main/Dockerfile)
  ( `docker build  --tag slack-logger .` )  
- Refer [Configuration](#configuration) for available config options
- Start the application in side the container, with required params
```bash
docker run -e SLACK_ACCESS_TOKEN=<SlackToken> slack-logger -NRAccountId=<xyz4> -NREndpoint=https://log-api.newrelic.com/log/v1 
 -logLevel=info -userLogs -channelDetails -accessLogs
```
#### Option 2: Standalone binary
- Build binary from source code
```bash
  git clone https://github.com/newrelic-experimental/SlackLogsIntegration.git
  cd SlackLogsIntegration
  go mod download
  GOARCH=amd64 GOOS=linux go build -o /slackLogger internal/main.go
```
- Refer [Configuration](#configuration) for available config options
- Export SLACK_ACCESS_TOKEN=<slackToken>
- Start application directly on host
```bash
  /slackLogger -NRAccountId=<xyz4> -NREndpoint=https://log-api.newrelic.com/log/v1 -logLevel=debug -channelDetails -userLogs -accessLogs
```

- Prepare, but DO NOT run, the application to run as a detached process that can survive user logout. How to do this is beyond the scope of this document, here are some useful references:
  - [systemd on Linux] (http://tuxgraphics.org/npa/systemd-scripts/)
  - [User defined Service on Windows] (https://learn.microsoft.com/en-us/troubleshoot/windows-client/deployment/create-user-defined-service)
  - [Docker on Linux] (https://linux.how2shout.com/how-to-start-docker-container-automatically-on-boot-in-linux/)

### Configuration
Configuration with defaults is self-describing for this application:
```bash
Usage of /slackLogger:
  -NRAccountId string
    	If set, sends logs to New Relic (default "xyz1234")
  -NREndpoint string
    	New Relic log endpoint (default "https://log-api.newrelic.com/log/v1")
  -accessLogs
    	Fetch access logs
  -channelDetails
    	Fetch channel details
  -flushInterval int
    	Flush interval in minutes (default 1440)
  -logLevel string
    	Golang slog log level: debug | info | warn | error (default "info")
  -userLogs
    	Fetch user logs
```

### Browse your Log data in NR
- [Login into One New Relic](https://one.newrelic.com)
- Open `Query Your Data` ![Alt text](./images/nr1-step-1.png)
- Query the data using NRQL ![Alt text](./images/nr1-step-2.png) 
  - select * from Log  where logtype='ChannelDetail' since 1 day ago

## Troubleshooting
- Please check whether Slack app has installed with proper permissions.
- Please refer FAQ document

## Development
- [Create new slack APP](https://api.slack.com/start/quickstart)
- [Slack access token authentication](https://api.slack.com/authentication/oauth-v2)
- Verify token validay by triggering respective [API call](https://api.slack.com/methods/conversations.list/test)    

## Support

New Relic has open-sourced this project. This project is provided AS-IS WITHOUT WARRANTY OR DEDICATED SUPPORT. Issues and contributions should be reported to the project here on GitHub.

We encourage you to bring your experiences and questions to the [Explorers Hub](https://discuss.newrelic.com) where our community members collaborate on solutions and new ideas.


## Contributing

We encourage your contributions to improve this project! Keep in mind when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project. If you have 
any questions, or to execute our corporate CLA, required if your contribution is on behalf of a company, please drop us an email at opensource@newrelic.com.

**A note about vulnerabilities**

As noted in our [security policy](../../security/policy), New Relic is committed to the privacy and security of our customers and their data. We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).


## License

This project is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.

