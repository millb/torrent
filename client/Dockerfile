FROM python:3.10-alpine

WORKDIR /

# Install useful packages
RUN apk add --update \
        bash \
        coreutils \
        curl \
        jq \
    && rm /var/cache/apk/*

COPY . .

CMD ["python", "main.py"]
# CMD exec /bin/bash -c "trap : TERM INT; sleep infinity & wait"
