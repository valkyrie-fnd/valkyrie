[Caleta Gaming System](https://caletagaming.com/) is a complete RGS + RNG system that has been designed to offer Game Providers and Operators an easy and quick solution for any kind of Game Platform: Slots, Bingo, Kenos, Video Poker, etc.

### Required Configuration
As always, contact the provider before you begin in order to get hold of URLs, keys and parameters. 

Caleta will provide:
- `url`
- `operator_id`
- `verification_key`

Caleta will request:
- An url for your Valkyrie deployment
- Your public key for your `signing_key` for verifying request signatures

```yaml
providers:
  - name: Caleta
    url: 'https://Caleta-Base-url' # base url used for request toward caleta
    auth:
      operator_id: YourCasino # Your operator id in Caleta system
      verification_key: | # public key provided by Caleta to verify their requests
        -----BEGIN PUBLIC KEY-----
        xyz-etc...
        -----END PUBLIC KEY-----
      signing_key: ${PRIVATE_KEY} # Private key used to sign requests toward Caleta.
```