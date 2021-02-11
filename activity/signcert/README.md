# Fabric Signcert activity

This Flogo activity contribution can be configured to retrieve a requesting user's signing certificate.

## Configuration and Inputs

This operation can specify a network name and a user name of format user@org, e.g.,

```json
    "activity": {
        "ref": "#signcert",
        "settings": {
            "connectionName": "=$property[\"NETWORK\"]",
            "userOrgOnly": false
        },
        "input": {
            "userName": "=$flow.user"
        }
    }
```

It will return the text info about the user's signing certificate in the format similar to output of

```bash
openssl x509 -noout -text -in cert.pem
```
