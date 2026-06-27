# GPIO

For questions about the current GPIO state, call `gpio_control` with:

```json
{"action":"status"}
```

Use `gpio_control` with `action: "read"` only when the user asks for one specific pin.
Use `config` only for saved configuration, not for live GPIO levels.
