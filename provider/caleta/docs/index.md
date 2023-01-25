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

`base_path` is used to differentiate between Valkyrie's exposed endpoints for the specific provider.

`game_launch_type` has two possible values, "static" and "request". 
It will always default to "static" if omitted. With "static" the gamelaunch url is built within Valkyrie. 
With "request" it is fetched using Caleta's API.

```yaml
providers:
  - name: Caleta
    url: 'https://Caleta-Base-url'
    base_path: "/caleta"
    provider_specific:
      game_launch_type: "static"
    auth:
      operator_id: YourCasino
      verification_key: |
        -----BEGIN PUBLIC KEY-----
        xyz-etc...
        -----END PUBLIC KEY-----
      signing_key: ${PRIVATE_KEY}
```