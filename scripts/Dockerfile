FROM python:3.7-slim

ENV PUBLIC_KEY_FILE=/keys/server/public/server_pubkey.pem
ENV PRIVATE_KEY_FILE=/keys/server/private/server_key.pem
ENV PAYLOAD=''

ADD app.py /app.py
ADD requirements.txt /requirements.txt

RUN pip install -r /requirements.txt

CMD ["sh", "-c", "python /app.py $PUBLIC_KEY_FILE $PRIVATE_KEY_FILE $PAYLOAD"]