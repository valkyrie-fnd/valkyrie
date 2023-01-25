[Evolution](https://www.evolution.com/) offers leading live games such as casino, cards & craps.

### Evolution API

See evolution API (https://studio.evolution.com/api/docs) for details on Evolution eco system.

### Valkyrie integration

Integrate your gaming lobby and wallet system to Valkyrie and you will be able to access games offered by Evolution Gaming.

Contact Evolution in order to setup an agreement and get needed configuration for the integration.

### Required configuration

Evolution will provide the following configuration.
- `url` - the base url valkyrie uses to send requests to evolution
- `casino_key` - Evolution provided key to identify your casino
- `api_key` - Api key to identify your integration with Evolution
- `casino_token` - Token used for game launch requests toward evolution backend

`base_path` is used to differentiate between Valkyrie's exposed endpoints for the specific provider.

```yaml
providers:
  - name: Evolution
    url: 'https://Evo-baseurl'
    base_path: "/evolution"
    auth:
      casino_key: ${EVO_CASINO_KEY}
      api_key: ${EVO_API_KEY}
      casino_token:  ${EVO_CASINO_API_TOKEN}
```