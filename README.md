# go-mailshield

## Motivation
This project was created for two main reasons:

* **Learning Project:** As a practical project to learn and deepen my Go programming skills. Working with IMAP, containers, and concurrent programming provides an excellent opportunity to apply various aspects of Go in practice.
* **Practical Problem:** My email provider unfortunately doesn't offer sufficient spam filtering mechanisms. With this tool, I can use SpamAssassin as an additional filtering layer to keep my inbox clean.

## What does go-mailshield do?
go-mailshield connects to your IMAP accounts, regularly checks for new emails, and forwards them to SpamAssassin for spam detection. Emails identified as spam are moved to the configured spam folder.

## Features
* Monitoring of multiple IMAP accounts
* Integration with SpamAssassin for reliable spam detection
* Fully containerized solution with Docker
* Configurable check intervals
* Option to check only unread emails

## Installation

### Prerequisites

* Docker and Docker Compose
* Go 1.22+ (for development only)

### Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/go-mailshield.git
cd go-mailshield
```

2. Create a configuration file:
```bash
cp config.yaml.dist config.yaml
```

3. Adjust the configuration (see Configuration section)

4. Start with Docker:
```bash
docker compose up -d
```

### Configuration
Edit the config.yaml according to your needs:
```yaml
poll_interval: 30s  # Check interval
workers: 2          # Number of parallel workers
accounts:
  - name: "primary"
    host: "imap.example.com:993"
    tls: true
    username: "user@example.com"
    password: "yourpassword"
    inbox: "INBOX"
    spam_folder: "Spam"
    search_unseen_only: true
```

### Development
For development, you can run SpamAssassin in a container and your Go app locally:
```bash
# Start only SpamAssassin
docker compose up spamassassin

# In another terminal
go run .
```

### Technical Background

* Uses the go-imap library for IMAP communication
* SpamAssassin runs in a separate Docker container
* Implements goroutines for parallel processing
* Uses Context for clean shutdown handling

### License
MIT

___
*This project is a personal learning project and is actively being developed.*