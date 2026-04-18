---

## Minimal Plugin Template

```go
package main

import (
    "C"
    "fmt"
    "unsafe"
    "github.com/fluent/fluent-bit-go/output"
)

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
    return output.FLBPluginRegister(def, "my_plugin", "My Plugin Description")
}

//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
    param := output.FLBPluginConfigKey(plugin, "MyParam")
    output.FLBPluginSetContext(plugin, param)
    return output.FLB_OK
}

//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, tag *C.char) int {
    dec := output.NewDecoder(data, int(length))
    for {
        ret, _, record := output.GetRecord(dec)
        if ret != 0 {
            break
        }
        fmt.Printf("Record: %v\n", record)
    }
    return output.FLB_OK
}

//export FLBPluginExitCtx
func FLBPluginExitCtx(ctx unsafe.Pointer) int {
    return output.FLB_OK
}

func main() {}
```

---

## Build

```bash
go mod init github.com/yourname/fluentbit-my-plugin
go get github.com/fluent/fluent-bit-go/output
go build -buildmode=c-shared -o out_my_plugin.so .
```

---

## Test with Docker

```bash
docker run --rm \
  -v $(pwd)/out_my_plugin.so:/out_my_plugin.so \
  -v $(pwd)/fluent-bit.conf:/fluent-bit/etc/fluent-bit.conf \
  fluent/fluent-bit:2.2.0 \
  /fluent-bit/bin/fluent-bit -c /fluent-bit/etc/fluent-bit.conf
```

---

## Return Codes

| Code | Constant | Meaning |
|------|----------|---------|
| 0 | FLB_OK | Success |
| 1 | FLB_ERROR | Fatal error, do not retry |
| 2 | FLB_RETRY | Temporary failure, retry |

Use FLB_RETRY for GCP API rate limits or timeouts.

---

## Contributing to FluentBit Upstream

1. Fork `github.com/fluent/fluent-bit`
2. `git checkout -b feat/out-my-plugin`
3. Add plugin under `plugins/out_<name>/`
4. Add entry in `CMakeLists.txt`
5. Write unit tests
6. Open PR — maintainers usually respond in 1–2 weeks
