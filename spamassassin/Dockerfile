FROM debian:stable-slim

# Install SpamAssassin and dependencies
RUN apt-get update && \
    apt-get install -y spamassassin && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Expose default SpamAssassin port
EXPOSE 783

# Start SpamAssassin in daemon mode
CMD ["spamd", "-c", "-m5", "--pidfile=/var/run/spamd.pid", "--syslog-socket=none"]