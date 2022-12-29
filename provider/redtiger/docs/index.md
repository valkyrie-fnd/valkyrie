[Red Tiger](https://www.redtiger.com/) offers leading games in the slot space.

### Red Tiger API

Contact Red Tiger to see their API and details on Red Tiger eco system.

:::note
Requires login to access
:::
### Valkyrie integration

Integrate your gaming lobby and wallet system to Valkyrie and you will be able to access games offered by Red Tiger.

Contact Red Tiger in order to setup an agreement and get access to their developer portal to see operator specific configuration.

### Required configuration

Red Tiger will provide the following configuration.
- `url` - Will be available at the settings page of Red Tigers developer portal.
- `api_key` - Api key available in the developer portal.
- `recon_token` - Reconciliation token used in some cases to resolve failed requests. Found in developer portal

```yaml
providers:
  - name: Red Tiger
    url: 'https://redtiger'
    auth:
      api_key: ${RED_TIGER_API_KEY}
      recon_token:  ${RECON_TOKEN}
```