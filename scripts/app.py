#!/usr/bin/env python
# -*- coding: utf-8 -*-

import sys
import getopt
import datetime
import json
import base64

import python_jwt as jwt
import jwcrypto.jwk as jwk

public_key = ""
private_key = ""
token = ""

obj = {}

def main(argv):

    public_key_file = argv[0]
    private_key_file = argv[1]
    payload = argv[2]

    x = json.loads(base64.b64decode(payload))

    with open(public_key_file, "rb") as pemfile:
        public_key = jwk.JWK.from_pem(pemfile.read())
        public_key = public_key.export()

    with open(private_key_file, "rb") as pemfile:
        private_key = jwk.JWK.from_pem(pemfile.read())
        private_key = private_key.export()

    token = jwt.generate_jwt(x, jwk.JWK.from_json(private_key), 'RS256', datetime.timedelta(minutes=500000))

    header, claims = jwt.verify_jwt(
        token, jwk.JWK.from_json(public_key), ['RS256'])

    print(token)


if __name__ == "__main__":
    main(sys.argv[1:])