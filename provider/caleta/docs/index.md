[Caleta Gaming System](https://caletagaming.com/) is a complete RGS + RNG system that has been designed to offer Game Providers and Operators an easy and quick solution for any kind of Game Platform: Slots, Bingo, Kenos, Video Poker, etc.

### Required configuration
As always, contact the provider before you begin in order to get hold of URLs, keys and parameters. 

Caleta will provide:
- `url` - Base url used for request toward caleta
- `operator_id` - Your operator id in Caleta system
- `verification_key` - Public key provided by Caleta to verify their requests

Caleta will request:
- A url for your Valkyrie deployment
- Your public key for your `signing_key` that is used for verifying request signatures

```yaml
providers:
  - name: Caleta
    url: 'https://Caleta-Base-url'
    auth:
      operator_id: YourCasino
      verification_key: |
        -----BEGIN PUBLIC KEY-----
        xyz-etc...
        -----END PUBLIC KEY-----
      signing_key: ${PRIVATE_KEY}
```